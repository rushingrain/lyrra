import { create } from 'zustand'
import type { LoginResult } from '@/types'

type AuthUser = LoginResult['user']

interface AuthState {
  /** 当前登录用户；null 表示未登录（或启动时尚未恢复会话） */
  user: AuthUser | null
  /** 启动恢复 / 登录成功 / 登出时更新 */
  setUser: (user: AuthUser | null) => void
}

/**
 * 登录态（全局唯一数据源）。
 * MainLayout 负责写入（启动恢复、登录成功、登出），页面按需订阅。
 */
export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
}))
