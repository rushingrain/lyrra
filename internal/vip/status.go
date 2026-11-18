package vip

import (
	"context"
	"errors"
	"time"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/kgapi"
)

const statusCacheTTL = 30 * time.Second

var (
	errDailyStatusUnknown = errors.New("暂时无法确认今日 VIP 状态，请稍后重试")
	errDailyStatusChanged = errors.New("查询期间日期已变化，请重新操作")
)

type dailyStatus struct {
	claimed      bool
	upgraded     bool
	upgradeKnown bool
}

// 查询缓存只在同一账号、同一北京时间日期内复用；失败同样冷却，避免连点重试。
type dailyStatusCache struct {
	userID string
	date   string
	at     time.Time
	status dailyStatus
	err    error
}

// syncDailyVipStatus 调用方持有 s.mu。只补充已确认的成功记录，避免查询延迟抹掉本地成功结果。
func (s *Service) syncDailyVipStatus(cookie map[string]string, userID string, now time.Time, st *state) (dailyStatus, error) {
	status, err := s.queryDailyVipStatus(cookie, userID, now)
	if err != nil {
		return dailyStatus{}, err
	}
	changed := false
	if status.claimed && (!st.Claim.covers(now, userID) || st.Claim.Date != claimDate(now)) {
		st.Claim = &record{
			Date:          claimDate(now),
			At:            s.now().UnixMilli(),
			Days:          claimedDays,
			UserID:        userID,
			Message:       "今天已领取过 VIP",
			ServerAlready: true,
		}
		changed = true
	}
	if status.upgraded && (!st.Upgrade.covers(now, userID) || st.Upgrade.Date != claimDate(now)) {
		st.Upgrade = &record{
			Date:          claimDate(now),
			At:            s.now().UnixMilli(),
			UserID:        userID,
			RewardType:    "svip",
			Message:       "今天已升级过 VIP",
			ServerAlready: true,
		}
		changed = true
	}
	if changed {
		if err := saveState(s.recordPath, *st); err != nil {
			s.log.Warn("failed to persist synced vip records", "err", err.Error())
		}
	}
	return status, nil
}

func (s *Service) queryDailyVipStatus(cookie map[string]string, userID string, now time.Time) (dailyStatus, error) {
	cache := s.statusCache
	ttl := statusCacheTTL
	if cache.err != nil {
		ttl = failureCooldown
	}
	elapsed := now.Sub(cache.at)
	if cache.userID == userID && cache.date == claimDate(now) && elapsed >= 0 && elapsed < ttl {
		return cache.status, cache.err
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()
	for key, value := range cookie {
		s.client.SetCookie(key, value)
	}
	resp, requestErr := s.client.YouthMonthVipRecord(ctx, kg.YouthMonthVipRecordRequest{Cookie: cookie})
	var status dailyStatus
	var err error
	switch {
	case requestErr != nil:
		err = errDailyStatusUnknown
		if isLoginError(requestErr) {
			err = ErrNotLoggedIn
		}
		// 不输出 SDK 原始错误或响应，避免带出登录态与账号信息。
		s.log.Warn("daily vip status query failed", "reason", "request failed")
	case resp == nil:
		err = errDailyStatusUnknown
		s.log.Warn("daily vip status query failed", "reason", "empty response")
	case !isBizSuccess(resp.Body) || kgapi.Scalar(resp.Body["error_code"]) != "0":
		err = errDailyStatusUnknown
		s.log.Warn("daily vip status query rejected",
			"status", kgapi.ScalarInt(resp.Body["status"]), "errorCode", kgapi.ScalarInt(resp.Body["error_code"]))
	default:
		var parseErr error
		status, parseErr = parseDailyVipStatus(resp.Body, claimDate(now))
		if parseErr != nil {
			err = errDailyStatusUnknown
			s.log.Warn("daily vip status response unrecognized", "reason", parseErr.Error())
		} else {
			s.log.Info("daily vip status queried", "date", claimDate(now),
				"claimed", status.claimed, "upgraded", status.upgraded, "upgradeKnown", status.upgradeKnown)
		}
	}
	finished := s.now()
	if claimDate(finished) != claimDate(now) {
		s.statusCache = dailyStatusCache{}
		return dailyStatus{}, errDailyStatusChanged
	}
	s.statusCache = dailyStatusCache{
		userID: userID,
		date:   claimDate(now),
		at:     finished,
		status: status,
		err:    err,
	}
	return status, err
}

// 字段依据 SDK 的 auto_daily_checkin 示例及本项目的领取、升级联调记录。
// 空数组表示没有记录；缺失数组、无效条目不能视为未领取。
func parseDailyVipStatus(body kgapi.Record, today string) (dailyStatus, error) {
	data := kgapi.RecordOf(body["data"])
	list, ok := data["list"].([]any)
	if !ok {
		return dailyStatus{}, errors.New("data.list is not an array")
	}
	status := dailyStatus{upgradeKnown: true}
	found := false
	for _, item := range list {
		entry := kgapi.RecordOf(item)
		day := kgapi.Text(entry["day"])
		if _, err := time.Parse("2006-01-02", day); err != nil {
			return dailyStatus{}, errors.New("invalid record day")
		}
		if day != today {
			continue
		}
		if found {
			return dailyStatus{}, errors.New("duplicate records for today")
		}
		found = true
		switch kgapi.Scalar(entry["receive_vip"]) {
		case "0":
			continue
		case "1":
			status.claimed = true
		default:
			return dailyStatus{}, errors.New("unrecognized receive_vip field")
		}
		switch kgapi.Scalar(entry["vip_type"]) {
		case "tvip":
		case "svip":
			status.upgraded = true
		default:
			// 可以确认领取，但无法确认是否已升级时禁止发送升级请求。
			status.upgradeKnown = false
		}
	}
	return status, nil
}
