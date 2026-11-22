import { useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { ImagePlus, Music } from 'lucide-react'
import toast from 'react-hot-toast'
import { usePlaylistMembership } from '@/hooks/usePlaylistMembership'
import { playlistError, updateUserPlaylist } from '@/services/userPlaylist'
import { useAuthStore } from '@/stores/authStore'
import { useUserPlaylistStore } from '@/stores/userPlaylistStore'
import type { UserPlaylist } from '@/types'

const INPUT_CLASS = 'mt-2 w-full rounded-md border border-[var(--color-border)] bg-transparent px-3 py-2 text-sm outline-none focus:border-[var(--color-accent)] disabled:opacity-50'

async function readImage(file: File): Promise<string> {
  const url = URL.createObjectURL(file)
  try {
    const image = new Image()
    await new Promise<void>((resolve, reject) => {
      image.onload = () => resolve()
      image.onerror = () => reject(new Error('读取封面失败，请重新选择'))
      image.src = url
    })
    if (!image.naturalWidth || !image.naturalHeight) throw new Error('封面图片尺寸无效')
    // 上传服务要求宽高均至少 400 像素；只放大小图，保留比例和完整画面。
    const scale = Math.max(1, 400 / image.naturalWidth, 400 / image.naturalHeight)
    if (scale > 1) {
      const width = Math.ceil(image.naturalWidth * scale)
      const height = Math.ceil(image.naturalHeight * scale)
      if (width > 16384 || height > 16384 || width * height > 16_000_000) {
        throw new Error('图片比例过于狭长，请更换一张封面')
      }
      const canvas = document.createElement('canvas')
      canvas.width = width
      canvas.height = height
      const context = canvas.getContext('2d')
      if (!context) throw new Error('处理封面失败，请重新选择')
      context.imageSmoothingQuality = 'high'
      context.drawImage(image, 0, 0, width, height)
      const result = canvas.toDataURL(file.type, 0.92).split(',')[1]
      if (!result) throw new Error('处理封面失败，请重新选择')
      return result
    }
    return await new Promise((resolve, reject) => {
      const reader = new FileReader()
      reader.onload = () => resolve(String(reader.result).split(',')[1] ?? '')
      reader.onerror = () => reject(new Error('读取封面失败，请重新选择'))
      reader.readAsDataURL(file)
    })
  } finally {
    URL.revokeObjectURL(url)
  }
}

function EditForm({ playlist }: { playlist: UserPlaylist }) {
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const [name, setName] = useState(playlist.name)
  const [description, setDescription] = useState(playlist.description)
  const [tags, setTags] = useState((playlist.tags ?? []).join('，'))
  const [file, setFile] = useState<File | null>(null)
  const [preview, setPreview] = useState('')
  const [busy, setBusy] = useState(false)
  const pending = useRef(false)
  const fileInput = useRef<HTMLInputElement>(null)
  const path = `/songs/playlist/${encodeURIComponent(playlist.globalCollectionId)}`
  useEffect(() => {
    if (!file) { setPreview(''); return }
    const url = URL.createObjectURL(file)
    setPreview(url)
    return () => URL.revokeObjectURL(url)
  }, [file])

  const save = async (event: FormEvent) => {
    event.preventDefault()
    if (pending.current || !name.trim() || !user) return
    pending.current = true
    setBusy(true)
    try {
      const coverData = file ? await readImage(file) : ''
      if (useAuthStore.getState().user !== user) return
      await updateUserPlaylist(playlist.globalCollectionId, {
        name: name.trim(), description,
        tags: [...new Set(tags.split(/[,，]/).map((tag) => tag.trim()).filter(Boolean))],
        coverData,
      })
      if (useAuthStore.getState().user !== user) return
      useUserPlaylistStore.getState().changed(playlist.globalCollectionId)
      toast.success('歌单信息已保存')
      navigate(path, { replace: true })
    } catch (error) {
      if (useAuthStore.getState().user === user) toast.error(playlistError(error, '保存歌单失败'))
    } finally { pending.current = false; setBusy(false) }
  }
  const cover = preview || playlist.coverUrl
  return <form onSubmit={(event) => void save(event)} className="max-w-[900px]">
    <h1 className="mb-6 text-2xl font-semibold text-[var(--color-text-primary)]">编辑歌单信息</h1>
    <div className="flex flex-col gap-10 md:flex-row">
      <div className="min-w-0 flex-1 space-y-5">
        <div><label htmlFor="playlist-name" className="text-[15px]">名称：</label><input id="playlist-name" required value={name} onChange={(event) => setName(event.target.value)} disabled={busy} className={INPUT_CLASS} /></div>
        <div><label htmlFor="playlist-description" className="text-[15px]">简介：</label><textarea id="playlist-description" rows={5} value={description} onChange={(event) => setDescription(event.target.value)} disabled={busy} className={`${INPUT_CLASS} resize-y`} placeholder="介绍一下这张歌单" /></div>
        <div><label htmlFor="playlist-tags" className="text-[15px]">标签：</label><input id="playlist-tags" value={tags} onChange={(event) => setTags(event.target.value)} disabled={busy} className={INPUT_CLASS} placeholder="用逗号分隔，例如：华语，流行" /></div>
      </div>
      <div className="w-44 shrink-0">
        <div className="flex h-44 w-44 items-center justify-center overflow-hidden rounded-lg bg-[var(--color-surface)]">
          {cover ? <img src={cover} alt="歌单封面预览" className="h-full w-full object-cover" /> : <Music size={40} className="text-[var(--color-text-secondary)]" />}
        </div>
        <input ref={fileInput} type="file" accept="image/jpeg,image/png" className="hidden" disabled={busy} onChange={(event) => {
          const selected = event.target.files?.[0]
          event.target.value = ''
          if (!selected) return
          if (!['image/jpeg', 'image/png'].includes(selected.type)) { toast.error('请选择 JPG 或 PNG 图片'); return }
          if (selected.size === 0 || selected.size > 10 * 1024 * 1024) { toast.error('请选择不超过 10 MB 的图片'); return }
          setFile(selected)
        }} />
        <button type="button" disabled={busy} onClick={() => fileInput.current?.click()} className="mt-3 flex w-full cursor-pointer items-center justify-center gap-2 rounded-md border border-[var(--color-border)] px-3 py-2 text-sm hover:bg-[var(--color-surface)] disabled:opacity-50"><ImagePlus size={16} />选择封面</button>
        <p className="mt-2 text-xs text-[var(--color-text-secondary)]">JPG / PNG，不超过 10 MB，小尺寸图片会自动等比例放大</p>
        {file && <button type="button" disabled={busy} onClick={() => setFile(null)} className="mt-2 cursor-pointer text-xs text-[var(--color-accent)]">撤销封面更换</button>}
      </div>
    </div>
    <div className="mt-8 flex gap-3">
      <button type="submit" disabled={busy || !name.trim()} className="cursor-pointer rounded-md bg-[var(--color-accent)] px-6 py-2 text-sm text-white disabled:opacity-50">{busy ? '保存中…' : '保存'}</button>
      <button type="button" disabled={busy} onClick={() => navigate(path)} className="cursor-pointer rounded-md border border-[var(--color-border)] px-6 py-2 text-sm hover:bg-[var(--color-surface)] disabled:opacity-50">取消</button>
    </div>
  </form>
}

export default function PlaylistEdit() {
  const { globalId = '' } = useParams<{ globalId: string }>()
  const user = useAuthStore((state) => state.user)
  // 编辑时保留表单草稿，避免从播放队列收藏歌曲后的列表刷新清空输入。
  const { playlist, loading, error, retry } = usePlaylistMembership(globalId, false)
  if (!user) return <p className="py-8 text-sm text-[var(--color-text-secondary)]">请先登录后编辑歌单</p>
  if (loading) return <p className="py-8 text-sm text-[var(--color-text-secondary)]">正在加载歌单…</p>
  if (error) return <div className="py-8 text-sm text-[var(--color-text-secondary)]">{error}<button type="button" onClick={retry} className="ml-3 cursor-pointer text-[var(--color-accent)]">重试</button></div>
  if (!playlist || playlist.type !== 0 || playlist.isDefault !== 0) return <p className="py-8 text-sm text-[var(--color-text-secondary)]">仅支持编辑自己创建的普通歌单</p>
  return <EditForm key={`${user.sessionExpiresAt}:${globalId}`} playlist={playlist} />
}
