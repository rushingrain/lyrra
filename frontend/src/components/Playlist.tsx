import type { Playlist } from '@/types'
import { Play } from 'lucide-react'

interface PlaylistProps {
  playlist: Playlist
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
      className={`group flex min-w-0 flex-col cursor-pointer text-left ${className}`}
    >
      {/* 封面：960:640 = 3:2，随容器宽度等比缩放；右下角叠加播放次数角标 */}
      <div className="relative mb-2 overflow-hidden rounded-xl">
        <img
          src={playlist.coverUrl}
          alt={playlist.name}
          className="aspect-[3/2] w-full object-cover transition-transform duration-300 group-hover:scale-105"
          loading="lazy"
        />
        <span className="absolute bottom-1.5 right-1.5 flex items-center gap-1 rounded-md bg-black/45 px-1.5 py-0.5 text-xs font-medium text-white backdrop-blur-sm">
          <Play size={12} fill="currentColor" />
          <span>{formatCount(playlist.playCount)}</span>
        </span>
      </div>
      {/* 名称：独占整行，超长省略，避免与播放次数挤压 */}
      <p className="min-w-0 truncate text-sm font-medium text-[var(--color-text-primary)]">
        {playlist.name}
      </p>
      {/* 描述 */}
      <p className="mt-1 truncate text-xs leading-relaxed text-[var(--color-text-secondary)]">
        {playlist.description}
      </p>
    </div>
  )
}
