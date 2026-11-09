import type { Album, Song } from '@/types'
import {
  GetAllAlbumSongs as GetAllAlbumSongsAPI,
  GetAlbumDetail as GetAlbumDetailAPI,
  GetAlbumSongs as GetAlbumSongsAPI,
} from '../../wailsjs/go/album/Service'

/** 专辑歌曲分页大小：与后端默认页大小一致 */
export const ALBUM_PAGE_SIZE = 30

export interface AlbumSongsPage {
  songs: Song[]
  total: number
  hasMore: boolean
}

function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.album?.Service)
}

const detailInflight = new Map<string, Promise<Album>>()

/** 获取专辑详情；同一专辑的并发请求共用一个 Promise。 */
export function getAlbumDetail(albumId: string): Promise<Album> {
  if (!isBackendAvailable()) return Promise.reject(new Error('专辑服务不可用'))

  const existing = detailInflight.get(albumId)
  if (existing) return existing

  const request = GetAlbumDetailAPI(albumId).finally(() => detailInflight.delete(albumId))
  detailInflight.set(albumId, request)
  return request
}

const songsInflight = new Map<string, Promise<AlbumSongsPage>>()

/** 获取专辑歌曲（分页）；只合并同一页的并发请求，不缓存已完成结果。 */
export function getAlbumSongs(
  albumId: string,
  page = 1,
  pageSize = ALBUM_PAGE_SIZE,
): Promise<AlbumSongsPage> {
  if (!isBackendAvailable()) return Promise.reject(new Error('专辑服务不可用'))

  const key = `${albumId}:${page}:${pageSize}`
  const existing = songsInflight.get(key)
  if (existing) return existing

  const request = GetAlbumSongsAPI(albumId, page, pageSize).finally(() => songsInflight.delete(key))
  songsInflight.set(key, request)
  return request
}

/** 获取专辑全部歌曲，供「播放全部」使用。 */
export function getAllAlbumSongs(albumId: string): Promise<Song[]> {
  if (!isBackendAvailable()) return Promise.reject(new Error('专辑服务不可用'))
  return GetAllAlbumSongsAPI(albumId)
}
