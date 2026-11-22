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
  Eye,
  Play,
  SquarePen,
  Trash,
  BookmarkMinus,
} from 'lucide-react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import toast from 'react-hot-toast'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'
import type { LoginResult, UserPlaylist } from '@/types'
import { useCallback, useEffect, useRef, useState } from 'react'
import { useClickStreak } from '@/hooks/useClickStreak'
import { useConfetti } from '@/hooks/useConfetti'
import { deleteUserPlaylist, playlistError } from '@/services/userPlaylist'
import { getAllPlaylistSongs } from '@/services/playlist'
import { useAuthStore } from '@/stores/authStore'
import { usePlayerStore } from '@/stores/playerStore'
import { useUserPlaylistStore } from '@/stores/userPlaylistStore'
import { ContextMenu, type MenuPosition } from '@/components/common/ContextMenu'
import { ConfirmDialog } from '@/components/common/ConfirmDialog'
import { claimDailyVip } from '@/services/vip'
import { DefaultPlaylistCover } from '@/components/common/DefaultPlaylistCover'
import eighthNote from '@/assets/icons/eighth-note.svg'

/** 隐藏入口：logo 区域连续点击多少次触发领 VIP */
const LOGO_CLAIM_CLICK_TIMES = 6

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

/** 创建和收藏的歌单项保留归属信息，用于菜单展示。 */
type SidebarPlaylist = UserPlaylist

interface PlaylistSectionProps {
  title: string
  playlists: SidebarPlaylist[]
  showCreate?: boolean
}

function PlaylistSection({ title, playlists, showCreate = false }: PlaylistSectionProps) {
  const [collapsed, setCollapsed] = useState(false)
  const [menu, setMenu] = useState<{ playlist: SidebarPlaylist; position: MenuPosition } | null>(null)
  const [deleting, setDeleting] = useState<SidebarPlaylist | null>(null)
  const [preparing, setPreparing] = useState(false)
  const pending = useRef(false)
  const user = useAuthStore((state) => state.user)
  const navigate = useNavigate()
  const location = useLocation()
  const closeMenu = useCallback(() => setMenu(null), [])
  useEffect(() => { setMenu(null); setDeleting(null) }, [user])
  const view = (playlist: SidebarPlaylist) => navigate(`/songs/playlist/${encodeURIComponent(playlist.globalCollectionId)}`, { state: { playlist } })
  const play = async (playlist: SidebarPlaylist) => {
    if (pending.current) return
    pending.current = true
    setPreparing(true)
    try {
      const songs = await getAllPlaylistSongs(playlist.globalCollectionId)
      if (useAuthStore.getState().user !== user) return
      if (!songs.length) { toast('歌单暂无歌曲'); return }
      await usePlayerStore.getState().playAll(songs, { replaceQueue: true })
    } catch (error) { toast.error(playlistError(error, '歌单歌曲加载失败，请稍后重试')) }
    finally { pending.current = false; setPreparing(false) }
  }
  const remove = async (playlist: SidebarPlaylist) => {
    if (!user || useAuthStore.getState().user !== user) throw new Error('请先登录')
    await deleteUserPlaylist(playlist.globalCollectionId)
    if (useAuthStore.getState().user !== user) return
    useUserPlaylistStore.getState().changed(playlist.globalCollectionId)
    toast.success(playlist.type === 1 ? '已取消收藏' : '歌单已删除')
    const path = `/songs/playlist/${encodeURIComponent(playlist.globalCollectionId)}`
    if (location.pathname === path || location.pathname.startsWith(`${path}/`)) navigate('/')
  }

  return (
    <>
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
              onClick={() => {
                if (!user) { toast.error('请先登录'); return }
                useUserPlaylistStore.getState().openCreate()
              }}
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
          {playlists.map((pl) => {
            const className =
              'flex w-[calc(100%-1.5rem)] items-center gap-3 rounded-lg px-3 py-1 mx-3 text-left text-xs transition-colors'
            const content = (
              <>
                {pl.isDefault === 1 || pl.isDefault === 2 ? (
                  <div className="h-8 w-8 shrink-0 overflow-hidden rounded">
                    <DefaultPlaylistCover isDefault={pl.isDefault} />
                  </div>
                ) : pl.coverUrl ? (
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
              </>
            )

            if (!pl.globalCollectionId || pl.globalCollectionId === '0') {
              return (
                <button
                  key={pl.id}
                  type="button"
                  className={`${className} cursor-pointer text-[var(--color-text-secondary)] hover:bg-gray-200`}
                  onClick={() => toast.error('暂无法获取该歌单歌曲')}
                  onContextMenu={(event) => { event.preventDefault(); setMenu({ playlist: pl, position: { x: event.clientX, y: event.clientY } }) }}
                >
                  {content}
                </button>
              )
            }

            return (
              <NavLink
                key={pl.id}
                to={`/songs/playlist/${encodeURIComponent(pl.globalCollectionId)}`}
                state={{ playlist: pl }}
                onContextMenu={(event) => { event.preventDefault(); setMenu({ playlist: pl, position: { x: event.clientX, y: event.clientY } }) }}
                className={({ isActive }) => `${className} ${isActive
                  ? 'text-white bg-[var(--color-accent)]'
                  : 'text-[var(--color-text-secondary)] hover:bg-gray-200'
                }`}
              >
                {content}
              </NavLink>
            )
          })}
        </div>
      )}
    </div>
    {menu && <ContextMenu position={menu.position} onClose={closeMenu} actions={[
      { label: '查看歌单', icon: Eye, disabled: !menu.playlist.globalCollectionId || menu.playlist.globalCollectionId === '0', onSelect: () => view(menu.playlist) },
      { label: '播放', icon: Play, disabled: preparing || !menu.playlist.globalCollectionId || menu.playlist.globalCollectionId === '0', onSelect: () => void play(menu.playlist) },
      ...(menu.playlist.type === 0 && menu.playlist.isDefault === 0 ? [{ label: '编辑歌单信息', icon: SquarePen, disabled: !menu.playlist.globalCollectionId || menu.playlist.globalCollectionId === '0', onSelect: () => navigate(`/songs/playlist/${encodeURIComponent(menu.playlist.globalCollectionId)}/edit`) }] : []),
      ...(menu.playlist.type === 1 || (menu.playlist.type === 0 && menu.playlist.isDefault === 0) ? [{ label: menu.playlist.type === 1 ? '取消收藏' : '删除歌单', icon: menu.playlist.type === 1 ? BookmarkMinus : Trash, disabled: !menu.playlist.globalCollectionId || menu.playlist.globalCollectionId === '0', onSelect: () => setDeleting(menu.playlist) }] : []),
    ]} />}
    {deleting && <ConfirmDialog
      title={deleting.type === 1 ? '取消收藏歌单' : '删除歌单'}
      message={deleting.type === 1 ? `确定取消收藏「${deleting.name}」吗？` : `确定删除「${deleting.name}」吗？此操作无法撤销。`}
      confirmLabel={deleting.type === 1 ? '取消收藏' : '删除'}
      onConfirm={() => remove(deleting)} onClose={() => setDeleting(null)}
    />}
    </>
  )
}

interface SidebarProps {
  /** 当前登录用户；未登录时为 null，此时不请求用户歌单 */
  user: LoginResult['user'] | null
}

export function Sidebar({ user }: SidebarProps) {
  const userPlaylists = useUserPlaylistStore((state) => state.playlists)
  const playlistErrorMessage = useUserPlaylistStore((state) => state.error)
  const playlistsLoading = useUserPlaylistStore((state) => state.loading)
  // 未知类型仍可查看，但菜单和后端均不授予编辑权限。
  const createdPlaylists = userPlaylists.filter((p) => p.type !== 1)
  const collectedPlaylists = userPlaylists.filter((p) => p.type === 1)

  const { celebrate, cry } = useConfetti()
  // 请求进行中时忽略后续触发：连点只会在第 7 次触发一次，避免并发请求
  const claimingRef = useRef(false)

  // 隐藏入口：在 logo 区域连续点 7 次领取当天 VIP。
  // 首次从服务器领取成功时撒花；命中“今天已领取”记录或服务端重复领取时撒 ❄️。
  // 接口本身在后端做了“一天只真正请求一次”的风控，所以重复触发不会打到酷狗。
  // 未登录时整条链路关闭：不计数、不请求、也不放动画。
  const handleLogoClick = useClickStreak({
    enabled: Boolean(user),
    times: LOGO_CLAIM_CLICK_TIMES,
    onTrigger: () => {
      if (claimingRef.current) return
      claimingRef.current = true
      claimDailyVip()
        .then((result) => {
          if (result.claimed && result.errorCode === 0) {
            celebrate()
          } else {
            cry()
          }
        })
        .catch((error: unknown) => {
          console.error('[vip] 领取当日 VIP 失败', error)
          cry()
        })
        .finally(() => {
          claimingRef.current = false
        })
    },
  })
  return (
    <aside className={`flex h-full ${SIDEBAR_WIDTH_CLASS} flex-col overflow-hidden bg-gray-50`}>
      {/* Logo 区域与 TopBar 共用高度，独立于下方滚动区域。 */}
      <div
        className="flex h-[var(--layout-topbar-height)] shrink-0 items-center gap-2.5 bg-[var(--color-accent)] px-6 select-none"
        onClick={handleLogoClick}
      >
        <span className="flex items-center gap-2 text-3xl font-normal text-white">
          <img src={eighthNote} alt="" aria-hidden="true" className="h-7 w-7 shrink-0" />
          <span>lyrra</span>
        </span>
      </div>
      <SimpleBar className="min-h-0 flex-1 scrollbar">

        {/* 主导航 */}
        <nav className="flex flex-col gap-1 px-2 pt-3">
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
        {playlistErrorMessage && <button type="button" disabled={playlistsLoading} onClick={() => void useUserPlaylistStore.getState().load()} className="mx-5 cursor-pointer text-left text-xs text-red-500 disabled:opacity-50">歌单加载失败，点击重试</button>}
        {playlistsLoading && !userPlaylists.length && <p className="mx-5 text-xs text-[var(--color-text-secondary)]">正在加载歌单…</p>}
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
