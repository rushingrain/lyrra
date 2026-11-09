import type { ReactNode } from 'react'
import { Heart } from 'lucide-react'
import { Cover } from '@/components/common/Cover'

interface CoverTextCellProps {
  /** 不传则不渲染封面（如歌词）；传空字符串则回退默认封面 */
  coverUrl?: string
  title: string
  /** 第二行内容：纯文本或自定义节点 */
  subtitle?: ReactNode
  /** 封面是否为圆形 */
  circle?: boolean
}

/** 封面 + 主/副标题单元格：左侧封面，右侧一行或两行文字 */
export function CoverTextCell({ coverUrl, title, subtitle, circle = false }: CoverTextCellProps) {
  return (
    <div className="flex min-w-0 items-center gap-3">
      {coverUrl !== undefined && (
        <Cover src={coverUrl} alt={title} variant={circle ? 'circle' : 'rounded'} />
      )}
      <div className="min-w-0">
        <div className="truncate text-base font-medium text-[var(--color-text-primary)]">{title}</div>
        {subtitle && (
          <div className="mt-0.5 min-w-0 truncate text-sm text-[var(--color-text-secondary)]">{subtitle}</div>
        )}
      </div>
    </div>
  )
}

/** 「喜欢」列占位：纯展示的空心心形 */
export function LikeCell() {
  return <Heart size={16} className="block text-[var(--color-text-secondary)]" />
}

/** 后端 duration 单位为毫秒 */
function formatDuration(ms?: number): string {
  if (!ms || ms <= 0) return ''
  const total = Math.floor(ms / 1000)
  const m = Math.floor(total / 60)
  const s = total % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

/** 时长单元格：毫秒格式化为 mm:ss，缺省显示为空 */
export function DurationCell({ ms }: { ms?: number }) {
  return <span className="tabular-nums">{formatDuration(ms)}</span>
}

function formatCount(value: number): string {
  const unit = value >= 1e8 ? 1e8 : value >= 1e4 ? 1e4 : 1
  if (unit === 1) return String(value)
  const result = value / unit
  return `${Number.isInteger(result) ? result : result.toFixed(1)}${unit === 1e8 ? '亿' : '万'}`
}

/** 数量单元格：万/亿缩写，可附带计量单位（如「首」），缺省显示为空 */
export function CountCell({ value, unit = '' }: { value?: number; unit?: string }) {
  return <span className="tabular-nums">{value == null ? '' : `${formatCount(value)}${unit}`}</span>
}
