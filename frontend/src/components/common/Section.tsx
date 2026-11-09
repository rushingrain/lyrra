// src/components/common/Section.tsx
import { ChevronRight } from 'lucide-react'
import type { MouseEvent, ReactNode } from 'react'

interface SectionProps {
  title: string
  children: ReactNode
  clickable?: boolean
  /** 点击标题执行操作；用于页面内切换等不涉及路由跳转的交互 */
  onClick?: (event: MouseEvent<HTMLButtonElement>) => void
}

const TITLE_ACTION_CLASS =
  'flex cursor-pointer items-center border-0 bg-transparent p-0 text-left font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80'

export function Section({ title, children, clickable = true, onClick }: SectionProps) {
  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        {clickable && onClick ? (
          <button type="button" className={TITLE_ACTION_CLASS} onClick={onClick}>
            {title}
            <ChevronRight size={20} />
          </button>
        ) : clickable ? (
          <a
            href="#"
            className={TITLE_ACTION_CLASS}
            onClick={(e) => e.preventDefault()}
          >
            {title}
            <ChevronRight size={20} />
          </a>
        ) : (
          <h2 className="font-bold text-lg text-[var(--color-text-primary)]">
            {title}
          </h2>
        )}
      </div>
      {children}
    </div>
  )
}
