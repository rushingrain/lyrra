import { useLayoutEffect, useRef, useState } from 'react'
import { ChevronRight } from 'lucide-react'
import { Playlist } from '@/components/common/Playlist'
import type { Playlist as PlaylistType } from '@/types'

interface PlaylistRecommendProps {
  playlists: PlaylistType[]
}

/**
 * 卡片「目标宽度」随容器宽度按比例缩放(clamp 思路):
 *  - 窄容器:目标小 → 列数少、卡片小
 *  - 宽容器:目标大 → 列数多、卡片大
 * 永远只渲染一行能放下的卡片,多余的不显示;
 * 卡片再通过 flex-1 均分剩余空间,占满整行并保留 gap 留白。
 */
const GAP = 20              // 与 className gap-5 保持同步
const MIN_CARD_WIDTH = 96  // 卡片目标宽度下限(窄窗口)
const MAX_CARD_WIDTH = 152 // 卡片目标宽度上限(宽窗口)
const CARD_RATIO = 0.12    // 卡片目标宽度约占容器宽度的 12%

export function PlayListGrid({ playlists }: PlaylistRecommendProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [visibleCount, setVisibleCount] = useState(0)

  useLayoutEffect(() => {
    const el = containerRef.current
    if (!el) return

    const update = () => {
      const width = el.clientWidth
      if (width === 0) return
      // 卡片目标宽度随容器缩放,实现「宽窗口卡片更大且列数更多」
      const targetCardWidth = Math.max(
        MIN_CARD_WIDTH,
        Math.min(MAX_CARD_WIDTH, width * CARD_RATIO)
      )
      // 单行可容纳的卡片数:N = floor((W + gap) / (cardW + gap))
      const n = Math.max(1, Math.floor((width + GAP) / (targetCardWidth + GAP)))
      setVisibleCount(n)
    }

    update()
    const ro = new ResizeObserver(update)
    ro.observe(el)
    return () => ro.disconnect()
    // 依赖 playlists.length:数据异步到达、容器 div 实际挂载后再装观察者
    // (length=0 时组件提前 return null,ref 为空,挂载那一次 effect 不生效)
  }, [playlists.length])

  if (playlists.length === 0) return null

  return (
    <div className="mb-10">
      

      <div ref={containerRef} className="flex gap-5 overflow-hidden">
        {playlists.slice(0, visibleCount).map((playlist) => (
          <Playlist key={playlist.id} playlist={playlist} className="flex-1" />
        ))}
      </div>
    </div>
  )
}
