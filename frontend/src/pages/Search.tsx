import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { SongList } from '@/components/common/SongList'
import { Section } from '@/components/common/Section'
import { PlaylistList } from '@/components/common/PlaylistList'
import { AlbumList } from '@/components/common/AlbumList'
import { ArtistList } from '@/components/common/ArtistList'
import { MvList } from '@/components/common/MvList'
import { ResultTable } from '@/components/common/ResultTable'
import { songColumns } from '@/components/search/songColumns'
import { playlistColumns } from '@/components/search/playlistColumns'
import { albumColumns } from '@/components/search/albumColumns'
import { lyricColumns } from '@/components/search/lyricColumns'
import { useSearchPagination } from '@/hooks/useSearchPagination'
import { SEARCH_MAX_ITEMS, type SearchResult, type SearchTabType } from '@/services/search'
import type { Album, Artist } from '@/types'

type SearchTab = SearchTabType

const TABS: { key: SearchTab; label: string }[] = [
  { key: 'complex', label: '综合' },
  { key: 'song', label: '单曲' },
  { key: 'special', label: '歌单' },
  { key: 'album', label: '专辑' },
  { key: 'author', label: '歌手' },
  { key: 'mv', label: 'MV' },
  { key: 'lyric', label: '歌词' },
]

function formatCount(value: number): string {
  const unit = value >= 1e8 ? 1e8 : value >= 1e4 ? 1e4 : 1
  if (unit === 1) return String(value)
  const result = value / unit
  return `${Number.isInteger(result) ? result : result.toFixed(1)}${unit === 1e8 ? '亿' : '万'}`
}

function ArtistSummary({ artist }: { artist: Artist }) {
  return (
    <div className="flex min-w-0 items-center gap-4 p-4">
      <img src={artist.avatarUrl ? artist.avatarUrl : 'src/assets/images/default_cover.jpg'} alt={artist.name} className="h-20 w-20 shrink-0 rounded-full object-cover" loading="lazy" />
      <div className="min-w-0">
        <div className="truncate font-bold text-[var(--color-text-primary)]">歌手：{artist.name}</div>
        <div className="mt-1 grid w-fit grid-cols-[auto_auto] gap-4 truncate text-sm text-[var(--color-text-secondary)]">
          <div>单曲：{artist.songCount}</div>
          <div>粉丝：{formatCount(artist.fans)}</div>
        </div>
      </div>
    </div>
  )
}

function AlbumSummary({ album }: { album: Album }) {
  return (
    <div className="flex min-w-0 items-center gap-4 p-4">
      <img
        src={album.coverUrl}
        alt={album.name}
        className="h-20 w-20 shrink-0 rounded-lg object-cover"
        loading="lazy"
      />
      <div className="min-w-0">
        <div className="truncate font-bold text-[var(--color-text-primary)]">专辑：{album.name}</div>
        <div className="mt-1 grid w-fit grid-cols-[auto_auto] gap-4 truncate text-sm text-[var(--color-text-secondary)]">
          <div>{album.artist}</div>
          <div>{album.date}</div>
        </div>
      </div>
    </div>
  )
}

function hasResults(result: SearchResult): boolean {
  return Object.values(result).some((items) => items.length > 0)
}

/**
 * 把最近的滚动容器滚回顶部。
 * 搜索页自身不滚动，滚动发生在外层 SimpleBar 的内容容器上，
 * 因此这里向上找第一个可滚动祖先，而不是直接操作 window。
 */
function scrollAncestorToTop(node: HTMLElement) {
  for (let el: HTMLElement | null = node.parentElement; el; el = el.parentElement) {
    const { overflowY } = getComputedStyle(el)
    if (overflowY === 'auto' || overflowY === 'scroll') {
      el.scrollTop = 0
      return
    }
  }
}

function TabSearchResults({ tab, result }: { tab: Exclude<SearchTab, 'complex'>; result: SearchResult }) {
  switch (tab) {
    case 'song':
      return <ResultTable columns={songColumns} rows={result.songs} rowKey={(song) => song.id} />
    case 'special':
      return <ResultTable columns={playlistColumns} rows={result.playlists} rowKey={(playlist) => playlist.id} />
    case 'album':
      return <ResultTable columns={albumColumns} rows={result.albums} rowKey={(album) => album.id} />
    case 'author':
      return <ArtistList artists={result.artists} />
    case 'mv':
      return <MvList mvs={result.mvs} />
    case 'lyric':
      return <ResultTable columns={lyricColumns} rows={result.songs} rowKey={(song) => song.id} />
  }
}

export default function Search() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q')?.trim() || ''
  // 分类状态带上它所属的关键词：新关键词下默认回到综合分类，用户点击后才落到新关键词上。
  // 这样比较在渲染期完成，不会像 effect 那样先按上一次的分类发一次请求再丢弃。
  const [tabState, setTabState] = useState({ query, tab: 'complex' as SearchTab })
  const activeTab = tabState.query === query ? tabState.tab : 'complex'

  const {
    result,
    isLoading,
    loadError,
    loadMoreError,
    hasMore,
    limitReached,
    sentinelRef,
    loadMore,
  } = useSearchPagination(query, activeTab)

  if (!query) {
    return (
      <div className="flex h-64 items-center justify-center text-[var(--color-text-secondary)]">
        <p className="text-lg">请输入搜索内容</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      {/* 标题与分类吸顶：列表滚到深处时仍能直接切换分类 */}
      <div className="sticky top-0 z-10 bg-[var(--color-background)] pb-4">
        <p className="mb-4 text-2xl font-semibold text-[var(--color-text-primary)]">{query}</p>

        <div className="flex gap-6">
          {TABS.map((tab) => (
            <button
              key={tab.key}
              type="button"
              className={`cursor-pointer border-b-3 text-base transition-[color,border-color] ${activeTab === tab.key
                  ? 'border-b-(--color-accent) font-bold text-(--color-text-primary)'
                  : 'border-b-transparent font-normal text-(--color-text-secondary) hover:text-(--color-text-primary)'
                }`}
              onClick={(event) => {
                setTabState({ query, tab: tab.key })
                // 新分类从头加载第一页，滚动位置也一并回到顶部
                scrollAncestorToTop(event.currentTarget)
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {loadError && (
        <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">{loadError}</p>
      )}

      {!isLoading && !loadError && !hasResults(result) && (
        <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">未找到相关结果</p>
      )}

      {activeTab === 'complex' && !isLoading && !loadError && hasResults(result) && (
        <div className="flex flex-col gap-8 pb-8">
          {(result.artists[0] || result.albums[0]) && (
            <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
              {result.artists[0] && <ArtistSummary artist={result.artists[0]} />}
              {result.albums[0] && <AlbumSummary album={result.albums[0]} />}
            </div>
          )}

          {result.songs.length > 0 && (
            // TODO 根据窗口宽度调整列数
            <Section title="单曲" children={<SongList columns={2} songs={result.songs.slice(0, 6)} />} />
          )}

          {result.playlists.length > 0 && (
            <Section title="歌单" children={<PlaylistList playlists={result.playlists} />} />
          )}

          {result.albums.length > 0 && (
            <Section title="专辑" children={<AlbumList albums={result.albums} />} />
          )}

          {result.artists.length > 0 && (
            <Section title="歌手" children={<ArtistList artists={result.artists} />} />
          )}

          {result.mvs.length > 0 && (
            <Section title="MV" children={<MvList mvs={result.mvs} />} />
          )}
        </div>
      )}

      {activeTab !== 'complex' && !isLoading && !loadError && hasResults(result) && (
        <div>
          <TabSearchResults tab={activeTab} result={result} />

          {/* 无限滚动哨兵：进入视口即追加下一页 */}
          {hasMore && <div ref={sentinelRef} className="h-px" aria-hidden />}

          <div className="pt-4 text-center text-sm text-[var(--color-text-secondary)]">
            {loadMoreError && (
              <button
                type="button"
                className="cursor-pointer hover:text-[var(--color-text-primary)]"
                onClick={loadMore}
              >
                加载失败，点击重试
              </button>
            )}
            {!loadMoreError && !hasMore && (
              <span>{limitReached ? `最多展示 ${SEARCH_MAX_ITEMS} 条` : '没有更多了'}</span>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
