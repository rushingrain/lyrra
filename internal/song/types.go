package song

// HomeSong 首页最新音乐（TopSong 新歌首发榜）单曲 DTO。
// json tag 与前端 frontend/src/types/index.ts 的 Song 类型对齐。
type HomeSong struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	CoverURL string `json:"coverUrl"`
	Duration int64  `json:"duration"`
	Hash     string `json:"hash"`
	// 最高音质标签："无损" / "Hi-Res"，无高音质时为空
	Quality string `json:"quality"`
	// 是否为原唱（上游 trans_param.is_original == 1）
	IsOriginal bool `json:"isOriginal"`
}
