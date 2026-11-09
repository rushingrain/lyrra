// Package vip 提供酷狗概念版“每日领取 VIP”能力。
package vip

// 每日 VIP 领取相关 DTO，json tag 与前端 frontend/src/types/index.ts 对齐
// （前端尚未接入，接入时在 types/index.ts 补 DailyVipResult 类型）。

// DailyVipResult 每日 VIP 领取结果。
type DailyVipResult struct {
	// AlreadyClaimed 当天是否已领取：本次领取成功、命中本地记录、
	// 或服务端返回“当天已领取”时都为 true。
	AlreadyClaimed bool `json:"alreadyClaimed"`
	// Claimed 本次是否真的向服务器发起了领取请求。
	// 当天已领过（命中本地记录）时为 false，前端据此可以确认风控生效。
	Claimed bool `json:"claimed"`
	// ClaimedAt 领取成功时间戳(ms)；无记录时为 0。
	ClaimedAt int64 `json:"claimedAt"`
	// Days 领取天数（固定 1 天）。
	Days int `json:"days"`
	// Message 面向用户的提示文案，优先取上游返回的文案。
	Message string `json:"message"`
	// ErrorCode 上游业务错误码，成功时为 0。
	ErrorCode int `json:"errorCode"`
}

// UpgradeVipResult 当天 VIP 升级结果。
type UpgradeVipResult struct {
	// Upgraded 当天是否已升级：本次升级成功、命中本地记录、服务端返回“已升级”
	// 时都为 true。
	Upgraded bool `json:"upgraded"`
	// Requested 本次是否真的向服务器发起了升级请求。
	// 当天已升级过（命中本地记录）时为 false，前端据此可以确认风控生效。
	Requested bool `json:"requested"`
	// UpgradedAt 升级时间戳(ms)；无记录时为 0。
	UpgradedAt int64 `json:"upgradedAt"`
	// RewardType 升级后的会员类型（如 svip），上游未返回时为空。
	RewardType string `json:"rewardType"`
	// RechargeHours 本次升级赠送的小时数（上游 recharge_hours），未返回时为 0。
	RechargeHours int `json:"rechargeHours"`
	// Message 面向用户的提示文案，优先取上游返回的文案。
	Message string `json:"message"`
	// ErrorCode 上游业务错误码，成功时为 0。
	ErrorCode int `json:"errorCode"`
}
