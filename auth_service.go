package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	kg "github.com/lfhy/kugou-music-api"
)

// ErrNotLoggedIn 未登录或会话已过期
var ErrNotLoggedIn = errors.New("未登录")

// AuthService 登录服务。
//
// 结构体实例常驻进程，持有 *kg.Client 作为全局登录态；
// 前端通过 Wails 绑定调用其导出方法（window.go.main.AuthService.xxx）。
type AuthService struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 内部维护 cookie 与登录态，非并发安全，
	// 用互斥锁串行化所有 client 调用。
	mu sync.Mutex
}

// NewAuthService 创建登录服务。
// 只做本地初始化（加载已保存的会话 cookie 注入 client），
// 不发起网络请求，因此不会阻塞应用启动。
// 会话有效性校验延迟到首次 GetCurrentUser 调用。
func NewAuthService(log *slog.Logger) (*AuthService, error) {
	client, err := kg.New(
		kg.WithCookie(loadSession()),
		// TODO: token 续期功能暂未启用；后续接入续期时改为 kg.WithAutoRefresh(true)，
		// 并处理会话过期（登录态失效）时的行为。
		kg.WithAutoRefresh(false),
	)
	if err != nil {
		return nil, err
	}
	return &AuthService{client: client, log: log}, nil
}

// GetLoginQrCode 获取扫码登录二维码。
// 对应 SDK 三步流程的 Step1 + Step2：LoginQrKey → LoginQrCreate。
func (s *AuthService) GetLoginQrCode() (*QrLoginResponse, error) {
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
func (s *AuthService) CheckQrLoginStatus(qrKey string) (string, error) {
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
		// 登录成功，token/userid 已由 SDK 自动注入 client，此刻持久化会话
		saveSession(s.client.Cookie())
		s.log.Info("QR login succeeded")
		return string(QrLoginConfirmed), nil
	default:
		return string(QrLoginPending), nil
	}
}

// GetCurrentUser 获取当前登录用户信息。
// 未登录或会话已过期时返回 ErrNotLoggedIn。
func (s *AuthService) GetCurrentUser() (*UserInfo, error) {
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
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, ErrNotLoggedIn
	}
	if code := strings.TrimSpace(fmt.Sprintf("%v", resp.Body["error_code"])); code != "" && code != "<nil>" && code != "0" && code != "0.0" {
		return nil, ErrNotLoggedIn
	}

	// UserDetail 的用户字段都在 data 子结构下（实测：data.nickname / data.pic，
	// 以及好友/关注/粉丝数 data.friends / data.follows / data.fans）
	data, _ := resp.Body["data"].(map[string]any)
	user := &UserInfo{
		Nickname: pickString(data, "nickname", "k_nickname", "fx_nickname"),
		Avatar:   normalizeAvatarUrl(pickString(data, "pic", "k_pic", "fx_pic", "avatar")),
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
	data, _ := vipResp.Body["data"].(map[string]any)
	list, _ := data["busi_vip"].([]any)
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if strings.TrimSpace(fmt.Sprintf("%v", m["busi_type"])) != "concept" {
			continue
		}
		if v := strings.TrimSpace(fmt.Sprintf("%v", m["is_vip"])); v == "1" || v == "1.0" {
			return true
		}
	}
	return false
}

// Logout 登出：清空登录态 cookie 并删除本地会话文件。
func (s *AuthService) Logout() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client.SetCookie("token", "")
	s.client.SetCookie("userid", "")
	s.client.SetCookie("vip_token", "")
	s.client.SetCookie("vip_type", "")
	clearSession()
	s.log.Info("logged out")
	return nil
}

// normalizeAvatarUrl 将酷狗图床地址统一为 https。
//
// 背景：wails dev 模式下应用以 .app 形式运行，受 App Transport Security 限制，
// Info.dev.plist 只放行本地网络（NSAllowsLocalNetworking），远程 http:// 资源会被
// WKWebView 拦截，导致头像加载失败；而 GoLand 直接运行裸二进制（无 bundle）时 ATS
// 不生效，同样的 http 地址却能加载，两种启动方式行为不一致。
// 酷狗图床同时支持 https（实测 https://imge.kugou.com 返回 200 + image/jpeg），
// 统一替换后各启动方式（wails dev / IDE / 生产构建）行为一致。
func normalizeAvatarUrl(raw string) string {
	if strings.HasPrefix(raw, "http://") {
		return "https://" + strings.TrimPrefix(raw, "http://")
	}
	return raw
}

// pickInt 在嵌套 map 中按 key 依次查找第一个可解析为整数的值。
// 兼容 JSON 反序列化后的 float64 / json.Number 及字符串形式的数字。
func pickInt(v any, keys ...string) int {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return 0
	}
	for _, k := range keys {
		x, ok := m[k]
		if !ok {
			continue
		}
		switch n := x.(type) {
		case float64:
			return int(n)
		case float32:
			return int(n)
		case int:
			return n
		case int64:
			return int(n)
		case json.Number:
			if i, err := n.Int64(); err == nil {
				return int(i)
			}
		case string:
			if i, err := strconv.Atoi(strings.TrimSpace(n)); err == nil {
				return i
			}
		}
	}
	return 0
}

// pickString 在嵌套 map 中按 key 依次查找第一个非空字符串，
// 兼容 SDK Response.Body 的 data 多层级结构。
func pickString(v any, keys ...string) string {
	m, ok := v.(map[string]any)
	if !ok || m == nil {
		return ""
	}
	for _, k := range keys {
		if x, ok := m[k]; ok {
			if s, ok := x.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}
	for _, x := range m {
		if mm, ok := x.(map[string]any); ok {
			if s := pickString(mm, keys...); s != "" {
				return s
			}
		}
	}
	return ""
}
