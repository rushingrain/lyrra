import { GetThemeColor, SaveThemeColor } from '../../wailsjs/go/main/App'

function isBackendAvailable(): boolean {
  return Boolean((window as Window & { go?: { main?: { App?: unknown } } }).go?.main?.App)
}

export async function getThemeColor(): Promise<string | null> {
  if (isBackendAvailable()) return GetThemeColor()
  // 仅供纯浏览器预览；桌面应用始终通过 Go 读写配置。
  return localStorage.getItem('lyrra:mock-theme-color')
}

export async function saveThemeColor(color: string): Promise<void> {
  if (isBackendAvailable()) return SaveThemeColor(color)
  localStorage.setItem('lyrra:mock-theme-color', color)
}
