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
		Degraded:     fileHash != "" && !strings.EqualFold(fileHash, expectedHash),
	}, nil
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
