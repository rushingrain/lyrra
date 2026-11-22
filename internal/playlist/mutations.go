package playlist

import (
	"context"
	"errors"
	"strconv"
	"strings"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/dto"
	"lyrra/internal/kgapi"
)

// CreateUserPlaylist 新建公开的自定义歌单。
func (s *Service) CreateUserPlaylist(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("请输入歌单名称")
	}
	cookie, err := s.loginCookie()
	if err != nil {
		return err
	}
	resp, err := s.client.PlaylistAdd(context.Background(), kg.PlaylistAddRequest{
		Name: name, Type: 0, IsPri: 0, Source: 1,
		ListCreateUserid: cookie["userid"], ListCreateListid: 0, ListCreateGid: "", Cookie: cookie,
	})
	if err != nil {
		return errors.New("创建歌单失败，请稍后重试")
	}
	return playlistResponseError(resp, "创建歌单失败，请稍后重试")
}

// CollectPlaylist 从上游详情读取原作者与原始 listid，不接收前端提供的归属信息。
func (s *Service) CollectPlaylist(globalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cookie, err := s.loginCookie()
	if err != nil {
		return err
	}
	item, _, err := s.findUserPlaylist(globalID, cookie)
	if err != nil {
		return err
	}
	if item != nil {
		return errors.New("该歌单已经在你的歌单列表中")
	}
	resp, err := s.client.PlaylistTrackAll(context.Background(), kg.PlaylistTrackAllRequest{
		Id: globalID, Page: 1, Pagesize: 1, Cookie: cookie,
	})
	if err != nil {
		return errors.New("获取原歌单信息失败")
	}
	if err := playlistResponseError(resp, "获取原歌单信息失败"); err != nil {
		return err
	}
	info := kgapi.RecordOf(kgapi.RecordOf(resp.Body["data"])["list_info"])
	creator := kgapi.Scalar(info["list_create_userid"])
	listID := kgapi.Scalar(info["list_create_listid"])
	name := kgapi.Text(info["name"])
	if kgapi.Int64(creator) <= 0 || kgapi.Int64(listID) <= 0 || name == "" {
		return errors.New("上游未返回收藏所需的原歌单信息")
	}
	if err := sameSession(cookie); err != nil {
		return err
	}
	added, err := s.client.PlaylistAdd(context.Background(), kg.PlaylistAddRequest{
		Name: name, Type: 1, Source: 1, ListCreateUserid: creator,
		ListCreateListid: listID, ListCreateGid: kgapi.Scalar(info["list_create_gid"]), Cookie: cookie,
	})
	if err != nil {
		return errors.New("收藏歌单失败，请稍后重试")
	}
	return playlistResponseError(added, "收藏歌单失败，请稍后重试")
}

// DeleteUserPlaylist 对自建歌单执行删除，对收藏歌单执行取消收藏。
func (s *Service) DeleteUserPlaylist(globalID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cookie, err := s.loginCookie()
	if err != nil {
		return err
	}
	item, _, err := s.findUserPlaylist(globalID, cookie)
	if err != nil {
		return err
	}
	kind, known := kgapi.IntOK(item["type"])
	if !known || (kind != 0 && kind != 1) || kgapi.Int64(item["listid"]) <= 0 {
		return errors.New("无法删除此歌单")
	}
	if kind == 0 {
		if err := requireCreated(item, true); err != nil {
			return err
		}
	}
	if err := sameSession(cookie); err != nil {
		return err
	}
	resp, err := s.client.PlaylistDel(context.Background(), kg.PlaylistDelRequest{
		Listid: kgapi.Int(item["listid"]), Cookie: cookie,
	})
	if err != nil {
		return errors.New("删除或取消收藏歌单失败，请稍后重试")
	}
	return playlistResponseError(resp, "删除或取消收藏歌单失败，请稍后重试")
}

// AddPlaylistSong 将歌曲加入当前账号的自建歌单（含两个特殊歌单）。
func (s *Service) AddPlaylistSong(globalID string, song dto.Song) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cookie, err := s.loginCookie()
	if err != nil {
		return err
	}
	item, _, err := s.findUserPlaylist(globalID, cookie)
	if err != nil {
		return err
	}
	if err := requireCreated(item, false); err != nil {
		return err
	}
	if strings.TrimSpace(song.Name) == "" || strings.TrimSpace(song.Hash) == "" || song.ID <= 0 {
		return errors.New("歌曲信息不完整，无法添加到歌单")
	}
	name := song.Name
	if song.Artist != "" {
		name = song.Artist + " - " + name
	}
	// SDK 的参数格式使用逗号与竖线分隔，拒绝会被拆成多条资源的输入。
	if strings.ContainsAny(name+song.Hash, ",|") {
		return errors.New("歌曲信息含接口不支持的分隔符，无法添加")
	}
	albumID, _ := strconv.ParseInt(song.AlbumID, 10, 64)
	data := name + "|" + song.Hash + "|" + strconv.FormatInt(albumID, 10) + "|" + strconv.FormatInt(song.ID, 10)
	if err := sameSession(cookie); err != nil {
		return err
	}
	resp, err := s.client.PlaylistTracksAdd(context.Background(), kg.PlaylistTracksAddRequest{
		Listid: kgapi.Int(item["listid"]), Data: data, Cookie: cookie,
	})
	if err != nil {
		return errors.New("添加歌曲失败，请稍后重试")
	}
	return playlistResponseError(resp, "添加歌曲失败，请稍后重试")
}

// RemovePlaylistSong 只接受歌单返回的 fileid，不能用 mixsongid 或 hash 代替。
func (s *Service) RemovePlaylistSong(globalID, fileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, err := strconv.ParseInt(fileID, 10, 64)
	if err != nil || id <= 0 {
		return errors.New("缺少歌单歌曲条目 ID，无法移除")
	}
	cookie, err := s.loginCookie()
	if err != nil {
		return err
	}
	item, _, err := s.findUserPlaylist(globalID, cookie)
	if err != nil {
		return err
	}
	if err := requireCreated(item, false); err != nil {
		return err
	}
	if err := sameSession(cookie); err != nil {
		return err
	}
	resp, err := s.client.PlaylistTracksDel(context.Background(), kg.PlaylistTracksDelRequest{
		Listid: kgapi.Int(item["listid"]), Fileids: fileID, Cookie: cookie,
	})
	if err != nil {
		return errors.New("移除歌曲失败，请稍后重试")
	}
	return playlistResponseError(resp, "移除歌曲失败，请稍后重试")
}
