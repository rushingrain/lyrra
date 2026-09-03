import type { Playlist } from '@/types'
import type { ResultColumn } from '@/components/common/ResultTable'
import { CountCell, CoverTextCell } from './cells'

export const playlistColumns: ResultColumn<Playlist>[] = [
  {
    key: 'index',
    title: '#',
    className: 'w-[6%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (_playlist, index) => index + 1,
  },
  {
    key: 'title',
    title: '标题',
    className: 'min-w-0 flex-1',
    render: (playlist) => (
      <CoverTextCell coverUrl={playlist.coverUrl || ''} title={playlist.name} subtitle={playlist.description} />
    ),
  },
  {
    key: 'songCount',
    title: '歌曲数',
    className: 'w-[12%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (playlist) => <CountCell value={playlist.songCount} unit="首" />,
  },
  {
    key: 'creator',
    title: '创建者',
    className: 'w-1/5 shrink-0 truncate text-sm text-[var(--color-text-secondary)]',
    render: (playlist) => playlist.creator,
  },
  {
    key: 'playCount',
    title: '播放量',
    className: 'w-[12%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (playlist) => <CountCell value={playlist.playCount} />,
  },
]
