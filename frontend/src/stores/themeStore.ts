import { create } from 'zustand'
import toast from 'react-hot-toast'
import { getThemeColor, saveThemeColor } from '@/services/theme'

export const PRESET_THEMES = [
  { name: '朱孔阳', color: '#B81A35' },
  { name: '龙膏烛', color: '#DE82A7' },
  { name: '郁金裙', color: '#D08635' },
  { name: '铜青', color: '#3D8E86' },
  { name: '芰荷', color: '#4F794A' },
  { name: '天水碧', color: '#5AA4AE' },
  { name: '孔雀蓝', color: '#4994C4' },
  { name: '凝夜紫', color: '#422256' },
  { name: '齐紫', color: '#6C216D' },
] as const

export const DEFAULT_THEME_COLOR = '#10B981'

export function normalizeHexColor(value: string): string | null {
  const hex = value.trim()
  return /^#?[\da-f]{6}$/i.test(hex) ? `#${hex.replace(/^#/, '').toUpperCase()}` : null
}

function applyThemeColor(color: string): void {
  const root = document.documentElement
  root.style.setProperty('--color-accent', color)

  // 强调色上的文字和图标统一为白色，切换主题时不随亮度变化。
  root.style.setProperty('--color-on-accent', '#FFFFFF')
}

interface ThemeState {
  color: string
  ready: boolean
  saving: boolean
  setColor: (value: string) => Promise<boolean>
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  color: DEFAULT_THEME_COLOR,
  ready: false,
  saving: false,
  setColor: async (value) => {
    const color = normalizeHexColor(value)
    if (!color || !get().ready || get().saving) return false
    const previousColor = get().color
    set({ color, saving: true })
    applyThemeColor(color)
    try {
      await saveThemeColor(color)
      return true
    } catch {
      set({ color: previousColor })
      applyThemeColor(previousColor)
      toast.error('主题色保存失败，请重试')
      return false
    } finally {
      set({ saving: false })
    }
  },
}))

let initialized = false

export async function initThemePersistence(): Promise<void> {
  if (initialized) return
  initialized = true
  try {
    const color = normalizeHexColor((await getThemeColor()) ?? '') ?? DEFAULT_THEME_COLOR
    applyThemeColor(color)
    useThemeStore.setState({ color })
  } catch {
    applyThemeColor(DEFAULT_THEME_COLOR)
    toast.error('主题设置读取失败，已使用默认颜色')
  } finally {
    useThemeStore.setState({ ready: true })
  }
}
