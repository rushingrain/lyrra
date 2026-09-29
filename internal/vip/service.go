package vip

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	kg "github.com/rushingrain/kugou-music-api"

	"kuwave/internal/auth"
)

// ErrNotLoggedIn 未登录或登录态已失效（语义与 auth.ErrNotLoggedIn 一致）。
var ErrNotLoggedIn = errors.New("未登录")

// ErrClaimRequired 当天还没成功领取 VIP 就调用了升级。
// 上游对“没领就升级”的请求没有定义（可能直接报错，也可能被判定为异常请求触发风控），
// 所以这里在本地直接拦住：一次请求都不发。
var ErrClaimRequired = errors.New("今天还没有领取 VIP，请先领取当天 VIP 再升级")

const (
	// claimedDays 每次成功领取的天数：酷狗“听歌领 VIP”每天发 1 天。
	claimedDays = 1
	// requestTimeout 单次上游请求超时时间。
	requestTimeout = 20 * time.Second
	// failureCooldown 失败后的冷却时间：期间重复调用直接返回错误，不再请求服务器，
	// 避免脚本/连点把失败请求打成轮询。
	failureCooldown = time.Minute
)

// alreadyClaimedErrorCodes 上游表示“当天已领取”的已知业务错误码。
// 130012 取自 SDK 的 auto_daily_checkin 示例（listen_song 当天重复领取）；
// 30002 表示当天奖励已领完（youth_vip 广告任务）。
var alreadyClaimedErrorCodes = []int{130012, 30002}

// alreadyUpgradedErrorCodes 上游表示“当天已升级”的已知业务错误码。
// 30002（当天奖励已领完）同样覆盖升级场景；真实错误码待实测后补齐。
var alreadyUpgradedErrorCodes = []int{30002}

// youthVipClient 只暴露本服务用到的 SDK 方法，便于测试注入桩实现。
type youthVipClient interface {
	YouthDayVip(ctx context.Context, req kg.YouthDayVipRequest) (*kg.YouthDayVipResponse, error)
	YouthDayVipUpgrade(ctx context.Context, req kg.YouthDayVipUpgradeRequest) (*kg.YouthDayVipUpgradeResponse, error)
	// SetCookie 用于把最新登录态同步进 client 内部 cookie 池：
	// SDK 的请求参数（dfid 等）取自池里的值，只传请求级 Cookie 不会生效。
	SetCookie(key, value string)
}

// Service 每日 VIP 服务（领取 + 升级）。
//
// 结构体实例常驻进程，持有独立的 *kg.Client（走 SDK 的 YouthDayVip / YouthDayVipUpgrade）；
// 前端通过 Wails 绑定调用其导出方法（window.go.vip.Service.xxx）。
//
// 风控对“领取”和“升级”各自独立生效：
//  1. 当天已成功过（本地记录命中）直接返回，不请求服务器；
//  2. 同一进程内失败后 1 分钟冷却，冷却期内不再请求服务器；
//  3. 互斥锁串行化，避免并发点击造成重复请求；
//  4. 升级额外要求当天已经成功领取过（见 ErrClaimRequired）。
type Service struct {
	client youthVipClient
	log    *slog.Logger

	// recordPath 本地风控状态文件路径（测试可覆盖）
	recordPath string
	// loadCookie 读取最新登录态：auth.Service 与 vip.Service 持有不同 SDK client，
	// 不能在创建时只读一次，否则登录后领不了、登出后仍带旧凭证。
	loadCookie func() map[string]string
	// now 便于测试注入时间（当天判断、冷却判断都依赖它）
	now func() time.Time

	mu sync.Mutex
	// claimCooldown / upgradeCooldown 两个动作各自独立的失败冷却
	claimCooldown   cooldown
	upgradeCooldown cooldown
}

// cooldown 记录某个动作最近一次失败的时间与原因，用于失败后的冷却。
type cooldown struct {
	at  time.Time
	err string
}

// remaining 返回距离冷却结束还剩多久，未处于冷却期时返回 0。
func (c *cooldown) remaining(now time.Time) time.Duration {
	if c.at.IsZero() {
		return 0
	}
	if left := failureCooldown - now.Sub(c.at); left > 0 {
		return left
	}
	return 0
}

// fail 记录失败时间与原因。
func (c *cooldown) fail(now time.Time, err error) {
	c.at = now
	c.err = err.Error()
}

// New 创建每日 VIP 领取服务。
// 只做本地初始化，不发起网络请求，因此不会阻塞应用启动。
func New(log *slog.Logger) (*Service, error) {
	client, err := kg.New(kg.WithAutoRefresh(false))
	if err != nil {
		return nil, err
	}
	return newService(client, log, recordPath(), auth.LoadSession), nil
}

// newService 用指定依赖构造服务，便于测试注入桩 client / 临时记录文件 / 固定时间。
func newService(client youthVipClient, log *slog.Logger, path string, loadCookie func() map[string]string) *Service {
	return &Service{
		client:     client,
		log:        log,
		recordPath: path,
		loadCookie: loadCookie,
		now:        time.Now,
	}
}

// ClaimDailyVip 领取当天 VIP（酷狗“听歌领 VIP”，每天一次）。
//
// 对外不接收任何参数：领取日由后端固定取北京时间当天（见 claimDate），
// 调用方（前端）无法指定日期或其它请求参数。
//
// 返回的 DailyVipResult.Claimed 表示本次是否真的请求了服务器：
// 当天已经领取成功过时直接返回本地记录，Claimed=false，不产生任何网络请求。
// 未登录返回 ErrNotLoggedIn；上游业务失败返回带 error_code 的错误。
//
// 实测记录（2026-09-28，lite 平台，脚本见 tmp/tmp_vip_claim）：
//
//	POST /youth/v1/recharge/receive_vip_listen_song
//	参数：source_id=90139、receive_day=2026-09-28（日期字符串）、dfid/userid/token 必带
//	成功：{"data":{"ad_vip_end_time":0,"ad_vip_num":0,"server_time":...},"error_code":0,"status":1}
//	失败：receive_day 传 int 1 → error_code=304001「日期格式错误」
//	领取后 youth_month_vip_record 新增当天记录 receive_vip=1 / vip_type=tvip
func (s *Service) ClaimDailyVip() (*DailyVipResult, error) {
	cookie, userID, err := s.loginCookie()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	st := loadState(s.recordPath)

	// 风控一：当天已领取过，直接返回本地记录，不请求服务器。
	if rec := st.Claim; rec.covers(now, userID) {
		s.log.Info("daily vip already claimed, skip request",
			"date", claimDate(now), "claimedAt", rec.At, "fromServer", rec.ServerAlready)
		return recordResult(rec, false), nil
	}

	// 风控二：失败冷却，避免失败请求被高频重放。
	if remaining := s.claimCooldown.remaining(now); remaining > 0 {
		s.log.Warn("daily vip claim rejected by cooldown",
			"remainingSeconds", int(remaining.Seconds())+1, "lastErr", s.claimCooldown.err)
		return nil, fmt.Errorf("操作过于频繁，请 %d 秒后重试", int(remaining.Seconds())+1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// 同步最新登录态到自己的 client：YouthDayVip 内部做完登录校验后走的是
	// client 池里的 cookie，dfid 缺失时 SDK 会退化成占位符 "-"。
	for key, value := range cookie {
		s.client.SetCookie(key, value)
	}

	resp, err := s.client.YouthDayVip(ctx, kg.YouthDayVipRequest{
		// receive_day 是“领取哪一天的 VIP”，必须是日期字符串。
		// 这里写死为后端算出的当天，不接受任何外部传入；实测传 int 1 会返回
		// error_code=304001「日期格式错误」。
		ReceiveDay: claimDate(now),
		Cookie:     cookie,
	})
	if err != nil {
		s.claimCooldown.fail(now, err)
		if isLoginError(err) {
			s.log.Warn("daily vip claim failed: not logged in", "err", err.Error())
			return nil, ErrNotLoggedIn
		}
		s.log.Warn("daily vip claim failed", "err", err.Error())
		return nil, err
	}

	code := intValue(resp.Body["error_code"])
	message := pickMessage(resp.Body)

	switch {
	case isBizSuccess(resp.Body):
		rec := &record{
			Date:    claimDate(now),
			At:      now.UnixMilli(),
			Days:    claimedDays,
			UserID:  userID,
			Message: firstNonEmpty(message, "领取成功"),
		}
		st.Claim = rec
		if err := saveState(s.recordPath, st); err != nil {
			// 落盘失败不影响本次领取结果，但要留下日志：下次调用会再打一次上游。
			s.log.Warn("failed to persist vip claim record", "err", err.Error())
		}
		s.log.Info("daily vip claimed", "date", rec.Date, "days", rec.Days)
		return &DailyVipResult{
			AlreadyClaimed: true,
			Claimed:        true,
			ClaimedAt:      rec.At,
			Days:           rec.Days,
			Message:        rec.Message,
		}, nil

	case isAlreadyClaimed(code, message):
		// 服务端认为当天已领（可能是在别的设备领的）：本地也记一笔，
		// 后续调用不再请求服务器。记录带有效期，跨过活动换日时刻即失效。
		rec := &record{
			Date:          claimDate(now),
			At:            now.UnixMilli(),
			Days:          claimedDays,
			UserID:        userID,
			Message:       firstNonEmpty(message, "今天已领取过 VIP"),
			ServerAlready: true,
			ExpiresAt:     nextActivityReset(now).UnixMilli(),
		}
		st.Claim = rec
		if err := saveState(s.recordPath, st); err != nil {
			s.log.Warn("failed to persist vip claim record", "err", err.Error())
		}
		s.log.Info("daily vip already claimed on server", "errorCode", code, "message", rec.Message)
		return &DailyVipResult{
			AlreadyClaimed: true,
			Claimed:        true,
			ClaimedAt:      rec.At,
			Days:           rec.Days,
			Message:        rec.Message,
			ErrorCode:      code,
		}, nil
	}

	s.claimCooldown.fail(now, fmt.Errorf("error_code=%d %s", code, message))
	s.log.Warn("daily vip claim rejected", "errorCode", code, "message", message)
	if message == "" {
		message = "上游未返回文案"
	}
	return nil, fmt.Errorf("领取 VIP 失败（error_code=%d）：%s", code, message)
}

// UpgradeDailyVip 升级当天领到的 VIP（酷狗“升级 VIP 奖励”，每天一次）。
//
// 前置条件：当天必须已经成功领取过 VIP。上游对“没领就升级”的请求行为未定义
// （可能直接报错，也可能被判定为异常请求触发风控），所以这里用本地记录拦住，
// 未领取时返回 ErrClaimRequired，一次请求都不发。
//
// 对外不接收任何参数：升级对象就是“当天”的奖励，由后端按北京时间当天判定。
// 返回的 UpgradeVipResult.Requested 表示本次是否真的请求了服务器：
// 当天已经升级过时直接返回本地记录，Requested=false，不产生任何网络请求。
// 未登录返回 ErrNotLoggedIn；上游业务失败返回带 error_code 的错误。
//
// 实测记录（2026-09-29，lite 平台，脚本见 tmp/tmp_vip_claim）：
//
//	POST /youth/v1/listen_song/upgrade_vip_reward?ad_type=1
//	成功：{"data":{"recharge_hours":24},"error_code":0,"status":1}
//	升级后 youth_month_vip_record 里当天记录由 vip_type=tvip 变为 svip
//	本次未先做广告上报也能成功；但正常使用场景应理解为“看完广告再升级”，
//	所以调用频率依旧严格限制在每天一次。
func (s *Service) UpgradeDailyVip() (*UpgradeVipResult, error) {
	cookie, userID, err := s.loginCookie()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	st := loadState(s.recordPath)

	// 前置条件：当天已经成功领取过 VIP（本地记录命中才放行）。
	if !st.Claim.covers(now, userID) {
		s.log.Warn("vip upgrade rejected: daily vip not claimed yet", "date", claimDate(now))
		return nil, ErrClaimRequired
	}

	// 风控一：当天已升级过，直接返回本地记录，不请求服务器。
	if rec := st.Upgrade; rec.covers(now, userID) {
		s.log.Info("daily vip upgrade already done, skip request",
			"date", claimDate(now), "upgradedAt", rec.At, "fromServer", rec.ServerAlready)
		return upgradeResult(rec, false), nil
	}

	// 风控二：失败冷却，避免失败请求被高频重放。
	if remaining := s.upgradeCooldown.remaining(now); remaining > 0 {
		s.log.Warn("vip upgrade rejected by cooldown",
			"remainingSeconds", int(remaining.Seconds())+1, "lastErr", s.upgradeCooldown.err)
		return nil, fmt.Errorf("操作过于频繁，请 %d 秒后重试", int(remaining.Seconds())+1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// 同步最新登录态到自己的 client（同领取：SDK 的请求参数取自 client 池）。
	for key, value := range cookie {
		s.client.SetCookie(key, value)
	}

	// 升级请求的参数（ad_type=1 等）全部由 SDK 兼容层内部补齐，
	// 这里只带登录态，不接受任何外部传入。
	resp, err := s.client.YouthDayVipUpgrade(ctx, kg.YouthDayVipUpgradeRequest{Cookie: cookie})
	if err != nil {
		s.upgradeCooldown.fail(now, err)
		if isLoginError(err) {
			s.log.Warn("vip upgrade failed: not logged in", "err", err.Error())
			return nil, ErrNotLoggedIn
		}
		s.log.Warn("vip upgrade failed", "err", err.Error())
		return nil, err
	}

	code := intValue(resp.Body["error_code"])
	message := pickMessage(resp.Body)

	switch {
	case isBizSuccess(resp.Body):
		rec := &record{
			Date:       claimDate(now),
			At:         now.UnixMilli(),
			UserID:     userID,
			Message:    firstNonEmpty(message, "升级成功"),
			RewardType: pickRewardType(resp.Body),
			Hours:      pickRechargeHours(resp.Body),
		}
		st.Upgrade = rec
		if err := saveState(s.recordPath, st); err != nil {
			s.log.Warn("failed to persist vip upgrade record", "err", err.Error())
		}
		s.log.Info("daily vip upgraded", "date", rec.Date, "rewardType", rec.RewardType)
		return upgradeResult(rec, true), nil

	case isAlreadyUpgraded(code, message):
		// 服务端认为当天已升级（可能是在别的设备升的）：本地也记一笔，
		// 后续调用不再请求服务器。
		rec := &record{
			Date:          claimDate(now),
			At:            now.UnixMilli(),
			UserID:        userID,
			Message:       firstNonEmpty(message, "今天已升级过 VIP"),
			ServerAlready: true,
			ExpiresAt:     nextActivityReset(now).UnixMilli(),
		}
		st.Upgrade = rec
		if err := saveState(s.recordPath, st); err != nil {
			s.log.Warn("failed to persist vip upgrade record", "err", err.Error())
		}
		s.log.Info("daily vip upgrade already done on server", "errorCode", code, "message", rec.Message)
		result := upgradeResult(rec, true)
		result.ErrorCode = code
		return result, nil
	}

	s.upgradeCooldown.fail(now, fmt.Errorf("error_code=%d %s", code, message))
	s.log.Warn("daily vip upgrade rejected", "errorCode", code, "message", message)
	if message == "" {
		message = "上游未返回文案"
	}
	return nil, fmt.Errorf("升级 VIP 失败（error_code=%d）：%s", code, message)
}

// GetDailyVipStatus 查询当天领取状态（只读本地记录，不请求服务器）。
// 前端可据此决定按钮是否可点；未登录返回 ErrNotLoggedIn。
func (s *Service) GetDailyVipStatus() (*DailyVipResult, error) {
	_, userID, err := s.loginCookie()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if rec := loadState(s.recordPath).Claim; rec.covers(s.now(), userID) {
		return recordResult(rec, false), nil
	}
	return &DailyVipResult{Days: claimedDays, Message: "今天还未领取 VIP"}, nil
}

// loginCookie 读取最新登录态并校验登录所需字段。
// 缺少 token/userid 时直接返回 ErrNotLoggedIn，不发起任何请求。
// 返回值 cookie 只带登录必需字段；userID 直接取自会话
// （auth.LoadSession 已把历史脏数据里的科学计数法 userid 规整成十进制）。
func (s *Service) loginCookie() (map[string]string, string, error) {
	session := s.loadCookie()
	cookie := make(map[string]string, 3)
	for _, key := range []string{"token", "userid", "dfid"} {
		if value := strings.TrimSpace(session[key]); value != "" {
			cookie[key] = value
		}
	}
	if cookie["token"] == "" || cookie["userid"] == "" || cookie["userid"] == "0" {
		return nil, "", ErrNotLoggedIn
	}
	return cookie, cookie["userid"], nil
}

// recordResult 把本地记录转换成对外的领取结果。claimed 表示本次是否真的请求过服务器。
func recordResult(rec *record, claimed bool) *DailyVipResult {
	days := rec.Days
	if days <= 0 {
		days = claimedDays
	}
	return &DailyVipResult{
		AlreadyClaimed: true,
		Claimed:        claimed,
		ClaimedAt:      rec.At,
		Days:           days,
		Message:        firstNonEmpty(rec.Message, "今天已领取过 VIP"),
	}
}

// upgradeResult 把本地记录转换成对外的升级结果。requested 表示本次是否真的请求过服务器。
func upgradeResult(rec *record, requested bool) *UpgradeVipResult {
	return &UpgradeVipResult{
		Upgraded:      true,
		Requested:     requested,
		UpgradedAt:    rec.At,
		RewardType:    rec.RewardType,
		RechargeHours: rec.Hours,
		Message:       firstNonEmpty(rec.Message, "今天已升级过 VIP"),
	}
}

// isBizSuccess 业务成功判断，与其它服务一致：body 里 status=="1" 且 error_code==0。
func isBizSuccess(body map[string]any) bool {
	if stringValue(body["status"]) != "1" {
		return false
	}
	switch stringValue(body["error_code"]) {
	case "", "0":
		return true
	default:
		return false
	}
}

// isAlreadyClaimed 判断上游是否在提示“当天已经领过”：
// 命中已知错误码，或文案里出现“已领取/已经领取/已领过/重复领取/已签到”。
// 关键词刻意收紧（不匹配泛化的“已领”），避免把“奖励已领完”之类的其它失败
// 误判成当天已领取；上游真实错误码待实测后在这里补齐。
func isAlreadyClaimed(code int, message string) bool {
	return isAlreadyDone(code, message, alreadyClaimedErrorCodes,
		"已领取", "已经领取", "已领过", "重复领取", "已签到")
}

// isAlreadyUpgraded 判断上游是否在提示“当天已经升级过”。
// 同样刻意收紧关键词，真实错误码待实测后补齐。
func isAlreadyUpgraded(code int, message string) bool {
	return isAlreadyDone(code, message, alreadyUpgradedErrorCodes,
		"已升级", "已经升级", "已升过", "重复升级")
}

// isAlreadyDone 命中已知错误码或文案关键词时返回 true。
func isAlreadyDone(code int, message string, codes []int, keywords ...string) bool {
	for _, known := range codes {
		if code == known {
			return true
		}
	}
	for _, keyword := range keywords {
		if strings.Contains(message, keyword) {
			return true
		}
	}
	return false
}

// pickRewardType 从升级响应里取升级后的会员类型（如 svip），取不到时返回空串。
func pickRewardType(body map[string]any) string {
	if body == nil {
		return ""
	}
	data, _ := body["data"].(map[string]any)
	nested, _ := data["data"].(map[string]any)
	for _, src := range []map[string]any{data, nested, body} {
		for _, key := range []string{"vip_type", "reward_type", "upgrade_type"} {
			if text := stringValue(src[key]); text != "" {
				return text
			}
		}
	}
	return ""
}

// pickRechargeHours 从升级响应里取赠送时长（recharge_hours），取不到时返回 0。
func pickRechargeHours(body map[string]any) int {
	if body == nil {
		return 0
	}
	data, _ := body["data"].(map[string]any)
	nested, _ := data["data"].(map[string]any)
	for _, src := range []map[string]any{data, nested, body} {
		if hours := intValue(src["recharge_hours"]); hours > 0 {
			return hours
		}
	}
	return 0
}

// isLoginError 识别 SDK 返回的登录态错误（login required / login expired 等英文文案）。
func isLoginError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrNotLoggedIn) {
		return true
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "login")
}

// pickMessage 依次从 body / body.data / body.data.data 中取上游文案。
func pickMessage(body map[string]any) string {
	if body == nil {
		return ""
	}
	data, _ := body["data"].(map[string]any)
	nested, _ := data["data"].(map[string]any)
	for _, src := range []map[string]any{body, data, nested} {
		for _, key := range []string{"error_msg", "err_msg", "msg", "message", "toast", "alert_msg"} {
			if value, ok := src[key]; ok {
				if text := stringValue(value); text != "" {
					return text
				}
			}
		}
	}
	// 少数接口直接返回字符串型 data，同样按文案处理
	return stringValue(body["data"])
}

// stringValue 把上游字段规整成字符串；map / 数组等复合类型返回空串。
func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		text := strings.TrimSpace(typed)
		if text == "<nil>" {
			return ""
		}
		// JSON 里的数字可能被解析成 float64 后再被格式化，这里统一去掉小数尾巴
		return text
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprintf("%d", int64(typed))
		}
		return fmt.Sprintf("%v", typed)
	case map[string]any, []any:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

// intValue 把上游字段规整成整数，无法解析时返回 0。
func intValue(value any) int {
	text := stringValue(value)
	if text == "" {
		return 0
	}
	var out int
	if _, err := fmt.Sscanf(text, "%d", &out); err != nil {
		return 0
	}
	return out
}

// firstNonEmpty 返回第一个非空字符串。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// claimDate 返回本次领取使用的“领取日”（北京时间当天）。
//
// 领取日完全由后端内部决定，对外的 ClaimDailyVip / GetDailyVipStatus 都不带参数，
// 调用方无法指定日期。请求参数 receive_day 与本地记录日期都取自这里，保证两者一致。
func claimDate(now time.Time) string {
	return chinaToday(now)
}
