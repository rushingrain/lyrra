import type { Song as SongType } from '@/types'
import defaultCover from '@/assets/images/default_cover.jpg'

interface SongListProps {
  songs: SongType[]
  /**
   * 列布局方式：
   * - 'auto'（默认）：auto-fit 按窗口自适应列数；
   * - 数字 N：固定 N 列。
   * 两种模式下每列均按 minmax(clamp(160px, 20vw, 220px), 1fr) 随窗口连续缩放，
   * 1fr 铺满剩余空间；columns 仅控制列数。
   */
  columns?: 'auto' | number
}

/**
 * 音乐卡片
 *
 * 行卡片布局：封面在左，右侧两行（歌名 / 音质标签 + 原唱标签 + 歌手）。
 * 多列 CSS Grid：每列 minmax(clamp(160px, 20vw, 220px), 1fr) 随窗口连续缩放；
 * columns='auto' 时列数自适应，指定数字时列数固定。
 */
export function SongList({ songs, columns = 'auto' }: SongListProps) {
  if (songs.length === 0) return null

  // 列数：'auto' → auto-fit 按窗口自适应；数字 → 固定列数。其余（每列尺寸、缩放）保持一致。
  const repeatCount = columns === 'auto' ? 'auto-fit' : columns
  const gridTemplateColumns = `repeat(${repeatCount}, minmax(clamp(160px, 20vw, 220px), 1fr))`

  return (
    <div className="mb-10">
      <div className="grid gap-8" style={{ gridTemplateColumns }}>
        {songs.map((song) => (
          <div key={song.id} className="flex min-w-0 items-center gap-4 rounded-lg">
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
        ))}
      </div>
    </div>
  )
}
