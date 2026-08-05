import { useCallback, useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'
import toast from 'react-hot-toast'
import { Sidebar } from './Sidebar'
import { MiniPlayer } from './MiniPlayer'
import { TopBar } from './TopBar'
import { getCurrentUser, logout } from '@/services/auth'
import type { LoginResult } from '@/types'
import { mockPlayHistory } from '@/data/mock'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'

const CONTENT_WRAPPER = 'mx-auto w-full max-w-[1400px]'

export function MainLayout() {
  const [playHistory] = useState(() => mockPlayHistory.length > 0 ? mockPlayHistory : [])
  const showPlayer = playHistory.length > 0
  const [user, setUser] = useState<LoginResult['user'] | null>(null)

  // 启动时恢复已保存的登录会话
  useEffect(() => {
    getCurrentUser().then(setUser)
  }, [])

  /** 登录成功：由 TopBar 透传 LoginModal 回填的用户信息 */
  const handleLoginSuccess = useCallback((u: LoginResult['user']) => {
    setUser(u)
  }, [])

  /** 退出登录：调用后端清空会话，无论成功与否都清除本地登录态 */
  const handleLogout = useCallback(() => {
    logout()
      .catch(() => {
        toast.error('退出登录失败，请稍后重试')
      })
      .finally(() => setUser(null))
  }, [])

  return (
    <div className="flex h-screen w-screen flex-col overflow-hidden bg-[var(--color-background)]">
      {/* 上部分：侧边栏 + 顶部栏/主内容区 */}
      <div className="flex flex-1 overflow-hidden">
        <Sidebar />

        <div className="flex flex-1 flex-col overflow-hidden">
          <header className="flex h-[72px] min-h-[72px] items-center bg-[var(--color-background)] px-10">
            <div className={CONTENT_WRAPPER}>
              <TopBar user={user} onLoginSuccess={handleLoginSuccess} onLogout={handleLogout} />
            </div>
          </header>

          <SimpleBar className="h-full scrollbar">
            <main className={`py-8 px-10 ${showPlayer ? 'pb-20' : 'pb-16'}`}>
              <div className={CONTENT_WRAPPER}>
                <Outlet />
              </div>
            </main>
          </SimpleBar>
        </div>
      </div>

      {/* 底部播放控制器：全宽 */}
      {showPlayer && (
        <div className="h-[82px] min-h-[82px] bg-[var(--color-background)]">
          <MiniPlayer />
        </div>
      )}
    </div>
  )
}
