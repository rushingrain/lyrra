import { useEffect, useState } from 'react'
import { BannerCarousel } from '@/components/common/BannerCarousel'
import { PlaylistList } from '@/components/common/PlaylistList'
import { Section } from '@/components/common/Section'
import { SongGrid } from '@/components/common/SongGrid'
import { useAuthStore } from '@/stores/authStore'
import type { Banner, Playlist, Song } from '@/types'
import { getHomeBanners } from '@/services/banner'
import {
  DEFAULT_PLAYLIST_WITH_SONG,
  GetHotPlaylists,
  OFFICIAL_PLAYLIST_PAGE_SIZE,
} from '@/services/playlist'
import {
  DEFAULT_SONG_PAGE,
  DEFAULT_SONG_PAGE_SIZE,
  getDailyRecommendSongs,
  getNewSongs,
  getHotSongs,
} from '@/services/song'

/** 首页只展示统一歌单数据的前 15 条，列表页展示完整结果 */
const MAX_PLAYLISTS = OFFICIAL_PLAYLIST_PAGE_SIZE

/** 首页歌曲板块固定展示 3 行（窄窗口 2 列 / 宽窗口 3 列） */
const HOME_SONG_ROWS = 3

export default function Home() {
  // 登录态：每日推荐属于用户级推荐，未登录时不展示、不请求
  const user = useAuthStore((s) => s.user)
  // 首页 banner：挂载时从后端拉取真实数据，失败时服务层回退（空数组/纯浏览器 mock）
  const [banners, setBanners] = useState<Banner[]>([])
  // 首页官方歌单：同上，服务层缓存 1h，导航切换不重复请求
  const [playlists, setPlaylists] = useState<Playlist[]>([])
  // 首页最新音乐：真实 TopSong 数据，服务层缓存 1h
  const [newSongs, setNewSongs] = useState<Song[]>([])
  // 首页热门音乐：真实 TopCardYouth 数据，服务层缓存 1h
  const [hotSongs, setHotSongs] = useState<Song[]>([])
  // 首页每日推荐：按登录账号返回，登录态变化时重新拉取
  const [dailySongs, setDailySongs] = useState<Song[]>([])

  useEffect(() => {
    if (!user) {
      setDailySongs([])
      return
    }
    let cancelled = false
    getDailyRecommendSongs().then((data) => {
      if (!cancelled) setDailySongs(data)
    })
    return () => {
      cancelled = true
    }
  }, [user])

  useEffect(() => {
    let cancelled = false
    getHomeBanners().then((data) => {
      if (!cancelled) setBanners(data)
    })
    GetHotPlaylists(OFFICIAL_PLAYLIST_PAGE_SIZE, DEFAULT_PLAYLIST_WITH_SONG).then((data) => {
      if (!cancelled) setPlaylists(data.slice(0, MAX_PLAYLISTS))
    })
    // 显式携带 page=1、pagesize 请求最新音乐
    getNewSongs(DEFAULT_SONG_PAGE, DEFAULT_SONG_PAGE_SIZE).then((data) => {
      if (!cancelled) setNewSongs(data)
    })
    // 热门音乐：TopCardYouth 固定返回一批，无分页参数。
    // 服务层缓存保留全量数据，页面只渲染前 3 行。
    getHotSongs().then((data) => {
      if (!cancelled) setHotSongs(data)
    })
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="flex flex-col gap-8 pb-4">
      <Section title="精选活动" clickable={false}>
        <BannerCarousel banners={banners} />
      </Section>
      <Section title="歌单推荐" to="/playlists">
        <PlaylistList minWidth={116} playlists={playlists} />
      </Section>
      {user && (
        <Section title="每日推荐" to="/songs/daily">
          <SongGrid songs={dailySongs} columns="responsive" rows={HOME_SONG_ROWS} />
        </Section>
      )}
      <Section title="最新音乐" to="/songs/new">
        <SongGrid songs={newSongs} columns="responsive" rows={HOME_SONG_ROWS} />
      </Section>
      <Section title="热门音乐" to="/songs/hot">
        <SongGrid songs={hotSongs} columns="responsive" rows={HOME_SONG_ROWS} />
      </Section>
    </div>
  )
}
