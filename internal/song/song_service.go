package song

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"kuwave/internal/imageurl"
)

// 新歌首发榜 rank_id：与 SDK TopSong 默认值一致。
const topSongRankID = 21608

// 默认每页条数：前端未传 pagesize 时使用。卡片网格下约两行，不多不少。
const defaultPageSize = 15

// SongService 首页最新音乐服务。
//
// 结构体实例常驻进程，持有独立的 *kg.Client；前端通过 Wails 绑定调用
// 其导出方法（window.go.song.SongService.xxx）。
type SongService struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化所有 client 调用（同 auth/banner/playlist 包）。
	mu sync.Mutex
}

// NewSongService 创建最新音乐服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func NewSongService(log *slog.Logger) (*SongService, error) {
	client, err := kg.New(kg.WithAutoRefresh(false))
	if err != nil {
		return nil, err
	}
	return &SongService{client: client, log: log}, nil
}

// GetHomeSongs 拉取首页最新音乐（TopSong 新歌首发榜）。
// page / pagesize 由前端可选传入：page<=0 回退 1，pagesize<=0 回退 15，
// 两个参数都会透传给上游请求。
func (s *SongService) GetHomeSongs(page int, pagesize int) ([]HomeSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if page <= 0 {
		page = 1
	}
	if pagesize <= 0 {
		pagesize = defaultPageSize
	}

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
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, errors.New("top_song 接口业务失败")
	}

	items, _ := resp.Body["data"].([]any)
	songs := make([]HomeSong, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// 过滤上游标记为失败（__status != 1）的条目
		if fmt.Sprintf("%v", m["__status"]) != "1" {
			continue
		}
		songs = append(songs, parseSong(m))
	}
	return songs, nil
}

// GetHotSongs 拉取首页热门音乐（TopCard 单曲卡片推荐）。
// TopCard 返回固定条数（song_list_size），无需分页参数；card_id 使用 SDK 默认 1。
func (s *SongService) GetHotSongs() ([]HomeSong, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	resp, err := s.client.TopCard(context.Background(), kg.TopCardRequest{CardId: 1})
	if err != nil {
		s.log.Warn("failed to fetch hot songs", "err", err.Error())
		return nil, err
	}
	// 业务成功判断同其他服务：body 里 status=="1" 且 error_code==0。
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, errors.New("top_card 接口业务失败")
	}

	data, _ := resp.Body["data"].(map[string]any)
	items, _ := data["song_list"].([]any)
	songs := make([]HomeSong, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		songs = append(songs, parseSong(m))
	}
	return songs, nil
}

// parseSong 把上游单曲条目统一解析为 HomeSong。
// TopSong（新歌首发榜）与 TopCard（热门推荐）字段略有差异，这里做兼容处理。
func parseSong(m map[string]any) HomeSong {
	return HomeSong{
		ID:         songID(m),
		Title:      pickString(m, "songname"),
		Artist:     songArtist(m),
		Album:      pickString(m, "album_name"),
		CoverURL:   songCover(m),
		Duration:   songDuration(m),
		Hash:       pickString(m, "hash"),
		Quality:    songQuality(m),
		IsOriginal: isOriginal(m),
	}
}

// songArtist 拼接歌手名：优先取结构化歌手列表（TopSong 的 authors / TopCard 的 singerinfo），
// 多个歌手用 "/" 分隔；列表缺失时回退顶层的 author_name。
func songArtist(m map[string]any) string {
	if names := collectNames(m, "authors", "author_name"); len(names) > 0 {
		return strings.Join(names, "/")
	}
	if names := collectNames(m, "singerinfo", "name"); len(names) > 0 {
		return strings.Join(names, "/")
	}
	return pickString(m, "author_name")
}

// collectNames 从 map 的数组字段里提取指定 key 的非空字符串。
func collectNames(m map[string]any, field string, key string) []string {
	arr, _ := m[field].([]any)
	names := make([]string, 0, len(arr))
	for _, item := range arr {
		if im, ok := item.(map[string]any); ok {
			if name := pickString(im, key); name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// songID 依次尝试 audio_id / songid / mixsongid / album_audio_id，
// 大数先转 int 再格式化避免科学计数法；全缺失时回退 hash。
func songID(m map[string]any) string {
	for _, key := range []string{"audio_id", "songid", "mixsongid", "album_audio_id"} {
		if v := strconv.FormatInt(int64(pickInt(m, key)), 10); v != "0" {
			return v
		}
	}
	return pickString(m, "hash")
}

// songCover 优先取专辑封面（TopSong 用 album_sizable_cover，TopCard 用 sizable_cover），
// 缺失时回退 trans_param.union_cover。
func songCover(m map[string]any) string {
	for _, key := range []string{"album_sizable_cover", "sizable_cover"} {
		if s := pickString(m, key); s != "" {
			return imageurl.Normalize(s)
		}
	}
	if trans, ok := m["trans_param"].(map[string]any); ok {
		return imageurl.Normalize(pickString(trans, "union_cover"))
	}
	return ""
}

// songQuality 按可用音质哈希推导最高音质标签：
// flac→无损，high→Hi-Res，320→高品（TopSong 有 flac/high，TopCard 有 flac/320）。
func songQuality(m map[string]any) string {
	switch {
	case pickString(m, "hash_flac") != "":
		return "无损"
	case pickString(m, "hash_high") != "":
		return "Hi-Res"
	case pickString(m, "hash_320") != "":
		return "高品"
	default:
		return ""
	}
}

// songDuration 统一为毫秒：TopSong 的 timelength 已是 ms；
// TopCard 的 time_length 是秒（浮点），需要 ×1000。
func songDuration(m map[string]any) int64 {
	if v := pickInt(m, "timelength"); v > 0 {
		return int64(v)
	}
	if v, ok := m["time_length"].(float64); ok && v > 0 {
		return int64(v * 1000)
	}
	return 0
}

// isOriginal 判断原唱：TopCard 的 is_original 在顶层，TopSong 的在 trans_param 里，都兼容。
func isOriginal(m map[string]any) bool {
	if fmt.Sprintf("%v", m["is_original"]) == "1" {
		return true
	}
	trans, _ := m["trans_param"].(map[string]any)
	return fmt.Sprintf("%v", trans["is_original"]) == "1"
}

// pickString 从 item map 提取字符串字段。
func pickString(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

// pickInt 从 item map 提取整数字段，兼容 JSON 反序列化后的 float64 及字符串形式。
func pickInt(m map[string]any, key string) int {
	switch n := m[key].(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case string:
		if v, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
			return v
		}
	}
	return 0
}
