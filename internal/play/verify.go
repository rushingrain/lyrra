package play

import (
	"errors"
	"net/url"
	"path"
	"regexp"
	"strings"
)

// playURLFileHash 匹配歌曲播放地址里的文件 hash 段：.../v3/<32位hash>/...
var playURLFileHash = regexp.MustCompile(`/v3/([0-9a-fA-F]{32})/`)
var playURLQuality = regexp.MustCompile(`(?:^|[/_])qu(128|320|flac|high)(?:_|\.)`)

// buildSongPlayURL 校验并组装歌曲播放地址，独立成纯函数便于单测。
func buildSongPlayURL(level string, qualityParam string, expectedHash string, primaryURL string) (*PlayURL, error) {
	primaryURL = strings.TrimSpace(primaryURL)
	if primaryURL == "" {
		return nil, errors.New("该歌曲在当前音质下没有可用地址")
	}
	// viper_* 这类当前账号拿不到的音质，上游响应里没有音频地址，SDK 会把封面图当候选返回。
	if isImageURL(primaryURL) {
		return nil, errors.New("上游未返回该音质的音频地址")
	}

	fileHash := fileHashFromPlayURL(primaryURL)
	return &PlayURL{
		URL:          primaryURL,
		Quality:      level,
		QualityParam: qualityParam,
		FileHash:     fileHash,
		Degraded:     isDegradedQuality(level, fileHash, expectedHash, primaryURL),
	}, nil
}

// isDegradedQuality 用于比较请求的音质与返回的 CDN 文件名中的音质标记
// 哈希比较仅作为回退方案，因为播放 API 现在可以独立于歌曲哈希接收音质意图
func isDegradedQuality(level, fileHash, expectedHash, rawURL string) bool {
	match := playURLQuality.FindStringSubmatch(strings.ToLower(rawURL))
	if len(match) == 2 {
		actual := match[1]
		parsed, _ := url.Parse(rawURL)
		extension := strings.ToLower(path.Ext(parsed.Path))
		switch level {
		case LevelRes:
			// The upstream keeps the quhigh marker when it cannot provide Hi-Res;
			// a non-FLAC file is the observable degraded form.
			return actual != "high" || extension != ".flac"
		case LevelSQ:
			return actual != "flac"
		case LevelHQ:
			return actual != "320"
		case LevelStandard:
			return actual != "128"
		}
	}
	return fileHash != "" && expectedHash != "" && !strings.EqualFold(fileHash, expectedHash)
}

// fileHashFromPlayURL 取播放地址里的文件 hash，取不到时返回空串。
func fileHashFromPlayURL(playURL string) string {
	match := playURLFileHash.FindStringSubmatch(playURL)
	if match == nil {
		return ""
	}
	return strings.ToLower(match[1])
}

// isImageURL 判断地址是不是图片（用于识别"没有音频地址、SDK 返回封面图"的情况）。
func isImageURL(raw string) bool {
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch strings.ToLower(path.Ext(parsed.Path)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	default:
		return false
	}
}
