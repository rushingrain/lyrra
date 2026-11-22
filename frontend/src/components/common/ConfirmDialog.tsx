import { useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { Modal } from './Modal'
import { playlistError } from '@/services/userPlaylist'

export function ConfirmDialog({ title, message, confirmLabel, onConfirm, onClose }: {
  title: string
  message: string
  confirmLabel: string
  onConfirm: () => Promise<void>
  onClose: () => void
}) {
  const [busy, setBusy] = useState(false)
  const pending = useRef(false)
  const confirm = async () => {
    if (pending.current) return
    pending.current = true
    setBusy(true)
    try { await onConfirm(); onClose() }
    catch (error) { toast.error(playlistError(error, '操作失败，请稍后重试')) }
    finally { pending.current = false; setBusy(false) }
  }
  return <Modal title={title} onClose={onClose} busy={busy}>
    <p className="text-sm leading-relaxed text-[var(--color-text-secondary)]">{message}</p>
    <div className="mt-6 flex justify-end gap-3">
      <button type="button" disabled={busy} onClick={onClose} className="cursor-pointer rounded-md px-4 py-2 text-sm hover:bg-[var(--color-surface)] disabled:opacity-50">取消</button>
      <button type="button" disabled={busy} onClick={() => void confirm()} className="cursor-pointer rounded-md bg-[var(--color-accent)] px-4 py-2 text-sm text-white disabled:opacity-50">{busy ? '处理中…' : confirmLabel}</button>
    </div>
  </Modal>
}
