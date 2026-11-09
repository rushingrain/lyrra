package search

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/auth"
	"lyrra/internal/imageurl"
	"lyrra/internal/kgapi"
	"lyrra/internal/paging"
	"lyrra/internal/play"
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

// Service 搜索服务。
//
// 普通搜索与综合搜索分别对应 SDK 的 Search 和 SearchComplex 接口；
// 前端通过 Wails 绑定调用其导出方法。
type Service struct {
	client *kg.Client
	log    *slog.Logger
	// kg.Client 非并发安全，用互斥锁串行化 client 调用。
	mu sync.Mutex
}

// New 创建搜索服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func New(log *slog.Logger) (*Service, error) {
	client, err := kgapi.NewClient(kg.WithAutoRefresh(false), kg.WithCookie(auth.LoadSession()))
	if err != nil {
		return nil, err
	}
	return &Service{client: client, log: log}, nil
}

// Search 执行普通搜索。
// searchType 支持 song、album、author、mv、lyric、special，值会原样透传给 SDK。
func (s *Service) Search(keyword string, pageNum int, pageSize int, searchType string) (*Result, error) {
	if _, ok := supportedSearchTypes[searchType]; !ok {
		return nil, fmt.Errorf("不支持的搜索类型: %s", searchType)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	pageNum, pageSize = paging.Normalize(pageNum, pageSize, defaultPageSize)
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

	return mapNormalSearch(searchType, searchData(resp.Body), searchHashPriority()), nil
}

// SearchComplex 执行综合搜索，不向 SDK 传递 type 参数。
func (s *Service) SearchComplex(keyword string, pageNum int, pageSize int) (*Result, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pageNum, pageSize = paging.Normalize(pageNum, pageSize, defaultPageSize)
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

	return mapComplexSearch(searchData(resp.Body), searchHashPriority()), nil
}

// loadSearchCookie 读取最新登录态，并同步到搜索 client。
// search.Service 与 auth.Service 持有不同 SDK client，因此不能只在创建时读取一次。
// dfid 不是必需项：服务端可能不下发（见 auth 的 ensureDfid 兜底），缺失时交给 SDK 用默认值。
func (s *Service) loadSearchCookie() (map[string]string, error) {
	cookie := auth.LoginCookie(auth.LoadSession())
	for _, key := range []string{"token", "userid"} {
		if cookie[key] == "" {
			return nil, errors.New("缺少搜索所需的登录身份信息")
		}
	}

	for key, value := range cookie {
		s.client.SetCookie(key, value)
	}
	return cookie, nil
}

func serializeSearchCookie(cookie map[string]string) string {
	return "token=" + cookie["token"] + ";userid=" + cookie["userid"] + ";dfid=" + cookie["dfid"]
}

// searchHashPriority 返回搜索结果解析 Song.Hash 的候选顺序。
//
// 搜索只需要一个能定位歌曲的 hash：播放文件由 play.GetSongPlayURL 的 quality 参数决定
// （见 internal/play/verify.go 的说明），所以缺了默认档位时可以继续往上找，
// 而不是把整条结果判成“没有可用音质”。
func searchHashPriority() []string {
	return play.PriorityForAny(play.DefaultQuality)
}

func validateResponse(body map[string]any) error {
	if !kgapi.StatusOK(body) {
		return errors.New("搜索接口业务失败")
	}
	return nil
}

func searchData(body map[string]any) map[string]any {
	data := kgapi.RecordOf(body["data"])
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
