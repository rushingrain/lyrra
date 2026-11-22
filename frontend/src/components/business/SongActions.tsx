import { createContext, useCallback, useContext, useState } from 'react'
import { MoreHorizontal, Play, SquarePlus, Trash } from 'lucide-react'
import toast from 'react-hot-toast'
import { ContextMenu, type MenuPosition } from '@/components/common/ContextMenu'
import { ConfirmDialog } from '@/components/common/ConfirmDialog'
import { removePlaylistSong } from '@/services/userPlaylist'
import { useAuthStore } from '@/stores/authStore'
import { usePlayerStore } from '@/stores/playerStore'
import { useUserPlaylistStore } from '@/stores/userPlaylistStore'
import type { LoginResult, Song } from '@/types'

/** 只在当前账号创建的歌单详情内提供移除入口，专辑和搜索等列表保持通用操作。 */
export const OwnedPlaylistContext = createContext<{ globalId: string; name: string } | null>(null)

export function SongActions({ song, showCollectAction = true }: { song: Song; showCollectAction?: boolean }) {
  const owner = useContext(OwnedPlaylistContext)
  const [position, setPosition] = useState<MenuPosition | null>(null)
  const [removal, setRemoval] = useState<{ globalId: string; fileId: string; user: LoginResult['user'] } | null>(null)
  const user = useAuthStore((state) => state.user)
  const close = useCallback(() => setPosition(null), [])
  const collect = () => {
    if (!useAuthStore.getState().user) { toast.error('请先登录'); return }
    useUserPlaylistStore.getState().collectSong(song)
  }
  const remove = async () => {
    if (!removal || !user || useAuthStore.getState().user !== removal.user) throw new Error('登录状态已变化，请重新操作')
    await removePlaylistSong(removal.globalId, removal.fileId)
    if (useAuthStore.getState().user !== user) return
    useUserPlaylistStore.getState().changed(removal.globalId)
    toast.success('歌曲已从歌单移除')
  }
  return <div className="flex shrink-0 items-center gap-1" onClick={(event) => event.stopPropagation()}>
    {showCollectAction && <button type="button" aria-label={`收藏到歌单：${song.name}`} title="收藏到歌单" onClick={collect} className="cursor-pointer rounded p-1 text-[var(--color-text-secondary)] hover:text-[var(--color-accent)]"><SquarePlus size={17} /></button>}
    <button type="button" aria-label={`更多：${song.name}`} title="更多" aria-haspopup="menu" aria-expanded={Boolean(position)} onPointerDown={(event) => event.stopPropagation()} onClick={(event) => {
      if (position) { close(); return }
      const rect = event.currentTarget.getBoundingClientRect()
      setPosition({ x: rect.left, y: rect.bottom + 4 })
    }} className="cursor-pointer rounded p-1 text-[var(--color-text-secondary)] hover:text-[var(--color-accent)]"><MoreHorizontal size={17} /></button>
    {position && <ContextMenu position={position} onClose={close} actions={[
      { label: '播放', icon: Play, onSelect: () => { void usePlayerStore.getState().playSong(song) } },
      { label: '收藏', icon: SquarePlus, onSelect: collect },
      ...(owner && user ? [{
        label: '从歌单中移除', icon: Trash, onSelect: () => {
          if (!song.fileId || song.fileId === '0') { toast.error('该歌曲缺少歌单条目 ID，暂时无法移除'); return }
          setRemoval({ globalId: owner.globalId, fileId: song.fileId, user })
        }
      }] : []),
    ]} />}
    {removal && owner && removal.globalId === owner.globalId && removal.fileId === song.fileId && removal.user === user && <ConfirmDialog
      title="从歌单中移除" message={`确定将「${song.name}」从「${owner.name}」移除吗？`}
      confirmLabel="移除" onConfirm={remove} onClose={() => setRemoval(null)}
    />}
  </div>
}
