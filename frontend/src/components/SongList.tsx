import { useCallback, useEffect, useLayoutEffect, useRef, useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { SongRow } from '@/components/SongRow'
import type { Song } from '@/types'

interface SongListProps {
  songs: Song[]
  /** 板块标题 */
  title: string
  /** 单列最小宽度（px），根据容器宽度计算一页展示多少列 */
  minColWidth?: number
  /** 固定行数 */
  rows?: number
  /** 一页总项数变化时回调，便于父级向后端请求对应数量的数据 */
  onPageSizeChange?: (count: number) => void
}

const GAP = 16 // gap-4

export function SongList({
  songs,
  title,
  minColWidth = 280,
  rows = 3,
  onPageSizeChange,
}: SongListProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [columns, setColumns] = useState(2)

  // 根据容器宽度计算列数
  const compute = useCallback(() => {
    const el = containerRef.current
    if (!el) return
    const width = el.clientWidth
    const cols = Math.max(2, Math.floor((width + GAP) / (minColWidth + GAP)))
    setColumns((prev) => (prev === cols ? prev : cols))
  }, [minColWidth])

  useLayoutEffect(() => {
    compute()
    const ro = new ResizeObserver(compute)
    const el = containerRef.current
    if (el) ro.observe(el)
    return () => ro.disconnect()
  }, [compute])

  // 通知父级当前一页最大项数（按最宽窗口算 = 列数 × 行数）
  const pageSize = columns * rows
  useEffect(() => {
    onPageSizeChange?.(pageSize)
  }, [pageSize, onPageSizeChange])

  if (songs.length === 0) return null

  // 只展示一页的数据，多出的丢弃
  const visible = songs.slice(0, pageSize)

  return (
    <section className="mb-10">
      <div className="mb-4 flex items-center justify-between">
        <a
          href="#"
          className="flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80"
        >
          {title}
          <ChevronRight size={20} />
        </a>
      </div>
      <div
        ref={containerRef}
        className="grid gap-4"
        style={{
          gridTemplateColumns: `repeat(${columns}, minmax(0, 1fr))`,
          gridTemplateRows: `repeat(${rows}, auto)`,
        }}
      >
        {visible.map((s) => (
          <SongRow key={s.id} song={s} />
        ))}
      </div>
    </section>
  )
}
