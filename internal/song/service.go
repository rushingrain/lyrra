package song

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/auth"
	"lyrra/internal/catalogmap"
	"lyrra/internal/dto"
	"lyrra/internal/kgapi"
	"lyrra/internal/paging"
	"lyrra/internal/play"
)

// 新歌首发榜 rank_id：与 SDK TopSong 默认值一致。
const topSongRankID = 21608

// 默认每页条数：前端未传 pagesize 时使用。卡片网格下约两行，不多不少。
const defaultPageSize = 15

// songClient 只暴露 song 服务用到的 SDK 方法，便于测试注入桩实现。
type songClient interface {
	TopSong(ctx context.Context, req kg.TopSongRequest) (*kg.TopSongResponse, error)
	TopCardYouth(ctx context.Context, req kg.TopCardYouthRequest) (*kg.TopCardYouthResponse, error)
	GetDailyRecommendGuest(ctx context.Context, cookie map[string]string) (*kg.Response, error)
	FavoriteCount(ctx context.Context, req kg.FavoriteCountRequest) (*kg.FavoriteCountResponse, error)
	CommentCount(ctx context.Context, req kg.CommentCountRequest) (*kg.CommentCountResponse, error)
	SearchLyric(ctx context.Context, req kg.SearchLyricRequest) (*kg.SearchLyricResponse, error)
	Lyric(ctx context.Context, req kg.LyricRequest) (*kg.LyricResponse, error)
}

// Service 音乐服务。
//
// 结构体实例常驻进程，持有独立的 *kg.Client；前端通过 Wails 绑定调用
// 其导出方法（window.go.song.Service.xxx）。
type Service struct {
	client songClient
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化所有 client 调用（同 auth/banner/playlist 包）。
	mu sync.Mutex
}

// New 创建概念版歌曲服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func New(log *slog.Logger) (*Service, error) {
	client, err := kgapi.NewClient(kg.WithAutoRefresh(false), kg.WithLite(true))
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// GetLatestSongs 拉取首页最新音乐（TopSong 新歌首发榜）。
// page / pagesize 由前端可选传入：page<=0 回退 1，pagesize<=0 回退 15，
// 两个参数都会透传给上游请求。
func (s *Service) GetLatestSongs(page int, pagesize int) ([]dto.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, pagesize = paging.Normalize(page, pagesize, defaultPageSize)

	resp, err := s.client.TopSong(context.Background(), kg.TopSongRequest{
		Page:     page,
		Pagesize: pagesize,
		Type:     topSongRankID,
	})
	if err != nil {
		s.log.Warn("failed to fetch home songs", "err", err.Error())
		return nil, err
	}
	// 业务成功判断同其他服务：body 里 status=="1" 且 error_code==0。
	if !kgapi.StatusOK(resp.Body) {
		return nil, errors.New("top_song 接口业务失败")
	}

	items := kgapi.Records(resp.Body["data"])
	priority := play.PriorityForAny(play.DefaultQuality)
	return catalogmap.LatestSongs(items, priority), nil
}

// GetRecommendedSongs 获取推荐歌曲（TopCardYouth 卡片推荐）。
// typeId 是上游卡片 ID，原样作为 SDK 的 cardId 参数；前端热门音乐传 3005。
// 该接口不支持分页，固定返回一批歌曲（song_list）。
func (s *Service) GetRecommendedSongs(typeId int) ([]dto.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.client.TopCardYouth(context.Background(), kg.TopCardYouthRequest{CardId: typeId})
	if err != nil {
		s.log.Warn("failed to fetch recommended songs", "typeId", typeId, "err", err.Error())
		return nil, err
	}
	// 业务成功判断同其他服务：body 里 status=="1" 且 error_code==0。
	if !kgapi.StatusOK(resp.Body) {
		return nil, errors.New("top_card_youth 接口业务失败")
	}

	data := kgapi.RecordOf(resp.Body["data"])
	items := kgapi.Records(data["song_list"])
	priority := play.PriorityForAny(play.DefaultQuality)
	return catalogmap.RecommendedSongs(items, priority), nil
}

// GetDailyRecommendSongs 获取每日推荐。
//
// 调用 SDK 的 GetDailyRecommendGuest：内部先请求 EverydayRecommend，
// 上游返回 200103 时由 SDK 回退到 RecommendSongs。
// 必须带当前会话 cookie，未登录时直接返回错误；前端也只在登录后才会调用。
func (s *Service) GetDailyRecommendSongs() ([]dto.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cookie := auth.LoadSession()
	if cookie["token"] == "" || cookie["userid"] == "" {
		return nil, errors.New("未登录，无法获取每日推荐")
	}

	ctx := context.Background()
	resp, err := s.client.GetDailyRecommendGuest(ctx, cookie)
	if err != nil {
		s.log.Warn("failed to fetch daily recommend songs", "err", err.Error())
		return nil, err
	}

	// 业务成功判断同其他服务：body 里 status=="1" 且 error_code==0。
	if !kgapi.StatusOK(resp.Body) {
		status := resp.Body["status"]
		code := resp.Body["error_code"]
		s.log.Warn("daily recommend business failure",
			"status", status,
			"error_code", code,
			"error_msg", resp.Body["error_msg"],
		)
		return nil, fmt.Errorf("everyday_recommend 接口业务失败（status=%v, error_code=%v）", status, code)
	}

	data := kgapi.RecordOf(resp.Body["data"])
	items := kgapi.Records(data["song_list"])
	priority := play.PriorityForAny(play.DefaultQuality)
	return catalogmap.DailyRecommendSongs(items, priority), nil
}

// getFavoriteCount 获取指定歌曲的收藏数，只供 GetSongStats 复用；
// 前端只调 GetSongStats，所以这里保持非导出，避免多出一个绑定。
//
// mixSongID 传 Song.ID（MixSongID / album_audio_id）。上游按 mixsongids 查询并返回列表，
// 这里只取与入参匹配的那条；列表里没有匹配项时按 0 处理（新歌可能还没有收藏数据）。
// 计数接口不需要登录态，因此和首页榜单一样只发匿名请求。
func (s *Service) getFavoriteCount(mixSongID int64) (*Count, error) {
	if mixSongID <= 0 {
		return nil, errors.New("缺少有效的歌曲 ID")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.client.FavoriteCount(context.Background(), kg.FavoriteCountRequest{Mixsongids: mixSongID})
	if err != nil {
		s.log.Warn("failed to fetch favorite count", "mixSongID", mixSongID, "err", err.Error())
		return nil, err
	}
	if err := checkFavoriteCountResponse(resp.Body); err != nil {
		return nil, err
	}

	entry := favoriteEntry(resp.Body, mixSongID)
	if entry == nil {
		return newCount(0), nil
	}
	return newCount(kgapi.Int64(entry["count"])), nil
}

// getCommentCount 获取指定歌曲的评论数，只供 GetSongStats 复用（同为非导出）。
//
// hash 传歌曲 hash（Song.Hash）。上游返回 {"<hash>": 数量}，键的大小写不一定与入参一致，
// 因此比较时忽略大小写；响应里没有该 hash 时按 0 处理（评论可能被关闭或尚未产生）。
func (s *Service) getCommentCount(hash string) (*Count, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return nil, errors.New("缺少歌曲 hash")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.client.CommentCount(context.Background(), kg.CommentCountRequest{Hash: hash})
	if err != nil {
		s.log.Warn("failed to fetch comment count", "hash", hash, "err", err.Error())
		return nil, err
	}
	if err := checkCommentCountResponse(resp.Body); err != nil {
		return nil, err
	}
	return newCount(commentCountOf(resp.Body, hash)), nil
}

// newCount 组装计数 DTO：Text 统一按展示规则格式化。
func newCount(n int64) *Count {
	return &Count{Count: n, Text: formatCount(n)}
}

// GetSongStats 一次性获取歌曲的收藏数与评论数，供播放器展示。
//
// mixSongID 传 Song.ID（MixSongID），hash 传 Song.Hash。
// 两项统计走不同接口，互不依赖：任一项失败只把该项置空，由前端显示占位符；
// 两项都取不到时才返回错误（前端同样静默降级）。
func (s *Service) GetSongStats(mixSongID int64, hash string) (*SongStats, error) {
	stats := &SongStats{}

	if favorite, err := s.getFavoriteCount(mixSongID); err == nil {
		stats.Favorite = favorite
	}
	if comment, err := s.getCommentCount(hash); err == nil {
		stats.Comment = comment
	}

	if stats.Favorite == nil && stats.Comment == nil {
		return nil, errors.New("歌曲统计数据获取失败")
	}
	return stats, nil
}

// checkFavoriteCountResponse 校验收藏数响应：该接口用 status + errcode/errmsg，
// 与 song/url 的 status + error_code 不同。
func checkFavoriteCountResponse(body map[string]any) error {
	if !kgapi.StatusOK(body) {
		return errors.New("收藏数接口业务失败")
	}
	if code := kgapi.Int(body["errcode"]); code != 0 {
		return fmt.Errorf("收藏数接口业务失败（errcode=%d）：%s", code, kgapi.Text(body["errmsg"]))
	}
	return nil
}

// checkCommentCountResponse 校验评论数响应。
//
// 正常响应直接是 {"<hash>": 数量}，没有状态字段，所以只在出现状态字段并判定失败时才报错，
// 避免把"评论数为 0"或"上游省略状态字段"误判成失败。
func checkCommentCountResponse(body map[string]any) error {
	if status, ok := body["status"]; ok && !kgapi.Flag(status) {
		return errors.New("评论数接口业务失败")
	}
	if code := kgapi.Int(body["errcode"]); code != 0 {
		return fmt.Errorf("评论数接口业务失败（errcode=%d）：%s", code, kgapi.Text(body["errmsg"]))
	}
	return nil
}

// favoriteEntry 从 data.list 里取与 mixSongID 匹配的条目。
// 只有一条且 id 对不上时也返回它（上游可能换用别的 id 字段或单位），其余情况返回 nil。
func favoriteEntry(body map[string]any, mixSongID int64) map[string]any {
	data := kgapi.RecordOf(body["data"])
	rawItems, _ := data["list"].([]any)

	var only kgapi.Record
	for _, entry := range kgapi.Records(rawItems) {
		if kgapi.Int64(entry["mixsongid"]) == mixSongID {
			return entry
		}
		if only == nil {
			only = entry
		}
	}
	if len(rawItems) == 1 {
		return only
	}
	return nil
}

// commentCountOf 取指定 hash 的评论数：上游用 hash 作键，比较时忽略大小写；
// 取不到时返回 0。
func commentCountOf(body map[string]any, hash string) int64 {
	for key, value := range body {
		if strings.EqualFold(strings.TrimSpace(key), hash) {
			return kgapi.Int64(value)
		}
	}
	return 0
}
