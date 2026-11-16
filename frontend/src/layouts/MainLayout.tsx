import { useCallback, useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'
import toast from 'react-hot-toast'
import { Sidebar, SIDEBAR_WIDTH_CLASS } from '@/layouts/Sidebar'
import { MiniPlayer } from '@/layouts/MiniPlayer'
import { TopBar } from '@/layouts/TopBar'
import { PlaylistPanel } from '@/components/business/PlaylistPanel'
import { LyricsView } from '@/components/business/LyricsView'
import { NowPlayingBackdrop } from '@/components/business/NowPlayingBackdrop'
import { WindowControls } from '@/layouts/WindowControls'
import { getCurrentUser, logout } from '@/services/auth'
import { clearDailyRecommendCache } from '@/services/song'
import { toggleWindowMaximise } from '@/services/window'
import { Environment } from '../../wailsjs/runtime'
import type { LoginResult } from '@/types'
import { mockPlayHistory } from '@/data/mock'
import { usePlayerStore } from '@/stores/playerStore'
import { initPlayerPersistence } from '@/stores/playerPersistence'
import { useAuthStore } from '@/stores/authStore'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'

const CONTENT_WRAPPER = 'mx-auto w-full max-w-[1400px]'

/**
 * 窗口顶部交通灯安全区高度：
 * main.go 里用 mac.TitleBarHiddenInset() 隐藏了标题文字但保留原生交通灯，
 * 交通灯会浮在网页内容之上，这里留出一条空白避免侧边栏 logo / 顶栏被压住。
 * 留白与顶部栏、侧边栏 Logo 区域使用同一强调色，避免出现颜色断层。
 */
const TITLEBAR_SAFE_AREA_CLASS = 'h-[28px] min-h-[28px] shrink-0 select-none titlebar-drag-region'

export function MainLayout() {
  const [isWindows, setIsWindows] = useState(false)
  const [playHistory] = useState(() => mockPlayHistory.length > 0 ? mockPlayHistory : [])
  const showPlayer = playHistory.length > 0
  // 登录态放在全局 store：MainLayout 负责写入，页面（如首页每日推荐）按需订阅
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  // 播放列表面板：贴在内容区右侧展开（高度从顶栏下沿到播放栏上沿）
  const [showPlaylist, setShowPlaylist] = useState(false)
  // 歌词详情页：覆盖侧边栏 / 顶栏 / 内容区（顶栏下沿到播放栏上沿），由播放栏封面唤出
  const [showLyrics, setShowLyrics] = useState(false)
  // 歌词展开时的沉浸背景取当前歌曲封面
  const currentCoverUrl = usePlayerStore((s) => s.song?.coverUrl)

  useEffect(() => {
    if (!('runtime' in window)) return
    let cancelled = false
    Environment()
      .then(({ platform }) => {
        if (!cancelled) setIsWindows(platform === 'windows')
      })
      .catch(() => { })
    return () => {
      cancelled = true
    }
  }, [])

  // 启动时恢复已保存的登录会话
  useEffect(() => {
    getCurrentUser().then((currentUser) => {
      clearDailyRecommendCache()
      setUser(currentUser)
    })
  }, [setUser])

  // 启动时恢复本地设备设置、播放队列与当前歌曲。
  useEffect(() => {
    initPlayerPersistence()
  }, [])

  /** 登录成功：由 TopBar 透传 LoginModal 回填的用户信息 */
  const handleLoginSuccess = useCallback((u: LoginResult['user']) => {
    clearDailyRecommendCache()
    setUser(u)
  }, [setUser])

  const clearAuthState = useCallback(() => {
    clearDailyRecommendCache()
    setUser(null)
    // 当前歌曲随登录态一起清掉，歌词页失去内容，同时收起
    setShowLyrics(false)
    // 后端已删除账号级播放数据，前端同步复位，避免旧队列继续显示或被回写
    usePlayerStore.getState().resetAccountData()
  }, [setUser])

  const closePlaylist = useCallback(() => setShowPlaylist(false), [])
  const togglePlaylist = useCallback(() => setShowPlaylist((open) => !open), [])
  const closeLyrics = useCallback(() => setShowLyrics(false), [])
  const toggleLyrics = useCallback(() => setShowLyrics((open) => !open), [])

  // 歌词页是覆盖整屏的沉浸视图，展开时收起播放列表面板，避免"面板开着却看不见"。
  // 反向不成立：歌词页展开时仍可唤出播放列表，面板会浮在歌词页之上。
  useEffect(() => {
    if (showLyrics) setShowPlaylist(false)
  }, [showLyrics])

  /** 退出登录：调用后端清空会话，无论成功与否都清除本地登录态 */
  const handleLogout = useCallback(() => {
    logout()
      .catch(() => {
        toast.error('退出登录失败，请稍后重试')
      })
      .finally(clearAuthState)
  }, [clearAuthState])

  /** 会话到期：清理本地状态并提示重新登录 */
  const handleSessionExpired = useCallback(() => {
    logout()
      .catch(() => { })
      .finally(() => {
        clearAuthState()
        toast.error('登录已过期，请重新登录')
      })
  }, [clearAuthState])

  // 前后端都会校验五天有效期；这里负责应用持续运行跨过过期时间点时立即退出登录
  useEffect(() => {
    const expiresAt = user?.sessionExpiresAt
    if (!expiresAt) return
    const remaining = expiresAt - Date.now()
    if (remaining <= 0) {
      handleSessionExpired()
      return
    }
    const timer = window.setTimeout(handleSessionExpired, remaining)
    return () => window.clearTimeout(timer)
  }, [user?.sessionExpiresAt, handleSessionExpired])

  return (
    <div className="relative flex h-screen w-screen flex-col overflow-hidden bg-[var(--color-background)]">
      {/*
        窗口顶部拖拽区：
        - Windows：自绘窗口按钮，歌词页时透明背景
        - macOS：保留原生交通灯安全区，使用 accent 背景
        双击切换窗口缩放（系统默认的标题栏双击行为）
      */}
      <div
        className={`${isWindows
          ? 'relative z-40 h-8 min-h-8 shrink-0 select-none titlebar-drag-region'
          : `${TITLEBAR_SAFE_AREA_CLASS} bg-[var(--color-accent)]`
          } flex ${isWindows && !showLyrics ? 'bg-[var(--color-accent)]' : ''} ${isWindows && showLyrics ? 'player-on-media' : ''}`}
        onDoubleClick={toggleWindowMaximise}
      >
        {isWindows ? (
          <>
            <div
              className={`${SIDEBAR_WIDTH_CLASS} ${showLyrics ? 'bg-transparent' : 'bg-[var(--color-accent)]'}`}
              aria-hidden="true"
            />
            <div className="flex-1" aria-hidden="true" />
            <WindowControls />
          </>
        ) : (
          <div className="flex-1" aria-hidden="true" />
        )}
      </div>

      {/* 上部分：侧边栏 + 顶部栏/主内容区 */}
      <div className="relative flex flex-1 overflow-hidden">
        <Sidebar user={user} />

        <div className="flex flex-1 flex-col overflow-hidden">
          <header className="flex h-[var(--layout-topbar-height)] shrink-0 items-center bg-[var(--color-accent)] px-10">
            <div className={CONTENT_WRAPPER}>
              <TopBar user={user} onLoginSuccess={handleLoginSuccess} onLogout={handleLogout} />
            </div>
          </header>

          {/* 内容区外层相对定位：播放列表面板贴右缘覆盖，不挤压正文 */}
          <div className="relative flex min-h-0 flex-1 flex-col overflow-hidden pt-3">
            <SimpleBar className="min-h-0 flex-1 scrollbar">
              <main className="pb-8 px-10">
                <div className={CONTENT_WRAPPER}>
                  <Outlet />
                </div>
              </main>
            </SimpleBar>
            {showPlaylist && <PlaylistPanel onClose={closePlaylist} />}
          </div>
        </div>

        {/* 歌词详情页：覆盖整窗上半部分（含侧边栏），底边紧贴播放栏 */}
        {showLyrics && <LyricsView onClose={closeLyrics} />}
      </div>

      {/* 沉浸背景：盖住上半区，也延伸到播放栏下方，让歌词与播放栏连成一体 */}
      {showLyrics && <NowPlayingBackdrop coverUrl={currentCoverUrl} />}

      {/* 底部播放控制器：全宽；歌词展开时背景透明并浮在沉浸背景之上 */}
      {showPlayer && (
        <div
          className={`h-[82px] min-h-[82px] ${showLyrics ? 'player-on-media relative z-40 bg-transparent' : 'bg-[var(--color-background)]'
            }`}
        >
          <MiniPlayer
            playlistOpen={showPlaylist}
            onTogglePlaylist={togglePlaylist}
            lyricsOpen={showLyrics}
            onToggleLyrics={toggleLyrics}
          />
        </div>
      )}
    </div>
  )
}
