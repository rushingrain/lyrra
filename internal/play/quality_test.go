package play

import (
	"strings"
	"testing"
)

func TestPriorityFor(t *testing.T) {
	tests := []struct {
		name      string
		preferred string
		want      []string
	}{
		{name: "默认（空值）落在标准", preferred: "", want: []string{LevelStandard}},
		{name: "未知值回落到默认档位", preferred: "viper", want: []string{LevelStandard}},
		{name: "标准", preferred: LevelStandard, want: []string{LevelStandard}},
		{name: "无损", preferred: LevelSQ, want: []string{LevelSQ, LevelHQ, LevelStandard}},
		{name: "HQ 不回头", preferred: LevelHQ, want: []string{LevelHQ, LevelStandard}},
		{name: "Hi-Res", preferred: LevelRes, want: []string{LevelRes, LevelSQ, LevelHQ, LevelStandard}},
		{name: "大小写与空白容忍", preferred: "  RES ", want: []string{LevelRes, LevelSQ, LevelHQ, LevelStandard}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PriorityFor(tt.preferred)
			if strings.Join(got, ">") != strings.Join(tt.want, ">") {
				t.Fatalf("PriorityFor(%q) = %v, want %v", tt.preferred, got, tt.want)
			}
		})
	}
}

func TestQualityParamFor(t *testing.T) {
	tests := []struct {
		level string
		want  string
	}{
		{level: LevelRes, want: "high"},
		{level: LevelSQ, want: "flac"},
		{level: LevelHQ, want: "320"},
		{level: LevelStandard, want: "128"},
		{level: "", want: "128"},
		{level: "viper_atmos", want: "128"},
	}

	for _, tt := range tests {
		t.Run("level="+tt.level, func(t *testing.T) {
			if got := QualityParamFor(tt.level); got != tt.want {
				t.Fatalf("QualityParamFor(%q) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestPreferredQuality(t *testing.T) {
	t.Cleanup(func() { SetPreferredQuality(DefaultPreferred) })

	if got := PreferredQuality(); got != DefaultPreferred {
		t.Fatalf("default preferred = %q, want %q", got, DefaultPreferred)
	}
	if got := strings.Join(Priority(), ">"); got != LevelStandard {
		t.Fatalf("default priority = %q, want %q", got, LevelStandard)
	}
	if got := QualityParam(); got != "128" {
		t.Fatalf("default quality param = %q, want 128", got)
	}

	SetPreferredQuality(" HQ ")
	if got := PreferredQuality(); got != LevelHQ {
		t.Fatalf("preferred = %q, want %q", got, LevelHQ)
	}
	if got := EffectiveQuality(); got != LevelHQ {
		t.Fatalf("effective = %q, want %q", got, LevelHQ)
	}
	if got := strings.Join(Priority(), ">"); got != "hq>standard" {
		t.Fatalf("priority after setting HQ = %q", got)
	}
	if got := QualityParam(); got != "320" {
		t.Fatalf("quality param after setting HQ = %q, want 320", got)
	}

	// 未知值：原样保留在 PreferredQuality，但解析/映射时回落到默认档位
	SetPreferredQuality("viper_atmos")
	if got := EffectiveQuality(); got != DefaultPreferred {
		t.Fatalf("effective for unknown = %q, want %q", got, DefaultPreferred)
	}
	if got := QualityParam(); got != "128" {
		t.Fatalf("quality param for unknown = %q, want 128", got)
	}
}

func TestResolve(t *testing.T) {
	// 各接口存 hash 的位置不同，用真实结构做样本。
	complexItem := map[string]any{
		"SQFileHash": "sq-flat-hash",
		"HQFileHash": "hq-flat-hash",
		"FileHash":   "standard-hash",
	}
	songItem := map[string]any{
		"SQ":       map[string]any{"Hash": "sq-nested-hash"},
		"HQ":       map[string]any{"Hash": "hq-nested-hash"},
		"FileHash": "standard-hash",
	}
	lyricItem := map[string]any{
		"SQHash":   "sq-flat-hash",
		"320Hash":  "hq-flat-hash",
		"FileHash": "standard-hash",
	}
	// 首页榜单：TopSong / TopCard 的字段名
	homeItem := map[string]any{
		"hash_high": "res-hash",
		"hash_flac": "sq-hash",
		"hash_320":  "hq-hash",
		"hash":      "standard-hash",
	}

	tests := []struct {
		name      string
		item      map[string]any
		preferred string
		want      string
	}{
		{name: "默认档位（标准）只取标准 hash", item: songItem, preferred: "", want: "standard-hash"},
		{name: "首页默认档位只取 hash", item: homeItem, preferred: "", want: "standard-hash"},
		{name: "complex 无损取扁平 SQ", item: complexItem, preferred: LevelSQ, want: "sq-flat-hash"},
		{name: "song 无损取嵌套 SQ", item: songItem, preferred: LevelSQ, want: "sq-nested-hash"},
		{name: "lyric 无损取扁平 SQHash", item: lyricItem, preferred: LevelSQ, want: "sq-flat-hash"},
		{name: "song 设 HQ 时取嵌套 HQ", item: songItem, preferred: LevelHQ, want: "hq-nested-hash"},
		{name: "song 设标准时只取 FileHash", item: songItem, preferred: LevelStandard, want: "standard-hash"},
		{name: "首页无损取 hash_flac", item: homeItem, preferred: LevelSQ, want: "sq-hash"},
		{name: "首页设 Hi-Res 取 hash_high", item: homeItem, preferred: LevelRes, want: "res-hash"},
		{name: "首页设 HQ 取 hash_320", item: homeItem, preferred: LevelHQ, want: "hq-hash"},
		{name: "首页设标准取 hash", item: homeItem, preferred: LevelStandard, want: "standard-hash"},
		{
			name:      "首页缺 hash_flac 时降到 hash_320",
			item:      map[string]any{"hash_320": "hq-hash", "hash": "standard-hash"},
			preferred: LevelSQ,
			want:      "hq-hash",
		},
		{
			name:      "没有 SQ 时降到 HQ 再降到标准",
			item:      map[string]any{"FileHash": "standard-hash"},
			preferred: LevelSQ,
			want:      "standard-hash",
		},
		{
			name:      "空串字段视作缺失继续往下",
			item:      map[string]any{"ResFileHash": "", "SQFileHash": "", "HQFileHash": "hq-flat-hash"},
			preferred: LevelRes,
			want:      "hq-flat-hash",
		},
		{
			name:      "Hi-Res 档优先取 Res",
			item:      map[string]any{"ResFileHash": "res-hash", "SQFileHash": "sq-flat-hash"},
			preferred: LevelRes,
			want:      "res-hash",
		},
		{
			name:      "偏好不好时不会回头取更高档位",
			item:      map[string]any{"ResFileHash": "res-hash", "FileHash": "standard-hash"},
			preferred: LevelHQ,
			want:      "standard-hash",
		},
		{name: "完全没有可用 hash", item: map[string]any{"SongName": "无音质"}, preferred: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.item, PriorityFor(tt.preferred))
			if got != tt.want {
				t.Fatalf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
