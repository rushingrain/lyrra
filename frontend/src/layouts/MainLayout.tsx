import { useCallback, useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'
import toast from 'react-hot-toast'
import { Sidebar } from '@/components/layout/Sidebar'
import { MiniPlayer } from '@/components/layout/MiniPlayer'
import { TopBar } from '@/components/layout/TopBar'
import { PlaylistPanel } from '@/components/business/PlaylistPanel'
import { getCurrentUser, logout } from '@/services/auth'
import type { LoginResult } from '@/types'
import { mockPlayHistory } from '@/data/mock'
import { usePlayerStore } from '@/stores/playerStore'
import { initPlayerPersistence } from '@/stores/playerPersistence'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'

const CONTENT_WRAPPER = 'mx-auto w-full max-w-[1400px]'

export function MainLayout() {
  const [playHistory] = useState(() => mockPlayHistory.length > 0 ? mockPlayHistory : [])
  const showPlayer = playHistory.length > 0
  const [user, setUser] = useState<LoginResult['user'] | null>(null)
  // 播放列表面板：贴在内容区右侧展开（高度从顶栏下沿到播放栏上沿）
  const [showPlaylist, setShowPlaylist] = useState(false)

  // 启动时恢复已保存的登录会话
  useEffect(() => {
    getCurrentUser().then(setUser)
  }, [])

  // 启动时恢复本地播放状态（播放模式 / 音量 / 播放队列 / 当前歌曲）
  useEffect(() => {
    initPlayerPersistence()
  }, [])

  /** 登录成功：由 TopBar 透传 LoginModal 回填的用户信息 */
  const handleLoginSuccess = useCallback((u: LoginResult['user']) => {
    setUser(u)
  }, [])

  const closePlaylist = useCallback(() => setShowPlaylist(false), [])
  const togglePlaylist = useCallback(() => setShowPlaylist((open) => !open), [])

  /** 退出登录：调用后端清空会话，无论成功与否都清除本地登录态 */
  const handleLogout = useCallback(() => {
    logout()
      .catch(() => {
        toast.error('退出登录失败，请稍后重试')
      })
      .finally(() => {
        setUser(null)
        // 后端已在 Logout 里删掉账号级播放数据，前端同步复位，避免旧队列继续显示或被回写
        usePlayerStore.getState().resetAccountData()
      })
  }, [])

  return (
    <div className="flex h-screen w-screen flex-col overflow-hidden bg-[var(--color-background)]">
      {/* 上部分：侧边栏 + 顶部栏/主内容区 */}
      <div className="flex flex-1 overflow-hidden">
        <Sidebar user={user} />

        <div className="flex flex-1 flex-col overflow-hidden">
          <header className="flex h-[72px] min-h-[72px] items-center bg-[var(--color-background)] px-10">
            <div className={CONTENT_WRAPPER}>
              <TopBar user={user} onLoginSuccess={handleLoginSuccess} onLogout={handleLogout} />
            </div>
          </header>

          {/* 内容区外层相对定位：播放列表面板贴右缘覆盖，不挤压正文 */}
          <div className="relative flex-1 overflow-hidden">
            <SimpleBar className="h-full scrollbar">
              <main className="py-8 px-10">
                <div className={CONTENT_WRAPPER}>
                  <Outlet />
                </div>
              </main>
            </SimpleBar>
            {showPlaylist && <PlaylistPanel onClose={closePlaylist} />}
          </div>
        </div>
      </div>

      {/* 底部播放控制器：全宽 */}
      {showPlayer && (
        <div className="h-[82px] min-h-[82px] bg-[var(--color-background)]">
          <MiniPlayer playlistOpen={showPlaylist} onTogglePlaylist={togglePlaylist} />
        </div>
      )}
    </div>
  )
}
