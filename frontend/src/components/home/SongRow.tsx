import type { Song } from '../../types'

interface SongRowProps {
  song: Song
}

export function SongRow({ song }: SongRowProps) {
  return (
    <div className="group flex cursor-pointer items-center gap-3 rounded-lg p-2 transition-colors hover:bg-gray-50">
      {/* 封面 */}
      <img
        src={song.coverUrl}
        alt={song.title}
        className="h-14 w-14 shrink-0 rounded-lg object-cover"
        loading="lazy"
      />
      {/* 文字区域 */}
      <div className="min-w-0 flex-1">
        {/* 第一行：歌曲名称 */}
        <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">
          {song.title}
        </p>
        {/* 第二行：音质标签 + 原唱标签 + 作者名 */}
        <div className="mt-0.5 flex items-center gap-1.5 text-xs text-[var(--color-text-secondary)]">
          {song.quality && (
            <span className="shrink-0 rounded border border-[var(--color-accent)] px-1 text-[10px] leading-relaxed text-[var(--color-accent)]">
              {song.quality}
            </span>
          )}
          {song.isOriginal && (
            <span className="shrink-0 rounded bg-[var(--color-accent)] px-1 text-[10px] leading-relaxed text-[var(--color-on-accent)]">
              原唱
            </span>
          )}
          <span className="truncate">{song.artist}</span>
        </div>
      </div>
    </div>
  )
}
