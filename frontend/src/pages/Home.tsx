import { useEffect, useState } from 'react'
import { BannerCarousel } from '@/components/common/BannerCarousel'
import { PlaylistList } from '@/components/common/PlaylistList'
import { ChevronRight } from 'lucide-react'
import { SongList } from '@/components/common/SongList'
import type { Banner, Playlist, Song } from '@/types'
import { getHomeBanners } from '@/services/banner'
import {
  DEFAULT_PLAYLIST_PAGE_SIZE,
  DEFAULT_PLAYLIST_WITH_SONG,
  getHomePlaylists,
} from '@/services/playlist'
import {
  DEFAULT_SONG_PAGE,
  DEFAULT_SONG_PAGE_SIZE,
  getNewSongs,
  getHotSongs,
} from '@/services/song'

/** 官方歌单最多展示条数：与请求的 pageSize 一致，后端返回过多时截断，避免页面过长 */
const MAX_PLAYLISTS = DEFAULT_PLAYLIST_PAGE_SIZE

export default function Home() {
  // 首页 banner：挂载时从后端拉取真实数据，失败时服务层回退（空数组/纯浏览器 mock）
  const [banners, setBanners] = useState<Banner[]>([])
  // 首页官方歌单：同上，服务层缓存 1h，导航切换不重复请求
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  // 首页最新音乐：真实 TopSong 数据，服务层缓存 1h
  const [newSongs, setNewSongs] = useState<Song[]>([])
  // 首页热门音乐：真实 TopCard 数据，服务层缓存 1h
  const [hotSongs, setHotSongs] = useState<Song[]>([])

  useEffect(() => {
    let cancelled = false
    getHomeBanners().then((data) => {
      if (!cancelled) setBanners(data)
    })
    getHomePlaylists(DEFAULT_PLAYLIST_PAGE_SIZE, DEFAULT_PLAYLIST_WITH_SONG).then((data) => {
      if (!cancelled) setPlaylists(data.slice(0, MAX_PLAYLISTS))
    })
    // 显式携带 page=1、pagesize=15 请求最新音乐
    getNewSongs(DEFAULT_SONG_PAGE, DEFAULT_SONG_PAGE_SIZE).then((data) => {
      if (!cancelled) setNewSongs(data)
    })
    // 热门音乐：TopCard 固定返回一批，无分页参数。
    // 后端可能返回多条：服务层缓存仍保留全量数据，
    // 页面只渲染前 DEFAULT_SONG_PAGE_SIZE 条，避免卡片过多。
    getHotSongs().then((data) => {
      if (!cancelled) setHotSongs(data.slice(0, DEFAULT_SONG_PAGE_SIZE))
    })
    return () => {
      cancelled = true
    }
  }, [])

  // clickable=false 时标题为纯文本,不带箭头、不可点击/跳转(如「精选活动」)
  const renderSection = (
    title: string,
    children: React.ReactNode,
    clickable = true
  ) => (
    <div>
      <div className="mb-4 flex items-center justify-between">
        {clickable ? (
          <a
            href="#"
            className="flex cursor-pointer items-center font-bold text-lg text-[var(--color-text-primary)] transition-colors hover:opacity-80"
            onClick={(e) => e.preventDefault()}
          >
            {title}
            <ChevronRight size={20} />
          </a>
        ) : (
          <h2 className="font-bold text-lg text-[var(--color-text-primary)]">
            {title}
          </h2>
        )}
      </div>
      {children}
    </div>
  )

  return (
    <div>
      {renderSection('精选活动', <BannerCarousel banners={banners} />, false)}
      {renderSection('官方歌单', <PlaylistList minWidth={104} playlists={playlists} />)}
      {renderSection('最新音乐', <SongList songs={newSongs} />)}
      {renderSection('热门音乐', <SongList songs={hotSongs} />)}
    </div>
  )
}
