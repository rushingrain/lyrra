import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { Music } from 'lucide-react'
import toast from 'react-hot-toast'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'
import { Modal } from '@/components/common/Modal'
import { DefaultPlaylistCover } from '@/components/common/DefaultPlaylistCover'
import { useAuthStore } from '@/stores/authStore'
import { useUserPlaylistStore } from '@/stores/userPlaylistStore'
import { addPlaylistSong, createUserPlaylist, playlistError } from '@/services/userPlaylist'
import type { Song, UserPlaylist } from '@/types'

function CreatePlaylistModal() {
  const nameInputRef = useRef<HTMLInputElement>(null)
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)
  const pending = useRef(false)
  const close = useUserPlaylistStore((state) => state.closeCreate)
  const user = useAuthStore((state) => state.user)
  const save = async (event: FormEvent) => {
    event.preventDefault()
    if (!name.trim() || pending.current || !user) return
    pending.current = true
    setBusy(true)
    try {
      await createUserPlaylist(name.trim())
      if (useAuthStore.getState().user !== user) return
      useUserPlaylistStore.getState().changed()
      toast.success('歌单已创建')
      close()
    } catch (error) {
      if (useAuthStore.getState().user === user) toast.error(playlistError(error, '创建歌单失败'))
    } finally { pending.current = false; setBusy(false) }
  }
  return <Modal title="创建歌单" onClose={close} busy={busy} initialFocusRef={nameInputRef}>
    <form onSubmit={(event) => void save(event)}>
      <label className="block text-sm" htmlFor="new-playlist-name">歌单标题</label>
      <input ref={nameInputRef} id="new-playlist-name" value={name} disabled={busy} onChange={(event) => setName(event.target.value)} placeholder="请输入歌单标题" className="mt-2 w-full rounded-md border border-[var(--color-border)] bg-transparent px-3 py-2 text-sm outline-none focus:border-[var(--color-accent)]" />
      <div className="mt-6 flex justify-end gap-3">
        <button type="button" disabled={busy} onClick={close} className="cursor-pointer rounded-md px-4 py-2 text-sm hover:bg-[var(--color-surface)] disabled:opacity-50">取消</button>
        <button type="submit" disabled={busy || !name.trim()} className="cursor-pointer rounded-md bg-[var(--color-accent)] px-4 py-2 text-sm text-white disabled:opacity-50">{busy ? '保存中…' : '保存'}</button>
      </div>
    </form>
  </Modal>
}

function CollectSongModal({ song }: { song: Song }) {
  const playlists = useUserPlaylistStore((state) => state.playlists)
  const loading = useUserPlaylistStore((state) => state.loading)
  const error = useUserPlaylistStore((state) => state.error)
  const close = () => useUserPlaylistStore.getState().collectSong(null)
  const user = useAuthStore((state) => state.user)
  const [busy, setBusy] = useState(false)
  const pending = useRef(false)
  const created = playlists.filter((playlist) => playlist.type === 0 && playlist.globalCollectionId && playlist.globalCollectionId !== '0')
  const select = async (playlist: UserPlaylist) => {
    if (pending.current || !user) return
    pending.current = true
    setBusy(true)
    try {
      await addPlaylistSong(playlist.globalCollectionId, song)
      if (useAuthStore.getState().user !== user) return
      useUserPlaylistStore.getState().changed(playlist.globalCollectionId)
      toast.success(`已添加到「${playlist.name}」`)
      close()
    } catch (error) {
      if (useAuthStore.getState().user === user) toast.error(playlistError(error, '添加歌曲失败'))
    } finally { pending.current = false; setBusy(false) }
  }
  return <Modal title="收藏到歌单" onClose={close} busy={busy}>
    <p className="mb-3 truncate text-sm text-[var(--color-text-secondary)]">{song.name}</p>
    {error && <div className="mb-3 text-sm text-red-500">{error}<button type="button" disabled={busy || loading} onClick={() => void useUserPlaylistStore.getState().load()} className="ml-2 cursor-pointer text-[var(--color-accent)]">重试</button></div>}
    <SimpleBar className="max-h-[45vh] scrollbar">
      {created.map((playlist) => <button key={playlist.id} type="button" disabled={busy} onClick={() => void select(playlist)} className="flex w-full cursor-pointer items-center gap-3 rounded-lg px-2 py-2 text-left hover:bg-[var(--color-surface)] disabled:opacity-50">
        <div className="h-10 w-10 shrink-0 overflow-hidden rounded bg-[var(--color-surface)]">
          {playlist.isDefault === 1 || playlist.isDefault === 2 ? <DefaultPlaylistCover isDefault={playlist.isDefault} /> : playlist.coverUrl ? <img src={playlist.coverUrl} alt="" className="h-full w-full object-cover" /> : <Music className="m-auto mt-3 h-4 w-4 text-[var(--color-text-secondary)]" />}
        </div>
        <span className="min-w-0 flex-1"><span className="block truncate text-sm">{playlist.name}</span><span className="text-xs text-[var(--color-text-secondary)]">{playlist.songCount} 首</span></span>
      </button>)}
      {!created.length && <p className="py-6 text-center text-sm text-[var(--color-text-secondary)]">{loading ? '正在加载歌单…' : error ? '歌单暂不可用' : '暂无可用歌单，请先在侧边栏新建歌单'}</p>}
    </SimpleBar>
    {busy && <p role="status" className="mt-3 text-center text-sm text-[var(--color-text-secondary)]">正在添加…</p>}
  </Modal>
}

/** 窗口内只挂载一份歌单弹窗；账号切换时同时清理数据和未完成的选择。 */
export function UserPlaylistDialogs() {
  const user = useAuthStore((state) => state.user)
  const createOpen = useUserPlaylistStore((state) => state.createOpen)
  const song = useUserPlaylistStore((state) => state.songToCollect)
  useEffect(() => {
    const store = useUserPlaylistStore.getState()
    store.reset()
    if (user) void store.load()
    return () => useUserPlaylistStore.getState().reset()
  }, [user])
  if (!user) return null
  return <>
    {createOpen && <CreatePlaylistModal />}
    {song && <CollectSongModal key={`${song.id}:${song.hash}`} song={song} />}
  </>
}
