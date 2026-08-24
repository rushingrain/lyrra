import { Song } from '@/components/common/Song'
import type { Song as SongType } from '@/types';

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
        {songs.map((s) => (
          <Song key={s.id} song={s} />
        ))}
      </div>
    </div>
  )
}
