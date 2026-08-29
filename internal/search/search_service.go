package search

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"kuwave/internal/auth"
	"kuwave/internal/imageurl"
)

const defaultPageSize = 15

var supportedSearchTypes = map[string]struct{}{
	"song":    {},
	"album":   {},
	"author":  {},
	"mv":      {},
	"lyric":   {},
	"special": {},
}

// SearchService 搜索服务。
//
// 普通搜索与综合搜索分别对应 SDK 的 Search 和 SearchComplex 接口；
// 前端通过 Wails 绑定调用其导出方法。
type SearchService struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化 client 调用。
	mu sync.Mutex
}

// NewSearchService 创建搜索服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func NewSearchService(log *slog.Logger) (*SearchService, error) {
	client, err := kg.New(kg.WithAutoRefresh(false), kg.WithCookie(auth.LoadSession()))
	if err != nil {
		return nil, err
	}
	return &SearchService{client: client, log: log}, nil
}

// Search 执行普通搜索。
// searchType 支持 song、album、author、mv、lyric、special，值会原样透传给 SDK。
func (s *SearchService) Search(keyword string, pageNum int, pageSize int, searchType string) (*SearchResult, error) {
	if _, ok := supportedSearchTypes[searchType]; !ok {
		return nil, fmt.Errorf("不支持的搜索类型: %s", searchType)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pageNum, pageSize = normalizePage(pageNum, pageSize)
	cookie, err := s.loadSearchCookie()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Search(context.Background(), kg.SearchRequest{
		Keywords: keyword,
		Page:     pageNum,
		Pagesize: pageSize,
		Type:     searchType,
		Cookie:   cookie,
	})
	if err != nil {
		s.log.Warn("failed to search", "keyword", keyword, "type", searchType, "err", err.Error())
		return nil, err
	}
	if err := validateResponse(resp.Body); err != nil {
		return nil, err
	}

	return mapNormalSearch(searchType, searchData(resp.Body)), nil
}

// SearchComplex 执行综合搜索，不向 SDK 传递 type 参数。
func (s *SearchService) SearchComplex(keyword string, pageNum int, pageSize int) (*SearchResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pageNum, pageSize = normalizePage(pageNum, pageSize)
	cookie, err := s.loadSearchCookie()
	if err != nil {
		return nil, err
	}
	resp, err := s.client.SearchComplex(context.Background(), kg.SearchComplexRequest{
		Keywords: keyword,
		Page:     pageNum,
		Pagesize: pageSize,
		Cookie:   cookie,
		Extra: map[string]any{
			"cookie": serializeSearchCookie(cookie),
		},
	})
	if err != nil {
		s.log.Warn("failed to search complex", "keyword", keyword, "err", err.Error())
		return nil, err
	}
	if err := validateResponse(resp.Body); err != nil {
		return nil, err
	}

	return mapComplexSearch(searchData(resp.Body)), nil
}

// loadSearchCookie 读取最新登录态，并同步到搜索 client。
// SearchService 与 AuthService 持有不同 SDK client，因此不能只在创建时读取一次。
func (s *SearchService) loadSearchCookie() (map[string]string, error) {
	cookie := auth.LoadSession()
	for _, key := range []string{"token", "userid", "dfid"} {
		if cookie[key] == "" {
			return nil, errors.New("缺少搜索所需的登录身份信息")
		}
		s.client.SetCookie(key, cookie[key])
	}
	return map[string]string{
		"token":  cookie["token"],
		"userid": cookie["userid"],
		"dfid":   cookie["dfid"],
	}, nil
}

func serializeSearchCookie(cookie map[string]string) string {
	return "token=" + cookie["token"] + ";userid=" + cookie["userid"] + ";dfid=" + cookie["dfid"]
}

func normalizePage(pageNum int, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	return pageNum, pageSize
}

func validateResponse(body map[string]any) error {
	if fmt.Sprintf("%v", body["status"]) != "1" {
		return errors.New("搜索接口业务失败")
	}
	return nil
}

func searchData(body map[string]any) map[string]any {
	data, _ := body["data"].(map[string]any)
	normalizeImageFields(data)
	return data
}

// normalizeImageFields 规范化搜索结果中的酷狗图片地址。
// 搜索响应保留 SDK 的动态结构，因此在递归遍历时只处理已知图片字段，
// 不改变其它字段的类型和值。
func normalizeImageFields(value any) {
	switch item := value.(type) {
	case map[string]any:
		for key, field := range item {
			if isImageField(key) {
				if raw, ok := field.(string); ok {
					item[key] = imageurl.Normalize(raw)
				}
				continue
			}
			normalizeImageFields(field)
		}
	case []any:
		for _, field := range item {
			normalizeImageFields(field)
		}
	}
}

func isImageField(key string) bool {
	switch strings.ToLower(key) {
	case "image", "img", "imgurl", "avatar", "pic", "erectpic":
		return true
	default:
		return false
	}
}
