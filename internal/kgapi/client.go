// Package kgapi 统一封装酷狗 SDK 动态响应的读取与 client 构造。
package kgapi

import kg "github.com/rushingrain/kugou-music-api"

// NewClient 创建各业务服务共用的 SDK client。
func NewClient(options ...kg.Option) (*kg.Client, error) {
	return kg.New(options...)
}
