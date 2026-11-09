package kgapi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Record 是 SDK 动态 JSON 响应中的对象。
type Record = map[string]any

// RecordOf 将值转换为对象；不是对象时返回 nil。
func RecordOf(value any) Record {
	record, _ := value.(map[string]any)
	return record
}

// Records 将数组中的对象提取出来，忽略其它类型。
func Records(value any) []Record {
	items, _ := value.([]any)
	records := make([]Record, 0, len(items))
	for _, item := range items {
		if record := RecordOf(item); record != nil {
			records = append(records, record)
		}
	}
	return records
}

// FirstRecord 取数组中的第一个对象；数组为空或首项不是对象时返回 nil。
func FirstRecord(value any) Record {
	items, _ := value.([]any)
	if len(items) == 0 {
		return nil
	}
	return RecordOf(items[0])
}

// Text 只接受字符串值并去掉首尾空白。
func Text(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

// Scalar 将标量值规整为字符串；复合类型返回空串。
func Scalar(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		text := strings.TrimSpace(typed)
		if text == "<nil>" {
			return ""
		}
		return text
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10)
		}
		return fmt.Sprintf("%v", typed)
	case map[string]any, []any:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", typed))
	}
}

// Int64 将标量值转换为 int64，无法解析时返回 0。
func Int64(value any) int64 {
	number, _ := Int64OK(value)
	return number
}

// Int64OK 将标量值转换为 int64，并返回是否成功解析。
func Int64OK(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case float32:
		return int64(typed), true
	case int:
		return int64(typed), true
	case int64:
		return typed, true
	case json.Number:
		number, err := typed.Int64()
		if err == nil {
			return number, true
		}
		floatNumber, floatErr := typed.Float64()
		if floatErr != nil {
			return 0, false
		}
		return int64(floatNumber), true
	case string:
		text := strings.TrimSpace(typed)
		if text == "" {
			return 0, false
		}
		if number, err := strconv.ParseInt(text, 10, 64); err == nil {
			return number, true
		}
		number, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, false
		}
		return int64(number), true
	default:
		return 0, false
	}
}

// Int 将标量值转换为 int，无法解析时返回 0。
func Int(value any) int {
	return int(Int64(value))
}

// IntOK 将标量值转换为 int，并返回是否成功解析。
func IntOK(value any) (int, bool) {
	number, ok := Int64OK(value)
	return int(number), ok
}

// ScalarInt 先按 Scalar 口径格式化再取整数，兼容带小数尾巴的数字字符串。
func ScalarInt(value any) int {
	text := Scalar(value)
	if text == "" {
		return 0
	}
	var number int
	if _, err := fmt.Sscanf(text, "%d", &number); err != nil {
		return 0
	}
	return number
}

// Flag 判断上游标记是否为 1。
func Flag(value any) bool {
	return Scalar(value) == "1"
}

// StatusOK 判断响应对象是否满足酷狗业务成功状态 status == 1。
func StatusOK(body Record) bool {
	return Flag(body["status"])
}

// FirstNonEmpty 返回第一个非空字符串。
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
