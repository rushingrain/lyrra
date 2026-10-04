package play

import (
	"strings"
)

// 音质档位标识。刻意与前端展示标签（Hi-Res / 无损 / HQ）解耦：
// 接入"用户设置播放品质"时，由设置那一层把 UI 文案映射成这里的键。
const (
	LevelRes      = "res"
	LevelSQ       = "sq"
	LevelHQ       = "hq"
	LevelStandard = "standard"
)

// DefaultQuality 默认播放档位：标准音质。
const DefaultQuality = LevelStandard

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

// QualityParamFor 返回指定档位对应的 quality 取值；未知档位回落到默认档位。
func QualityParamFor(level string) string {
	if normalized := normalizeLevel(level); normalized != "" {
		return qualityParams[normalized]
	}
	return qualityParams[DefaultQuality]
}

// PriorityForAny 返回“任意档位有 hash 就算命中”的候选顺序：先目标档位，
// 再逐级向更高档位兜底，最后才回落到比目标更低的档位。
// 默认档位是最低档，所以默认结果是：标准 → HQ → 无损 → Hi-Res。
//
// hash 只用于定位歌曲，实际播放文件由请求里的 quality 参数决定（见 verify.go 的说明），
// 因此调用方只关心“能不能拿到一个可用 hash”，缺档时继续换档找比直接判定
// “没有可用音质”更合适。播放品质到上游参数的映射由 QualityParamFor 负责。
func PriorityForAny(quality string) []string {
	index := levelIndex(quality)
	if index < 0 {
		index = levelIndex(DefaultQuality)
	}
	if index < 0 {
		return nil
	}

	out := make([]string, 0, len(ladder))
	out = append(out, ladder[index])
	// ladder 从高到低排列：index 之前是更高档位，之后是更低档位。
	for i := index - 1; i >= 0; i-- {
		out = append(out, ladder[i])
	}
	for i := index + 1; i < len(ladder); i++ {
		out = append(out, ladder[i])
	}
	return out
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
