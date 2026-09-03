import type { ReactNode } from 'react'

interface TagProps {
  /** outline: 描边；solid: 实心 */
  variant?: 'outline' | 'solid'
  className?: string
  children: ReactNode
}

/** 通用小标签 */
export function Tag({ variant = 'outline', className = '', children }: TagProps) {
  return (
    <span
      className={`shrink-0 rounded-xs border px-0.5 text-[9px] leading-relaxed ${variant === 'solid'
        ? 'border-[var(--color-accent)] bg-[var(--color-accent)] text-[var(--color-on-accent)]'
        : 'border-[var(--color-accent)] text-[var(--color-accent)]'} ${className}`}
    >
      {children}
    </span>
  )
}
