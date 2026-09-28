package play

import (
	"strings"
	"sync"
)

// 音质档位标识。刻意与前端展示标签（Hi-Res / 无损 / HQ）解耦：
// 接入"用户设置播放品质"时，由设置那一层把 UI 文案映射成这里的键。
const (
	LevelRes      = "res"
	LevelSQ       = "sq"
	LevelHQ       = "hq"
	LevelStandard = "standard"
)

// DefaultPreferred 默认偏好档位：标准音质。
const DefaultPreferred = LevelStandard

// ladder 是唯一的档位顺序定义（从高到低）。
var ladder = []string{LevelRes, LevelSQ, LevelHQ, LevelStandard}

// fieldsByLevel 每个档位对应的候选 hash 字段，覆盖各接口的嵌套 / 扁平写法：
//
//	search 单曲 / 综合搜索：SQ.Hash、SQFileHash
//	search 歌词搜索：SQHash、320Hash
//	首页榜单（TopSong / TopCard）：hash_flac、hash_320、hash_high
//
// 含 "." 的项表示取一层嵌套字段。
var fieldsByLevel = map[string][]string{
	LevelRes:      {"Res.Hash", "ResFileHash", "hash_high"},
	LevelSQ:       {"SQ.Hash", "SQFileHash", "SQHash", "hash_flac"},
	LevelHQ:       {"HQ.Hash", "HQFileHash", "320Hash", "hash_320"},
	LevelStandard: {"FileHash", "hash"},
}

// qualityParams 档位 → GetSongPlayURL 的 quality 取值（实测）。
//
// 注意 res 对应 "high"：歌曲没有 Hi-Res 时，服务端会把 high 降级成别的文件
// （标记仍是 quhigh），所以拿到地址后必须用 URL 里的文件 hash 校验，见 GetSongPlayURL。
// 多轨（multitrack）与 viper_* 系音质不属于常规档位，不在这里维护。
var qualityParams = map[string]string{
	LevelRes:      "high",
	LevelSQ:       "flac",
	LevelHQ:       "320",
	LevelStandard: "128",
}

var (
	preferredMu sync.RWMutex
	// preferred 是进程内共享的播放品质偏好，属于用户级设置，搜索 / 首页 / 播放共用一份。
	preferred = DefaultPreferred
)

// SetPreferredQuality 设置播放品质偏好（res / sq / hq / standard）。
// 值未知或为空时，解析阶段会回落到 DefaultPreferred。
func SetPreferredQuality(quality string) {
	preferredMu.Lock()
	defer preferredMu.Unlock()
	preferred = strings.ToLower(strings.TrimSpace(quality))
}

// PreferredQuality 返回当前偏好（可能是未知值，由 PriorityFor 负责回落）。
func PreferredQuality() string {
	preferredMu.RLock()
	defer preferredMu.RUnlock()
	return preferred
}

// EffectiveQuality 返回归一化后的档位：偏好为空或未知时回落到 DefaultPreferred。
func EffectiveQuality() string {
	if level := normalizeLevel(PreferredQuality()); level != "" {
		return level
	}
	return DefaultPreferred
}

// Priority 按当前偏好返回档位顺序。
func Priority() []string {
	return PriorityFor(PreferredQuality())
}

// QualityParam 返回当前偏好对应的 GetSongPlayURL 的 quality 取值。
func QualityParam() string {
	return QualityParamFor(EffectiveQuality())
}

// QualityParamFor 返回指定档位对应的 quality 取值；未知档位回落到默认档位。
func QualityParamFor(level string) string {
	if normalized := normalizeLevel(level); normalized != "" {
		return qualityParams[normalized]
	}
	return qualityParams[DefaultPreferred]
}

// PriorityFor 返回从偏好档位开始、按音质从高到低截取的档位列表（不回头：
// 设 HQ 就只走 HQ → standard，不会再去试更高档位）。偏好为空或不在阶梯中时回落到默认档位。
func PriorityFor(preferred string) []string {
	index := levelIndex(preferred)
	if index < 0 {
		index = levelIndex(DefaultPreferred)
	}
	if index < 0 {
		return nil
	}
	return append([]string(nil), ladder[index:]...)
}

func levelIndex(level string) int {
	level = normalizeLevel(level)
	for i, candidate := range ladder {
		if candidate == level {
			return i
		}
	}
	return -1
}

// normalizeLevel 归一化档位标识；不是有效档位时返回空串。
func normalizeLevel(level string) string {
	level = strings.ToLower(strings.TrimSpace(level))
	for _, candidate := range ladder {
		if candidate == level {
			return level
		}
	}
	return ""
}

// Resolve 按档位顺序取第一个非空 hash；字段缺失或为空串都继续往下一档找。
// 全部档位都取不到时返回空串，前端据此提示该条目没有可用音质。
func Resolve(item map[string]any, priority []string) string {
	for _, level := range priority {
		for _, field := range fieldsByLevel[level] {
			if value := textValue(readNestedField(item, field)); value != "" {
				return value
			}
		}
	}
	return ""
}

// readNestedField 读取 "SQ.Hash" 这类一层嵌套字段；不含 "." 时直接取顶层。
// 外层缺失或不是对象时返回 nil（对 nil map 取值返回零值，不会 panic）。
func readNestedField(item map[string]any, field string) any {
	outer, inner, nested := strings.Cut(field, ".")
	if !nested {
		return item[outer]
	}
	nestedItem, _ := item[outer].(map[string]any)
	return nestedItem[inner]
}

// textValue 只接受字符串值并去掉首尾空白，与 mapper 侧的口径保持一致。
func textValue(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}
