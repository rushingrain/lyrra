import { useEffect, useRef, useState } from 'react'
import { useLocation, useNavigate, useParams } from 'react-router-dom'
import { BookmarkCheck, BookmarkPlus, SquarePen } from 'lucide-react'
import toast from 'react-hot-toast'
import { SongCollectionDetail } from '@/components/business/SongCollectionDetail'
import { OwnedPlaylistContext } from '@/components/business/SongActions'
import { DefaultPlaylistCover } from '@/components/common/DefaultPlaylistCover'
import { ConfirmDialog } from '@/components/common/ConfirmDialog'
import { usePlaylistSongs } from '@/hooks/usePlaylistSongs'
import { usePlaylistMembership } from '@/hooks/usePlaylistMembership'
import { getAllPlaylistSongs } from '@/services/playlist'
import { collectPlaylist, deleteUserPlaylist, playlistError } from '@/services/userPlaylist'
import { useAuthStore } from '@/stores/authStore'
import { useUserPlaylistStore } from '@/stores/userPlaylistStore'
import type { LoginResult, Playlist as PlaylistType } from '@/types'

/** 歌单详情共用歌曲布局，编辑与移除权限只来自当前账号的后端查询。 */
export default function Playlist() {
  const { globalId = '' } = useParams<{ globalId: string }>()
  const location = useLocation()
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const revision = useUserPlaylistStore((state) => state.revisions[globalId] ?? 0)
  const membership = usePlaylistMembership(globalId)
  const own = membership.playlist
  const [collecting, setCollecting] = useState(false)
  const [uncollecting, setUncollecting] = useState<{ globalId: string; user: LoginResult['user'] } | null>(null)
  const pending = useRef(false)
  const statePlaylist = (location.state as {
    playlist?: Partial<PlaylistType> & { isDefault?: number }
  } | null)?.playlist
  const songCollectionId = own?.globalCollectionId || statePlaylist?.globalCollectionId || globalId
  const { info, songs, hasMore, loadError, loadMoreError, sentinelRef, loadMore } =
    usePlaylistSongs(songCollectionId, `${user?.sessionExpiresAt ?? 0}:${revision}`)

  const base = info?.name ? info : statePlaylist
  const playlist = own?.type === 0 && own.isDefault === 0
    ? { ...base, name: own.name, coverUrl: own.coverUrl, description: own.description, tags: own.tags }
    : base
  const isDefault = own?.isDefault ?? statePlaylist?.isDefault
  const editable = own?.type === 0 && own.isDefault === 0

  useEffect(() => {
    if (own?.type === 1 && own.globalCollectionId && own.globalCollectionId !== globalId) {
      navigate(`/songs/playlist/${encodeURIComponent(own.globalCollectionId)}`, { replace: true, state: { playlist: own } })
    }
  }, [own, globalId, navigate])

  const collect = async () => {
    if (!user) { toast.error('请先登录'); return }
    if (pending.current || membership.loading || membership.error) return
    pending.current = true
    setCollecting(true)
    try {
      await collectPlaylist(globalId)
      if (useAuthStore.getState().user !== user) return
      useUserPlaylistStore.getState().changed(globalId)
      toast.success('歌单已收藏')
    } catch (error) {
      if (useAuthStore.getState().user === user) toast.error(playlistError(error, '收藏歌单失败'))
    } finally { pending.current = false; setCollecting(false) }
  }
  const uncollect = async () => {
    if (!uncollecting || !user || useAuthStore.getState().user !== uncollecting.user) throw new Error('请先登录')
    await deleteUserPlaylist(uncollecting.globalId)
    if (useAuthStore.getState().user !== user) return
    useUserPlaylistStore.getState().changed(uncollecting.globalId)
    toast.success('已取消收藏')
  }

  if (loadError) return <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">{loadError}</p>

  return <>
    {membership.error && <button type="button" onClick={membership.retry} className="mb-4 cursor-pointer text-sm text-[var(--color-text-secondary)]">歌单操作暂不可用，点击重试</button>}
    <OwnedPlaylistContext.Provider value={own?.type === 0 ? { globalId, name: own.name } : null}>
      <SongCollectionDetail
        cover={isDefault === 1 || isDefault === 2 ? <DefaultPlaylistCover isDefault={isDefault} /> : undefined}
        coverUrl={playlist?.coverUrl}
        title={playlist?.name}
        titleAction={editable ? <button type="button" title="编辑歌单信息" aria-label="编辑歌单信息" onClick={() => navigate(`/songs/playlist/${encodeURIComponent(globalId)}/edit`)} className="shrink-0 cursor-pointer rounded p-1 text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]"><SquarePen size={16} /></button> : undefined}
        actions={!membership.loading && !membership.error && (!own || own.type === 1) ? <button
          type="button" title={own?.type === 1 ? '取消收藏歌单' : '收藏歌单'} aria-label={own?.type === 1 ? '取消收藏歌单' : '收藏歌单'}
          disabled={collecting || membership.loading || Boolean(membership.error)}
          onClick={() => { if (own?.type === 1 && user) setUncollecting({ globalId, user }); else void collect() }}
          className="flex cursor-pointer items-center justify-center rounded-md border border-[var(--color-border)] p-2 text-[var(--color-accent)] hover:bg-[var(--color-surface)] disabled:cursor-not-allowed disabled:opacity-50"
        >{own?.type === 1 ? <BookmarkCheck size={18} /> : <BookmarkPlus size={18} />}</button> : undefined}
        description={playlist?.description}
        author={playlist?.creator}
        authorAvatar={playlist?.creatorAvatar}
        tags={playlist?.tags}
        publishTime={playlist?.publishTime}
        publishLabel="创建"
        playCount={statePlaylist?.playCount ?? 0}
        songs={songs}
        hasMore={hasMore}
        loadMoreError={loadMoreError}
        sentinelRef={sentinelRef}
        loadMore={loadMore}
        loadAllSongs={() => getAllPlaylistSongs(songCollectionId)}
        loadAllErrorMessage="歌单歌曲加载失败，请稍后重试"
      />
    </OwnedPlaylistContext.Provider>
    {uncollecting?.globalId === globalId && uncollecting?.user === user && own?.type === 1 && user && <ConfirmDialog title="取消收藏歌单" message={`确定取消收藏「${own.name}」吗？`} confirmLabel="取消收藏" onConfirm={uncollect} onClose={() => setUncollecting(null)} />}
  </>
}
