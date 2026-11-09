package vip

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"lyrra/internal/appdir"
	"lyrra/internal/atomicfile"
)

// 本地状态记录：风控的落地点。
//
// 存储位置与登录会话同基址（应用数据目录下的 vip.json）：
// 保存“当天领取 / 当天升级”两条最近记录，用于当天重复调用时直接短路、
// 不再请求服务器。与 session.json 一样按用户维度隔离（目录本身属于当前用户）。
const recordFileName = "vip.json"

// activityDayResetHour 酷狗“听歌领 VIP”活动侧的换日时间（北京时间 01:00）。
//
// 依据：SDK 自带的 examples/auto_daily_checkin 用 youth_month_vip_record 判断
// “今天是否已签到”，并在北京时间 01:00 前视为“尚未换日”。
// 记录以日历日（北京时间 00:00 换日）为主，只有“服务端返回当天已领取”这一种
// 推断记录会额外带上这个换日时间作为有效期，避免在 00:00~01:00 之间误挡住
// 新一天真正可领的一次机会。
const activityDayResetHour = 1

var chinaZone = time.FixedZone("CST", 8*3600)

// record 一次“当天 VIP 相关动作”（领取 / 升级）的本地记录。
type record struct {
	// Date 动作日期（北京时间 YYYY-MM-DD），风控按它判断“当天”
	Date string `json:"date"`
	// At 动作时间戳(ms)
	At int64 `json:"at"`
	// Days 本次领取天数（仅领取动作，升级动作不填）
	Days int `json:"days,omitempty"`
	// RewardType 升级后的会员类型（仅升级动作，如 svip）
	RewardType string `json:"rewardType,omitempty"`
	// Hours 升级赠送的时长/小时数（仅升级动作，来自上游 recharge_hours）
	Hours int `json:"hours,omitempty"`
	// UserID 动作发生时的登录用户，避免同一台机器换账号后误判“今天已领过”
	UserID string `json:"userId,omitempty"`
	// Message 结果文案
	Message string `json:"message,omitempty"`
	// ServerAlready 该记录来自服务端返回“当天已领取”，并非本次请求成功
	ServerAlready bool `json:"serverAlready,omitempty"`
	// ExpiresAt 记录有效期(ms)，0 表示按 Date 判断（整个自然日有效）
	ExpiresAt int64 `json:"expiresAt,omitempty"`
}

// state 本地风控状态：领取与升级各保留一条最近记录。
type state struct {
	Claim   *record `json:"claim,omitempty"`
	Upgrade *record `json:"upgrade,omitempty"`
}

// recordPath 返回本地状态文件路径。
func recordPath() string {
	return filepath.Join(appdir.Dir(), recordFileName)
}

// loadState 读取本地风控状态。文件缺失或内容损坏时返回空状态（按“今天什么都没做”处理）。
func loadState(path string) state {
	data, err := os.ReadFile(path)
	if err != nil {
		return state{}
	}
	var st state
	if err := json.Unmarshal(data, &st); err == nil && (st.Claim != nil || st.Upgrade != nil) {
		return st
	}
	// 兼容旧格式：顶层就是一条领取记录（时间戳字段名是 claimedAt）
	var legacy struct {
		Date          string `json:"date"`
		ClaimedAt     int64  `json:"claimedAt"`
		Days          int    `json:"days"`
		UserID        string `json:"userId"`
		Message       string `json:"message"`
		ServerAlready bool   `json:"serverAlready"`
		ExpiresAt     int64  `json:"expiresAt"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil || legacy.Date == "" {
		return state{}
	}
	return state{Claim: &record{
		Date:          legacy.Date,
		At:            legacy.ClaimedAt,
		Days:          legacy.Days,
		UserID:        legacy.UserID,
		Message:       legacy.Message,
		ServerAlready: legacy.ServerAlready,
		ExpiresAt:     legacy.ExpiresAt,
	}}
}

// saveState 写入本地风控状态，权限与 session.json 一致（仅当前用户可读写）。
//
// 采用“同目录临时文件 + rename”的原子写：进程在写入过程中退出（崩溃/强杀/断电）
// 只会留下临时文件，不会把主文件截断成半截 JSON。读取侧虽然对损坏文件有兜底
// （退化成空状态），但那会导致多打一次上游请求，能避免就避免。
func saveState(path string, st state) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return atomicfile.Write(path, data)
}

// covers 判断记录是否仍然代表“当前这次调用无需请求服务器”。
// userID 为当前登录用户，用于排除换账号后复用旧记录的情况。
func (r *record) covers(now time.Time, userID string) bool {
	if r == nil {
		return false
	}
	// 记录里的用户与当前登录用户不一致时视为无效（空 userID 为历史记录，宽容处理）
	if r.UserID != "" && userID != "" && r.UserID != userID {
		return false
	}
	if r.ExpiresAt > 0 {
		return now.UnixMilli() < r.ExpiresAt
	}
	return r.Date == chinaToday(now)
}

// chinaToday 返回北京时间（CST, UTC+8）当天日期。
// 与 SDK 的 auto_daily_checkin 示例保持一致：酷狗活动按北京时间切天。
func chinaToday(now time.Time) string {
	return now.In(chinaZone).Format("2006-01-02")
}

// nextActivityReset 返回 now 之后最近一次活动换日时刻（北京时间 01:00）。
func nextActivityReset(now time.Time) time.Time {
	local := now.In(chinaZone)
	reset := time.Date(local.Year(), local.Month(), local.Day(), activityDayResetHour, 0, 0, 0, chinaZone)
	if !reset.After(local) {
		reset = reset.AddDate(0, 0, 1)
	}
	return reset
}
