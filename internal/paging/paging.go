// Package paging 统一分页参数的回退规则。
package paging

// Normalize 将非正数的页码和页大小替换为默认值。
func Normalize(page, pageSize, defaultPageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	return page, pageSize
}

// NormalizeSize 将非正数的页大小替换为默认值。
func NormalizeSize(pageSize, defaultPageSize int) int {
	if pageSize <= 0 {
		return defaultPageSize
	}
	return pageSize
}
