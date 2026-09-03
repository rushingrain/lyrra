import type { SearchLyric } from '@/services/search'
import type { ResultColumn } from '@/components/common/ResultTable'
import { CoverTextCell, DurationCell, LikeCell } from './cells'

export const lyricColumns: ResultColumn<SearchLyric>[] = [
  {
    key: 'index',
    title: '#',
    className: 'w-[6%] shrink-0 text-sm text-[var(--color-text-secondary)]',
    render: (_lyric, index) => index + 1,
  },
  {
    key: 'title',
    title: '标题',
    className: 'min-w-0 flex-1',
    // 歌词暂无封面数据，仅展示文字
    render: (lyric) => <CoverTextCell title={lyric.title} subtitle={lyric.artist} />,
  },
  {
    key: 'album',
    title: '专辑',
    className: 'w-1/5 shrink-0 truncate text-sm text-[var(--color-text-secondary)]',
    render: (lyric) => lyric.album,
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
    // 歌词暂无时长数据，显示为空
    render: () => <DurationCell />,
  },
]
