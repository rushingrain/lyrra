import type { Playlist as PlaylistType } from '@/types'
import { Headphones } from 'lucide-react'

interface PlaylistProps {
  playlist: PlaylistType
  /** 外层布局注入的尺寸控制（如 flex/grid 中的 min/max-width），默认不设随父容器 */
  className?: string
}

function formatCount(count: number): string {
  if (count >= 100000000) {
    return `${(count / 100000000).toFixed(1)}亿`
  }
  if (count >= 10000) {
    return `${(count / 10000).toFixed(1)}w`
  }
  return `${count}`
}

export function Playlist({ playlist, className = '' }: PlaylistProps) {
  return (
    <div
      role="button"
      tabIndex={0}
      // ✅ 关键修改：加上 w-full 让卡片填满 Grid 单元格
      className={`group flex w-full min-w-0 flex-col cursor-pointer text-left ${className}`}
    >
      <div className="relative mb-2 overflow-hidden rounded-t-md">
        <img
          src={playlist.coverUrl}
          alt={playlist.name}
          className="aspect-square w-full object-cover transition-transform duration-300 group-hover:scale-105"
          loading="lazy"
        />
        <span className="absolute left-1.5 top-1.5 flex items-center gap-1 text-xs font-medium text-white">
          <Headphones size={12} />
          <span>{formatCount(playlist.playCount)}</span>
        </span>
      </div>
      <p className="min-w-0 truncate text-sm font-medium text-[var(--color-text-primary)]">
        {playlist.name}
      </p>
      <p className="mt-1 min-w-0 truncate text-xs leading-relaxed text-[var(--color-text-secondary)]">
        {playlist.description}
      </p>
    </div>
  )
}