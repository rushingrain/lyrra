import { Fragment, useLayoutEffect, useState } from 'react'
import type { CSSProperties, Key, ReactNode } from 'react'

/**
 * 卡片宽度随容器宽度按比例缩放(clamp 思路定「目标宽度」,再算单行容量 N):
 *  - 窄容器:目标小 → 列数少、卡片小
 *  - 宽容器:目标大 → 列数多、卡片大
 * 单行最多渲染 N 张(N 为单行可容纳数),超出不显示。
 * 卡片实际宽度 = 单行 N 张均分整行时的单卡宽度,两种情况同尺寸:
 *  - 数据充足(length ≥ N):flex-1 均分占满整行
 *  - 数据较少(length < N):按 fillCardWidth 定宽,不再拉伸,行末留白
 */
const GAP = 20              // 与下方 flex gap 内联样式保持同步
const MIN_CARD_WIDTH = 120  // 卡片目标宽度下限(窄窗口) 96
const MAX_CARD_WIDTH = 156 // 卡片目标宽度上限(宽窗口) 152
const CARD_RATIO = 0.12    // 卡片目标宽度约占容器宽度的 12%

export interface SingleRowCardItemLayout {
  className: string
  style?: CSSProperties
}

interface SingleRowCardListProps<T> {
  items: T[]
  className?: string
  getItemKey: (item: T) => Key
  renderItem: (item: T, layout: SingleRowCardItemLayout) => ReactNode
}

export function SingleRowList<T>({
  items,
  className,
  getItemKey,
  renderItem,
}: SingleRowCardListProps<T>) {
  const [container, setContainer] = useState<HTMLDivElement | null>(null)
  const [layout, setLayout] = useState({
    containerWidth: 0,
    capacity: 0,
  })

  useLayoutEffect(() => {
    if (!container) return

    const update = () => {
      const width = container.clientWidth
      if (width === 0) return

      const target = Math.max(
        MIN_CARD_WIDTH,
        Math.min(MAX_CARD_WIDTH, width * CARD_RATIO),
      )
      const capacity = Math.max(
        1,
        Math.floor((width + GAP) / (target + GAP)),
      )

      setLayout({ containerWidth: width, capacity })
    }

    update()

    const observer = new ResizeObserver(update)
    observer.observe(container)

    return () => observer.disconnect()
  }, [container, items.length])

  if (items.length === 0) return null

  const { containerWidth, capacity } = layout
  const canFillRow = items.length >= capacity
  const visibleCount = canFillRow ? capacity : items.length
  const cardWidth =
    capacity > 0
      ? (containerWidth - (capacity - 1) * GAP) / capacity
      : MIN_CARD_WIDTH

  const itemLayout: SingleRowCardItemLayout = canFillRow
    ? { className: 'flex-1' }
    : {
      className: 'flex-none',
      style: { width: cardWidth },
    }

  return (
    <div className={className}>
      <div
        ref={setContainer}
        className="flex overflow-hidden"
        style={{ gap: GAP }}
      >
        {items.slice(0, visibleCount).map((item) => (
          <Fragment key={getItemKey(item)}>
            {renderItem(item, itemLayout)}
          </Fragment>
        ))}
      </div>
    </div>
  )
}
