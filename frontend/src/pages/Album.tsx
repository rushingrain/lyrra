import { useLocation, useParams } from 'react-router-dom'
import { SongCollectionDetail } from '@/components/business/SongCollectionDetail'
import { useAlbumDetail } from '@/hooks/useAlbumDetail'
import { useAlbumSongs } from '@/hooks/useAlbumSongs'
import { getAllAlbumSongs } from '@/services/album'
import type { Album as AlbumType } from '@/types'

/** 专辑详情页：与歌单详情共用布局、播放全部和分页歌曲列表。 */
export default function Album() {
  const { albumId = '' } = useParams<{ albumId: string }>()
  const location = useLocation()
  const stateAlbum = (location.state as { album?: AlbumType } | null)?.album
  const detail = useAlbumDetail(albumId)
  const { songs, hasMore, loadError, loadMoreError, sentinelRef, loadMore } = useAlbumSongs(albumId)

  const album = detail?.name ? detail : stateAlbum
  const albumDetails = [
    album?.type ? `类型：${album.type}` : '',
    album?.language ? `语言：${album.language}` : '',
  ].filter(Boolean)

  if (loadError) {
    return <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">{loadError}</p>
  }

  return (
    <SongCollectionDetail
      coverUrl={album?.coverUrl}
      title={album?.name}
      description={album?.description}
      author={album?.artist}
      details={albumDetails}
      publishTime={album?.date}
      publishLabel="发行"
      songs={songs}
      hasMore={hasMore}
      loadMoreError={loadMoreError}
      sentinelRef={sentinelRef}
      loadMore={loadMore}
      loadAllSongs={() => getAllAlbumSongs(albumId)}
      loadAllErrorMessage="专辑歌曲加载失败，请稍后重试"
    />
  )
}
