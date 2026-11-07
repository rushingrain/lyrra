import {
  Music,
  UserRound,
  Radio,
  Heart,
  Clock,
  Cloud,
  Plus,
  ChevronUp,
  ChevronDown,
} from 'lucide-react'
import { NavLink } from 'react-router-dom'
import toast from 'react-hot-toast'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'
import type { LoginResult, UserPlaylist } from '@/types'
import { useEffect, useRef, useState } from 'react'
import { useClickStreak } from '@/hooks/useClickStreak'
import { useConfetti } from '@/hooks/useConfetti'
import { getUserPlaylists } from '@/services/userPlaylist'
import { claimDailyVip } from '@/services/vip'

/** 隐藏入口：logo 区域连续点击多少次触发领 VIP */
const LOGO_CLAIM_CLICK_TIMES = 7

/** 侧边栏宽度：MainLayout 的窗口顶部交通灯留白要与它对齐，所以在同一处定义 */
export const SIDEBAR_WIDTH_CLASS = 'w-[200px] min-w-[200px]'

interface NavItemProps {
  icon: React.ReactNode
  label: string
  to: string
}

function NavItem({ icon, label, to }: NavItemProps) {
  return (
    <NavLink
      to={to}
      className={({ isActive }) =>
        `flex items-center rounded-lg text-sm font-medium transition-all duration-200 gap-3 ml-3 mr-3 py-2 px-3 ${isActive
          ? 'text-white bg-[var(--color-accent)]'
          : 'text-[var(--color-text-secondary)] bg-transparent hover:bg-gray-200'
        }`
      }
    >
      {icon}
      <span>{label}</span>
    </NavLink>
  )
}

/** 侧边栏歌单项的最小结构：创建/收藏的歌单共用 */
interface SidebarPlaylist {
  id: string
  name: string
  coverUrl: string
}

interface PlaylistSectionProps {
  title: string
  playlists: SidebarPlaylist[]
  showCreate?: boolean
}

function PlaylistSection({ title, playlists, showCreate = false }: PlaylistSectionProps) {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <div className="flex flex-col px-2 py-2">
      <div className="flex items-center justify-between px-3 mx-3">
        <span className="text-xs font-medium text-[var(--color-text-secondary)]">
          {title}
        </span>
        <div className="flex items-center gap-0.5">
          {showCreate && (
            <button
              className="flex h-6 w-6 cursor-pointer items-center justify-center rounded text-[var(--color-text-secondary)] transition-colors hover:bg-gray-200"
              aria-label="新建歌单"
            >
              <Plus size={14} />
            </button>
          )}
          <button
            className="flex h-6 w-6 cursor-pointer items-center justify-center rounded text-[var(--color-text-secondary)] transition-colors hover:bg-gray-200"
            onClick={() => setCollapsed((v) => !v)}
            aria-label={collapsed ? '展开' : '收起'}
          >
            {collapsed ? <ChevronDown size={14} /> : <ChevronUp size={14} />}
          </button>
        </div>
      </div>
      {!collapsed && (
        <div className="mt-2">
          {playlists.map((pl) => (
            <a
              key={pl.id}
              href="#"
              className="flex items-center gap-3 rounded-lg px-3 py-1 mx-3 text-xs text-[var(--color-text-secondary)] transition-colors hover:bg-gray-200"
            >
              {pl.coverUrl ? (
                <img
                  src={pl.coverUrl}
                  alt={pl.name}
                  className="h-8 w-8 shrink-0 rounded object-cover"
                  loading="lazy"
                />
              ) : (
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded bg-gray-200">
                  <Music size={14} className="text-[var(--color-text-secondary)]" />
                </div>
              )}
              <span className="truncate">{pl.name}</span>
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

interface SidebarProps {
  /** 当前登录用户；未登录时为 null，此时不请求用户歌单 */
  user: LoginResult['user'] | null
}

export function Sidebar({ user }: SidebarProps) {
  // 用户歌单：SideBar 常驻挂载，登录后拉取一次即可，无需缓存。
  // 依赖 user 变化重新拉取（登录/登出），卸载或切换时用 cancelled 标志防泄漏。
  const [userPlaylists, setUserPlaylists] = useState<UserPlaylist[]>([])
  // type=1 为收藏的歌单，其余为创建的歌单
  const createdPlaylists = userPlaylists.filter((p) => p.type !== 1)
  const collectedPlaylists = userPlaylists.filter((p) => p.type === 1)

  const { celebrate, cry } = useConfetti()
  // 请求进行中时忽略后续触发：连点只会在第 7 次触发一次，避免并发请求
  const claimingRef = useRef(false)

  // 隐藏入口：在 logo 区域连续点 7 次领取当天 VIP。
  // 成功（含“今天已领过”的后端缓存结果）撒花，失败撒 ❄️；接口本身在后端做了
  // “一天只真正请求一次”的风控，所以这里重复触发也不会打到酷狗。
  // 未登录时整条链路关闭：不计数、不请求、也不放动画。
  const handleLogoClick = useClickStreak({
    enabled: Boolean(user),
    times: LOGO_CLAIM_CLICK_TIMES,
    onTrigger: () => {
      if (claimingRef.current) return
      claimingRef.current = true
      claimDailyVip()
        .then(() => celebrate())
        .catch((error: unknown) => {
          console.error('[vip] 领取当日 VIP 失败', error)
          cry()
        })
        .finally(() => {
          claimingRef.current = false
        })
    },
  })

  useEffect(() => {
    if (!user) {
      setUserPlaylists([])
      return
    }
    let cancelled = false
    getUserPlaylists()
      .then((data) => {
        if (!cancelled) setUserPlaylists(data)
      })
      .catch(() => {
        if (!cancelled) toast.error('获取歌单失败，请稍后重试')
      })
    return () => {
      cancelled = true
    }
  }, [user])

  return (
    <aside className={`h-full ${SIDEBAR_WIDTH_CLASS} bg-gray-50`}>
      <SimpleBar className="h-full scrollbar">
        {/* Logo：吸顶在侧边栏顶部，下方导航/歌单滚动时不会被带走 */}
        <div
          className="sticky top-0 z-10 flex h-[68px] min-h-[68px] items-center gap-2.5 bg-gray-50 px-6 select-none"
          onClick={handleLogoClick}
        >
          {/* 品牌字标：SVG 轮廓作为遮罩，颜色随主题色变量切换 */}
          <span
            className="sidebar-brand-wordmark h-9 w-[100px] shrink-0"
            role="img"
            aria-label="lyrra"
          />
        </div>

        {/* 主导航 */}
        <nav className="flex flex-col gap-1 px-2">
          <NavItem icon={<Music size={20} />} label="精选" to="/" />
          <NavItem icon={<UserRound size={20} />} label="关注" to="/following" />
          <NavItem icon={<Radio size={20} />} label="FM电台" to="/fm" />
        </nav>

        {/* 分隔线 */}
        <div className="border-t border-solid border-[var(--color-border)] my-4 mx-5" />
        <nav className="flex flex-col gap-1 px-2">
          <NavItem icon={<Heart size={20} />} label="我的喜欢" to="/likes" />
          <NavItem icon={<Clock size={20} />} label="最近播放" to="/history" />
          <NavItem icon={<Cloud size={20} />} label="我的云盘" to="/cloud" />
        </nav>

        {/* 分隔线 */}
        <div className="border-t border-solid border-[var(--color-border)] my-4 mx-5" />
        {/* 创建的歌单 */}
        <PlaylistSection title="创建的歌单" playlists={createdPlaylists} showCreate />
        {/* 分隔线 */}
        <div className="border-t border-solid border-[var(--color-border)] my-4 mx-5" />
        {/* 收藏的歌单 */}
        <PlaylistSection title="收藏的歌单" playlists={collectedPlaylists} />
      </SimpleBar>
    </aside>
  )
}
