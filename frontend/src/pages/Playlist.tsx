import { useLocation, useParams } from 'react-router-dom'
import { SongCollectionDetail } from '@/components/business/SongCollectionDetail'
import { usePlaylistSongs } from '@/hooks/usePlaylistSongs'
import { getAllPlaylistSongs } from '@/services/playlist'
import type { Playlist as PlaylistType } from '@/types'

/**
 * 歌单详情页：左侧封面（含播放次数），右侧标题/描述/作者信息与播放全部，
 * 下方是支持滚动分页的歌曲列表。
 */
export default function Playlist() {
  const { globalId = '' } = useParams<{ globalId: string }>()
  const location = useLocation()
  const statePlaylist = (location.state as { playlist?: PlaylistType } | null)?.playlist
  const { info, songs, hasMore, loadError, loadMoreError, sentinelRef, loadMore } =
    usePlaylistSongs(globalId)

  // 接口返回的 list_info 为准；拿不到有效信息时回退首页传入的歌单
  const playlist = info?.name ? info : statePlaylist

  if (loadError) {
    return <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">{loadError}</p>
  }

  return (
    <SongCollectionDetail
      coverUrl={playlist?.coverUrl}
      title={playlist?.name}
      description={playlist?.description}
      author={playlist?.creator}
      authorAvatar={playlist?.creatorAvatar}
      tags={playlist?.tags}
      publishTime={playlist?.publishTime}
      publishLabel="创建"
      playCount={statePlaylist?.playCount ?? 0}
      songs={songs}
      hasMore={hasMore}
      loadMoreError={loadMoreError}
      sentinelRef={sentinelRef}
      loadMore={loadMore}
      loadAllSongs={() => getAllPlaylistSongs(globalId)}
      loadAllErrorMessage="歌单歌曲加载失败，请稍后重试"
    />
  )
}
