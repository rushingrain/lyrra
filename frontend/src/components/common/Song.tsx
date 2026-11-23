import type { Song as SongType } from '@/types'
import type { ReactNode } from 'react'
import { Pause, Play } from 'lucide-react'
import { Cover } from '@/components/common/Cover'
import { Tag } from '@/components/common/Tag'

interface SongProps {
  song: SongType
  /** sm: 表格行内（44px 封面）；md: 卡片网格（56px 封面） */
  size?: 'sm' | 'md'
  className?: string
  /** 传入即开启播放：点击封面触发（同一首再点由调用方决定播放/暂停） */
  onPlay?: (song: SongType) => void
  /** 是否是当前正在播放的歌曲：遮罩常显并显示暂停图标 */
  playing?: boolean
  /** 是否是当前曲目：标题高亮（播放列表用来标记正在播放的那一行） */
  active?: boolean
  actions?: ReactNode
}

const PRESETS = {
  sm: { root: 'min-h-12 gap-3', cover: 'h-11 w-11', meta: 'mt-0.5' },
  md: { root: 'min-h-15 gap-4', cover: 'h-14 w-14', meta: 'mt-1.5' },
} as const

/** 歌曲条目：封面 + 标题 + 元信息行（风格标签 + 音质 + 原唱 + 歌手） */
export function Song({
  song,
  size = 'sm',
  className = '',
  onPlay,
  playing = false,
  active = false,
  actions,
}: SongProps) {
  const preset = PRESETS[size]
  const titleColor = active ? 'text-[var(--color-accent)]' : 'text-[var(--color-text-primary)]'
  return (
    <div className={`group group/song flex min-w-0 items-center ${preset.root} ${className}`}>
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
        <p className={`truncate text-sm text-[15px] ${titleColor}`}>{song.name}</p>
        <div className={`${preset.meta} flex h-4 min-w-0 items-center gap-1 overflow-hidden whitespace-nowrap text-[13px] leading-4 text-[var(--color-text-secondary)]`}>
          {song.tags?.map((tag) => (
            <span key={tag} className="inline-flex h-4 shrink-0 items-center rounded-xs bg-[var(--color-accent-bg)] px-px py-0 text-[9px] leading-none text-[var(--color-accent)]">
              {tag}
            </span>
          ))}
          {song.quality && <Tag variant="outline" className="inline-flex h-4 items-center px-0.5 py-0">{song.quality}</Tag>}
          {song.isOriginal && <Tag variant="solid" className="inline-flex h-4 items-center px-0.5 py-0">原唱</Tag>}
          <span className="min-w-0 truncate">{song.artist}</span>
        </div>
      </div>
      {actions && (
        <div className="pointer-events-none shrink-0 opacity-0 transition-opacity group-hover/song:pointer-events-auto group-hover/song:opacity-100 group-hover/song-row:pointer-events-auto group-hover/song-row:opacity-100 group-focus-within/song:pointer-events-auto group-focus-within/song:opacity-100 focus-within:pointer-events-auto focus-within:opacity-100 has-[[aria-expanded=true]]:pointer-events-auto has-[[aria-expanded=true]]:opacity-100">
          {actions}
        </div>
      )}
    </div>
  )
}
