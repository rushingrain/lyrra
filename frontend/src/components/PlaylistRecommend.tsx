import { ChevronRight } from 'lucide-react'
import { PlaylistCard } from '@/components/PlaylistCard'
import type { Playlist } from '@/types'

interface PlaylistRecommendProps {
  playlists: Playlist[]
}

/**
 * 首页官方歌单板块。
 *
 * 多列自适应布局：CSS Grid 按容器宽度自动决定列数，
 * 每列 minmax(clamp(110px, 15vw, 180px), 1fr)：
 * - 1fr 保证卡片铺满整行，任何窗口宽度都没有行尾留白；
 * - 窗口小时下限收紧到 110px → 列数变多，避免卡片过大；
 * - 窗口大时下限放宽到 180px → 卡片保持舒适尺寸，不随行宽无限撑大；
 * 窗口变化时列数与卡片尺寸连续缩放。
 * 纯 CSS 实现，无 ResizeObserver，不存在监听器泄漏；窗口变化也不重新请求。
 */
export function PlaylistRecommend({ playlists }: PlaylistRecommendProps) {
  if (playlists.length === 0) return null

  return (
    <section className="mb-10">
      <div className="mb-4 flex items-center justify-between">
        <a
          href="#"
          className="flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80"
        >
          官方歌单
          <ChevronRight size={20} />
        </a>
      </div>
      {/* 多列 grid：auto-fit + 1fr 铺满整行，最小卡片宽度随窗口 clamp 动态缩放 */}
      <div className="grid grid-cols-[repeat(auto-fit,minmax(clamp(110px,15vw,180px),1fr))] gap-5">
        {playlists.map((p) => (
          <PlaylistCard key={p.id} playlist={p} />
        ))}
      </div>
    </section>
  )
}
