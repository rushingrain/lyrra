import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { PlaylistCard } from '@/components/PlaylistCard'
import type { Playlist } from '@/types'

interface PlaylistRecommendProps {
  playlists: Playlist[]
  /** 单个歌单项的最小宽度（px），用于根据容器宽度计算一页展示多少项 */
  minCardWidth?: number
  /** 一页项数变化时回调，便于父级向后端请求对应数量的数据 */
  onPageSizeChange?: (count: number) => void
}

const GAP = 20 // gap-5 = 1.25rem

export function PlaylistRecommend({
  playlists,
  minCardWidth = 180,
  onPageSizeChange,
}: PlaylistRecommendProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [pageSize, setPageSize] = useState(1)

  // 根据容器宽度计算一页能容纳多少项：
  // N = floor((容器宽度 + GAP) / (minCardWidth + GAP))，每项实际宽度由 flex-1 平分
  useLayoutEffect(() => {
    const el = containerRef.current
    if (!el) return
    const compute = () => {
      const width = el.clientWidth
      const count = Math.max(1, Math.floor((width + GAP) / (minCardWidth + GAP)))
      setPageSize((prev) => (prev === count ? prev : count))
    }
    // 同步计算一次，避免首帧闪现错误数量；再监听后续尺寸变化
    compute()
    const ro = new ResizeObserver(compute)
    ro.observe(el)
    return () => ro.disconnect()
  }, [minCardWidth, playlists.length])

  // 通知父级当前一页的项数，便于向后端请求对应数量的歌单
  useEffect(() => {
    onPageSizeChange?.(pageSize)
  }, [pageSize, onPageSizeChange])

  if (playlists.length === 0) return null

  // 一页只展示 pageSize 项
  const visible = playlists.slice(0, pageSize)

  return (
    <section className="mb-10">
      <div className="mb-4 flex items-center justify-between">
        <a
          href="#"
          className="flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80"
        >
          官方歌单
          <ChevronRight size={20} />
        </a>
      </div>
      <div ref={containerRef} className="flex items-start gap-5">
        {visible.map((p) => (
          <PlaylistCard key={p.id} playlist={p} />
        ))}
      </div>
    </section>
  )
}
