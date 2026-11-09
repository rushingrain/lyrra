import { Headphones } from 'lucide-react'

interface PlaylistPlayCountProps {
  count: number
  className?: string
}

function formatCount(count: number): string {
  if (count >= 100000000) {
    return `${(count / 100000000).toFixed(1)}亿`
  }
  if (count >= 10000) {
    return `${(count / 10000).toFixed(1)}w`
  }
  return `${count}`
}

/** 歌单播放次数：首页与歌单列表页共用同一套图标和缩写规则。 */
export function PlaylistPlayCount({ count, className = '' }: PlaylistPlayCountProps) {
  return (
    <span className={`flex items-center gap-1 text-xs font-medium text-white ${className}`}>
      <Headphones size={12} />
      <span>{formatCount(count)}</span>
    </span>
  )
}
