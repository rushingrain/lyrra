import type { Song as SongType } from '@/types'
import { Cover } from '@/components/common/Cover'
import { Tag } from '@/components/common/Tag'

interface SongProps {
  song: SongType
  /** sm: 表格行内（48px 封面）；md: 卡片网格（60px 封面） */
  size?: 'sm' | 'md'
  className?: string
}

const PRESETS = {
  sm: { root: 'gap-3', cover: 'h-12 w-12', meta: 'mt-0.5' },
  md: { root: 'gap-4', cover: 'h-15 w-15', meta: 'mt-1.5' },
} as const

/** 歌曲条目：封面 + 标题 + 元信息行（音质/原唱标签 + 歌手）+ 风格标签 */
export function Song({ song, size = 'sm', className = '' }: SongProps) {
  const preset = PRESETS[size]
  return (
    <div className={`flex min-w-0 items-center ${preset.root} ${className}`}>
      <Cover src={song.coverUrl || ''} alt={song.name} className={preset.cover} />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm text-[15px] text-[var(--color-text-primary)]">{song.name}</p>
        <div className={`${preset.meta} flex min-w-0 items-center gap-1.5 text-[13px] text-[var(--color-text-secondary)]`}>
          {song.quality && <Tag variant="outline">{song.quality}</Tag>}
          {song.isOriginal && <Tag variant="solid">原唱</Tag>}
          <span className="min-w-0 truncate text-[13px]">{song.artist}</span>
        </div>
        {song.tags && song.tags.length > 0 && (
          <div className={`${preset.meta} flex flex-wrap items-center gap-1.5`}>
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
