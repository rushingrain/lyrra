import type { Song } from '@/types'
import defaultCover from '@/assets/images/default_cover.jpg'

interface SongProps {
  song: Song
}

export function Song({ song }: SongProps) {
  return (
    <div className="flex min-w-0 items-center gap-4 rounded-lg">
      <img src={song.coverUrl ? song.coverUrl : defaultCover} alt={song.title} className="h-15 w-15 shrink-0 rounded-lg object-cover" loading="lazy" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">{song.title}</p>
        <div className="mt-1.5 flex items-center gap-1.5 text-xs text-[var(--color-text-secondary)]">
          {song.quality && (
            <span className="shrink-0 rounded-xs border border-[var(--color-accent)] px-0.5 text-[8px] leading-relaxed text-[var(--color-accent)]">
              {song.quality}
            </span>
          )}
          {song.isOriginal && (
            <span className="shrink-0 rounded-xs bg-[var(--color-accent)] px-0.5 text-[8px] leading-relaxed text-[var(--color-on-accent)]">原唱</span>
          )}
          <span className="min-w-0 truncate text-xs">{song.artist}</span>
        </div>
        {song.tags && song.tags.length > 0 && (
          <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
            {song.tags.map((tag) => (
              <span key={tag} className="shrink-0 rounded bg-[var(--color-accent-bg)] px-1 py-0.5 text-[10px] leading-relaxed text-[var(--color-accent)]">
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
