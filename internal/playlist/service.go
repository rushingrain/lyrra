package playlist

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/auth"
	"lyrra/internal/catalogmap"
	"lyrra/internal/dto"
	"lyrra/internal/imageurl"
	"lyrra/internal/kgapi"
	"lyrra/internal/paging"
	"lyrra/internal/play"
)

// 歌单歌曲分页大小：与上游默认页大小一致。
const defaultSongPageSize = 30

// 上游单页上限：拉取整张歌单时使用，减少请求次数。
const maxSongPageSize = 300

// 全量拉取的安全上限（300 × 10 = 3000 首），防止异常数据导致无限请求。
const maxAllSongsPages = 10

// Service 歌单服务，提供官方歌单、用户歌单和歌单歌曲。
//
// 结构体实例常驻进程，持有独立的 *kg.Client；用户歌单请求读取最新登录态。前端通过 Wails
// 绑定调用其导出方法（window.go.playlist.Service.xxx）。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化所有 client 调用（同 auth/banner 包）。
	mu sync.Mutex
}

// New 创建歌单服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func New(log *slog.Logger) (*Service, error) {
	client, err := kgapi.NewClient(kg.WithAutoRefresh(false))
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

	pageSize = paging.NormalizeSize(pageSize, 15)
	resp, err := s.client.TopPlaylist(context.Background(), kg.TopPlaylistRequest{
		Pagesize: pageSize,
		Withsong: withSong,
	})
	if err != nil {
		s.log.Warn("failed to fetch home playlists", "err", err.Error())
		return nil, err
	}
	// 业务成功判断同 auth 包：body 里 status=="1" 且 error_code==0。
	if !kgapi.StatusOK(resp.Body) {
		return nil, errors.New("top/playlist 接口业务失败")
	}

	data := kgapi.RecordOf(resp.Body["data"])
	return catalogmap.HotPlaylists(kgapi.Records(data["special_list"])), nil
}

// GetUserPlaylists 拉取当前登录用户的歌单列表（/v7/get_all_list）。
// page / pagesize 由前端传入：page<=0 回退 1，pagesize<=0 回退 20。
func (s *Service) GetUserPlaylists(page int, pagesize int) ([]UserPlaylist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, pagesize = paging.Normalize(page, pagesize, 20)
	// 每次读取最新会话；历史科学计数法 userid 由 auth.LoadSession 统一规整。
	cookie := auth.LoginCookie(auth.LoadSession())
	// SDK 请求参数会读取 client 内部 cookie；缺失字段也要覆盖，避免残留旧账号身份。
	for _, key := range []string{"token", "userid", "dfid"} {
		s.client.SetCookie(key, cookie[key])
	}
	userid, err := strconv.Atoi(cookie["userid"])
	if cookie["token"] == "" || err != nil || userid <= 0 {
		return nil, auth.ErrNotLoggedIn
	}

	resp, err := s.client.UserPlaylist(context.Background(), kg.UserPlaylistRequest{
		Page:     page,
		Pagesize: pagesize,
		Userid:   userid,
		Cookie:   cookie,
	})
	if err != nil {
		s.log.Warn("failed to fetch user playlists", "err", err.Error())
		return nil, err
	}
	if !kgapi.StatusOK(resp.Body) {
		s.log.Warn("user playlists business failure", "error_code", resp.Body["error_code"])
		return nil, auth.ErrNotLoggedIn
	}

	data := kgapi.RecordOf(resp.Body["data"])
	items := kgapi.Records(data["info"])
	playlists := make([]UserPlaylist, 0, len(items))
	for _, item := range items {
		playlists = append(playlists, UserPlaylist{
			ID:        strconv.FormatInt(kgapi.Int64(item["listid"]), 10),
			Name:      kgapi.Text(item["name"]),
			SongCount: kgapi.Int(item["count"]),
			CoverURL:  imageurl.Normalize(kgapi.Text(item["pic"])),
			IsDefault: kgapi.Int(item["is_def"]),
			Type:      kgapi.Int(item["type"]),
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
	page, pageSize = paging.Normalize(page, pageSize, defaultSongPageSize)

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
	if !kgapi.StatusOK(resp.Body) {
		return nil, errors.New("playlist_track_all 接口业务失败")
	}

	data := kgapi.RecordOf(resp.Body["data"])
	rawItems, _ := data["songs"].([]any)
	total := kgapi.Int64(data["count"])
	beginIdx := kgapi.Int(data["begin_idx"])

	result := &SongsResult{
		Info:  catalogmap.PlaylistInfo(data),
		Songs: make([]dto.Song, 0, len(rawItems)),
		Total: total,
	}
	priority := play.PriorityForAny(play.DefaultQuality)
	result.Songs = catalogmap.PlaylistSongs(kgapi.Records(rawItems), priority)
	if total > 0 {
		result.HasMore = int64(beginIdx+len(rawItems)) < total
	} else {
		result.HasMore = len(rawItems) >= pageSize
	}
	return result, nil
}
