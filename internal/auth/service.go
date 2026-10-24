package auth

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"
	"time"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/imageurl"
	"lyrra/internal/kgapi"
	"lyrra/internal/paging"
	"lyrra/internal/playerstate"
)

// ErrNotLoggedIn 未登录或会话已过期
var ErrNotLoggedIn = errors.New("未登录")

// Service 登录服务。
//
// 结构体实例常驻进程，持有 *kg.Client 作为全局登录态；
// 前端通过 Wails 绑定调用其导出方法（window.go.auth.Service.xxx）。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 内部维护 cookie 与登录态，非并发安全，
	// 用互斥锁串行化所有 client 调用。
	mu sync.Mutex
}

// New 创建登录服务。
// 只做本地初始化（加载已保存的会话 cookie 注入 client），
// 不发起网络请求，因此不会阻塞应用启动。
// 会话有效性校验延迟到首次 GetCurrentUser 调用。
func New(log *slog.Logger) (*Service, error) {
	client, err := kgapi.NewClient(
		kg.WithCookie(LoadSession()),
		// TODO: token 续期功能暂未启用；后续接入续期时改为 kg.WithAutoRefresh(true)，
		// 并处理会话过期（登录态失效）时的行为。
		kg.WithAutoRefresh(false),
	)
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// GetLoginQrCode 获取扫码登录二维码。
// 对应 SDK 三步流程的 Step1 + Step2：LoginQrKey → LoginQrCreate。
func (s *Service) GetLoginQrCode() (*QrLoginResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	keyResp, err := s.client.LoginQrKey(context.Background(), kg.LoginQrKeyRequest{})
	if err != nil {
		s.log.Warn("failed to get login QR key", "err", err.Error())
		return nil, err
	}
	key := keyResp.QRCodeKey()
	if key == "" {
		return nil, errors.New("获取二维码 key 失败")
	}

	qrResp, err := s.client.LoginQrCreate(context.Background(), kg.LoginQrCreateRequest{
		Key:   key,
		Qrimg: true, // 本地生成 base64 PNG，返回 data:image/png;base64,...
	})
	if err != nil {
		s.log.Warn("failed to create login QR code", "err", err.Error())
		return nil, err
	}

	return &QrLoginResponse{
		QrKey:     key,
		QrUrl:     qrResp.Base64(),
		ExpiresAt: time.Now().Add(5 * time.Minute).UnixMilli(),
	}, nil
}

// CheckQrLoginStatus 查询扫码登录状态（单次查询，由前端每 2~3 秒轮询调用）。
// 单次请求失败时返回 pending，交由前端下一个轮询 tick 重试。
// 返回 string 而非 QrLoginStatus：Wails 的 TS 绑定生成器不产出命名 string 类型的声明，
// 返回 string 可避免生成的 .d.ts 引用不存在的 main.QrLoginStatus 导致前端类型检查失败。
func (s *Service) CheckQrLoginStatus(qrKey string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	checkResp, err := s.client.LoginQrCheck(context.Background(), kg.LoginQrCheckRequest{Key: qrKey})
	if err != nil {
		s.log.Debug("QR login status check failed", "err", err.Error())
		return string(QrLoginPending), nil
	}
	switch checkResp.StatusCode() {
	case 0:
		// 二维码已过期，前端会刷新二维码
		s.log.Info("login QR code expired")
		return string(QrLoginExpired), nil
	case 2:
		// 已扫码，等待用户在手机上确认
		s.log.Info("QR code scanned, awaiting confirmation")
		return string(QrLoginScanned), nil
	case 4:
		// 这里不再调用 SDK 的 RegisterDev 注册设备：
		// 它内部会在登录态校验失败时刷新 token，而刷新走的是明文 http
		// （login.user.kugou.com，见 SDK 的 RouteLoginToken），会把 token 暴露在网上；
		// 而设备注册本来就只是拿 dfid，服务端还可能"业务成功但不下发"（返回空 data）。
		// 现在改为：复用会话里已有的 dfid，没有就本地生成一个（见 dfid.go），
		// 搜索、用户歌单等依赖 dfid 的接口照常可用。
		s.client.SetCookie("userid", normalizeUserIDString(s.client.Cookie()["userid"]))
		s.ensureDfid()
		saveSession(s.client.Cookie())
		s.log.Info("QR login succeeded")
		return string(QrLoginConfirmed), nil
	default:
		return string(QrLoginPending), nil
	}
}

// GetCurrentUser 获取当前登录用户信息。
// 未登录或会话已过期时返回 ErrNotLoggedIn。
func (s *Service) GetCurrentUser() (*UserInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: 会话过期处理行为暂未定义（token 续期功能未启用）。
	// 当前行为：UserDetail 失败即视为未登录，前端静默进入未登录态。
	resp, err := s.client.UserDetail(context.Background(), kg.UserDetailRequest{})
	if err != nil {
		s.log.Warn("failed to fetch user info, treating as not logged in", "err", err.Error())
		return nil, ErrNotLoggedIn
	}

	// 业务成功判断：酷狗 API 用 body 里的 status=="1" 且 error_code==0 表示成功，
	// resp.Status 是 HTTP 状态码，不能用于业务判断。
	if !kgapi.StatusOK(resp.Body) {
		return nil, ErrNotLoggedIn
	}
	if code := kgapi.Scalar(resp.Body["error_code"]); code != "" && code != "0" {
		return nil, ErrNotLoggedIn
	}

	// UserDetail 的用户字段都在 data 子结构下（实测：data.nickname / data.pic，
	// 以及好友/关注/粉丝数 data.friends / data.follows / data.fans）
	data := kgapi.RecordOf(resp.Body["data"])
	user := &UserInfo{
		Nickname: pickString(data, "nickname", "k_nickname", "fx_nickname"),
		Avatar:   imageurl.Normalize(pickString(data, "pic", "k_pic", "fx_pic", "avatar")),
		Friends:  pickInt(data, "friends"),
		Follows:  pickInt(data, "follows"),
		Fans:     pickInt(data, "fans"),
	}
	// VIP 判断用权威接口 UserVipDetail：概念版 VIP 在 data.busi_vip 中 busi_type=="concept" 的条目。
	// （UserDetail 的 data.vip_type / cookie 的 vip_type 对概念版账号恒为 0，不能用于判断）
	vipResp, vipErr := s.client.UserVipDetail(context.Background(), kg.UserVipDetailRequest{})
	user.IsVip = vipErr == nil && isConceptVip(vipResp)
	return user, nil
}

// isConceptVip 判断用户是否为酷狗概念版 VIP。
// UserVipDetail 响应 data.busi_vip 为各业务线的会员状态数组；
// 概念版对应 busi_type=="concept"，其中任一条目 is_vip==1 即为 VIP。
func isConceptVip(vipResp *kg.UserVipDetailResponse) bool {
	if vipResp == nil || vipResp.Body == nil {
		return false
	}
	data := kgapi.RecordOf(vipResp.Body["data"])
	for _, item := range kgapi.Records(data["busi_vip"]) {
		if kgapi.Scalar(item["busi_type"]) != "concept" {
			continue
		}
		if value := kgapi.Scalar(item["is_vip"]); value == "1" || value == "1.0" {
			return true
		}
	}
	return false
}

// Logout 登出：清空登录态 cookie 并删除本地会话文件。
func (s *Service) Logout() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client.SetCookie("token", "")
	s.client.SetCookie("userid", "")
	s.client.SetCookie("dfid", "")
	s.client.SetCookie("vip_token", "")
	s.client.SetCookie("vip_type", "")
	clearSession()
	// 账号级播放数据（播放队列 / 当前歌曲）随登出清除；设备级设置（音量 / 播放模式）保留。
	if err := playerstate.ClearAccount(); err != nil {
		s.log.Warn("failed to clear player account state", "err", err.Error())
	}
	s.log.Info("logged out")
	return nil
}

// GetUserPlaylists 拉取当前登录用户的歌单列表（/v7/get_all_list）。
// page / pagesize 由前端传入：page<=0 回退 1，pagesize<=0 回退 20。
//
// 兼容说明：历史登录流程可能把 userid 存成科学计数法字符串（如 "8.63006722e+08"），
// 直接放入 Cookie 会导致接口返回 20010 未登录。这里规整为十进制整数后
// 通过请求参数 + 请求 Cookie 覆盖两种方式显式传入，与手动排查时验证有效的请求一致。
func (s *Service) GetUserPlaylists(page int, pagesize int) ([]UserPlaylist, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	page, pagesize = paging.Normalize(page, pagesize, 20)
	userid := normalizeUserID(s.client.Cookie()["userid"])

	resp, err := s.client.UserPlaylist(context.Background(), kg.UserPlaylistRequest{
		Page:     page,
		Pagesize: pagesize,
		Userid:   userid,
		Cookie: map[string]string{
			"userid": strconv.FormatInt(int64(userid), 10),
		},
	})
	if err != nil {
		s.log.Warn("failed to fetch user playlists", "err", err.Error())
		return nil, err
	}
	if !kgapi.StatusOK(resp.Body) {
		s.log.Warn("user playlists business failure", "error_code", resp.Body["error_code"])
		return nil, ErrNotLoggedIn
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

// normalizeUserID 把可能为科学计数法（如 "8.63006722e+08"）的 userid 规整为十进制整数。
func normalizeUserID(raw string) int {
	uid := normalizeUserIDString(raw)
	if uid == "" {
		return 0
	}
	i, err := strconv.ParseInt(uid, 10, 64)
	if err != nil {
		return 0
	}
	return int(i)
}

// pickInt 在嵌套 map 中按 key 依次查找第一个可解析为整数的值。
// 兼容 JSON 反序列化后的 float64 / json.Number 及字符串形式的数字。
func pickInt(v any, keys ...string) int {
	record := kgapi.RecordOf(v)
	if record == nil {
		return 0
	}
	for _, key := range keys {
		value, ok := record[key]
		if !ok {
			continue
		}
		if number, ok := kgapi.IntOK(value); ok {
			return number
		}
	}
	return 0
}

// pickString 在嵌套 map 中按 key 依次查找第一个非空字符串，
// 兼容 SDK Response.Body 的 data 多层级结构。
func pickString(v any, keys ...string) string {
	record := kgapi.RecordOf(v)
	if record == nil {
		return ""
	}
	for _, key := range keys {
		if value, ok := record[key]; ok {
			if text := kgapi.Text(value); text != "" {
				return value.(string)
			}
		}
	}
	for _, value := range record {
		if text := pickString(value, keys...); text != "" {
			return text
		}
	}
	return ""
}
