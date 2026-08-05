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
import logo from '../../assets/images/logo-universal.png'
import type { LoginResult, UserPlaylist } from '@/types'
import { useEffect, useState } from 'react'
import { getUserPlaylists } from '@/services/userPlaylist'

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
        `flex items-center rounded-lg text-sm font-medium transition-all duration-200 gap-3 ml-3 mr-3 py-2 px-3 ${
          isActive
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
    <aside className="h-full w-[200px] min-w-[200px] bg-gray-50">
      <SimpleBar className="h-full scrollbar">
        {/* Logo */}
        <div className="flex items-center h-[72px] min-h-[72px] gap-2.5 px-6">
          <img src={logo} alt="kuwave" className="h-8 w-8 rounded-lg object-cover" />
          <span className="text-lg font-bold tracking-tight text-[var(--color-text-primary)]">
            kuwave
          </span>
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
