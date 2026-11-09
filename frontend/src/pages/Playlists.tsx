import { useEffect, useState } from 'react'
import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import defaultCover from '@/assets/images/default_cover.jpg'
import { PlaylistPlayCount } from '@/components/common/PlaylistPlayCount'
import { useCoverColor } from '@/hooks/useCoverColor'
import {
  DEFAULT_PLAYLIST_WITH_SONG,
  GetHotPlaylists,
  OFFICIAL_PLAYLIST_PAGE_SIZE,
} from '@/services/playlist'
import type { Playlist } from '@/types'

interface PlaylistCardProps {
  playlist: Playlist
}

function PlaylistCard({ playlist }: PlaylistCardProps) {
  const coverUrl = playlist.coverUrl || defaultCover
  const backgroundColor = useCoverColor(coverUrl)

  return (
    <Link
      to={`/songs/playlist/${encodeURIComponent(playlist.globalCollectionId)}`}
      state={{ playlist }}
      className="@container group flex min-w-0 flex-col overflow-hidden rounded-md bg-[var(--color-surface)] shadow-sm transition-shadow duration-200 hover:shadow-md"
    >
      <div className="relative aspect-square w-full overflow-hidden">
        <img
          src={coverUrl}
          alt={playlist.name}
          className="h-full w-full object-cover"
          loading="lazy"
        />
        <PlaylistPlayCount
          count={playlist.playCount}
          className="absolute right-2 top-2"
        />
      </div>

      <div
        className="flex aspect-[3/1] w-full items-center overflow-hidden transition-colors duration-300"
        style={{ backgroundColor }}
      >
        <p className="line-clamp-2 w-full px-[6cqi] text-[clamp(14px,6.5cqi,15px)] font-medium leading-[1.3] text-white">
          {playlist.name}
        </p>
      </div>
    </Link>
  )
}

/**
 * 歌单推荐列表页：多行展示歌单卡片，卡片点击进入对应歌单的歌曲详情页。
 */
export default function Playlists() {
  const [playlists, setPlaylists] = useState<Playlist[]>([])

  useEffect(() => {
    let cancelled = false
    GetHotPlaylists(OFFICIAL_PLAYLIST_PAGE_SIZE, DEFAULT_PLAYLIST_WITH_SONG).then((data) => {
      if (!cancelled) setPlaylists(data)
    })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-2xl font-semibold text-[var(--color-text-primary)]">歌单推荐</h1>

      {playlists.length > 0 && (
        <div className="@container">
          <div
            className="grid grid-cols-[repeat(auto-fill,minmax(var(--card-w),1fr))] gap-6"
            style={{ '--card-w': 'clamp(120px, 11cqi, 180px)' } as CSSProperties}
          >
            {playlists.map((playlist) => (
              <PlaylistCard key={playlist.id} playlist={playlist} />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
