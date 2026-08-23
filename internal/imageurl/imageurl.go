// Package imageurl 提供酷狗图床地址的统一规范化处理，
// 供 auth / banner / playlist / song 等服务复用。
package imageurl

import "strings"

// Normalize 将酷狗图床地址统一为 https，并把尺寸占位符 {size} 替换为 400。
//
// 背景：wails dev 模式下应用以 .app 形式运行，受 App Transport Security 限制，
// Info.dev.plist 只放行本地网络，远程 http:// 资源会被 WKWebView 拦截，
// 统一替换后各启动方式行为一致。
// 不含 {size} 的地址（如用户头像、banner 图）替换为 no-op，可安全复用。
func Normalize(raw string) string {
	raw = strings.ReplaceAll(raw, "{size}", "")
	if strings.HasPrefix(raw, "http://") {
		return "https://" + strings.TrimPrefix(raw, "http://")
	}
	return raw
}
