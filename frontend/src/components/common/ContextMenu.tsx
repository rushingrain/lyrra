import { useLayoutEffect, useRef } from 'react'
import { createPortal } from 'react-dom'
import type { LucideIcon } from 'lucide-react'

export interface MenuAction {
  label: string
  icon?: LucideIcon
  onSelect: () => void
  disabled?: boolean
}

export interface MenuPosition { x: number; y: number }

/** 侧边栏右键菜单与歌曲更多菜单共用，支持窗口边缘避让和键盘选择。 */
export function ContextMenu({ position, actions, onClose }: {
  position: MenuPosition
  actions: MenuAction[]
  onClose: () => void
}) {
  const ref = useRef<HTMLDivElement>(null)
  useLayoutEffect(() => {
    const menu = ref.current
    if (!menu) return
    const previous = document.activeElement as HTMLElement | null
    const rect = menu.getBoundingClientRect()
    menu.style.left = `${Math.max(8, Math.min(position.x, window.innerWidth - rect.width - 8))}px`
    menu.style.top = `${Math.max(8, Math.min(position.y, window.innerHeight - rect.height - 8))}px`
    menu.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus()
    const outside = (event: PointerEvent) => { if (!menu.contains(event.target as Node)) onClose() }
    const close = () => onClose()
    document.addEventListener('pointerdown', outside)
    window.addEventListener('resize', close)
    window.addEventListener('scroll', close, true)
    return () => {
      document.removeEventListener('pointerdown', outside)
      window.removeEventListener('resize', close)
      window.removeEventListener('scroll', close, true)
      if (menu.contains(document.activeElement)) previous?.focus()
    }
  }, [position.x, position.y, onClose])

  return createPortal(
    <div
      ref={ref} role="menu" aria-label="操作菜单"
      className="fixed z-[90] min-w-40 rounded-lg bg-[var(--color-background)] p-1 shadow-[0_4px_12px_rgba(0,0,0,0.08),0_12px_32px_rgba(0,0,0,0.12)]"
      style={{ left: position.x, top: position.y }}
      onContextMenu={(event) => event.preventDefault()}
      onKeyDown={(event) => {
        if (event.key === 'Escape' || event.key === 'Tab') { event.preventDefault(); onClose(); return }
        if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
        event.preventDefault()
        const buttons = [...(ref.current?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)') ?? [])]
        const current = buttons.indexOf(document.activeElement as HTMLButtonElement)
        const next = event.key === 'Home' ? 0 : event.key === 'End' ? buttons.length - 1 : (current + (event.key === 'ArrowDown' ? 1 : -1) + buttons.length) % buttons.length
        buttons[next]?.focus()
      }}
    >
      {actions.map(({ icon: Icon, ...action }) => <button
        key={action.label} type="button" role="menuitem" disabled={action.disabled}
        onClick={() => { onClose(); action.onSelect() }}
        className="flex w-full cursor-pointer items-center gap-2.5 rounded px-3 py-2 text-left text-sm font-normal text-[var(--color-text-secondary)] outline-none hover:bg-[color-mix(in_srgb,var(--color-text-secondary)_4%,transparent)] focus-visible:outline-1 focus-visible:outline-[var(--color-border)] disabled:cursor-not-allowed disabled:opacity-40"
      >
        {Icon && <Icon size={16} className="shrink-0" aria-hidden="true" />}
        <span>{action.label}</span>
      </button>)}
    </div>, document.body,
  )
}
