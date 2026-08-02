import { useCallback } from 'react'
import { BannerCarousel } from '@/components/BannerCarousel'
import { PlaylistRecommend } from '@/components/PlaylistRecommend'
import { SongList } from '@/components/SongList'
import { mockBanners, mockPlaylists, mockSongs } from '@/data/mock'

export default function Home() {
  const handlePlaylistPageSizeChange = useCallback((count: number) => {
    console.log('[PlaylistRecommend] 一页项数:', count)
  }, [])

  const handleSongPageSizeChange = useCallback((count: number) => {
    console.log('[SongList] 一页项数:', count)
  }, [])

  return (
    <>
      <BannerCarousel banners={mockBanners} />
      <div className="mt-10">
        <PlaylistRecommend
          playlists={mockPlaylists}
          onPageSizeChange={handlePlaylistPageSizeChange}
        />
      </div>
      <SongList title="最新音乐" songs={mockSongs} onPageSizeChange={handleSongPageSizeChange} />
      <SongList title="热门音乐" songs={mockSongs} onPageSizeChange={handleSongPageSizeChange} />
    </>
  )
}
