package playlist

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net"
	"strings"

	kg "github.com/rushingrain/kugou-music-api"

	"lyrra/internal/kgapi"
)

// UpdateUserPlaylist 保存用户输入；封面先上传，成功后一次提交信息修改。
func (s *Service) UpdateUserPlaylist(globalID string, edit PlaylistEdit) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	edit.Name = strings.TrimSpace(edit.Name)
	if edit.Name == "" {
		return errors.New("请输入歌单名称")
	}
	cookie, err := s.loginCookie()
	if err != nil {
		return err
	}
	item, version, err := s.findUserPlaylist(globalID, cookie)
	if err != nil {
		return err
	}
	if err := requireCreated(item, true); err != nil {
		return err
	}
	tags := make([]string, 0, len(edit.Tags))
	seen := make(map[string]bool)
	for _, tag := range edit.Tags {
		tag = strings.TrimSpace(tag)
		if tag != "" && !seen[tag] {
			if strings.ContainsAny(tag, ",，") {
				return errors.New("单个标签不能包含逗号")
			}
			seen[tag] = true
			tags = append(tags, tag)
		}
	}
	pic := ""
	if edit.CoverData != "" {
		pic, err = s.uploadCover(edit.CoverData)
		if err != nil {
			return err
		}
		// 图片上传期间可能发生账号切换或远端编辑，重新读取归属与版本。
		if err := sameSession(cookie); err != nil {
			return err
		}
		item, version, err = s.findUserPlaylist(globalID, cookie)
		if err != nil {
			return err
		}
		if err := requireCreated(item, true); err != nil {
			return err
		}
	}
	if err := sameSession(cookie); err != nil {
		return err
	}
	resp, err := s.client.PlaylistUpdate(context.Background(), kg.PlaylistUpdateRequest{
		Listid: kgapi.Int(item["listid"]), TotalVer: version, Type: 0,
		Sort: kgapi.Int(item["sort"]), Name: edit.Name, Intro: edit.Description,
		Tags: strings.Join(tags, ","), Pic: pic, Cookie: cookie,
	})
	if err != nil {
		return errors.New("保存歌单信息失败，请稍后重试")
	}
	return playlistResponseError(resp, "保存歌单信息失败，请稍后重试")
}

// uploadCover 不接收本机路径，只上传本次由文件选择器读取的 JPG/PNG 内容。
func (s *Service) uploadCover(encoded string) (string, error) {
	const maxBytes = 10 * 1024 * 1024
	if len(encoded) > base64.StdEncoding.EncodedLen(maxBytes) {
		s.log.Warn("playlist cover rejected locally", "reason", "too_large", "encoded_bytes", len(encoded))
		return "", errors.New("封面大小不能超过 10 MB")
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(data) == 0 || len(data) > maxBytes {
		s.log.Warn("playlist cover rejected locally", "reason", "invalid_data", "encoded_bytes", len(encoded))
		return "", errors.New("封面图片数据无效")
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") || config.Width <= 0 || config.Height <= 0 {
		s.log.Warn("playlist cover rejected locally", "reason", "invalid_image", "bytes", len(data))
		return "", errors.New("请选择有效的 JPG 或 PNG 图片")
	}
	if config.Width < 400 || config.Height < 400 {
		s.log.Warn("playlist cover rejected locally", "reason", "dimensions_too_small", "width", config.Width, "height", config.Height, "format", format, "bytes", len(data))
		return "", errors.New("封面的宽度和高度均不能小于 400 像素")
	}
	logger := s.log.With("width", config.Width, "height", config.Height, "format", format, "bytes", len(data))
	extension := ".jpg"
	if format == "png" {
		extension = ".png"
	}
	resp, err := s.client.PlaylistUpload(context.Background(), kg.PlaylistUploadRequest{
		Data: data, ExtendName: extension, Type: "custom",
	})
	if err != nil {
		// URL 错误可能包含查询参数；不记录原始错误或响应体。
		var networkError net.Error
		timeout := errors.As(err, &networkError) && networkError.Timeout()
		var dnsError *net.DNSError
		var syntaxError *json.SyntaxError
		reason := "request_failed"
		if errors.As(err, &dnsError) {
			reason = "dns_failed"
		} else if errors.As(err, &syntaxError) {
			reason = "invalid_json_response"
		}
		logger.Warn("playlist cover upload request failed", "reason", reason, "error_type", fmt.Sprintf("%T", err), "timeout", timeout)
		if timeout {
			return "", errors.New("上传封面超时，请检查网络后重试")
		}
		if reason == "invalid_json_response" {
			return "", errors.New("封面上传服务响应异常，请稍后重试")
		}
		return "", errors.New("上传封面失败，请稍后重试")
	}
	if err := playlistResponseError(resp, "上传封面失败，请稍后重试"); err != nil {
		message := "上传服务拒绝了这张图片，请尝试其他 JPG 或 PNG 图片"
		reason := "upstream_rejected"
		if resp != nil {
			// 只映射已知原因，不把上游任意消息写入日志或透传。
			switch strings.ToLower(kgapi.Text(resp.Body["msg"])) {
			case "bad width":
				reason, message = "bad_width", "上传服务不支持这张封面的宽度，请调整图片尺寸后重试"
			case "bad height":
				reason, message = "bad_height", "上传服务不支持这张封面的高度，请调整图片尺寸后重试"
			}
			logger = logger.With("status", resp.Status, "upstream_status", kgapi.Int(resp.Body["status"]))
		}
		logger.Warn("playlist cover upload rejected", "reason", reason)
		return "", errors.New(message)
	}
	filename := kgapi.Text(resp.Body["FileName"])
	if filename == "" || strings.ContainsAny(filename, "\\?#|,") || strings.Contains(filename, "..") || strings.HasPrefix(filename, "/") || strings.Contains(filename, ":") {
		logger.Warn("playlist cover upload response invalid", "reason", "invalid_filename")
		return "", errors.New("上传接口未返回有效的封面文件名")
	}
	return "custom/" + filename, nil
}
