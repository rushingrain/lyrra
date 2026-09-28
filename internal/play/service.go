// Package play 负责播放相关的对外接口与策略：
// 音质档位与用户偏好（quality.go）、播放地址解析（本文件）、地址校验（verify.go）。
//
// 歌曲与之后的 MV 播放都放在这里：两者共享偏好存取、SDK client 构建与错误语义，
// 但档位阶梯与地址校验各自独立（音频按文件 hash，视频按清晰度）。
//
// 关于播放地址的协议：酷狗 CDN 返回的是 http:// 地址，且实测 fs*.kugou.com
// 的 https 证书无效（证书校验失败，无法改走 https）。因此桌面端在
// build/darwin/Info.plist 与 Info.dev.plist 里只对 kugou.com 及其子域放行 http
// （NSExceptionDomains，不做全局放行）。如果该放行被移除，前端会播放失败
// （console 里表现为 audio error code 4 = MEDIA_ERR_SRC_NOT_SUPPORTED，toast「播放失败，请重试」）。
//
// 另外，解析播放地址时关闭了 SDK 回退到 /v6/priv_url 的行为（WithSongURLFallback(false)）：
// 那条回退走明文 http 且携带 token / vip_token / userid，不适合带着登录凭证发出去。
package play

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"kuwave/internal/auth"
)

// Service 播放服务。
// 前端通过 Wails 绑定调用其导出方法（window.go.play.Service.xxx）。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化 client 调用（同其它服务）。
	mu sync.Mutex
}

// New 创建播放服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
//
// 使用概念版（lite）平台：当前账号的会员权益只在概念版平台上生效
// （普通平台拿不到无损 / Hi-Res 地址，实测见 internal/play 的档位探测记录）。
func New(log *slog.Logger) (*Service, error) {
	client, err := kg.New(kg.WithAutoRefresh(false), kg.WithLite(true))
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// SetPlaybackQuality 设置播放品质偏好（res / sq / hq / standard）。
// 偏好是进程内共享的用户级设置，影响所有返回 Song 的接口里 Hash 的解析起点
// （搜索结果、首页榜单）以及播放地址的 quality 参数：从该档位起按音质从高到低找第一个
// 可用的文件，不回退到更高档位。值未知或为空时按默认偏好（标准）处理。
func (s *Service) SetPlaybackQuality(quality string) {
	SetPreferredQuality(quality)
}

// PlayURL 是播放地址解析结果。
type PlayURL struct {
	// URL 上游返回的播放地址（含时间戳签名，按账号 + 设备签发，会过期）
	URL string `json:"url"`
	// Quality 本次使用的音质档位（res / sq / hq / standard）
	Quality string `json:"quality"`
	// QualityParam 发给上游的 quality 参数，便于排查
	QualityParam string `json:"qualityParam"`
	// FileHash URL 实际指向的文件 hash；无法解析时为空
	FileHash string `json:"fileHash,omitempty"`
	// Degraded 上游返回的不是期望档位的文件（例如歌曲没有 Hi-Res 时 high 被降级成标准文件）
	Degraded bool `json:"degraded"`
}

// GetSongPlayURL 按当前播放品质偏好解析歌曲播放地址。
//
// hash 与 quality 必须成对：hash 传"目标档位的文件 hash"（搜索结果里的 Song.Hash 就是按
// 同一偏好解析出来的），quality 由偏好映射（见 QualityParam）。上游是按 quality 选文件的
// （传哪个档位的 hash 都一样），所以拿到地址后用 URL 里的文件 hash 校验是否真的命中目标
// 档位；被降级时 Degraded 为 true，由前端决定提示还是照常播放。
// albumAudioID 传 Song.ID（MixSongID / album_audio_id），可为 0。
func (s *Service) GetSongPlayURL(hash string, albumAudioID int) (*PlayURL, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expectedHash := strings.TrimSpace(hash)
	if expectedHash == "" {
		return nil, errors.New("该条目没有可用的文件 hash")
	}

	level := EffectiveQuality()
	qualityParam := QualityParamFor(level)

	result, err := s.client.ResolveSongPlayURL(context.Background(), kg.SongPlayURLRequest{
		Hash:         expectedHash,
		AlbumAudioID: albumAudioID,
		Quality:      qualityParam,
		Cookie:       auth.LoadSession(),
	},
		kg.WithSongURLAll(true),
		// 关闭回退到 /v6/priv_url（song_url_new）：那条路走明文 http
		// （http://tracker.kugou.com），请求体与 Cookie 里带着 token / vip_token / userid。
		// 宁可少一条回退，也不把登录凭证放到明文网络上。
		kg.WithSongURLFallback(false),
	)
	if err != nil {
		s.log.Warn("failed to resolve song play url", "hash", expectedHash, "quality", level, "err", err.Error())
		return nil, err
	}
	if result == nil {
		// 主路由没解析出地址。评估影响面用：如果这条日志经常出现，说明需要
		// 我们自己在 https 上实现原本由回退路由提供的解析（而不是让 SDK 走 http）。
		s.log.Warn("song url not resolved by primary route", "hash", expectedHash, "quality", level)
		return nil, errors.New("上游未返回该歌曲的播放地址")
	}

	playURL, err := buildSongPlayURL(level, qualityParam, expectedHash, result.Primary)
	if err != nil {
		s.log.Warn("no playable url for quality", "hash", expectedHash, "quality", level, "err", err.Error())
		return nil, err
	}
	if playURL.Degraded {
		// 典型场景：歌曲没有 Hi-Res，上游把 high 降级成标准文件（标记仍是 quhigh）。
		s.log.Info("play url degraded by upstream",
			"hash", expectedHash, "fileHash", playURL.FileHash, "quality", level)
	}
	return playURL, nil
}
