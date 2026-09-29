package vip

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	kg "github.com/rushingrain/kugou-music-api"
)

// stubClient 记录调用次数并可返回预设响应，避免单测打真实网络。
type stubClient struct {
	calls int

	// lastReq 最近一次请求，用于断言领取参数与登录态同步
	lastReq kg.YouthDayVipRequest
	// lastUpgradeReq 最近一次升级请求
	lastUpgradeReq kg.YouthDayVipUpgradeRequest
	// cookies 收到的 SetCookie 调用
	cookies map[string]string

	// upgradeCalls 升级接口调用次数
	upgradeCalls int

	resp        *kg.YouthDayVipResponse
	err         error
	upgradeResp *kg.YouthDayVipUpgradeResponse
	upgradeErr  error
}

func (s *stubClient) YouthDayVip(_ context.Context, req kg.YouthDayVipRequest) (*kg.YouthDayVipResponse, error) {
	s.calls++
	s.lastReq = req
	return s.resp, s.err
}

func (s *stubClient) SetCookie(key, value string) {
	if s.cookies == nil {
		s.cookies = map[string]string{}
	}
	s.cookies[key] = value
}

func (s *stubClient) YouthDayVipUpgrade(_ context.Context, req kg.YouthDayVipUpgradeRequest) (*kg.YouthDayVipUpgradeResponse, error) {
	s.upgradeCalls++
	s.lastUpgradeReq = req
	return s.upgradeResp, s.upgradeErr
}

// okUpgradeResponse 构造升级业务成功的响应体。
func okUpgradeResponse(data map[string]any) *kg.YouthDayVipUpgradeResponse {
	body := map[string]any{"status": float64(1), "error_code": float64(0)}
	if data != nil {
		body["data"] = data
	}
	return &kg.YouthDayVipUpgradeResponse{Body: body}
}

// failUpgradeResponse 构造升级业务失败的响应体。
func failUpgradeResponse(code int, message string) *kg.YouthDayVipUpgradeResponse {
	body := map[string]any{"status": float64(1), "error_code": float64(code)}
	if message != "" {
		body["error_msg"] = message
	}
	return &kg.YouthDayVipUpgradeResponse{Body: body}
}

// okResponse 构造业务成功的响应体。
func okResponse(message string) *kg.YouthDayVipResponse {
	body := map[string]any{"status": float64(1), "error_code": float64(0)}
	if message != "" {
		body["data"] = map[string]any{"msg": message}
	}
	return &kg.YouthDayVipResponse{Body: body}
}

// failResponse 构造带业务错误码的响应体。
func failResponse(code int, message string) *kg.YouthDayVipResponse {
	body := map[string]any{"status": float64(1), "error_code": float64(code)}
	if message != "" {
		body["error_msg"] = message
	}
	return &kg.YouthDayVipResponse{Body: body}
}

const testUserID = "863006722"

func testSession() map[string]string {
	return map[string]string{"token": "test-token", "userid": testUserID, "dfid": "test-dfid"}
}

// newTestService 构造指向临时记录文件的被测服务，返回服务与固定“当前时间”的引用。
func newTestService(t *testing.T, client youthVipClient, at *time.Time, session map[string]string) *Service {
	t.Helper()
	service := newService(
		client,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		filepath.Join(t.TempDir(), recordFileName),
		func() map[string]string { return session },
	)
	service.now = func() time.Time { return *at }
	return service
}

// 北京时间固定基准：2026-09-28 10:00。
func baseTime() time.Time {
	return time.Date(2026, 9, 28, 10, 0, 0, 0, chinaZone)
}

func TestClaimDailyVipClaimsOnceThenSkips(t *testing.T) {
	client := &stubClient{resp: okResponse("领取成功")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	first, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("首次领取失败: %v", err)
	}
	if !first.AlreadyClaimed || !first.Claimed {
		t.Fatalf("首次领取结果异常: %+v", first)
	}
	if first.Days != claimedDays || first.ClaimedAt == 0 {
		t.Fatalf("首次领取天数/时间戳异常: %+v", first)
	}

	// 同一天内再次调用：不再请求服务器
	now = now.Add(2 * time.Hour)
	second, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("重复领取不应报错: %v", err)
	}
	if second.Claimed {
		t.Fatalf("当天已领取过时不应再请求服务器: %+v", second)
	}
	if !second.AlreadyClaimed || second.ClaimedAt != first.ClaimedAt {
		t.Fatalf("重复领取结果应与首次一致: %+v", second)
	}
	if client.calls != 1 {
		t.Fatalf("上游应只被调用 1 次，实际 %d 次", client.calls)
	}
}

func TestClaimDailyVipSkipsWithoutNetworkWhenRecordExists(t *testing.T) {
	client := &stubClient{resp: okResponse("")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	// 预置“今天已领取”的本地记录
	if err := saveState(service.recordPath, state{Claim: &record{
		Date:    chinaToday(now),
		At:      now.Add(-time.Hour).UnixMilli(),
		Days:    claimedDays,
		UserID:  testUserID,
		Message: "领取成功",
	}}); err != nil {
		t.Fatalf("预置记录失败: %v", err)
	}

	result, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	if result.Claimed || !result.AlreadyClaimed {
		t.Fatalf("命中本地记录时应直接返回: %+v", result)
	}
	if client.calls != 0 {
		t.Fatalf("命中本地记录时不应请求服务器，实际 %d 次", client.calls)
	}
}

func TestClaimDailyVipClaimsAgainNextDay(t *testing.T) {
	client := &stubClient{resp: okResponse("")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("首日领取失败: %v", err)
	}

	// 跨到第二天（北京时间）后应重新请求
	now = now.AddDate(0, 0, 1)
	result, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("次日领取失败: %v", err)
	}
	if !result.Claimed {
		t.Fatalf("次日应重新请求服务器: %+v", result)
	}
	if client.calls != 2 {
		t.Fatalf("上游应被调用 2 次，实际 %d 次", client.calls)
	}
	// 领取日始终跟着后端时钟走，没有外部入参可以改它
	if got, ok := client.lastReq.ReceiveDay.(string); !ok || got != claimDate(now) || got == "2026-09-28" {
		t.Fatalf("次日 receive_day = %#v, want %q", client.lastReq.ReceiveDay, claimDate(now))
	}
}

func TestClaimDailyVipRecordsServerAlreadyClaimed(t *testing.T) {
	client := &stubClient{resp: failResponse(130012, "您今天已经领取过了")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	first, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("服务端提示已领取不应报错: %v", err)
	}
	if !first.AlreadyClaimed || !first.Claimed || first.ErrorCode != 130012 {
		t.Fatalf("服务端已领取结果异常: %+v", first)
	}

	// 服务端已领取同样计入本地记录，后续调用不再请求
	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("重复调用失败: %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("上游应只被调用 1 次，实际 %d 次", client.calls)
	}
}

func TestClaimDailyVipServerAlreadyClaimedExpiresAtActivityReset(t *testing.T) {
	client := &stubClient{resp: failResponse(130012, "您今天已经领取过了")}
	// 北京时间 00:30：此时若服务端仍认为归属前一天，本地记录只应有效到 01:00
	now := time.Date(2026, 9, 29, 0, 30, 0, 0, chinaZone)
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}

	// 01:30 之后（已跨过活动换日）应重新请求服务器
	client.resp = okResponse("")
	now = time.Date(2026, 9, 29, 1, 30, 0, 0, chinaZone)
	result, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("换日后领取失败: %v", err)
	}
	if !result.Claimed || client.calls != 2 {
		t.Fatalf("换日后应重新请求服务器: result=%+v calls=%d", result, client.calls)
	}
}

func TestClaimDailyVipSendsDateAndLoginCookie(t *testing.T) {
	client := &stubClient{resp: okResponse("")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}

	// receive_day 必须是日期字符串：实测传 int 1 会被上游拒绝（304001 日期格式错误）
	got, ok := client.lastReq.ReceiveDay.(string)
	if !ok || got != chinaToday(now) {
		t.Fatalf("receive_day = %#v (%T), want %q",
			client.lastReq.ReceiveDay, client.lastReq.ReceiveDay, chinaToday(now))
	}
	if client.lastReq.Cookie["token"] == "" || client.lastReq.Cookie["userid"] != testUserID {
		t.Fatalf("请求未带上登录态: %+v", client.lastReq.Cookie)
	}
	// 登录态要同步进 client 池，否则 SDK 会把请求参数里的 dfid 退化成占位符 "-"
	if client.cookies["dfid"] != "test-dfid" {
		t.Fatalf("dfid 未同步进 client: %+v", client.cookies)
	}
}

func TestClaimDailyVipRequiresLogin(t *testing.T) {
	client := &stubClient{resp: okResponse("")}
	now := baseTime()
	service := newTestService(t, client, &now, map[string]string{})

	if _, err := service.ClaimDailyVip(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("未登录应返回 ErrNotLoggedIn，实际 %v", err)
	}
	if client.calls != 0 {
		t.Fatalf("未登录不应请求服务器，实际 %d 次", client.calls)
	}
}

func TestClaimDailyVipMapsLoginError(t *testing.T) {
	client := &stubClient{err: errors.New("login expired: token refresh failed")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("登录态失效应返回 ErrNotLoggedIn，实际 %v", err)
	}
}

func TestClaimDailyVipCooldownAfterFailure(t *testing.T) {
	client := &stubClient{err: errors.New("boom")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err == nil {
		t.Fatal("请求失败时应返回错误")
	}

	// 冷却期内：直接报错，不请求服务器
	now = now.Add(10 * time.Second)
	_, err := service.ClaimDailyVip()
	if err == nil || !strings.Contains(err.Error(), "操作过于频繁") {
		t.Fatalf("冷却期内应返回频率限制错误，实际 %v", err)
	}
	if client.calls != 1 {
		t.Fatalf("冷却期内不应请求服务器，实际 %d 次", client.calls)
	}

	// 冷却结束后：允许重试
	now = now.Add(failureCooldown)
	client.err = nil
	client.resp = okResponse("")
	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("冷却结束后应允许重试: %v", err)
	}
	if client.calls != 2 {
		t.Fatalf("上游应被调用 2 次，实际 %d 次", client.calls)
	}
}

func TestClaimDailyVipBusinessFailure(t *testing.T) {
	client := &stubClient{resp: failResponse(20018, "登录已过期")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	_, err := service.ClaimDailyVip()
	if err == nil || !strings.Contains(err.Error(), "20018") {
		t.Fatalf("业务失败应带上 error_code，实际 %v", err)
	}
	if loadState(service.recordPath).Claim != nil {
		t.Fatal("业务失败不应写入领取记录")
	}
}

func TestClaimDailyVipIgnoresRecordOfOtherUser(t *testing.T) {
	client := &stubClient{resp: okResponse("")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if err := saveState(service.recordPath, state{Claim: &record{
		Date:   chinaToday(now),
		At:     now.UnixMilli(),
		Days:   claimedDays,
		UserID: "123456",
	}}); err != nil {
		t.Fatalf("预置记录失败: %v", err)
	}

	result, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	if !result.Claimed || client.calls != 1 {
		t.Fatalf("换了账号后应重新领取: result=%+v calls=%d", result, client.calls)
	}
}

func TestGetDailyVipStatus(t *testing.T) {
	client := &stubClient{resp: okResponse("")}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	status, err := service.GetDailyVipStatus()
	if err != nil {
		t.Fatalf("查询状态失败: %v", err)
	}
	if status.AlreadyClaimed || status.Claimed {
		t.Fatalf("尚未领取时状态异常: %+v", status)
	}

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	status, err = service.GetDailyVipStatus()
	if err != nil {
		t.Fatalf("查询状态失败: %v", err)
	}
	if !status.AlreadyClaimed || status.Claimed {
		t.Fatalf("已领取时状态异常: %+v", status)
	}
	if client.calls != 1 {
		t.Fatalf("状态查询不应请求服务器，上游调用 %d 次", client.calls)
	}
}

func TestUpgradeDailyVipRequiresClaimFirst(t *testing.T) {
	client := &stubClient{resp: okResponse(""), upgradeResp: okUpgradeResponse(nil)}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	// 没领过就当升级：本地直接拦住，一个请求都不发
	if _, err := service.UpgradeDailyVip(); !errors.Is(err, ErrClaimRequired) {
		t.Fatalf("未领取应返回 ErrClaimRequired，实际 %v", err)
	}
	if client.calls != 0 || client.upgradeCalls != 0 {
		t.Fatalf("未领取不应请求服务器，claim=%d upgrade=%d", client.calls, client.upgradeCalls)
	}
}

func TestUpgradeDailyVipAfterClaimThenSkips(t *testing.T) {
	client := &stubClient{
		resp:        okResponse("领取成功"),
		upgradeResp: okUpgradeResponse(map[string]any{"vip_type": "svip", "recharge_hours": 24}),
	}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	first, err := service.UpgradeDailyVip()
	if err != nil {
		t.Fatalf("升级失败: %v", err)
	}
	if !first.Upgraded || !first.Requested {
		t.Fatalf("首次升级结果异常: %+v", first)
	}
	if first.RewardType != "svip" || first.RechargeHours != 24 || first.UpgradedAt == 0 {
		t.Fatalf("首次升级类型/时间戳异常: %+v", first)
	}

	// 同一天内再次升级：不再请求服务器
	now = now.Add(2 * time.Hour)
	second, err := service.UpgradeDailyVip()
	if err != nil {
		t.Fatalf("重复升级不应报错: %v", err)
	}
	if second.Requested || !second.Upgraded || second.UpgradedAt != first.UpgradedAt {
		t.Fatalf("重复升级应命中本地记录: %+v", second)
	}
	// 命中本地记录时也要带上上次升级的时长
	if second.RechargeHours != 24 {
		t.Fatalf("重复升级应保留赠送时长: %+v", second)
	}
	if client.upgradeCalls != 1 {
		t.Fatalf("升级接口应只被调用 1 次，实际 %d 次", client.upgradeCalls)
	}

	// 升级记录不能覆盖领取记录：升级后领取仍应命中本地缓存
	st := loadState(service.recordPath)
	if st.Claim == nil || st.Upgrade == nil {
		t.Fatalf("本地状态应同时保留领取与升级记录: %+v", st)
	}
	claim, err := service.ClaimDailyVip()
	if err != nil {
		t.Fatalf("升级后重复领取失败: %v", err)
	}
	if claim.Claimed || client.calls != 1 {
		t.Fatalf("升级后重复领取应命中本地记录: result=%+v calls=%d", claim, client.calls)
	}
}

func TestUpgradeDailyVipNeedsNewClaimNextDay(t *testing.T) {
	client := &stubClient{resp: okResponse(""), upgradeResp: okUpgradeResponse(nil)}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	if _, err := service.UpgradeDailyVip(); err != nil {
		t.Fatalf("升级失败: %v", err)
	}

	// 第二天：昨天的升级记录失效，且必须先重新领取
	now = now.AddDate(0, 0, 1)
	if _, err := service.UpgradeDailyVip(); !errors.Is(err, ErrClaimRequired) {
		t.Fatalf("次日未领取应返回 ErrClaimRequired，实际 %v", err)
	}
	if client.upgradeCalls != 1 {
		t.Fatalf("次日未领取不应请求升级接口，实际 %d 次", client.upgradeCalls)
	}
}

func TestUpgradeDailyVipRecordsServerAlreadyUpgraded(t *testing.T) {
	client := &stubClient{
		resp:        okResponse(""),
		upgradeResp: failUpgradeResponse(30002, "今天已经升级过了"),
	}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	first, err := service.UpgradeDailyVip()
	if err != nil {
		t.Fatalf("服务端提示已升级不应报错: %v", err)
	}
	if !first.Upgraded || !first.Requested || first.ErrorCode != 30002 {
		t.Fatalf("服务端已升级结果异常: %+v", first)
	}

	if _, err := service.UpgradeDailyVip(); err != nil {
		t.Fatalf("重复升级失败: %v", err)
	}
	if client.upgradeCalls != 1 {
		t.Fatalf("升级接口应只被调用 1 次，实际 %d 次", client.upgradeCalls)
	}
}

func TestUpgradeDailyVipRequiresLogin(t *testing.T) {
	client := &stubClient{}
	now := baseTime()
	service := newTestService(t, client, &now, map[string]string{})

	if _, err := service.UpgradeDailyVip(); !errors.Is(err, ErrNotLoggedIn) {
		t.Fatalf("未登录应返回 ErrNotLoggedIn，实际 %v", err)
	}
	if client.upgradeCalls != 0 {
		t.Fatalf("未登录不应请求服务器，实际 %d 次", client.upgradeCalls)
	}
}

func TestUpgradeDailyVipCooldownAfterFailure(t *testing.T) {
	client := &stubClient{
		resp:       okResponse(""),
		upgradeErr: errors.New("boom"),
	}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	if _, err := service.UpgradeDailyVip(); err == nil {
		t.Fatal("升级失败时应返回错误")
	}

	now = now.Add(10 * time.Second)
	_, err := service.UpgradeDailyVip()
	if err == nil || !strings.Contains(err.Error(), "操作过于频繁") {
		t.Fatalf("冷却期内应返回频率限制错误，实际 %v", err)
	}
	if client.upgradeCalls != 1 {
		t.Fatalf("冷却期内不应请求服务器，实际 %d 次", client.upgradeCalls)
	}

	now = now.Add(failureCooldown)
	client.upgradeErr = nil
	client.upgradeResp = okUpgradeResponse(nil)
	if _, err := service.UpgradeDailyVip(); err != nil {
		t.Fatalf("冷却结束后应允许重试: %v", err)
	}
	if client.upgradeCalls != 2 {
		t.Fatalf("升级接口应被调用 2 次，实际 %d 次", client.upgradeCalls)
	}
}

func TestUpgradeDailyVipBusinessFailure(t *testing.T) {
	client := &stubClient{
		resp:        okResponse(""),
		upgradeResp: failUpgradeResponse(30001, "请先观看广告"),
	}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	_, err := service.UpgradeDailyVip()
	if err == nil || !strings.Contains(err.Error(), "30001") {
		t.Fatalf("业务失败应带上 error_code，实际 %v", err)
	}
	if st := loadState(service.recordPath); st.Upgrade != nil {
		t.Fatalf("业务失败不应写入升级记录: %+v", st.Upgrade)
	}
}

func TestUpgradeDailyVipSendsNoExternalParams(t *testing.T) {
	client := &stubClient{resp: okResponse(""), upgradeResp: okUpgradeResponse(nil)}
	now := baseTime()
	service := newTestService(t, client, &now, testSession())

	if _, err := service.ClaimDailyVip(); err != nil {
		t.Fatalf("领取失败: %v", err)
	}
	if _, err := service.UpgradeDailyVip(); err != nil {
		t.Fatalf("升级失败: %v", err)
	}

	// 升级只依赖登录态：userid / Extra 等参数一律不由外部传入
	if client.lastUpgradeReq.Userid != 0 || len(client.lastUpgradeReq.Extra) != 0 {
		t.Fatalf("升级请求不应带外部参数: %+v", client.lastUpgradeReq)
	}
	if client.lastUpgradeReq.Cookie["token"] == "" || client.lastUpgradeReq.Cookie["userid"] != testUserID {
		t.Fatalf("升级请求未带上登录态: %+v", client.lastUpgradeReq.Cookie)
	}
	// 登录态要同步进 client 池，否则 SDK 会把请求参数里的 dfid 退化成占位符 "-"
	if client.cookies["dfid"] != "test-dfid" {
		t.Fatalf("dfid 未同步进 client: %+v", client.cookies)
	}
}

func TestLoadStateMigratesLegacyClaimRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), recordFileName)
	legacy := `{"date":"2026-09-28","claimedAt":1790000000000,"days":1,` +
		`"userId":"863006722","message":"领取成功"}`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatalf("写入旧格式记录失败: %v", err)
	}

	st := loadState(path)
	if st.Claim == nil {
		t.Fatalf("旧格式记录应被读成领取记录: %+v", st)
	}
	if st.Claim.Date != "2026-09-28" || st.Claim.At != 1790000000000 {
		t.Fatalf("旧格式迁移结果异常: %+v", st.Claim)
	}
	if st.Upgrade != nil {
		t.Fatalf("旧格式不应产生升级记录: %+v", st.Upgrade)
	}
}

func TestChinaTodayAndNextActivityReset(t *testing.T) {
	// UTC 2026-09-28 17:00 == 北京时间 2026-09-29 01:00
	utc := time.Date(2026, 9, 28, 17, 0, 0, 0, time.UTC)
	if got := chinaToday(utc); got != "2026-09-29" {
		t.Fatalf("chinaToday() = %q, want 2026-09-29", got)
	}

	// 北京时间 00:30 的下一次换日是当天 01:00
	before := time.Date(2026, 9, 29, 0, 30, 0, 0, chinaZone)
	if got := nextActivityReset(before).In(chinaZone); got.Hour() != activityDayResetHour || got.Day() != 29 {
		t.Fatalf("nextActivityReset() = %v, want 2026-09-29 01:00", got)
	}

	// 北京时间 01:30 的下一次换日是次日 01:00
	after := time.Date(2026, 9, 29, 1, 30, 0, 0, chinaZone)
	if got := nextActivityReset(after).In(chinaZone); got.Hour() != activityDayResetHour || got.Day() != 30 {
		t.Fatalf("nextActivityReset() = %v, want 2026-09-30 01:00", got)
	}
}

func TestRecordPersistenceRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", recordFileName)
	want := state{
		Claim: &record{
			Date:          "2026-09-28",
			At:            1790000000000,
			Days:          1,
			UserID:        testUserID,
			Message:       "领取成功",
			ServerAlready: true,
			ExpiresAt:     1790032800000,
		},
	}
	if err := saveState(path, want); err != nil {
		t.Fatalf("写入记录失败: %v", err)
	}
	// 原子写不应留下临时文件
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("读取目录失败: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(path) {
		t.Fatalf("目录里应只剩 %s，实际 %v", filepath.Base(path), entries)
	}
	got := loadState(path)
	if got.Claim == nil || *got.Claim != *want.Claim {
		t.Fatalf("读取记录 = %+v, want %+v", got.Claim, want.Claim)
	}

	if st := loadState(filepath.Join(t.TempDir(), "missing.json")); st.Claim != nil || st.Upgrade != nil {
		t.Fatalf("缺失文件应返回空状态，实际 %+v", st)
	}
}

func TestLoadStateToleratesCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), recordFileName)
	if err := os.WriteFile(path, []byte(`{"claim":{"date":`), 0o600); err != nil {
		t.Fatalf("写入损坏文件失败: %v", err)
	}
	if st := loadState(path); st.Claim != nil || st.Upgrade != nil {
		t.Fatalf("损坏文件应退化成空状态，实际 %+v", st)
	}
}
