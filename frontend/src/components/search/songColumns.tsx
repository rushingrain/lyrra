import type { Song as SongType } from '@/types'
import type { ResultColumn } from '@/components/common/ResultTable'
import { PlayableSong } from '@/components/business/PlayableSong'
import { DurationCell, LikeCell } from './cells'

export const songColumns: ResultColumn<SongType>[] = [
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
    render: (song) => <PlayableSong song={song} />,
  },
  {
    key: 'album',
    title: '专辑',
    className: 'w-1/5 shrink-0 truncate text-sm text-[var(--color-text-secondary)]',
    render: (song) => song.album ?? '',
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
