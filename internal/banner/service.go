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

	"kuwave/internal/dto"
	"kuwave/internal/imageurl"
)

// Service 首页 banner 服务。
//
// 结构体实例常驻进程，持有独立的 *kg.Client（banner 接口无需登录态，
// 不复用 auth 的 client，避免跨服务耦合）；前端通过 Wails 绑定调用
// 其导出方法（window.go.banner.Service.xxx）。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化所有 client 调用（同 auth 包）。
	mu sync.Mutex
}

// New 创建 banner 服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func New(log *slog.Logger) (*Service, error) {
	client, err := kg.New(kg.WithAutoRefresh(false))
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// GetHomeBanners 拉取首页轮播 banner（/pc/diantai 电台 banner 接口）。
//
// 该接口返回广告与内容混合列表，全部透传返回：广告项（直播/HiFi 等）与
// 内容项共用 code（图片地址）、url（跳转链接）字段。
func (s *Service) GetHomeBanners() ([]dto.Banner, error) {
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
	banners := make([]dto.Banner, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		banners = append(banners, dto.Banner{
			ID:       fmt.Sprintf("%v", m["id"]),
			ImageURL: bannerImageURL(pickString(m, "code")),
			Title:    pickString(m, "title"),
			LinkURL:  pickString(m, "url"),
		})
	}
	return banners, nil
}

// bannerImageURL 规范化 banner 图片地址：
// adsfile 图床（广告位）不支持 https（证书与域名不匹配，WKWebView 会拒绝），
// 保留 http 原样返回，配合 Info.plist 的 ATS 例外加载；
// 其余图床（mdpfile 等）统一转 https，避免被 ATS 拦截。
func bannerImageURL(raw string) string {
	if strings.HasPrefix(raw, "http://adsfile") {
		return raw
	}
	return imageurl.Normalize(raw)
}

// pickString 从 item map 提取字符串字段。
func pickString(m map[string]any, key string) string {
	if s, ok := m[key].(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}
