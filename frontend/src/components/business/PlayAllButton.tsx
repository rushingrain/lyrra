import { Play } from 'lucide-react'

interface PlayAllButtonProps {
  onClick: () => void
  disabled?: boolean
  /** 正在准备队列（如歌单全量拉取），按钮显示加载文案并禁用 */
  loading?: boolean
}

/** 播放全部按钮：主题色背景 + 白字 + 实心播放图标 */
export function PlayAllButton({ onClick, disabled = false, loading = false }: PlayAllButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled || loading}
      onClick={onClick}
      className="flex cursor-pointer items-center gap-1 rounded-md bg-[var(--color-accent)] px-3 py-2 text-sm text-[var(--color-on-accent)] transition-colors hover:bg-[var(--color-accent-hover)] disabled:cursor-not-allowed disabled:opacity-50"
    >
      {/* Play 图标本身是描边图形，fill 后即为实心三角 */}
      <Play size={16} fill="currentColor" />
      {loading ? '加载中...' : '播放全部'}
    </button>
  )
}
