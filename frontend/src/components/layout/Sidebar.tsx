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
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'
import logo from '../../assets/images/logo-universal.png'
import { mockPlaylists } from '../../data/mock'
import type { Playlist } from '../../types'
import { useState } from 'react'

interface NavItemProps {
  icon: React.ReactNode
  label: string
  active?: boolean
}

function NavItem({ icon, label, active = false }: NavItemProps) {
  return (
    <a
      href="#"
      className={`flex items-center rounded-lg text-sm font-medium transition-all duration-200 gap-3 ml-3 mr-3 py-2 px-3 ${active ? 'text-white bg-[var(--color-accent)]' : 'text-[var(--color-text-secondary)] bg-transparent hover:bg-gray-200'}`}
    >
      {icon}
      <span>{label}</span>
    </a>
  )
}

interface PlaylistSectionProps {
  title: string
  playlists: Playlist[]
  showCreate?: boolean
}

function PlaylistSection({ title, playlists, showCreate = false }: PlaylistSectionProps) {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <div className="flex flex-col px-2">
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
              <img
                src={pl.coverUrl}
                alt={pl.name}
                className="h-8 w-8 shrink-0 rounded object-cover"
                loading="lazy"
              />
              <span className="truncate">{pl.name}</span>
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

interface SidebarProps {
  showPlayer?: boolean
}

export function Sidebar({ showPlayer = false }: SidebarProps) {
  return (
    <aside
      className={`h-screen w-[200px] min-w-[200px] bg-gray-50 ${showPlayer ? 'pb-30' : 'pb-16'}`}
    >
      <SimpleBar
        className="h-full scrollbar"
      >
        {/* Logo */}
        <div
          className="flex items-center h-[72px] min-h-[72px] gap-2.5 px-6"
        >
          <img src={logo} alt="kuwave" className="h-8 w-8 rounded-lg object-cover" />
          <span className="text-lg font-bold tracking-tight text-[var(--color-text-primary)]">
            kuwave
          </span>
        </div>

        {/* 主导航 */}
        <nav className="flex flex-col gap-1 px-2">
          <NavItem icon={<Music size={20} />} label="精选" active />
          <NavItem icon={<UserRound size={20} />} label="关注" />
          <NavItem icon={<Radio size={20} />} label="FM电台" />
        </nav>

        {/* 分隔线 */}
        <div className="border-t border-solid border-[var(--color-border)] my-4 mx-5" />
        <nav className="flex flex-col gap-1 px-2">
          <NavItem icon={<Heart size={20} />} label="我的喜欢" />
          <NavItem icon={<Clock size={20} />} label="最近播放" />
          <NavItem icon={<Cloud size={20} />} label="我的云盘" />
        </nav>

        {/* 分隔线 */}
        <div className="border-t border-solid border-[var(--color-border)] my-4 mx-5" />
        {/* 创建的歌单 */}
        <PlaylistSection title="创建的歌单" playlists={mockPlaylists} showCreate />
        {/* 分隔线 */}
        <div className="border-t border-solid border-[var(--color-border)] my-4 mx-5" />
        {/* 收藏的歌单 */}
        <PlaylistSection title="收藏的歌单" playlists={mockPlaylists} />
      </SimpleBar>
    </aside>
  )
}
