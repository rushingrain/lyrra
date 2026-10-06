import type { Song as SongType } from '@/types'
import { Song } from '@/components/common/Song'
import { usePlayerStore } from '@/stores/playerStore'

interface PlayableSongProps {
  song: SongType
  size?: 'sm' | 'md'
  className?: string
  /** 当前曲目是否高亮标题（播放列表用于标记正在播放的那一行；列表页默认不标） */
  markCurrent?: boolean
}

/**
 * 把播放器状态接到通用 Song 组件上：
 * 传入当前歌曲是否正在播放，并把点击交给 store（同一首再点即暂停）。
 *
 * 之所以单独做一层包装，是因为搜索表格的列配置是在 ResultTable 的 render 里调用的，
 * 那里不是组件渲染，不能直接用 hook。
 */
export function PlayableSong({ song, size, className, markCurrent = false }: PlayableSongProps) {
  const isCurrent = usePlayerStore((state) => state.song?.id === song.id)
  const isPlaying = usePlayerStore((state) => state.status === 'playing')
  const playSong = usePlayerStore((state) => state.playSong)

  return (
    <Song
      song={song}
      size={size}
      className={className}
      onPlay={playSong}
      playing={isCurrent && isPlaying}
      active={markCurrent && isCurrent}
    />
  )
}
