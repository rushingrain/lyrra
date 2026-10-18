import type { Song as SongType } from '@/types'
import { PlayableSong } from '@/components/business/PlayableSong'

interface SongListProps {
  songs: SongType[]
  /**
   * 列布局方式：
   * - 'auto'（默认）：auto-fit 按窗口自适应列数；
   * - 'responsive'：小窗口 2 列、大窗口（≥1280px）3 列，列宽均分整行；
   * - 数字 N：固定 N 列。
   * auto / 数字模式下每列均按 minmax(clamp(160px, 20vw, 220px), 1fr) 随窗口连续缩放，
   * 1fr 铺满剩余空间；columns 仅控制列数。
   */
  columns?: 'auto' | 'responsive' | number
  /** 固定展示行数（仅 responsive 模式生效）：窄窗口每行 2 条、宽窗口每行 3 条 */
  rows?: number
}

/**
 * 音乐卡片
 *
 * 行卡片布局：封面在左，右侧两行（歌名 / 音质标签 + 原唱标签 + 歌手）。
 * 多列 CSS Grid：columns='auto' 时列数自适应，'responsive' 按断点切换 2/3 列
 * 并按 1fr 均分整行，指定数字时列数固定；auto / 数字模式的列宽为
 * minmax(clamp(160px, 20vw, 220px), 1fr)，随窗口连续缩放。
 */
export function SongList({ songs, columns = 'auto', rows }: SongListProps) {
  if (songs.length === 0) return null

  // 首页歌曲板块：小窗口 2 列、大窗口 3 列，列宽 1fr 均分整行
  if (columns === 'responsive') {
    const narrowPerRow = 2
    const widePerRow = 3
    // 行数固定时换算成条数：窄窗口按 2 条/行，宽窗口按 3 条/行
    const narrowCount = rows === undefined ? songs.length : rows * narrowPerRow
    const wideCount = rows === undefined ? songs.length : rows * widePerRow
    // 数据够满行时按宽窗口的条数渲染，多出来的只在宽窗口出现；
    // 不足时退到窄窗口的条数，保证两种列数下都不会出现半行
    const count = songs.length >= wideCount ? wideCount : Math.min(songs.length, narrowCount)

    return (
      <div className="grid grid-cols-2 gap-8 xl:grid-cols-3">
        {songs.slice(0, count).map((song, index) => (
          <PlayableSong
            key={song.id}
            song={song}
            size="md"
            // 窄窗口只保留前 rows 行的条目，避免多出一行
            className={`rounded-lg${index >= narrowCount ? ' max-xl:hidden' : ''}`}
          />
        ))}
      </div>
    )
  }

  const items = songs.map((song) => (
    <PlayableSong key={song.id} song={song} size="md" className="rounded-lg" />
  ))

  // 列数：'auto' → auto-fit 按窗口自适应；数字 → 固定列数。其余（每列尺寸、缩放）保持一致。
  const repeatCount = columns === 'auto' ? 'auto-fit' : columns
  const gridTemplateColumns = `repeat(${repeatCount}, minmax(clamp(160px, 20vw, 220px), 1fr))`

  return (
    <div className="grid gap-8" style={{ gridTemplateColumns }}>
      {items}
    </div>
  )
}
