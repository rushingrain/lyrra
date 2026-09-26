package search

// Song 是搜索结果中的单曲 DTO。
// JSON 字段与前端 frontend/src/types/index.ts 的 Song 类型对齐。
type Song struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Artist     string   `json:"artist"`
	Album      string   `json:"album,omitempty"`
	CoverURL   string   `json:"coverUrl"`
	Duration   int64    `json:"duration,omitempty"`
	Hash       string   `json:"hash,omitempty"`
	Quality    string   `json:"quality,omitempty"`
	IsOriginal bool     `json:"isOriginal"`
	Tags       []string `json:"tags,omitempty"`
}

// Playlist 是搜索结果中的歌单 DTO。
type Playlist struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	CoverURL    string `json:"coverUrl"`
	Description string `json:"description"`
	Creator     string `json:"creator"`
	PlayCount   int64  `json:"playCount"`
	SongCount   int64  `json:"songCount,omitempty"`
}

// Album 是搜索结果中的专辑 DTO。
type Album struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	CoverURL string `json:"coverUrl"`
	Date     string `json:"date"`
}

// Artist 是搜索结果中的歌手 DTO。
type Artist struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
	SongCount int64  `json:"songCount"`
	Genre     string `json:"genre"`
	Fans      int64  `json:"fans"`
}

// MV 是搜索结果中的 MV DTO。
type MV struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Artist   string `json:"artist"`
	CoverURL string `json:"coverUrl"`
}

// Lyric 是搜索结果中的歌词条目 DTO。
type Lyric struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

// SearchResult 是搜索服务的稳定响应 DTO。
// 普通搜索只填充其中一个分类；综合搜索按上游分组填充多个分类。
type SearchResult struct {
	Songs     []Song     `json:"songs"`
	Playlists []Playlist `json:"playlists"`
	Albums    []Album    `json:"albums"`
	Artists   []Artist   `json:"artists"`
	MVs       []MV       `json:"mvs"`
	Lyrics    []Lyric    `json:"lyrics"`
}
