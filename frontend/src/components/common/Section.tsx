// src/components/common/Section.tsx
import { ChevronRight } from 'lucide-react'
import { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface SectionProps {
  title: string
  children: ReactNode
  clickable?: boolean
  /** 点击标题跳转的路由；不传时标题只是可点击的占位（目标页面尚未实现） */
  to?: string
}

const TITLE_LINK_CLASS =
  'flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80'

export function Section({ title, children, clickable = true, to }: SectionProps) {
  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        {clickable && to ? (
          <Link to={to} className={TITLE_LINK_CLASS}>
            {title}
            <ChevronRight size={20} />
          </Link>
        ) : clickable ? (
          <a
            href="#"
            className={TITLE_LINK_CLASS}
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
