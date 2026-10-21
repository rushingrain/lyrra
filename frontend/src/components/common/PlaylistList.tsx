import type { Playlist as PlaylistType } from '@/types'
import { Summary } from '@/components/common/Summary'
import type { SummaryLayout } from '@/components/common/Summary'
import { Headphones } from 'lucide-react'
import { Link } from 'react-router-dom'

interface PlaylistListProps extends SummaryLayout {
  playlists: PlaylistType[]
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

/**
 * 歌单列表:负责歌单业务语义,单行响应式布局由 Summary 复用。
 */
export function PlaylistList({ playlists, ...layout }: PlaylistListProps) {
  return (
    <Summary
      {...layout}
      items={playlists}
      getItemKey={(playlist) => playlist.id}
      renderItem={(playlist, { className, style }) => (
        <Link
          to={`/songs/playlist/${encodeURIComponent(playlist.globalCollectionId)}`}
          // 把歌单信息带到详情页：接口不返回播放次数，需要用这里的 playCount 补齐
          state={{ playlist }}
          className={`flex w-full min-w-0 flex-col cursor-pointer text-left ${className}`}
          style={style}
        >
          <div className="relative mb-2 overflow-hidden rounded-t-md">
            <img
              src={playlist.coverUrl}
              alt={playlist.name}
              className="aspect-square w-full object-cover transition-transform duration-300"
              loading="lazy"
            />
            <span className="absolute left-1.5 top-1.5 flex items-center gap-1 text-xs font-medium text-white">
              <Headphones size={12} />
              <span>{formatCount(playlist.playCount)}</span>
            </span>
          </div>
          <p className="min-w-0 truncate text-[15px] font-medium text-[var(--color-text-primary)]">
            {playlist.name}
          </p>
          <p className="mt-1 min-w-0 truncate text-[13px] leading-relaxed text-[var(--color-text-secondary)]">
            {playlist.description}
          </p>
        </Link>
      )}
    />
  )
}
