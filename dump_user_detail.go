//go:build ignore

// 自己调用 /v3/get_my_info 的独立脚本（不参与应用编译）。
// 用法： go run dump_user_detail.go
// 它会：
//  1. 打印一条与 SDK 构造完全一致的 curl 命令（signature/p/mid 已现场算好）
//  2. 实际执行请求，并美化打印完整响应 JSON
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lfhy/kugou-music-api/core/kugou"
	"github.com/lfhy/kugou-music-api/core/util"
)

// 与 SDK sdk/user_detail_manual.go 保持一致（JSON 字段顺序影响签名）
type userDetailSignature struct {
	Token      string `json:"token"`
	ClientTime int64  `json:"clienttime"`
}
type userDetailPayload struct {
	VisitTime int64  `json:"visit_time"`
	UserType  int    `json:"usertype"`
	P         string `json:"p"`
	UserID    int    `json:"userid"`
}

func main() {
	session, err := loadSession()
	if err != nil {
		fatal("读取会话失败（请先在应用内扫码登录）: %v", err)
	}
	token := session["token"]
	userID := parseUserID(session["userid"])

	// ---- 以下与 SDK lite 平台请求构造完全一致 ----
	clientTime := time.Now().Unix()
	p, err := kugou.CryptoRSAEncryptRawHex(
		userDetailSignature{Token: token, ClientTime: clientTime},
		kugou.PublicLiteRASKey,
	)
	if err != nil {
		fatal("RSA 加密失败: %v", err)
	}
	p = strings.ToUpper(p)

	appid, clientver := "3116", "11440" // lite 平台

	// 设备身份：guid → mid，与 sdk injectPlatformCookies 一致
	guid := util.MD5Hex(util.RandomString(16))
	mid := util.CalculateMid(guid)
	dev := strings.ToUpper(util.RandomString(10))
	cookieMap := map[string]string{
		"token":              token,
		"userid":             session["userid"],
		"KUGOU_API_PLATFORM": "lite",
		"KUGOU_API_MID":      mid,
		"KUGOU_API_GUID":     guid,
		"KUGOU_API_DEV":      dev,
		"KUGOU_API_MAC":      "02:00:00:00:00:00",
	}

	params := map[string]any{
		"dfid":       "-",
		"mid":        mid,
		"uuid":       "-",
		"appid":      appid,
		"clientver":  clientver,
		"clienttime": strconv.FormatInt(clientTime, 10),
		"plat":       1,
		"token":      token,
		"userid":     strconv.Itoa(userID),
	}

	rawBody, _ := json.Marshal(userDetailPayload{VisitTime: clientTime, UserType: 1, P: p, UserID: userID})
	dataString := string(rawBody)
	params["signature"] = kugou.SignatureAndroidParams(params, dataString, true)

	u, _ := url.Parse("https://gateway.kugou.com/v3/get_my_info")
	q := u.Query()
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		q.Set(k, fmt.Sprintf("%v", params[k]))
	}
	u.RawQuery = q.Encode()

	headers := map[string]string{
		"User-Agent":   "Android15-1070-11083-46-0-DiscoveryDRADProtocol-wifi",
		"Content-Type": "application/json",
		"dfid":         "-",
		"clienttime":   strconv.FormatInt(clientTime, 10),
		"mid":          mid,
		"kg-rc":        "1",
		"kg-thash":     "5d816a0",
		"kg-rec":       "1",
		"kg-rf":        "B9EDA08A64250DEFFBCADDEE00F8F25F",
		"x-router":     "usercenter.kugou.com",
	}

	cookieKeys := make([]string, 0, len(cookieMap))
	for k, v := range cookieMap {
		if strings.TrimSpace(k) != "" && strings.TrimSpace(v) != "" {
			cookieKeys = append(cookieKeys, k)
		}
	}
	sort.Strings(cookieKeys)
	cookieParts := make([]string, 0, len(cookieKeys))
	for _, k := range cookieKeys {
		cookieParts = append(cookieParts, k+"="+cookieMap[k])
	}
	cookieHeader := strings.Join(cookieParts, "; ")

	// ---- 打印可运行的 curl ----
	var hdr strings.Builder
	order := []string{"User-Agent", "Content-Type", "dfid", "clienttime", "mid", "kg-rc", "kg-thash", "kg-rec", "kg-rf", "x-router"}
	for _, h := range order {
		fmt.Fprintf(&hdr, " -H '%s: %s'", h, headers[h])
	}
	fmt.Fprintf(&hdr, " -H 'Cookie: %s'", cookieHeader)
	fmt.Printf("== 可直接运行的 curl ==\ncurl -X POST '%s'%s --data-raw '%s'\n\n",
		u.String(), hdr.String(), dataString)

	// ---- 实际执行并美化输出 ----
	req, _ := http.NewRequestWithContext(context.Background(), "POST", u.String(), bytes.NewReader(rawBody))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Cookie", cookieHeader)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fatal("请求失败: %v", err)
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(resp.Body)

	fmt.Printf("== 响应（HTTP %d）==\n", resp.StatusCode)
	var pretty any
	if json.Unmarshal(rb, &pretty) == nil {
		out, _ := json.MarshalIndent(pretty, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Println(string(rb))
	}
}

func loadSession() (map[string]string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	data, err := os.ReadFile(filepath.Join(dir, "kuwave", "session.json"))
	if err != nil {
		return nil, err
	}
	var cookie map[string]string
	if err := json.Unmarshal(data, &cookie); err != nil {
		return nil, err
	}
	return cookie, nil
}

func parseUserID(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if fv, err := strconv.ParseFloat(s, 64); err == nil {
		return int(fv)
	}
	if iv, err := strconv.Atoi(s); err == nil {
		return iv
	}
	return 0
}

func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
