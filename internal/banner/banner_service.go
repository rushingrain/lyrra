package banner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	kg "github.com/rushingrain/kugou-music-api"
)

// BannerService 首页 banner 服务。
//
// 结构体实例常驻进程，持有独立的 *kg.Client（banner 接口无需登录态，
// 不复用 auth 的 client，避免跨服务耦合）；前端通过 Wails 绑定调用
// 其导出方法（window.go.main.BannerService.xxx）。
type BannerService struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化所有 client 调用（同 auth 包）。
	mu sync.Mutex
}

// NewBannerService 创建 banner 服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func NewBannerService(log *slog.Logger) (*BannerService, error) {
	client, err := kg.New(kg.WithAutoRefresh(false))
	if err != nil {
		return nil, err
	}
	return &BannerService{client: client, log: log}, nil
}

// GetHomeBanners 拉取首页轮播 banner（/pc/diantai 电台 banner 接口）。
//
// 该接口返回广告与内容混合列表，isAd 为 true 的第三方推广项（直播/HiFi 等）
// 直接过滤；剩余项 code 为 banner 图片地址、url 为点击跳转链接。
func (s *BannerService) GetHomeBanners() ([]HomeBanner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 上游 adservice.kugou.com 偶发瞬时网络故障（连接被重置/EOF），
	// 做有限重试避免首页冷启动时误报“轮播加载失败”。
	var resp *kg.PcDiantaiResponse
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		resp, err = s.client.PcDiantai(context.Background(), kg.PcDiantaiRequest{})
		if err == nil {
			break
		}
		s.log.Warn("failed to fetch home banners", "attempt", attempt, "err", err.Error())
		if attempt < 3 {
			time.Sleep(time.Duration(attempt) * 300 * time.Millisecond)
		}
	}
	if err != nil {
		return nil, err
	}
	// 业务成功判断同 auth 包：酷狗 API 用 body 里 status=="1" 且 error_code==0
	// 表示成功，resp.Status 是 HTTP 状态码，不能用于业务判断。
	if fmt.Sprintf("%v", resp.Body["status"]) != "1" {
		return nil, errors.New("diantai 接口业务失败")
	}

	data, _ := resp.Body["data"].(map[string]any)
	items, _ := data["data"].([]any)
	banners := make([]HomeBanner, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// 过滤第三方广告
		if isAd, _ := m["isAd"].(bool); isAd {
			continue
		}
		banners = append(banners, HomeBanner{
			ID:       fmt.Sprintf("%v", m["id"]),
			ImageURL: normalizeImageURL(pickString(m, "code")),
			Title:    pickString(m, "title"),
			LinkURL:  pickString(m, "url"),
		})
	}
	return banners, nil
}

// pickString 从 item map 提取字符串字段。
func pickString(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

// normalizeImageURL 将酷狗图床地址统一为 https。
//
// 背景同 auth 包 normalizeAvatarUrl：wails dev 模式下应用以 .app 形式运行，
// 受 App Transport Security 限制，Info.dev.plist 只放行本地网络，远程 http://
// 资源会被 WKWebView 拦截，导致 banner 图加载失败。非广告 banner 图均在
// mdpfile 图床（实测 https 可用），统一替换后各启动方式行为一致。
// 广告位用的 adsfile 图床不支持 https，但广告已在 GetHomeBanners 中过滤。
func normalizeImageURL(raw string) string {
	if strings.HasPrefix(raw, "http://") {
		return "https://" + strings.TrimPrefix(raw, "http://")
	}
	return raw
}
