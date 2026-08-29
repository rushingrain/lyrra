import type { Album as AlbumType } from '@/types'
import { SingleRowList } from '@/components/common/SingleRowList'

interface AlbumListProps {
  albums: AlbumType[]
}

/**
 * 专辑列表:负责专辑业务语义,单行响应式布局由 SingleRowList 复用。
 */
export function AlbumList({ albums }: AlbumListProps) {
  return (
    <SingleRowList
      className="mb-10"
      items={albums}
      getItemKey={(album) => album.id}
      renderItem={(album, { className, style }) => (
        <div className={`min-w-0 cursor-pointer ${className}`} style={style}>
          <div className="relative overflow-hidden rounded-t-md">
            <img src={album.coverUrl} alt={album.name} className="aspect-square w-full object-cover transition-transform duration-300" loading="lazy" />
          </div>
          <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{album.name}</p>
          <p className="mt-0.5 truncate text-xs text-[var(--color-text-secondary)]">{album.artist} · {album.date}</p>
        </div>
      )}
    />
  )
}
