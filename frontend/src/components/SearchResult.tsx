import { useState } from 'react'
import { Heart } from 'lucide-react'
import type { Song } from '@/types'
import { mockSongs } from '@/data/mock'

type SearchTab = 'song' | 'playlist' | 'album' | 'artist' | 'user'

interface SearchResultsProps {
  query: string
  onPageSizeChange?: (count: number) => void
}

function formatDuration(seconds: number): string {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

const TABS: { key: SearchTab; label: string }[] = [
  { key: 'song', label: '单曲' },
  { key: 'playlist', label: '歌单' },
  { key: 'album', label: '专辑' },
  { key: 'artist', label: '歌手' },
  { key: 'user', label: '用户' },
]

export function SearchResult({ query, onPageSizeChange }: SearchResultsProps) {
  const [activeTab, setActiveTab] = useState<SearchTab>('song')
  const [liked, setLiked] = useState<Set<string>>(new Set())

  const results: Song[] = mockSongs.filter(
    (s) =>
      s.title.includes(query) ||
      s.artist.includes(query) ||
      s.album.includes(query),
  )

  const toggleLike = (id: string) => {
    setLiked((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  return (
    <div className="flex flex-col">
      <p className="mb-4 text-xl font-bold text-[var(--color-text-primary)]">
        {query}
      </p>

      <div className="mb-4 flex gap-6">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            className={`cursor-pointer text-base transition-colors ${activeTab === tab.key
              ? 'font-bold text-[var(--color-text-primary)] border-b-3 border-b-[color:var(--color-accent)]'
              : 'font-normal text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'
              }`}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {activeTab === 'song' && results.length > 0 && (
        <div>
          <div className="mb-1 flex items-center gap-2 px-2 text-xs text-[var(--color-text-secondary)]">
            <span className="w-8 shrink-0 text-start">#</span>
            <span className="min-w-0 flex-1 text-start">标题</span>
            <span className="w-[140px] shrink-0 text-start">专辑</span>
            <span className="w-10 shrink-0 text-start">喜欢</span>
            <span className="w-14 shrink-0 text-start">时长</span>
          </div>

          {results.map((song, i) => (
            <div
              key={song.id}
              className="flex items-center gap-2 rounded-lg px-2 py-2 transition-colors hover:bg-gray-50"
            >
              <span className="w-8 shrink-0 text-start text-sm tabular-nums text-[var(--color-text-secondary)]">
                {i + 1}
              </span>

              <div className="flex min-w-0 flex-1 cursor-pointer items-center gap-3">
                <img
                  src={song.coverUrl}
                  alt={song.title}
                  className="h-11 w-11 shrink-0 rounded-lg object-cover"
                  loading="lazy"
                />
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">
                    {song.title}
                  </p>
                  <div className="mt-0.5 flex items-center gap-1.5 text-xs text-[var(--color-text-secondary)]">
                    {song.quality && (
                      <span className="shrink-0 rounded border border-[var(--color-accent)] px-1 text-[10px] leading-relaxed text-[var(--color-accent)]">
                        {song.quality}
                      </span>
                    )}
                    {song.isOriginal && (
                      <span className="shrink-0 rounded bg-[var(--color-accent)] px-1 text-[10px] leading-relaxed text-[var(--color-on-accent)]">
                        原唱
                      </span>
                    )}
                    <span className="truncate">{song.artist}</span>
                  </div>
                </div>
              </div>

              <span className="w-[140px] shrink-0 truncate text-sm text-[var(--color-text-secondary)]">
                {song.album}
              </span>

              <button
                className="flex w-10 shrink-0 cursor-pointer items-center justify-start"
                onClick={() => toggleLike(song.id)}
                aria-label={liked.has(song.id) ? '取消喜欢' : '喜欢'}
              >
                <Heart
                  size={16}
                  className={liked.has(song.id) ? 'fill-red-500 text-red-500' : 'text-[var(--color-text-secondary)]'}
                />
              </button>

              <span className="w-14 shrink-0 text-start text-sm tabular-nums text-[var(--color-text-secondary)]">
                {formatDuration(song.duration)}
              </span>
            </div>
          ))}
        </div>
      )}

      {activeTab === 'song' && results.length === 0 && (
        <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">
          未找到相关歌曲
        </p>
      )}

      {activeTab !== 'song' && (
        <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">
          暂无{TABS.find((t) => t.key === activeTab)?.label}搜索结果
        </p>
      )}
    </div>
  )
}
