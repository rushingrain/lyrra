package play

import (
	"strings"
	"testing"
)

func TestPriorityForAny(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  []string
	}{
		{
			name:  "默认（空值）从标准档向上兜底",
			level: "",
			want:  []string{LevelStandard, LevelHQ, LevelSQ, LevelRes},
		},
		{
			name:  "未知值同样回落到默认档位",
			level: "viper",
			want:  []string{LevelStandard, LevelHQ, LevelSQ, LevelRes},
		},
		{
			name:  "HQ 先向上兜底，再回落到更低档位",
			level: LevelHQ,
			want:  []string{LevelHQ, LevelSQ, LevelRes, LevelStandard},
		},
		{
			name:  "Hi-Res 已是最高档，其余按序回退",
			level: LevelRes,
			want:  []string{LevelRes, LevelSQ, LevelHQ, LevelStandard},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PriorityForAny(tt.level)
			if strings.Join(got, ">") != strings.Join(tt.want, ">") {
				t.Fatalf("PriorityForAny(%q) = %v, want %v", tt.level, got, tt.want)
			}
		})
	}
}

// TestResolveAnyFallsBackToHigherQuality 覆盖搜索场景：条目没有标准档 hash 时，
// 仍取到更高档位的 hash（播放文件由 quality 参数决定，hash 只用于定位歌曲）。
func TestResolveAnyFallsBackToHigherQuality(t *testing.T) {
	item := map[string]any{"SQFileHash": "sq-flat-hash"}

	if got := Resolve(item, []string{LevelStandard}); got != "" {
		t.Fatalf("单档候选只取标准档 hash，实际 %q", got)
	}
	if got := Resolve(item, PriorityForAny(LevelStandard)); got != "sq-flat-hash" {
		t.Fatalf("PriorityForAny(标准) 应向上兜底取到 SQ，实际 %q", got)
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

func TestDefaultQuality(t *testing.T) {
	// 默认档位是最低档：候选顺序从标准档开始，逐级向上兜底。
	priority := PriorityForAny("")
	if len(priority) == 0 || priority[0] != LevelStandard {
		t.Fatalf("default priority = %v, want head %q", priority, LevelStandard)
	}
	if got := QualityParamFor(""); got != "128" {
		t.Fatalf("default quality param = %q, want 128", got)
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
		name   string
		item   map[string]any
		levels []string
		want   string
	}{
		{name: "song 标准档取根级 FileHash", item: songItem, levels: []string{LevelStandard}, want: "standard-hash"},
		{name: "首页标准档取根级 hash", item: homeItem, levels: []string{LevelStandard}, want: "standard-hash"},
		{name: "complex 无损取扁平 SQFileHash", item: complexItem, levels: []string{LevelSQ}, want: "sq-flat-hash"},
		{name: "song 无损取嵌套 SQ.Hash", item: songItem, levels: []string{LevelSQ}, want: "sq-nested-hash"},
		{name: "lyric 无损取扁平 SQHash", item: lyricItem, levels: []string{LevelSQ}, want: "sq-flat-hash"},
		{name: "song HQ 取嵌套 HQ.Hash", item: songItem, levels: []string{LevelHQ}, want: "hq-nested-hash"},
		{name: "首页无损取 hash_flac", item: homeItem, levels: []string{LevelSQ}, want: "sq-hash"},
		{name: "首页 Hi-Res 取 hash_high", item: homeItem, levels: []string{LevelRes}, want: "res-hash"},
		{name: "首页 HQ 取 hash_320", item: homeItem, levels: []string{LevelHQ}, want: "hq-hash"},
		{
			name:   "默认候选顺序优先标准档",
			item:   homeItem,
			levels: PriorityForAny(DefaultQuality),
			want:   "standard-hash",
		},
		{
			name:   "按候选顺序取第一个有值的档位",
			item:   map[string]any{"hash_320": "hq-hash", "hash": "standard-hash"},
			levels: []string{LevelSQ, LevelHQ, LevelStandard},
			want:   "hq-hash",
		},
		{
			name:   "空串字段视作缺失继续往下",
			item:   map[string]any{"ResFileHash": "", "SQFileHash": "", "HQFileHash": "hq-flat-hash"},
			levels: []string{LevelRes, LevelSQ, LevelHQ},
			want:   "hq-flat-hash",
		},
		{name: "完全没有可用 hash", item: map[string]any{"SongName": "无音质"}, levels: PriorityForAny(DefaultQuality), want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Resolve(tt.item, tt.levels)
			if got != tt.want {
				t.Fatalf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
