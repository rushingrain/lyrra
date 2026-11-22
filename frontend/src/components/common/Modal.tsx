import { useEffect, useId, useRef } from 'react'
import type { ReactNode, RefObject } from 'react'
import { createPortal } from 'react-dom'
import { X } from 'lucide-react'

interface ModalProps {
  title: string
  onClose: () => void
  busy?: boolean
  initialFocusRef?: RefObject<HTMLElement | null>
  children: ReactNode
}

/** 原生 dialog 提供焦点锁定与关闭后的焦点恢复，portal 避免被滚动容器裁剪。 */
export function Modal({ title, onClose, busy = false, initialFocusRef, children }: ModalProps) {
  const ref = useRef<HTMLDialogElement>(null)
  const titleRef = useRef<HTMLHeadingElement>(null)
  const titleId = useId()
  useEffect(() => {
    const dialog = ref.current
    dialog?.showModal()
    // showModal 会重新分配焦点，需在打开后聚焦输入框或标题，避免选中关闭按钮。
    const initialFocus = initialFocusRef?.current ?? titleRef.current
    initialFocus?.focus({ preventScroll: true })
    return () => dialog?.close()
  }, [initialFocusRef])

  return createPortal(
    <dialog
      ref={ref}
      aria-labelledby={titleId}
      className="m-auto w-[420px] max-w-[calc(100vw-32px)] rounded-xl border border-[var(--color-border)] bg-[var(--color-background)] p-0 text-[var(--color-text-primary)] shadow-xl backdrop:bg-black/40"
      onCancel={(event) => { event.preventDefault(); if (!busy) onClose() }}
      onClick={(event) => { if (event.target === event.currentTarget && !busy) onClose() }}
    >
      <div className="p-6" onClick={(event) => event.stopPropagation()}>
        <div className="mb-5 flex items-center justify-between gap-4">
          <h2 ref={titleRef} id={titleId} tabIndex={-1} className="text-lg font-semibold outline-none">{title}</h2>
          <button type="button" aria-label="关闭" disabled={busy} onClick={onClose} className="cursor-pointer rounded p-1 hover:bg-[var(--color-surface)] disabled:opacity-50"><X size={18} /></button>
        </div>
        {children}
      </div>
    </dialog>, document.body,
  )
}
