package dto

// UserPlaylist 是当前账号创建或收藏的歌单。
type UserPlaylist struct {
	// ID 是当前账号的 listid；GlobalCollectionID 用于读取歌曲，收藏项指向原歌单。
	ID                 string   `json:"id"`
	GlobalCollectionID string   `json:"globalCollectionId"`
	Name               string   `json:"name"`
	SongCount          int      `json:"songCount"`
	CoverURL           string   `json:"coverUrl"`
	Description        string   `json:"description"`
	Tags               []string `json:"tags"`
	// IsDefault：1=默认收藏，2=我喜欢，0=普通歌单。
	IsDefault int `json:"isDefault"`
	// Type：0=创建，1=收藏；未知类型不可编辑。
	Type int `json:"type"`
}
