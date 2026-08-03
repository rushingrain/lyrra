package main

// 登录相关 DTO，json tag 与前端 frontend/src/types/index.ts 对齐。

// QrLoginResponse 扫码登录 - 获取二维码的响应
type QrLoginResponse struct {
	// 二维码唯一标识，用于轮询登录状态
	QrKey string `json:"qrKey"`
	// 二维码图片地址（后端直接返回 data:image/png;base64,...，前端 <img> 可直接渲染）
	QrUrl string `json:"qrUrl"`
	// 二维码过期时间戳(ms)
	ExpiresAt int64 `json:"expiresAt"`
}

// QrLoginStatus 扫码登录状态：未扫码 / 已扫码待确认 / 已确认 / 已过期
type QrLoginStatus string

const (
	QrLoginPending   QrLoginStatus = "pending"
	QrLoginScanned   QrLoginStatus = "scanned"
	QrLoginConfirmed QrLoginStatus = "confirmed"
	QrLoginExpired   QrLoginStatus = "expired"
)

// UserInfo 当前登录用户信息
type UserInfo struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	IsVip    bool   `json:"isVip"`
	// 好友 / 关注 / 粉丝数（来自 user_detail 接口 data.friends/follows/fans）
	Friends int `json:"friends"`
	Follows int `json:"follows"`
	Fans    int `json:"fans"`
}
