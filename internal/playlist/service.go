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

	"kuwave/internal/dto"
	"kuwave/internal/imageurl"
)

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
			ID:          int64(pickInt(m, "specialid")),
			Name:        pickString(m, "specialname"),
			Creator:     pickString(m, "nickname"),
			CoverURL:    imageurl.Normalize(pickString(m, "imgurl")),
			Description: pickString(m, "intro"),
			PlayCount:   int64(pickInt(m, "play_count")),
			SongCount:   int64(pickInt(m, "song_count")),
		})
	}
	return playlists, nil
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
