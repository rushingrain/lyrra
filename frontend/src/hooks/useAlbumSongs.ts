import { usePaginatedSongs, type PaginatedSongsState } from '@/hooks/usePaginatedSongs'
import { ALBUM_PAGE_SIZE, getAlbumSongs } from '@/services/album'

export type AlbumSongsState = PaginatedSongsState<never>

/** 专辑歌曲分页；与歌单共用分页 hook。 */
export function useAlbumSongs(albumId: string): AlbumSongsState {
  return usePaginatedSongs<never>({
    collectionId: albumId,
    pageSize: ALBUM_PAGE_SIZE,
    fetchPage: getAlbumSongs,
    firstPageError: '专辑歌曲加载失败，请稍后重试',
    nextPageError: '加载更多失败，请重试',
  })
}
