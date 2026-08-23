package search

// SearchResponse 是普通搜索接口的暂时响应模型。
// Data 保留 SDK 返回的动态结构，待拿到真实响应示例后再细化为具体 DTO。
type SearchResponse struct {
	Status    string         `json:"status"`
	ErrorCode any            `json:"errorCode"`
	Data      map[string]any `json:"data"`
}

// SearchComplexResponse 是综合搜索接口的暂时响应模型。
// Data 保留 SDK 返回的动态结构，待拿到真实响应示例后再细化为具体 DTO。
type SearchComplexResponse struct {
	Status    string         `json:"status"`
	ErrorCode any            `json:"errorCode"`
	Data      map[string]any `json:"data"`
}
