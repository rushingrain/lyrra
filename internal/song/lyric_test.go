package song

import (
	"encoding/base64"
	"errors"
	"testing"

	kg "github.com/rushingrain/kugou-music-api"
)

// sampleLRC 取自实测歌词详情：带 BOM、CRLF 换行与 [ar:]/[ti:]/[offset:] 等元信息行。
const sampleLRC = "\ufeff[id:$00000000]\r\n" +
	"[ar:梁静茹]\r\n" +
	"[ti:给未来的自己]\r\n" +
	"[offset:0]\r\n" +
	"[00:00.52]梁静茹 - 给未来的自己\r\n" +
	"[00:04.00]作词：黄婷\r\n" +
	"\r\n" +
	"[01:04.41]找一个人惺惺相惜\r\n" +
	"[02:16.85]给未来的自己\r\n"

func TestParseLRCFiltersMetaAndConvertsTime(t *testing.T) {
	got := parseLRC(sampleLRC)
	want := []LyricLine{
		{Time: 520, Text: "梁静茹 - 给未来的自己"},
		{Time: 4000, Text: "作词：黄婷"},
		{Time: 64410, Text: "找一个人惺惺相惜"},
		{Time: 136850, Text: "给未来的自己"},
	}
	if len(got) != len(want) {
		t.Fatalf("parseLRC() 返回 %d 行，want %d：%+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseLRC()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseLRCMultipleTagsAndFraction(t *testing.T) {
	text := "[ar:测试]\n" +
		"[00:01.00][00:05.00]同一句\n" +
		"[00:07]无小数\n" +
		"[00:09.5]十分之一秒\n" +
		"[00:11.123]毫秒\n" +
		"[00:13.00]\n" +
		"无时间标签的说明行\n" +
		"[02:00.00]结束\n"

	got := parseLRC(text)
	want := []LyricLine{
		{Time: 1000, Text: "同一句"},
		{Time: 5000, Text: "同一句"},
		{Time: 7000, Text: "无小数"},
		{Time: 9500, Text: "十分之一秒"},
		{Time: 11123, Text: "毫秒"},
		{Time: 120000, Text: "结束"},
	}
	if len(got) != len(want) {
		t.Fatalf("parseLRC() 返回 %d 行，want %d：%+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseLRC()[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestGetLyrics 覆盖两步流程：搜索候选 -> 拉取详情（base64 编码的 lrc）。
func TestGetLyrics(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte(sampleLRC))
	client := &stubClient{
		searchLyricResp: &kg.SearchLyricResponse{Body: map[string]any{
			"status": float64(200),
			"candidates": []any{
				map[string]any{"id": "190523224", "accesskey": "ACCESSKEY"},
			},
		}},
		lyricResp: &kg.LyricResponse{Body: map[string]any{
			"status":      float64(200),
			"contenttype": float64(1),
			"content":     encoded,
		}},
	}
	service := newCountService(client)

	lines, err := service.GetLyrics(" 1DA15BA0B42B8DF19C27FC99A04B7829 ")
	if err != nil {
		t.Fatalf("GetLyrics() error = %v", err)
	}
	if len(lines) != 4 {
		t.Fatalf("GetLyrics() 返回 %d 行，want 4：%+v", len(lines), lines)
	}
	if lines[0].Time != 520 || lines[0].Text != "梁静茹 - 给未来的自己" {
		t.Fatalf("首行不符：%+v", lines[0])
	}
	if client.searchLyricReq.Hash != "1DA15BA0B42B8DF19C27FC99A04B7829" {
		t.Fatalf("歌词搜索 hash = %q", client.searchLyricReq.Hash)
	}
	if client.lyricReq.Id != "190523224" || client.lyricReq.Accesskey != "ACCESSKEY" {
		t.Fatalf("Lyric 入参不符：%+v", client.lyricReq)
	}
	if client.lyricReq.Fmt != "lrc" || client.lyricReq.Decode != true {
		t.Fatalf("Lyric 应固定 fmt=lrc、decode=true：%+v", client.lyricReq)
	}
}

func TestGetLyricsNoCandidateReturnsEmpty(t *testing.T) {
	client := &stubClient{
		searchLyricResp: &kg.SearchLyricResponse{Body: map[string]any{
			"status":     float64(200),
			"candidates": []any{},
		}},
	}
	service := newCountService(client)

	lines, err := service.GetLyrics("hash")
	if err != nil {
		t.Fatalf("GetLyrics() error = %v, want nil", err)
	}
	if lines == nil || len(lines) != 0 {
		t.Fatalf("GetLyrics() = %#v, want 非 nil 空列表", lines)
	}
	if client.lyricCalls != 0 {
		t.Fatalf("无候选时不应请求歌词详情，lyricCalls = %d", client.lyricCalls)
	}
}

func TestGetLyricsEmptyHash(t *testing.T) {
	service := newCountService(&stubClient{})
	if _, err := service.GetLyrics("   "); err == nil {
		t.Fatal("GetLyrics() 对空 hash 应返回错误")
	}
}

func TestGetLyricsSearchError(t *testing.T) {
	client := &stubClient{searchLyricErr: errors.New("boom")}
	service := newCountService(client)
	if _, err := service.GetLyrics("hash"); err == nil {
		t.Fatal("搜索失败时应返回错误")
	}
}

func TestGetLyricsStatusFailure(t *testing.T) {
	client := &stubClient{
		searchLyricResp: &kg.SearchLyricResponse{Body: map[string]any{"status": float64(500)}},
	}
	service := newCountService(client)
	if _, err := service.GetLyrics("hash"); err == nil {
		t.Fatal("业务失败时应返回错误")
	}
}
