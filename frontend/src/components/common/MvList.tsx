import type { Mv as MvType } from '@/types'
import { SingleRowList } from '@/components/common/SingleRowList'
import type { SingleRowListLayout } from '@/components/common/SingleRowList'

interface MvListProps extends SingleRowListLayout {
  mvs: MvType[]
}

/**
 * MV 列表:负责 MV 业务语义, 单行响应式布局由 SingleRowList 复用。
 */
export function MvList({ mvs, ...layout }: MvListProps) {
  return (
    <SingleRowList
      {...layout}
      items={mvs}
      getItemKey={(mv) => mv.id}
      renderItem={(mv, { className, style }) => (
        <div className={`min-w-0 ${className}`} style={style}>
          <img src={mv.coverUrl} alt={mv.name} className="aspect-[3/2] w-full rounded-lg object-cover" loading="lazy" />
          <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{mv.name}</p>
          <p className="mt-0.5 truncate text-xs text-[var(--color-text-secondary)]">{mv.artist}</p>
        </div>
      )}
    />
  )
}
