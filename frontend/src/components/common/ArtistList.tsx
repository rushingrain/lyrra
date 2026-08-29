import type { Artist as ArtistType } from '@/types'
import { Summary } from '@/components/common/Summary'
import type { SummaryLayout } from '@/components/common/Summary'

interface ArtistListProps extends SummaryLayout {
  artists: ArtistType[]
}

/**
 * 歌手列表:负责歌手业务语义, 单行响应式布局由 Summary 复用。
 */
export function ArtistList({ artists, ...layout }: ArtistListProps) {
  return (
    <Summary
      {...layout}
      items={artists}
      getItemKey={(artist) => artist.id}
      renderItem={(artist, { className, style }) => (
        <div className={`min-w-0 text-center ${className}`} style={style}>
          <img src={artist.avatarUrl ? artist.avatarUrl : 'src/assets/images/default_cover.jpg'} alt={artist.name} className="mx-auto aspect-square w-full rounded-full object-cover" loading="lazy" />
          <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{artist.name}</p>
          <p className="mt-0.5 text-xs text-[var(--color-text-secondary)]">单曲: {artist.songCount}</p>
        </div>
      )}
    />
  )
}
