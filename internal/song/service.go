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

	"kuwave/internal/dto"
	"kuwave/internal/imageurl"
	"kuwave/internal/play"
)

// 新歌首发榜 rank_id：与 SDK TopSong 默认值一致。
const topSongRankID = 21608

// 默认每页条数：前端未传 pagesize 时使用。卡片网格下约两行，不多不少。
const defaultPageSize = 15

// songClient 只暴露 song 服务用到的 SDK 方法，便于测试注入桩实现。
type songClient interface {
	TopSong(ctx context.Context, req kg.TopSongRequest) (*kg.TopSongResponse, error)
	TopCard(ctx context.Context, req kg.TopCardRequest) (*kg.TopCardResponse, error)
	FavoriteCount(ctx context.Context, req kg.FavoriteCountRequest) (*kg.FavoriteCountResponse, error)
	CommentCount(ctx context.Context, req kg.CommentCountRequest) (*kg.CommentCountResponse, error)
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
	client, err := kg.New(kg.WithAutoRefresh(false), kg.WithLite(true))
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
	songs := make([]dto.Song, 0, len(items))
	priority := play.PriorityForAny(play.DefaultQuality)
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// 过滤上游标记为失败（__status != 1）的条目
		if fmt.Sprintf("%v", m["__status"]) != "1" {
			continue
		}
		songs = append(songs, parseSong(m, priority))
	}
	return songs, nil
}

// GetHotSongs 获取热门音乐
// TopCard 返回固定条数（song_list_size），无需分页参数；card_id 使用 SDK 默认 1。
func (s *Service) GetHotSongs() ([]dto.Song, error) {
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
	songs := make([]dto.Song, 0, len(items))
	priority := play.PriorityForAny(play.DefaultQuality)
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		songs = append(songs, parseSong(m, priority))
	}
	return songs, nil
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
	return newCount(int64(pickInt(entry, "count"))), nil
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
	if fmt.Sprintf("%v", body["status"]) != "1" {
		return errors.New("收藏数接口业务失败")
	}
	if code := pickInt(body, "errcode"); code != 0 {
		return fmt.Errorf("收藏数接口业务失败（errcode=%d）：%s", code, pickString(body, "errmsg"))
	}
	return nil
}

// checkCommentCountResponse 校验评论数响应。
//
// 正常响应直接是 {"<hash>": 数量}，没有状态字段，所以只在出现状态字段并判定失败时才报错，
// 避免把"评论数为 0"或"上游省略状态字段"误判成失败。
func checkCommentCountResponse(body map[string]any) error {
	if status, ok := body["status"]; ok && fmt.Sprintf("%v", status) != "1" {
		return errors.New("评论数接口业务失败")
	}
	if code := pickInt(body, "errcode"); code != 0 {
		return fmt.Errorf("评论数接口业务失败（errcode=%d）：%s", code, pickString(body, "errmsg"))
	}
	return nil
}

// favoriteEntry 从 data.list 里取与 mixSongID 匹配的条目。
// 只有一条且 id 对不上时也返回它（上游可能换用别的 id 字段或单位），其余情况返回 nil。
func favoriteEntry(body map[string]any, mixSongID int64) map[string]any {
	data, _ := body["data"].(map[string]any)
	items, _ := data["list"].([]any)

	var only map[string]any
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if int64(pickInt(entry, "mixsongid")) == mixSongID {
			return entry
		}
		if only == nil {
			only = entry
		}
	}
	if len(items) == 1 {
		return only
	}
	return nil
}

// commentCountOf 取指定 hash 的评论数：上游用 hash 作键，比较时忽略大小写；
// 取不到时返回 0。
func commentCountOf(body map[string]any, hash string) int64 {
	for key, value := range body {
		if strings.EqualFold(strings.TrimSpace(key), hash) {
			return int64(pickIntValue(value))
		}
	}
	return 0
}

// parseSong 把上游单曲条目统一解析为 dto.Song。
// TopSong（新歌首发榜）与 TopCard（热门推荐）字段略有差异，这里做兼容处理。
// priority 是 Hash 的档位候选顺序，取其中第一个非空值（见 internal/play）。
func parseSong(m map[string]any, priority []string) dto.Song {
	return dto.Song{
		ID:         songID(m),
		Name:       pickString(m, "songname"),
		Artist:     songArtist(m),
		Album:      pickString(m, "album_name"),
		CoverURL:   songCover(m),
		Duration:   songDuration(m),
		Hash:       play.Resolve(m, priority),
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
func songID(m map[string]any) int64 {
	for _, key := range []string{"audio_id", "songid", "mixsongid", "album_audio_id"} {
		if v := pickInt(m, key); v != 0 {
			return int64(v)
		}
	}
	return 0
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

// songQuality 按可用音质哈希推导最高音质标签
func songQuality(m map[string]any) string {
	switch {
	case pickString(m, "hash_high") != "":
		return "Hi-Res"
	case pickString(m, "hash_flac") != "":
		return "无损"
	case pickString(m, "hash_320") != "":
		return "HQ"
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
	return pickIntValue(m[key])
}

// pickIntValue 把单个字段值转成整数，兼容 JSON 反序列化后的 float64 及字符串形式。
func pickIntValue(value any) int {
	switch n := value.(type) {
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
