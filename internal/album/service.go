package album

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/catalogmap"
	"lyrra/internal/dto"
	"lyrra/internal/kgapi"
	"lyrra/internal/paging"
	"lyrra/internal/play"
)

const defaultSongPageSize = 30
const maxSongPageSize = 300
const maxAllSongsPages = 10

// Service 专辑服务。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	mu     sync.Mutex
}

// New 创建专辑服务；只做本地初始化，不发起网络请求。
func New(log *slog.Logger) (*Service, error) {
	client, err := kgapi.NewClient(kg.WithAutoRefresh(false))
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// GetAlbumDetail 获取专辑基本信息。
func (s *Service) GetAlbumDetail(albumID string) (*dto.Album, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	albumID = strings.TrimSpace(albumID)
	if albumID == "" {
		return nil, errors.New("缺少专辑 ID")
	}
	resp, err := s.client.AlbumDetail(context.Background(), kg.AlbumDetailRequest{Id: albumID})
	if err != nil {
		s.log.Warn("failed to fetch album detail", "albumID", albumID, "err", err.Error())
		return nil, err
	}
	if !kgapi.StatusOK(resp.Body) {
		return nil, errors.New("album_detail 接口业务失败")
	}

	item := kgapi.FirstRecord(resp.Body["data"])
	if item == nil {
		return nil, errors.New("专辑详情为空")
	}
	detail := catalogmap.AlbumDetail(item)
	return &detail, nil
}

// GetAlbumSongs 获取专辑歌曲，按页返回。
func (s *Service) GetAlbumSongs(albumID string, page int, pageSize int) (*SongsResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.fetchAlbumSongs(albumID, page, pageSize)
}

// GetAllAlbumSongs 获取专辑全部歌曲，供“播放全部”使用。
func (s *Service) GetAllAlbumSongs(albumID string) ([]dto.Song, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	songs := make([]dto.Song, 0, maxSongPageSize)
	seen := make(map[int64]struct{})
	for page := 1; page <= maxAllSongsPages; page++ {
		result, err := s.fetchAlbumSongs(albumID, page, maxSongPageSize)
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
	s.log.Warn("album song fetch hit page cap", "pages", maxAllSongsPages)
	return songs, nil
}

func (s *Service) fetchAlbumSongs(albumID string, page int, pageSize int) (*SongsResult, error) {
	albumID = strings.TrimSpace(albumID)
	if albumID == "" {
		return nil, errors.New("缺少专辑 ID")
	}
	page, pageSize = paging.Normalize(page, pageSize, defaultSongPageSize)

	resp, err := s.client.AlbumSongs(context.Background(), kg.AlbumSongsRequest{
		Id:       albumID,
		Page:     page,
		Pagesize: pageSize,
	})
	if err != nil {
		s.log.Warn("failed to fetch album songs", "albumID", albumID, "err", err.Error())
		return nil, err
	}
	if !kgapi.StatusOK(resp.Body) {
		return nil, errors.New("album_songs 接口业务失败")
	}

	data := kgapi.RecordOf(resp.Body["data"])
	rawItems := kgapi.Records(data["songs"])
	total := kgapi.Int64(data["total"])
	if total <= 0 {
		total = kgapi.Int64(resp.Body["total"])
	}

	result := &SongsResult{
		Songs: catalogmap.AlbumSongs(rawItems, play.PriorityForAny(play.DefaultQuality)),
		Total: total,
	}
	if total > 0 {
		result.HasMore = int64((page-1)*pageSize+len(rawItems)) < total
	} else {
		result.HasMore = len(rawItems) >= pageSize
	}
	return result, nil
}
