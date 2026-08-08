import { ChevronRight } from 'lucide-react'
import { SongCard } from '@/components/SongCard'
import type { Song } from '@/types'

interface SongListProps {
  songs: Song[]
  title?: string
}

/**
 * 首页音乐板块。
 *
 * 行卡片布局：封面在左，右侧两行（歌名 / 音质标签 + 原唱标签 + 歌手）。
 * 多列 CSS Grid 自适应，每列 minmax(clamp(160px, 20vw, 220px), 1fr)，
 * 窗口变化时列数与卡片尺寸连续缩放。
 */
export function SongList({ songs, title }: SongListProps) {
  if (songs.length === 0) return null

  return (
    <section className="mb-10">
      <div className="flex items-center justify-between mb-4">
        <a
          href="#"
          className="flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80"
        >
          {title}
          <ChevronRight size={20} />
        </a>
      </div>
      {/* 多列 grid：auto-fit + 1fr 铺满整行，行卡片在 clamp 下限与 220px 上限之间缩放 */}
      <div className="grid grid-cols-[repeat(auto-fit,minmax(clamp(160px,20vw,220px),1fr))] gap-4">
        {songs.map((s) => (
          <SongCard key={s.id} song={s} />
        ))}
      </div>
    </section>
  )
}
