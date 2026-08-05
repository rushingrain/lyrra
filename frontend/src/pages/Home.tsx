import { useEffect, useState } from 'react'
import { BannerCarousel } from '@/components/BannerCarousel'
import { PlaylistRecommend } from '@/components/PlaylistRecommend'
import { SongRecommend } from '@/components/SongRecommend'
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
  const [songs, setSongs] = useState<Song[]>([])
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
      if (!cancelled) setSongs(data)
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

  return (
    <>
      <BannerCarousel banners={banners} />
      <div className="mt-10">
        <PlaylistRecommend playlists={playlists} />
      </div>
      <SongRecommend songs={songs} title="最新音乐" />
      <SongRecommend songs={hotSongs} title="热门音乐" />
    </>
  )
}
