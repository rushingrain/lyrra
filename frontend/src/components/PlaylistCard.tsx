import type { Playlist } from '@/types'
import { Play } from 'lucide-react'

interface PlaylistCardProps {
  playlist: Playlist
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

export function PlaylistCard({ playlist }: PlaylistCardProps) {
  return (
    <button className="group min-w-0 flex-1 cursor-pointer text-left">
      {/* 封面：960:640 = 3:2，随容器宽度等比缩放 */}
      <div className="relative mb-2 overflow-hidden rounded-xl">
        <img
          src={playlist.coverUrl}
          alt={playlist.name}
          className="aspect-[3/2] w-full object-cover transition-transform duration-300 group-hover:scale-105"
          loading="lazy"
        />
      </div>
      {/* 名称 + 播放次数 */}
      <div className="flex items-center gap-2">
        <p className="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text-primary)]">
          {playlist.name}
        </p>
        <span className="flex shrink-0 items-center gap-1 text-xs text-[var(--color-text-secondary)]">
          <Play size={12} fill="currentColor" />
          <span>{formatCount(playlist.playCount)}</span>
        </span>
      </div>
      {/* 描述 */}
      <p className="mt-1 line-clamp-2 text-xs leading-relaxed text-[var(--color-text-secondary)]">
        {playlist.description}
      </p>
    </button>
  )
}
