package auth

import (
	"crypto/rand"
	"strconv"
	"strings"
	"time"
)

// dfid 是酷狗的设备标识，搜索、用户歌单等接口会带上它。
// 服务端注册设备成功时会下发；下发失败时本地生成一个兜底，做法与 SDK 自身登录流程一致
// （sdk/login_flows.go 在缺少 dfid 时用 util.RandomString(24)）。
const (
	dfidLength = 24
	dfidChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// ensureDfid 保证 client 的 cookie 池里有一个可用的 dfid，返回最终生效的值。
// 服务端下发（或会话里已有）的优先；缺失、为空、或只是 SDK 的 "-" 占位时本地生成。
func (s *Service) ensureDfid() string {
	if dfid := strings.TrimSpace(s.client.Cookie()["dfid"]); dfid != "" && dfid != "-" {
		return dfid
	}

	dfid := randomDfid()
	s.client.SetCookie("dfid", dfid)
	s.log.Warn("device dfid missing, generated locally", "dfid", dfid)
	return dfid
}

// randomDfid 生成 24 位随机设备标识（大小写字母 + 数字）。
func randomDfid() string {
	buf := make([]byte, dfidLength)
	if _, err := rand.Read(buf); err != nil {
		return timestampDfid()
	}
	for i, value := range buf {
		buf[i] = dfidChars[int(value)%len(dfidChars)]
	}
	return string(buf)
}

// timestampDfid 是 crypto/rand 不可用时的退化实现，几乎不会走到。
func timestampDfid() string {
	value := strconv.FormatInt(time.Now().UnixNano(), 10)
	for len(value) < dfidLength {
		value += value
	}
	return value[:dfidLength]
}
