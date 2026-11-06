import { usePaginatedSongs, type PaginatedSongsState } from '@/hooks/usePaginatedSongs'
import { PLAYLIST_PAGE_SIZE, getPlaylistSongs } from '@/services/playlist'
import type { Playlist } from '@/types'

export type PlaylistSongsState = PaginatedSongsState<Playlist>

/**
 * 歌单歌曲分页：首屏取第一页，滚动到底部自动追加下一页。
 *
 * 与搜索分页相同的处理：请求代次丢弃过期响应、在途请求去重、按歌曲 id 去重；
 * 追加失败时保留已加载内容，并停止自动加载，等用户点重试。
 */
export function usePlaylistSongs(globalCollectionId: string): PlaylistSongsState {
  return usePaginatedSongs<Playlist>({
    collectionId: globalCollectionId,
    pageSize: PLAYLIST_PAGE_SIZE,
    fetchPage: getPlaylistSongs,
    firstPageError: '歌单歌曲加载失败，请稍后重试',
    nextPageError: '加载更多失败，请重试',
  })
}
