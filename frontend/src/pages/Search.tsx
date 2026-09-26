import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import toast from 'react-hot-toast'
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
import {
  EMPTY_RESULT,
  getSearch,
  type SearchResult,
  type SearchTabType,
} from '@/services/search'
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
      return (
        <ResultTable
          columns={lyricColumns}
          rows={result.lyrics}
          rowKey={(lyric, index) => lyric.id || `${lyric.title}-${lyric.artist}-${index}`}
        />
      )
  }
}

export default function Search() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q')?.trim() || ''
  const [activeTab, setActiveTab] = useState<SearchTab>('complex')
  const [result, setResult] = useState<SearchResult>(EMPTY_RESULT)
  const [isLoading, setIsLoading] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)

  useEffect(() => {
    if (query) setActiveTab('complex')
  }, [query])

  useEffect(() => {
    if (!query) return

    let cancelled = false
    setIsLoading(true)
    setLoadError(null)
    setResult(EMPTY_RESULT)

    getSearch(query, activeTab)
      .then((data) => {
        if (!cancelled) setResult(data)
      })
      .catch(() => {
        if (cancelled) return
        const message = '搜索失败，请稍后重试'
        setLoadError(message)
        toast.error(message)
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [query, activeTab])

  if (!query) {
    return (
      <div className="flex h-64 items-center justify-center text-[var(--color-text-secondary)]">
        <p className="text-lg">请输入搜索内容</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      <p className="mb-4 text-xl font-bold text-[var(--color-text-primary)]">{query}</p>

      <div className="mb-4 flex gap-6">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            type="button"
            className={`cursor-pointer border-b-3 text-base transition-[color,border-color] ${activeTab === tab.key
                ? 'border-b-(--color-accent) font-bold text-(--color-text-primary)'
                : 'border-b-transparent font-normal text-(--color-text-secondary) hover:text-(--color-text-primary)'
              }`}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {isLoading && (
        <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">正在搜索...</p>
      )}

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
        <div className="pb-8">
          <TabSearchResults tab={activeTab} result={result} />
        </div>
      )}
    </div>
  )
}
