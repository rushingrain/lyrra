package song

// Count 是一首歌的计数（收藏数 / 评论数）。
//
// Count 是原始数量；Text 是按展示规则格式化后的文案（<1000 原样、k、w、亿，向下取整），
// 前端直接展示 Text 即可，需要精确值时用 Count。
type Count struct {
	Count int64  `json:"count"`
	Text  string `json:"text"`
}

// SongStats 是歌曲统计数据：收藏数与评论数。
// 单项为 nil 表示该项取不到（上游失败或没有对应入参），前端显示占位符，不提示错误。
type SongStats struct {
	Favorite *Count `json:"favorite,omitempty"`
	Comment  *Count `json:"comment,omitempty"`
}
