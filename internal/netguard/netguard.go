// Package netguard 在进程级接管 http.DefaultTransport：默认拒绝非 loopback 的明文 http 请求。
//
// 背景：项目使用的 SDK 里有若干接口的 BaseURL 写的是 http://，其中两条会携带登录凭证
// （token / vip_token / userid），例如 token 刷新（login.user.kugou.com）与播放地址回退
// （tracker.kugou.com）。调用侧已经尽量避免走到它们，但 SDK 内部仍可能触发；
// 这里做一层兜底：宁可让请求明确失败，也不把账号凭证放到明文网络上。
//
// 注意：本护栏只作用于 Go 侧的 HTTP 客户端。WebView 直接加载的音频流（http://fs*.kugou.com）
// 不经过这里，由 macOS 的 ATS 域名白名单控制（见 build/darwin/*.plist）。
package netguard

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
)

// allowedCleartextHosts 允许明文 http 的 host 白名单（不含端口，精确匹配）。
//
// 目前为空：我们需要的接口全部走 https。若将来确有必须走明文的接口，
// 在这里显式放行并写明理由，而不是放开整个护栏。
var allowedCleartextHosts = map[string]struct{}{}

var (
	installMu sync.Mutex
	installed bool
)

// Install 用受控 Transport 包装当前的 http.DefaultTransport。重复调用只生效一次。
//
// 需要尽早调用（在创建各业务服务之前），因为 SDK 的 http.Client 未设置 Transport，
// 请求时会取 http.DefaultTransport。
func Install(logger *slog.Logger) {
	installMu.Lock()
	defer installMu.Unlock()

	if installed {
		return
	}
	http.DefaultTransport = &guard{
		base:    http.DefaultTransport,
		logger:  logger,
		allowed: allowedCleartextHosts,
	}
	installed = true
}

// guard 是按 host 策略放行/拒绝明文请求的 RoundTripper。
type guard struct {
	base    http.RoundTripper
	logger  *slog.Logger
	allowed map[string]struct{}
}

func (g *guard) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := g.check(req); err != nil {
		if g.logger != nil {
			g.logger.Warn("blocked cleartext http request", "host", hostOf(req), "url", req.URL.String(), "err", err.Error())
		}
		return nil, err
	}
	return g.base.RoundTrip(req)
}

// check 返回非 nil 表示拒绝该请求。
func (g *guard) check(req *http.Request) error {
	if req == nil || req.URL == nil {
		return nil
	}
	if !strings.EqualFold(req.URL.Scheme, "http") {
		return nil
	}
	host := hostOf(req)
	if host == "" || isLoopback(host) {
		return nil
	}
	if _, ok := g.allowed[host]; ok {
		return nil
	}
	return fmt.Errorf("明文 http 请求被拒绝（host=%s）：账号凭证不允许走 http；如确需明文请显式加入白名单", host)
}

func hostOf(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	return strings.ToLower(req.URL.Hostname())
}

// isLoopback 判断 host 是否为本机回环地址（localhost / 127.0.0.0/8 / ::1）。
func isLoopback(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
