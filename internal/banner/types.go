package banner

// 首页 banner DTO，json tag 与 frontend/src/types/index.ts 的 Banner 对齐。

// HomeBanner 首页轮播 banner。
type HomeBanner struct {
	ID       string `json:"id"`
	ImageURL string `json:"imageUrl"`
	Title    string `json:"title"`
	LinkURL  string `json:"linkUrl"`
}
