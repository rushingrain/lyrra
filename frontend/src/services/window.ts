import { ToggleWindowMaximise } from '../../wailsjs/go/main/App'

/** Wails 运行时是否可用（纯浏览器开发模式没有 window.go） */
function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.main?.App)
}

/**
 * 切换窗口的缩放（最大化）/ 还原，即 macOS 标题栏双击的默认行为。
 *
 * 隐藏标题栏后系统不再响应标题栏双击，由顶部留白处的双击调用；
 * 纯浏览器开发模式下没有绑定，静默跳过。
 */
export function toggleWindowMaximise(): void {
  if (!isBackendAvailable()) return
  ToggleWindowMaximise().catch((error) => {
    console.error('[window] 切换窗口缩放失败', error)
  })
}
