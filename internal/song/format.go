package song

import "strconv"

// 数量级阈值：<1000 原样显示，[1000, 1万) 用 k，[1万, 1亿) 用 w，≥1亿 用亿。
const (
	thousand       = 1000
	myriad         = 10000
	hundredMillion = 100000000
)

// formatCount 按统一的展示规则格式化数量：一律向下取整，不保留小数。
//
//	999      -> "999"
//	1000     -> "1k"
//	10000    -> "1w"
//	8552140  -> "855w"
//	100000000 -> "1亿"
func formatCount(n int64) string {
	switch {
	case n < thousand:
		return strconv.FormatInt(n, 10)
	case n < myriad:
		return strconv.FormatInt(n/thousand, 10) + "k"
	case n < hundredMillion:
		return strconv.FormatInt(n/myriad, 10) + "w"
	default:
		return strconv.FormatInt(n/hundredMillion, 10) + "亿"
	}
}
