import type { Song as SongType } from '@/types'
import { Pause, Play } from 'lucide-react'
import { Cover } from '@/components/common/Cover'
import { Tag } from '@/components/common/Tag'

interface SongProps {
  song: SongType
  /** sm: 表格行内（48px 封面）；md: 卡片网格（60px 封面） */
  size?: 'sm' | 'md'
  className?: string
  /** 传入即开启播放：点击封面触发（同一首再点由调用方决定播放/暂停） */
  onPlay?: (song: SongType) => void
  /** 是否是当前正在播放的歌曲：遮罩常显并显示暂停图标 */
  playing?: boolean
}

const PRESETS = {
  sm: { root: 'gap-3', cover: 'h-12 w-12', meta: 'mt-0.5' },
  md: { root: 'gap-4', cover: 'h-15 w-15', meta: 'mt-1.5' },
} as const

/** 歌曲条目：封面 + 标题 + 元信息行（音质/原唱标签 + 歌手）+ 风格标签 */
export function Song({ song, size = 'sm', className = '', onPlay, playing = false }: SongProps) {
  const preset = PRESETS[size]
  return (
    <div className={`group flex min-w-0 items-center ${preset.root} ${className}`}>
      <div className="relative shrink-0">
        <Cover src={song.coverUrl || ''} alt={song.name} className={preset.cover} />
        {onPlay && (
          <button
            type="button"
            onClick={() => onPlay(song)}
            aria-label={playing ? `暂停 ${song.name}` : `播放 ${song.name}`}
            title={playing ? '暂停' : '播放'}
            className={`absolute inset-0 flex cursor-pointer items-center justify-center rounded-lg bg-black/40 transition-opacity ${
              playing ? 'opacity-100' : 'opacity-0 group-hover:opacity-100'
            }`}
          >
            {playing ? (
              <Pause size={20} className="text-white" fill="currentColor" />
            ) : (
              // 纯白三角：Play 图标本身是描边图形，fill 后即为实心三角
              <Play size={20} className="ml-0.5 text-white" fill="currentColor" />
            )}
          </button>
        )}
      </div>
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
