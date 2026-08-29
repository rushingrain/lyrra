import { Fragment, useLayoutEffect, useState } from 'react'
import type { CSSProperties, Key, ReactNode } from 'react'

/**
 * 卡片宽度随容器宽度按比例缩放(clamp 思路定「目标宽度」,再算单行容量 N):
 *  - 窄容器:目标小 → 列数少、卡片小
 *  - 宽容器:目标大 → 列数多、卡片大
 * 单行最多渲染 N 张(N 为单行可容纳数),超出不显示。
 * 卡片实际宽度 = 单行 N 张均分整行时的单卡宽度,两种情况同尺寸:
 *  - 数据充足(length ≥ N):flex-1 均分占满整行
 *  - 数据较少(length < N):按实际展示数量均分定宽,行末留白
 */
const DEFAULT_GAP = 20
const DEFAULT_MIN_WIDTH = 120
const DEFAULT_MAX_WIDTH = 156
const DEFAULT_RATIO = 0.12

export interface SummaryItemLayout {
  className: string
  style?: CSSProperties
}

export interface SummaryLayout {
  /** 卡片间距，单位 px */
  gap?: number
  /** 卡片目标宽度下限，单位 px */
  minWidth?: number
  /** 卡片目标宽度上限，单位 px */
  maxWidth?: number
  /** 卡片目标宽度约占容器宽度的比例 */
  ratio?: number
}

interface SummaryProps<T> extends SummaryLayout {
  items: T[]
  className?: string
  getItemKey: (item: T) => Key
  renderItem: (item: T, layout: SummaryItemLayout) => ReactNode
}

export function Summary<T>({
  items,
  className,
  gap = DEFAULT_GAP,
  minWidth: minCardWidth = DEFAULT_MIN_WIDTH,
  maxWidth: maxCardWidth = DEFAULT_MAX_WIDTH,
  ratio: cardRatio = DEFAULT_RATIO,
  getItemKey,
  renderItem,
}: SummaryProps<T>) {
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

      // 计算目标宽度（受 min/max 约束）
      const target = Math.max(
        minCardWidth,
        Math.min(maxCardWidth, width * cardRatio),
      )
      // 计算单行可容纳数量
      const capacity = Math.max(
        1,
        Math.floor((width + gap) / (target + gap)),
      )

      setLayout({ containerWidth: width, capacity })
    }

    update()

    const observer = new ResizeObserver(update)
    observer.observe(container)

    return () => observer.disconnect()
  }, [container, gap, minCardWidth, maxCardWidth, cardRatio])

  if (items.length === 0) return null

  const { containerWidth, capacity } = layout
  const canFillRow = items.length >= capacity
  // 实际展示数量：能填满则填满，不能填满则展示所有
  const visibleCount = canFillRow ? capacity : items.length

  // 关键修复：用实际展示数量来计算均分宽度
  // 数据充足时用 capacity，数据不足时用 items.length
  const effectiveCount = canFillRow ? capacity : items.length
  const cardWidth =
    effectiveCount > 0 && containerWidth > 0
      ? Math.max(
        minCardWidth,
        Math.min(
          maxCardWidth,
          (containerWidth - (effectiveCount - 1) * gap) / effectiveCount,
        ),
      )
      : minCardWidth

  // 数据充足时用 flex-1 均分，数据不足时用固定宽度（留空白）
  const itemLayout: SummaryItemLayout = canFillRow
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
        style={{ gap }}
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