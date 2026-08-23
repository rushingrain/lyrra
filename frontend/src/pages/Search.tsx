import { useEffect, useState, type ReactNode } from 'react'
import { useSearchParams } from 'react-router-dom'
import { ChevronRight } from 'lucide-react'
import toast from 'react-hot-toast'
import { Playlist } from '@/components/Playlist'
import {
  EMPTY_RESULT,
  getSearch,
  type ComplexSearchAlbum,
  type ComplexSearchArtist,
  type ComplexSearchMv,
  type ComplexSearchSong,
  type SearchResult,
  type SearchTabType,
} from '@/services/search'
import type { Playlist as PlaylistData } from '@/types'

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

function SectionHeader({ title }: { title: string }) {
  return (
    <div className="mb-3 flex items-center gap-1 font-bold text-lg text-[var(--color-text-primary)]">
      {title}
      <ChevronRight size={20} />
    </div>
  )
}

function ResponsiveCardGrid({ children }: { children: ReactNode }) {
  return <div className="grid grid-cols-4 gap-4 lg:grid-cols-5 2xl:grid-cols-6">{children}</div>
}

function ArtistSummaryCard({ artist }: { artist: ComplexSearchArtist }) {
  return (
    <div className="flex min-w-0 items-center gap-4 p-4">
      <img src={artist.avatarUrl ? artist.avatarUrl : 'src/assets/images/default_cover.jpg'} alt={artist.name} className="h-20 w-20 shrink-0 rounded-full object-cover" loading="lazy" />
      <div className="min-w-0">
        <p className="truncate font-bold text-[var(--color-text-primary)]">歌手：{artist.name}</p>
        <p className="mt-1 flex items-center gap-4 truncate text-sm text-[var(--color-text-secondary)]">
          <span className="shrink-0">单曲: {artist.songCount}</span>
          <span className="shrink-0">粉丝: {formatCount(artist.fans)}</span>
        </p>
      </div>
    </div>
  )
}

function AlbumSummaryCard({ album }: { album: ComplexSearchAlbum }) {
  return (
    <div className="flex min-w-0 items-center gap-4 p-4">
      <img src={album.coverUrl} alt={album.name} className="h-20 w-20 shrink-0 rounded-lg object-cover" loading="lazy" />
      <div className="min-w-0">
        <p className="truncate font-bold text-[var(--color-text-primary)]">专辑：{album.name}</p>
        <p className="mt-1 flex items-center gap-4 truncate text-sm text-[var(--color-text-secondary)]">
          <span className="shrink-0">{album.artist}</span>
          <span className="shrink-0">{album.date}</span>
        </p>
      </div>
    </div>
  )
}

function ComplexSongCard({ song }: { song: ComplexSearchSong }) {
  return (
    <div className="flex min-w-0 items-center gap-4 rounded-lg">
      <img src={song.coverUrl ? song.coverUrl : 'src/assets/images/default_cover.jpg'} alt={song.title} className="h-16 w-16 shrink-0 rounded-lg object-cover" loading="lazy" />
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">{song.title}</p>
        <div className="mt-1.5 flex items-center gap-1.5 text-xs text-[var(--color-text-secondary)]">
          {song.quality && (
            <span className="shrink-0 rounded-xs border border-[var(--color-accent)] px-0.5 text-[8px] leading-relaxed text-[var(--color-accent)]">
              {song.quality}
            </span>
          )}
          {song.isOriginal && (
            <span className="shrink-0 rounded-xs bg-[var(--color-accent)] px-0.5 text-[8px] leading-relaxed text-[var(--color-on-accent)]">原唱</span>
          )}
          <span className="min-w-0 truncate text-xs">{song.artist}</span>
        </div>
        {song.tags.length > 0 && (
          <div className="mt-1.5 flex flex-wrap items-center gap-1.5">
            {song.tags.map((tag) => (
              <span key={tag} className="shrink-0 rounded bg-[var(--color-accent-bg)] px-1 py-0.5 text-[8px] leading-relaxed text-[var(--color-accent)]">
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function AlbumCard({ album }: { album: ComplexSearchAlbum }) {
  return (
    <div className="min-w-0">
      <img src={album.coverUrl} alt={album.name} className="aspect-square w-4/5 rounded-lg object-cover" loading="lazy" />
      <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{album.name}</p>
      <p className="mt-0.5 truncate text-xs text-[var(--color-text-secondary)]">{album.artist} · {album.date}</p>
    </div>
  )
}

function ArtistCard({ artist }: { artist: ComplexSearchArtist }) {
  return (
    <div className="min-w-0 text-center">
      <img src={artist.avatarUrl ? artist.avatarUrl : 'src/assets/images/default_cover.jpg'} alt={artist.name} className="mx-auto aspect-square w-4/5 rounded-full object-cover" loading="lazy" />
      <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{artist.name}</p>
      <p className="mt-0.5 text-xs text-[var(--color-text-secondary)]">单曲: {artist.songCount}</p>
    </div>
  )
}

function MvCard({ mv }: { mv: ComplexSearchMv }) {
  return (
    <div className="min-w-0">
      <img src={mv.coverUrl} alt={mv.name} className="aspect-video w-full rounded-lg object-cover" loading="lazy" />
      <p className="mt-2 truncate text-sm font-medium text-[var(--color-text-primary)]">{mv.name}</p>
      <p className="mt-0.5 truncate text-xs text-[var(--color-text-secondary)]">{mv.artist}</p>
    </div>
  )
}

function toPlaylist(item: SearchResult['playlists'][number]): PlaylistData {
  return {
    id: item.id,
    name: item.name,
    coverUrl: item.coverUrl,
    description: item.description,
    playCount: item.playCount,
  }
}

function hasResults(result: SearchResult): boolean {
  return Object.values(result).some((items) => items.length > 0)
}

function TabSearchResults({ tab, result }: { tab: Exclude<SearchTab, 'complex'>; result: SearchResult }) {
  switch (tab) {
    case 'song':
      return <div className="grid grid-cols-1 gap-x-3 gap-y-5 lg:grid-cols-2">{result.songs.map((song) => <ComplexSongCard key={song.id} song={song} />)}</div>
    case 'special':
      return <ResponsiveCardGrid>{result.playlists.map((playlist) => <Playlist key={playlist.id} playlist={toPlaylist(playlist)} />)}</ResponsiveCardGrid>
    case 'album':
      return <ResponsiveCardGrid>{result.albums.map((album) => <AlbumCard key={album.id} album={album} />)}</ResponsiveCardGrid>
    case 'author':
      return <ResponsiveCardGrid>{result.artists.map((artist) => <ArtistCard key={artist.id} artist={artist} />)}</ResponsiveCardGrid>
    case 'mv':
      return <ResponsiveCardGrid>{result.mvs.map((mv) => <MvCard key={mv.id} mv={mv} />)}</ResponsiveCardGrid>
    case 'lyric':
      return (
        <div className="divide-y divide-[var(--color-border)]">
          {result.lyrics.map((lyric) => (
            <div key={lyric.id || `${lyric.title}-${lyric.artist}`} className="flex min-w-0 items-center gap-4 py-3">
              <p className="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text-primary)]">{lyric.title}</p>
              <p className="w-1/4 truncate text-sm text-[var(--color-text-secondary)]">{lyric.artist}</p>
              <p className="w-1/4 truncate text-sm text-[var(--color-text-secondary)]">{lyric.album}</p>
            </div>
          ))}
        </div>
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
            className={`cursor-pointer text-base transition-[color] ${activeTab === tab.key
              ? 'border-b-3 border-b-[color:var(--color-accent)] font-bold text-[var(--color-text-primary)]'
              : 'font-normal text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'}`}
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
              {result.artists[0] && <ArtistSummaryCard artist={result.artists[0]} />}
              {result.albums[0] && <AlbumSummaryCard album={result.albums[0]} />}
            </div>
          )}

          {result.songs.length > 0 && (
            <section>
              <SectionHeader title="单曲" />
              <div className="grid grid-cols-1 gap-x-3 gap-y-5 lg:grid-cols-2">
                {result.songs.slice(0, 6).map((song) => <ComplexSongCard key={song.id} song={song} />)}
              </div>
            </section>
          )}

          {result.playlists.length > 0 && (
            <section>
              <SectionHeader title="歌单" />
              <ResponsiveCardGrid>
                {result.playlists.map((playlist) => <Playlist key={playlist.id} playlist={toPlaylist(playlist)} />)}
              </ResponsiveCardGrid>
            </section>
          )}

          {result.albums.length > 0 && (
            <section>
              <SectionHeader title="专辑" />
              <ResponsiveCardGrid>
                {result.albums.map((album) => <AlbumCard key={album.id} album={album} />)}
              </ResponsiveCardGrid>
            </section>
          )}

          {result.artists.length > 0 && (
            <section>
              <SectionHeader title="歌手" />
              <ResponsiveCardGrid>
                {result.artists.map((artist) => <ArtistCard key={artist.id} artist={artist} />)}
              </ResponsiveCardGrid>
            </section>
          )}

          {result.mvs.length > 0 && (
            <section>
              <SectionHeader title="MV" />
              <ResponsiveCardGrid>
                {result.mvs.map((mv) => <MvCard key={mv.id} mv={mv} />)}
              </ResponsiveCardGrid>
            </section>
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
