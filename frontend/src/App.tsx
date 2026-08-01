import { useCallback } from 'react'
import { MainLayout } from './components/layout/MainLayout'
import { BannerCarousel } from './components/home/BannerCarousel'
import { PlaylistRecommend } from './components/home/PlaylistRecommend'
import { SongList } from './components/home/SongList'
import { mockBanners, mockPlaylists, mockSongs } from './data/mock'

function App() {
  // 一页项数变化时通知父级，便于向后端请求对应数量的官方歌单
  const handlePlaylistPageSizeChange = useCallback((count: number) => {
    // TODO: 根据 count 调用后端接口请求对应数量的官方歌单
    console.log('[PlaylistRecommend] 一页项数:', count)
  }, [])

  // 最新音乐一页项数变化时通知父级，便于向后端请求对应数量的歌曲
  const handleSongPageSizeChange = useCallback((count: number) => {
    // TODO: 根据 count 调用后端接口请求对应数量的歌曲
    console.log('[SongList] 一页项数:', count)
  }, [])

  return (
    <MainLayout>
      <BannerCarousel banners={mockBanners} />
      <div className="mt-10">
        <PlaylistRecommend
          playlists={mockPlaylists}
          onPageSizeChange={handlePlaylistPageSizeChange}
        />
      </div>
      {/* 最新音乐 */}
      <SongList title="最新音乐" songs={mockSongs} onPageSizeChange={handleSongPageSizeChange} />
      {/* 热门音乐 */}
      <SongList title="热门音乐" songs={mockSongs} onPageSizeChange={handleSongPageSizeChange} />
    </MainLayout>
  )
}

export default App
