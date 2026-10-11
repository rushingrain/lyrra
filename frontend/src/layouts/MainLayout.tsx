import { useCallback, useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'
import toast from 'react-hot-toast'
import { Sidebar, SIDEBAR_WIDTH_CLASS } from '@/components/layout/Sidebar'
import { MiniPlayer } from '@/components/layout/MiniPlayer'
import { TopBar } from '@/components/layout/TopBar'
import { PlaylistPanel } from '@/components/business/PlaylistPanel'
import { LyricsView } from '@/components/business/LyricsView'
import { NowPlayingBackdrop } from '@/components/business/NowPlayingBackdrop'
import { getCurrentUser, logout } from '@/services/auth'
import { toggleWindowMaximise } from '@/services/window'
import type { LoginResult } from '@/types'
import { mockPlayHistory } from '@/data/mock'
import { usePlayerStore } from '@/stores/playerStore'
import { initPlayerPersistence } from '@/stores/playerPersistence'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'

const CONTENT_WRAPPER = 'mx-auto w-full max-w-[1400px]'

/**
 * 窗口顶部交通灯安全区高度：
 * main.go 里用 mac.TitleBarHiddenInset() 隐藏了标题文字但保留原生交通灯，
 * 交通灯会浮在网页内容之上，这里留出一条空白避免侧边栏 logo / 顶栏被压住。
 * 留白按下方分区拆成两段（左段与侧边栏同色），避免出现颜色断层。
 */
const TITLEBAR_SAFE_AREA_CLASS = 'h-[28px] min-h-[28px] shrink-0 select-none titlebar-drag-region'

export function MainLayout() {
  const [playHistory] = useState(() => mockPlayHistory.length > 0 ? mockPlayHistory : [])
  const showPlayer = playHistory.length > 0
  const [user, setUser] = useState<LoginResult['user'] | null>(null)
  // 播放列表面板：贴在内容区右侧展开（高度从顶栏下沿到播放栏上沿）
  const [showPlaylist, setShowPlaylist] = useState(false)
  // 歌词详情页：覆盖侧边栏 / 顶栏 / 内容区（顶栏下沿到播放栏上沿），由播放栏封面唤出
  const [showLyrics, setShowLyrics] = useState(false)
  // 歌词展开时的沉浸背景取当前歌曲封面
  const currentCoverUrl = usePlayerStore((s) => s.song?.coverUrl)

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
      .finally(() => {
        setUser(null)
        // 当前歌曲随登出一起清掉，歌词页失去内容，同时收起
        setShowLyrics(false)
        // 后端已在 Logout 里删掉账号级播放数据，前端同步复位，避免旧队列继续显示或被回写
        usePlayerStore.getState().resetAccountData()
      })
  }, [])

  return (
    <div className="relative flex h-screen w-screen flex-col overflow-hidden bg-[var(--color-background)]">
      {/* 交通灯安全区：留白作为窗口拖拽区，双击切换窗口缩放（系统默认的标题栏双击行为） */}
      <div className={`${TITLEBAR_SAFE_AREA_CLASS} flex`} onDoubleClick={toggleWindowMaximise} aria-hidden>
        <div className={`${SIDEBAR_WIDTH_CLASS} bg-gray-50`} />
        <div className="flex-1" />
      </div>

      {/* 上部分：侧边栏 + 顶部栏/主内容区 */}
      <div className="relative flex flex-1 overflow-hidden">
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

        {/* 歌词详情页：覆盖整窗上半部分（含侧边栏），底边紧贴播放栏 */}
        {showLyrics && <LyricsView onClose={closeLyrics} />}
      </div>

      {/* 沉浸背景：盖住上半区，也延伸到播放栏下方，让歌词与播放栏连成一体 */}
      {showLyrics && <NowPlayingBackdrop coverUrl={currentCoverUrl} />}

      {/* 底部播放控制器：全宽；歌词展开时背景透明并浮在沉浸背景之上 */}
      {showPlayer && (
        <div
          className={`h-[82px] min-h-[82px] ${
            showLyrics ? 'player-on-media relative z-40 bg-transparent' : 'bg-[var(--color-background)]'
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
