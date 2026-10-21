package playlist

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/dto"
	"lyrra/internal/imageurl"
	"lyrra/internal/play"
)

// 歌单歌曲分页大小：与上游默认页大小一致。
const defaultSongPageSize = 30

// 上游单页上限：拉取整张歌单时使用，减少请求次数。
const maxSongPageSize = 300

// 全量拉取的安全上限（300 × 10 = 3000 首），防止异常数据导致无限请求。
const maxAllSongsPages = 10

// Service 首页官方歌单服务。
//
// 结构体实例常驻进程，持有独立的 *kg.Client（无需登录态）；前端通过 Wails
// 绑定调用其导出方法（window.go.playlist.Service.xxx）。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化所有 client 调用（同 auth/banner 包）。
	mu sync.Mutex
}

// New 创建官方歌单服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func New(log *slog.Logger) (*Service, error) {
	client, err := kg.New(kg.WithAutoRefresh(false))
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// GetHotPlaylists 拉取首页官方歌单（/top/playlist）。
// pageSize 为期望返回条数（<=0 时回退默认 15）；withSong 为是否携带歌曲
// （0 不携带 / 1 携带），两个参数都会透传给上游请求。
func (s *Service) GetHotPlaylists(pageSize int, withSong int) ([]dto.Playlist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pageSize <= 0 {
		pageSize = 15
	}
	resp, err := s.client.TopPlaylist(context.Background(), kg.TopPlaylistRequest{
		Pagesize: pageSize,
		Withsong: withSong,
	})
	if err != nil {
		s.log.Warn("failed to fetch home playlists", "err", err.Error())
		return nil, err
	}
	// 业务成功判断同 auth 包：body 里 status=="1" 且 error_code==0。
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, errors.New("top/playlist 接口业务失败")
	}

	data, _ := resp.Body["data"].(map[string]any)
	items, _ := data["special_list"].([]any)
	playlists := make([]dto.Playlist, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		playlists = append(playlists, dto.Playlist{
			ID:                 int64(pickInt(m, "specialid")),
			GlobalCollectionID: pickString(m, "global_collection_id"),
			Name:               pickString(m, "specialname"),
			Creator:            pickString(m, "nickname"),
			// 部分歌单只返回 pic（与 user_avatar 同为用户头像），缺失时回退
			CreatorAvatar: imageurl.Normalize(firstNonEmpty(pickString(m, "user_avatar"), pickString(m, "pic"))),
			Tags:          pickTagNames(m["tags"]),
			PublishTime:   pickString(m, "publishtime"),
			CoverURL:      imageurl.Normalize(pickString(m, "imgurl")),
			Description:   pickString(m, "intro"),
			PlayCount:     int64(pickInt(m, "play_count")),
			SongCount:     int64(pickInt(m, "song_count")),
		})
	}
	return playlists, nil
}

// GetPlaylistSongs 获取歌单歌曲（PlaylistTrackAll），按页返回。
//
// globalCollectionID 是歌单的 global_collection_id；page 从 1 开始。
// SDK 目前只透传 page，上游按 begin_idx 取偏移，这里显式补上 begin_idx，
// 避免翻页时重复返回第一页。
func (s *Service) GetPlaylistSongs(globalCollectionID string, page int, pageSize int) (*SongsResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.fetchPlaylistSongs(globalCollectionID, page, pageSize)
}

// GetAllPlaylistSongs 获取歌单全部歌曲，供“播放全部”使用。
//
// 按上游单页上限循环拉取，并按歌曲 ID 去重；请求次数有安全上限。
func (s *Service) GetAllPlaylistSongs(globalCollectionID string) ([]dto.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	songs := make([]dto.Song, 0, maxSongPageSize)
	seen := make(map[int64]struct{})
	for page := 1; page <= maxAllSongsPages; page++ {
		result, err := s.fetchPlaylistSongs(globalCollectionID, page, maxSongPageSize)
		if err != nil {
			return nil, err
		}
		for _, song := range result.Songs {
			if song.ID != 0 {
				if _, ok := seen[song.ID]; ok {
					continue
				}
				seen[song.ID] = struct{}{}
			}
			songs = append(songs, song)
		}
		if !result.HasMore {
			return songs, nil
		}
	}
	s.log.Warn("playlist song fetch hit page cap", "pages", maxAllSongsPages)
	return songs, nil
}

// fetchPlaylistSongs 拉取单页歌单歌曲；调用方需持有 s.mu。
func (s *Service) fetchPlaylistSongs(globalCollectionID string, page int, pageSize int) (*SongsResult, error) {
	globalCollectionID = strings.TrimSpace(globalCollectionID)
	if globalCollectionID == "" {
		return nil, errors.New("缺少歌单 ID")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultSongPageSize
	}

	resp, err := s.client.PlaylistTrackAll(context.Background(), kg.PlaylistTrackAllRequest{
		Id:       globalCollectionID,
		Page:     page,
		Pagesize: pageSize,
		Extra:    map[string]any{"begin_idx": (page - 1) * pageSize},
	})
	if err != nil {
		s.log.Warn("failed to fetch playlist songs", "err", err.Error())
		return nil, err
	}
	// 业务成功判断同其他服务：body 里 status=="1" 且 error_code==0。
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, errors.New("playlist_track_all 接口业务失败")
	}

	data, _ := resp.Body["data"].(map[string]any)
	items, _ := data["songs"].([]any)
	total := int64(pickInt(data, "count"))
	beginIdx := pickInt(data, "begin_idx")

	result := &SongsResult{
		Info:  parsePlaylistInfo(data),
		Songs: make([]dto.Song, 0, len(items)),
		Total: total,
	}
	priority := play.PriorityForAny(play.DefaultQuality)
	result.Songs = appendValidSongs(result.Songs, items, priority)
	if total > 0 {
		result.HasMore = int64(beginIdx+len(items)) < total
	} else {
		result.HasMore = len(items) >= pageSize
	}
	return result, nil
}

// appendValidSongs 解析并追加有效歌曲，跳过没有歌名的条目。
//
// 上游偶尔会混入只有 fileid / shield 的屏蔽占位条目，映射后歌名为空，
// 这类数据不应返回给前端。
func appendValidSongs(dst []dto.Song, items []any, priority []string) []dto.Song {
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		song := parsePlaylistSong(m, priority)
		if song.Name == "" {
			continue
		}
		dst = append(dst, song)
	}
	return dst
}

// parsePlaylistInfo 从 PlaylistTrackAll 的 list_info 解析歌单元信息。
// 该接口不返回播放次数，PlayCount 保持 0，由前端用首页数据补齐。
func parsePlaylistInfo(data map[string]any) dto.Playlist {
	info, _ := data["list_info"].(map[string]any)
	return dto.Playlist{
		ID:                 int64(pickInt(info, "specialid")),
		GlobalCollectionID: pickString(info, "global_collection_id"),
		Name:               pickString(info, "name"),
		CoverURL:           imageurl.Normalize(pickString(info, "pic")),
		Description:        pickString(info, "intro"),
		Creator:            pickString(info, "list_create_username"),
		CreatorAvatar:      imageurl.Normalize(pickString(info, "create_user_pic")),
		Tags:               pickTagNames(info["musiclib_tags"]),
		PublishTime:        pickString(info, "publish_date"),
		SongCount:          int64(pickInt(info, "count")),
	}
}

// parsePlaylistSong 解析歌单歌曲条目。
//
// 上游 name 是 "歌手 - 歌曲名" 的拼接格式，展示时取第一个 " - " 之后的部分。
func parsePlaylistSong(m map[string]any, priority []string) dto.Song {
	artist, name := splitPlaylistSongName(pickString(m, "name"))
	if singer := pickString(firstRecord(m["singerinfo"]), "name"); singer != "" {
		artist = singer
	}

	album := ""
	if albumInfo, ok := m["albuminfo"].(map[string]any); ok {
		album = pickString(albumInfo, "name")
	}

	cover := pickString(m, "cover")
	if cover == "" {
		if trans, ok := m["trans_param"].(map[string]any); ok {
			cover = pickString(trans, "union_cover")
		}
	}

	return dto.Song{
		ID:         int64(pickInt(m, "mixsongid")),
		Name:       name,
		Artist:     artist,
		Album:      album,
		CoverURL:   imageurl.Normalize(cover),
		Duration:   int64(pickInt(m, "timelen")),
		Hash:       play.Resolve(m, priority),
		Quality:    playlistSongQuality(m),
		IsOriginal: isOriginal(m),
	}
}

// splitPlaylistSongName 按第一个 " - " 拆出歌手与歌曲名；没有分隔符时整串作为歌曲名。
func splitPlaylistSongName(full string) (string, string) {
	artist, name, ok := strings.Cut(strings.TrimSpace(full), " - ")
	if !ok {
		return "", strings.TrimSpace(full)
	}
	return strings.TrimSpace(artist), strings.TrimSpace(name)
}

// playlistSongQuality 按 relate_goods 的最高码率推导展示标签。
// 标准音质不展示标签，与其他歌曲列表保持一致。
func playlistSongQuality(m map[string]any) string {
	maxBitrate := 0
	goods, _ := m["relate_goods"].([]any)
	for _, item := range goods {
		g, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if bitrate := pickInt(g, "bitrate"); bitrate > maxBitrate {
			maxBitrate = bitrate
		}
	}
	switch {
	case maxBitrate >= 1400:
		return "Hi-Res"
	case maxBitrate >= 900:
		return "无损"
	case maxBitrate >= 320:
		return "HQ"
	default:
		return ""
	}
}

// isOriginal 判断原唱：is_original 可能在顶层，也可能在 trans_param 里。
func isOriginal(m map[string]any) bool {
	if fmt.Sprintf("%v", m["is_original"]) == "1" {
		return true
	}
	trans, _ := m["trans_param"].(map[string]any)
	return fmt.Sprintf("%v", trans["is_original"]) == "1"
}

// pickTagNames 取标签数组里的 tag_name，兼容顶层 tags 与 list_info.musiclib_tags。
func pickTagNames(value any) []string {
	items, _ := value.([]any)
	names := make([]string, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if name := pickString(m, "tag_name"); name != "" {
			names = append(names, name)
		}
	}
	return names
}

// firstRecord 取数组里的第一个对象，缺失时返回 nil（对 nil map 取值返回零值）。
func firstRecord(value any) map[string]any {
	items, _ := value.([]any)
	if len(items) == 0 {
		return nil
	}
	m, _ := items[0].(map[string]any)
	return m
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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
