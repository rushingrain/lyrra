import type { Song as SongType } from '@/types'
import type { ResultColumn } from '@/components/common/ResultTable'
import { Song } from '@/components/common/Song'
import { DurationCell, LikeCell } from './cells'

/**
 * 歌词搜索的展示列：结果本身是歌曲，标题列复用 Song 组件
 * （封面 + 音质/原唱标签 + 歌手 + 风格标签）。
 */
export const lyricColumns: ResultColumn<SongType>[] = [
  {
    key: 'index',
    title: '#',
    className: 'w-[2%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (_song, index) => String(index + 1).padStart(2, '0'),
  },
  {
    key: 'title',
    title: '标题',
    className: 'min-w-0 flex-1',
    render: (song) => <Song song={song} />,
  },
  {
    key: 'like',
    title: '喜欢',
    className: 'w-[6%] shrink-0',
    render: () => <LikeCell />,
  },
  {
    key: 'duration',
    title: '时长',
    className: 'w-[10%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (song) => <DurationCell ms={song.duration} />,
  },
]
