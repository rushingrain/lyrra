import { Heart, Star } from 'lucide-react'

interface DefaultPlaylistCoverProps {
  isDefault: 1 | 2
}

/** 特殊歌单使用固定配色，避免封面随导航的悬停或选中颜色变化。 */
export function DefaultPlaylistCover({ isDefault }: DefaultPlaylistCoverProps) {
  const Icon = isDefault === 1 ? Star : Heart
  const colors = isDefault === 1
    ? 'from-amber-50 to-amber-100 text-amber-600 ring-amber-200/70'
    : 'from-rose-50 to-rose-100 text-rose-500 ring-rose-200/70'

  return (
    <div
      aria-hidden="true"
      className={`flex h-full w-full items-center justify-center bg-linear-to-br ring-1 ring-inset ${colors}`}
    >
      <Icon className="h-[45%] w-[45%]" strokeWidth={1.75} fill="currentColor" fillOpacity={0.15} />
    </div>
  )
}
