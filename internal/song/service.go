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

	"lyrra/internal/auth"
	"lyrra/internal/dto"
	"lyrra/internal/imageurl"
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
		songs = append(songs, parseLatestSong(m, priority))
	}
	return songs, nil
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
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, errors.New("top_card_youth 接口业务失败")
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
		songs = append(songs, parseRecommendedSong(m, priority))
	}
	return songs, nil
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
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		status := resp.Body["status"]
		code := resp.Body["error_code"]
		s.log.Warn("daily recommend business failure",
			"status", status,
			"error_code", code,
			"error_msg", resp.Body["error_msg"],
		)
		return nil, fmt.Errorf("everyday_recommend 接口业务失败（status=%v, error_code=%v）", status, code)
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
		songs = append(songs, parseDailyRecommendSong(m, priority))
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

// parseFlatSong 解析字段平铺在条目顶层的榜单歌曲（TopSong / 每日推荐）。
// idKey 是播放地址接口用来定位歌曲的 ID 字段；priority 是 Hash 的档位候选顺序，
// 取其中第一个非空值（见 internal/play）。
func parseFlatSong(m map[string]any, idKey string, name string, priority []string) dto.Song {
	return dto.Song{
		ID:         int64(pickInt(m, idKey)),
		Name:       name,
		Artist:     songArtist(m),
		Album:      pickString(m, "album_name"),
		CoverURL:   songCover(m),
		Duration:   songDuration(m),
		Hash:       play.Resolve(m, priority),
		Quality:    songQuality(m),
		IsOriginal: isOriginal(m),
	}
}

// parseLatestSong 解析 TopSong（新歌首发榜）条目。
//
// ID 取 album_audio_id：播放地址接口按它定位歌曲，顶层的 audio_id 是另一个值，
// 用 audio_id 请求会拿不到正确的播放地址。歌名直接用 songname（已带版本后缀）。
func parseLatestSong(m map[string]any, priority []string) dto.Song {
	return parseFlatSong(m, "album_audio_id", pickString(m, "songname"), priority)
}

// parseDailyRecommendSong 解析每日推荐条目。
//
// 字段与 TopSong 一样平铺在顶层：ID 取 mixsongid；歌名优先 songname
// （上游已带上版本后缀），缺失时回退 ori_audio_name。
func parseDailyRecommendSong(m map[string]any, priority []string) dto.Song {
	name := pickString(m, "songname")
	if name == "" {
		name = pickString(m, "ori_audio_name")
	}
	return parseFlatSong(m, "mixsongid", name, priority)
}

// parseRecommendedSong 解析 TopCardYouth（卡片推荐）条目。
//
// 与 TopSong 不同：音质 hash、时长、原唱标记都在 audio_info 下，专辑信息在 album_info 下。
// 顶层的 hash 不是音质 hash（取它播放会失败），所以只在 audio_info 里找播放用的 hash。
func parseRecommendedSong(m map[string]any, priority []string) dto.Song {
	audioInfo, _ := m["audio_info"].(map[string]any)
	albumInfo, _ := m["album_info"].(map[string]any)
	trans, _ := audioInfo["trans_param"].(map[string]any)

	// audio_info.timelength 已是毫秒；整个 audio_info 缺失时再按顶层字段兜底
	duration := int64(pickInt(audioInfo, "timelength"))
	if duration <= 0 {
		duration = songDuration(m)
	}

	return dto.Song{
		ID:         int64(pickInt(m, "mixsongid")),
		Name:       joinSongNameSuffix(pickString(m, "ori_audio_name"), pickString(trans, "songname_suffix")),
		Artist:     songArtist(m),
		Album:      pickString(albumInfo, "album_name"),
		CoverURL:   recommendedCover(m, albumInfo, audioInfo),
		Duration:   duration,
		Hash:       play.Resolve(audioInfo, priority),
		Quality:    songQuality(audioInfo),
		IsOriginal: isOriginal(audioInfo),
	}
}

// joinSongNameSuffix 用空格把歌名与后缀拼起来（如 "冷雨夜 R&B版"）；后缀为空时返回原名。
func joinSongNameSuffix(name string, suffix string) string {
	if name == "" || suffix == "" {
		return name
	}
	return name + " " + suffix
}

// recommendedCover 取推荐歌曲封面：优先专辑封面，其次卡片自带的封面图。
func recommendedCover(item map[string]any, albumInfo map[string]any, audioInfo map[string]any) string {
	if cover := pickString(albumInfo, "sizable_cover"); cover != "" {
		return imageurl.Normalize(cover)
	}
	for _, key := range []string{"cover_pic", "pic"} {
		if cover := pickString(item, key); cover != "" {
			return imageurl.Normalize(cover)
		}
	}
	trans, _ := audioInfo["trans_param"].(map[string]any)
	return imageurl.Normalize(pickString(trans, "union_cover"))
}

// songArtist 拼接歌手名：优先取结构化歌手列表（TopSong / TopCardYouth 的 authors），
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

// songCover 优先取专辑封面（TopSong 用 album_sizable_cover），
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

// songDuration 统一为毫秒：timelength 已是 ms；
// 旧推荐接口的 time_length 是秒（浮点），需要 ×1000。
func songDuration(m map[string]any) int64 {
	if v := pickInt(m, "timelength"); v > 0 {
		return int64(v)
	}
	if v, ok := m["time_length"].(float64); ok && v > 0 {
		return int64(v * 1000)
	}
	return 0
}

// isOriginal 判断原唱：is_original 可能在顶层，也可能在 trans_param 里，两种都兼容。
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
