import type { Album } from '@/types'
import type { ResultColumn } from '@/components/common/ResultTable'
import { CoverTextCell } from './cells'

export const albumColumns: ResultColumn<Album>[] = [
  {
    key: 'index',
    title: '#',
    className: 'w-[6%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (_album, index) => index + 1,
  },
  {
    key: 'title',
    title: '标题',
    className: 'min-w-0 flex-1',
    render: (album) => <CoverTextCell coverUrl={album.coverUrl || ''} title={album.name} />,
  },
  {
    key: 'artist',
    title: '作者',
    className: 'w-1/5 shrink-0 truncate text-sm text-[var(--color-text-secondary)]',
    render: (album) => album.artist,
  },
  {
    key: 'date',
    title: '时间',
    className: 'w-[12%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (album) => album.date,
  },
]
