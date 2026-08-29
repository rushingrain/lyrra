// src/components/common/Section.tsx
import { ChevronRight } from 'lucide-react'
import { ReactNode } from 'react'

interface SectionProps {
  title: string
  children: ReactNode
  clickable?: boolean
}

export function Section({ title, children, clickable = true }: SectionProps) {
  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        {clickable ? (
          <a
            href="#"
            className="flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80"
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