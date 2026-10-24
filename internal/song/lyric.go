package song

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/kgapi"
)

// 歌词获取分两步：SearchLyric 按歌曲 hash 搜索候选，取第一条的 id + accesskey；
// 再用 Lyric 拉取歌词详情（本次固定 fmt=lrc、decode=true）。
//
// 上游没有歌词（纯音乐、尚未收录等）属于正常结果，返回空列表而不是错误。

var (
	// LRC 时间标签：`[mm:ss]`、`[mm:ss.xx]`（两位为厘秒）或 `[mm:ss.xxx]`，一行可能出现多个。
	lrcTimeTagPattern = regexp.MustCompile(`\[(\d+):(\d{1,2})(?:[.:](\d{1,3}))?\]`)
	// 整行只有 `[ar:xxx]`、`[offset:0]` 这类元信息标签（无时间轴），解析时跳过。
	lrcMetaLinePattern = regexp.MustCompile(`^\[[a-zA-Z]+:.*\]$`)
)

// GetLyrics 按歌曲 hash 获取结构化歌词（时间轴 + 文本），按时间升序返回。
//
// hash 传歌曲文件 hash（dto.Song.Hash）。上游没有对应歌词时返回空列表且 err 为 nil；
// 只有请求本身失败（网络错误、上游报错等）才返回非空 error。
func (s *Service) GetLyrics(hash string) ([]LyricLine, error) {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return nil, errors.New("缺少歌曲 hash")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 第一步：搜索歌词候选，取第一条可用的 id + accesskey。
	searchResp, err := s.client.SearchLyric(context.Background(), kg.SearchLyricRequest{Hash: hash})
	if err != nil {
		s.log.Warn("failed to search lyric", "hash", hash, "err", err.Error())
		return nil, err
	}
	if err := checkLyricResponse(searchResp.Body, "歌词搜索"); err != nil {
		return nil, err
	}
	id, accesskey := firstLyricCandidate(searchResp.Body)
	if id == "" || accesskey == "" {
		// 没有候选：返回空列表，前端显示"暂无歌词"。
		return []LyricLine{}, nil
	}

	// 第二步：拉取歌词详情，固定要 lrc 并让 SDK 直接解码。
	lyricResp, err := s.client.Lyric(context.Background(), kg.LyricRequest{
		Id:        id,
		Accesskey: accesskey,
		Fmt:       "lrc",
		Decode:    true,
	})
	if err != nil {
		s.log.Warn("failed to fetch lyric", "hash", hash, "id", id, "err", err.Error())
		return nil, err
	}
	if err := checkLyricResponse(lyricResp.Body, "歌词详情"); err != nil {
		return nil, err
	}

	// DecodedContent 负责 base64 / BOM / KRC 解密。个别响应缺 contenttype 会让 SDK 解码得到空串，
	// 此时退回原始 content：若它本就是 LRC 文本可正常解析，若是 base64 等非 LRC 内容，
	// parseLRC 找不到时间标签，同样收敛为空列表。
	text := lyricResp.DecodedContent()
	if strings.TrimSpace(text) == "" {
		text = kgapi.Text(lyricResp.Body["content"])
	}
	return parseLRC(text), nil
}

// firstLyricCandidate 取 candidates 里第一条同时带 id 与 accesskey 的候选。
// 上游已按匹配度排序，取第一条即可；没有可用候选时返回空字符串。
func firstLyricCandidate(body map[string]any) (id string, accesskey string) {
	for _, candidate := range kgapi.Records(body["candidates"]) {
		candidateID := kgapi.Text(candidate["id"])
		candidateKey := kgapi.Text(candidate["accesskey"])
		if candidateID == "" || candidateKey == "" {
			continue
		}
		return candidateID, candidateKey
	}
	return "", ""
}

// checkLyricResponse 校验歌词接口的业务状态。
// 歌词接口用 HTTP 语义的 status（成功为 200），个别实现返回 1，两者都算成功；
// status 缺失时不判断——无歌词属于正常业务结果，交由调用方按空列表处理。
func checkLyricResponse(body map[string]any, action string) error {
	status, ok := body["status"]
	if !ok {
		return nil
	}
	switch fmt.Sprintf("%v", status) {
	case "1", "200":
		return nil
	default:
		return fmt.Errorf("%s接口业务失败（status=%v）", action, status)
	}
}

// parseLRC 把 LRC 文本解析成按时间升序排列的歌词行，Time 单位为毫秒。
//
// 规则：跳过空行与 [ar:]/[ti:]/[offset:] 等元信息行；一行带多个时间标签时展开成多行；
// 文本取最后一个时间标签之后的部分，trim 后为空的行丢弃；忽略 [offset:] 偏移。
func parseLRC(text string) []LyricLine {
	lines := make([]LyricLine, 0, 64)
	for _, raw := range strings.Split(normalizeNewlines(text), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || lrcMetaLinePattern.MatchString(line) {
			continue
		}
		matches := lrcTimeTagPattern.FindAllStringSubmatchIndex(line, -1)
		if len(matches) == 0 {
			continue
		}
		content := strings.TrimSpace(line[matches[len(matches)-1][1]:])
		if content == "" {
			continue
		}
		for _, m := range matches {
			minutes, _ := strconv.Atoi(line[m[2]:m[3]])
			seconds, _ := strconv.Atoi(line[m[4]:m[5]])
			millis := int64(minutes*60+seconds) * 1000
			if m[6] >= 0 {
				millis += fractionToMillis(line[m[6]:m[7]])
			}
			lines = append(lines, LyricLine{Time: millis, Text: content})
		}
	}
	sort.SliceStable(lines, func(i, j int) bool { return lines[i].Time < lines[j].Time })
	return lines
}

// fractionToMillis 把时间标签的小数部分换算成毫秒：1 位为十分之一秒，2 位为厘秒，3 位为毫秒。
func fractionToMillis(fraction string) int64 {
	value, _ := strconv.Atoi(fraction)
	switch len(fraction) {
	case 1:
		return int64(value) * 100
	case 2:
		return int64(value) * 10
	default:
		return int64(value)
	}
}

// normalizeNewlines 统一换行符并去掉可能存在的 UTF-8 BOM。
func normalizeNewlines(text string) string {
	text = strings.TrimPrefix(text, "\ufeff")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}
