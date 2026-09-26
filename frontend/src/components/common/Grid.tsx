import type { CSSProperties, Key } from 'react'

/**
 * 封面卡片网格：多行布局，一行装不下自动换行。
 * 列最小宽度随容器宽度按 ratio 比例缩放（clamp 限定上下限），列数基本恒定，
 * 卡片随窗口增大而增大，不出现「窗口大卡片反而小」的现象。
 * clamp 数值来自运行时 props，Tailwind 无法生成动态类名，
 * 故通过 CSS 变量注入，由静态工具类 grid-cols-[...var(--card-w)...] 引用。
 */
interface GridProps<T> {
  items: T[]
  getItemKey: (item: T) => Key
  /** 封面图地址与描述，图片兜底逻辑由调用方处理 */
  getCover: (item: T) => { src: string; alt: string }
  /** 封面宽高比（CSS aspect-ratio 值，如 '1 / 1'、'3 / 2'），与卡片宽度共同决定高度 */
  aspect: string
  /** 封面圆角等附加样式，如 'rounded-full'、'rounded-lg' */
  coverClassName?: string
  /** 卡片宽度下限（px），窄窗口时生效 */
  minWidth: number
  /** 卡片宽度上限（px），宽窗口时生效 */
  maxWidth: number
  /** 卡片目标宽度占容器宽度的比例（如 0.12 表示 12%），决定随窗口缩放的幅度 */
  ratio: number
  /** 封面下方文案：主标题与可选副标题 */
  getInfo: (item: T) => { title: string; subtitle?: string }
  /** 文案居中显示 */
  center?: boolean
}

export function Grid<T>({
  items,
  getItemKey,
  getCover,
  aspect,
  coverClassName,
  minWidth,
  maxWidth,
  ratio,
  getInfo,
  center,
}: GridProps<T>) {
  if (items.length === 0) return null

  return (
    <div className="@container">
      <div
        className="grid grid-cols-[repeat(auto-fill,minmax(var(--card-w),1fr))] gap-5"
        style={{ '--card-w': `clamp(${minWidth}px, ${ratio * 100}cqi, ${maxWidth}px)` } as CSSProperties}
      >
        {items.map((item) => {
          const cover = getCover(item)
          const info = getInfo(item)
          return (
            <div key={getItemKey(item)} className={`min-w-0 ${center ? 'text-center' : ''}`}>
              <img
                src={cover.src}
                alt={cover.alt}
                className={`w-full object-cover ${coverClassName ?? ''}`}
                style={{ aspectRatio: aspect }}
                loading="lazy"
              />
              <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{info.title}</p>
              {info.subtitle && (
                <p className="mt-0.5 truncate text-xs text-[var(--color-text-secondary)]">{info.subtitle}</p>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
