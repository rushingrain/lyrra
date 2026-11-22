package playlist

import (
	"context"
	"errors"
	"strconv"
	"strings"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/auth"
	"lyrra/internal/catalogmap"
	"lyrra/internal/kgapi"
)

// 账号歌单读取与归属校验；所有内部方法均由持有 s.mu 的调用方使用。
func (s *Service) loginCookie() (map[string]string, error) {
	cookie := s.syncLoginCookie()
	userid, err := strconv.ParseInt(cookie["userid"], 10, 64)
	if err != nil || userid <= 0 || cookie["token"] == "" {
		return nil, auth.ErrNotLoggedIn
	}
	return cookie, nil
}

func sameSession(cookie map[string]string) error {
	latest := auth.LoginCookie(auth.LoadSession())
	if cookie["userid"] != latest["userid"] || cookie["token"] != latest["token"] {
		return errors.New("登录账号已变化，请重新操作")
	}
	return nil
}

func (s *Service) fetchUserPlaylists(page, size int, cookie map[string]string) ([]kgapi.Record, int, error) {
	userid, _ := strconv.Atoi(cookie["userid"])
	resp, err := s.client.UserPlaylist(context.Background(), kg.UserPlaylistRequest{
		Page: page, Pagesize: size, Userid: userid, Cookie: cookie,
	})
	if err != nil {
		return nil, 0, errors.New("获取用户歌单失败，请稍后重试")
	}
	if err := playlistResponseError(resp, "获取用户歌单失败，请重新登录或稍后重试"); err != nil {
		return nil, 0, err
	}
	data := kgapi.RecordOf(resp.Body["data"])
	if _, ok := data["info"].([]any); !ok {
		return nil, 0, errors.New("用户歌单响应缺少列表数据")
	}
	return kgapi.Records(data["info"]), kgapi.Int(data["total_ver"]), sameSession(cookie)
}

// findUserPlaylist 从当前账号的列表验证归属，不能信任前端传来的 type/listid。
// 权限查询允许查找后续页，不改变侧边栏只展示首批 150 条的约定。
func (s *Service) findUserPlaylist(globalID string, cookie map[string]string) (kgapi.Record, int, error) {
	globalID = strings.TrimSpace(globalID)
	if globalID == "" || globalID == "0" {
		return nil, 0, errors.New("缺少有效的歌单 ID")
	}
	const size = 150
	seen := make(map[string]bool)
	for page := 1; page <= 100; page++ {
		items, version, err := s.fetchUserPlaylists(page, size, cookie)
		if err != nil {
			return nil, 0, err
		}
		fresh := false
		for _, item := range items {
			// 收藏项展示原歌单，写操作仍定位当前账号的条目；兼容旧详情路由。
			if kgapi.Scalar(item["global_collection_id"]) == globalID || catalogmap.UserPlaylist(item).GlobalCollectionID == globalID {
				return item, version, nil
			}
			key := kgapi.Scalar(item["listid"])
			if !seen[key] {
				seen[key] = true
				fresh = true
			}
		}
		if len(items) < size {
			return nil, version, nil
		}
		if !fresh {
			return nil, 0, errors.New("用户歌单分页未前进，无法确认歌单归属")
		}
	}
	return nil, 0, errors.New("用户歌单过多，无法确认歌单归属")
}

func requireCreated(item kgapi.Record, editInfo bool) error {
	playlist := catalogmap.UserPlaylist(item)
	if playlist.Type != 0 || kgapi.Int64(item["listid"]) <= 0 {
		return errors.New("只能修改自己创建的歌单")
	}
	if editInfo && playlist.IsDefault != 0 {
		return errors.New("默认收藏和我喜欢不允许编辑信息或删除")
	}
	return nil
}

// GetUserPlaylist 查询当前账号与指定歌单的关系；未创建或收藏时返回 null。
func (s *Service) GetUserPlaylist(globalID string) (*UserPlaylist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cookie, err := s.loginCookie()
	if err != nil {
		return nil, err
	}
	item, _, err := s.findUserPlaylist(globalID, cookie)
	if err != nil || item == nil {
		return nil, err
	}
	playlist := catalogmap.UserPlaylist(item)
	_, hasIntro := item["intro"]
	_, hasTags := item["tags"]
	if playlist.Type == 0 && playlist.IsDefault == 0 && (!hasIntro || !hasTags) {
		result, err := s.fetchPlaylistSongs(globalID, 1, 1)
		if err != nil {
			return nil, err
		}
		playlist = catalogmap.UserPlaylistWithInfo(item, result.Info)
	}
	if err := sameSession(cookie); err != nil {
		return nil, err
	}
	return &playlist, nil
}

func playlistResponseError(resp *kg.Response, message string) error {
	if resp == nil || resp.Status >= 400 || !kgapi.StatusOK(resp.Body) {
		return errors.New(message)
	}
	for _, key := range []string{"error_code", "errcode"} {
		if code := kgapi.Scalar(resp.Body[key]); code != "" && code != "0" {
			return errors.New(message)
		}
	}
	return nil
}
