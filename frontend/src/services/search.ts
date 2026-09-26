import { Search, SearchComplex } from '../../wailsjs/go/search/Service'
import { search } from '../../wailsjs/go/models'

export type SearchLyric = search.Lyric
export type SearchResult = search.Result

export type NormalSearchType = 'song' | 'special' | 'album' | 'author' | 'mv' | 'lyric'
export type SearchTabType = 'complex' | NormalSearchType

const EMPTY_RESULT = new search.Result({
  songs: [],
  playlists: [],
  albums: [],
  artists: [],
  mvs: [],
  lyrics: [],
})

const CACHE_LIMIT = 20
const cache = new Map<string, Promise<SearchResult>>()

function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.search?.Service)
}

/** 获取综合搜索结果，后端已返回稳定 DTO。 */
export function getComplexSearch(keyword: string, pageNum = 1, pageSize = 10): Promise<SearchResult> {
  if (!isBackendAvailable()) return Promise.reject(new Error('搜索服务不可用'))

  const key = `complex:${keyword}:${pageNum}:${pageSize}`
  const cached = cache.get(key)
  if (cached) return cached

  const request = SearchComplex(keyword, pageNum, pageSize).catch((error) => {
    cache.delete(key)
    throw error
  })

  cache.set(key, request)
  if (cache.size > CACHE_LIMIT) {
    const oldestKey = cache.keys().next().value
    if (oldestKey) cache.delete(oldestKey)
  }
  return request
}

/** 根据搜索分类获取结果，后端已返回稳定 DTO。 */
export function getSearch(keyword: string, searchType: SearchTabType, pageNum = 1, pageSize = 10): Promise<SearchResult> {
  if (searchType === 'complex') return getComplexSearch(keyword, pageNum, pageSize)
  if (!isBackendAvailable()) return Promise.reject(new Error('搜索服务不可用'))

  const key = `${searchType}:${keyword}:${pageNum}:${pageSize}`
  const cached = cache.get(key)
  if (cached) return cached

  const request = Search(keyword, pageNum, pageSize, searchType).catch((error) => {
    cache.delete(key)
    throw error
  })

  cache.set(key, request)
  if (cache.size > CACHE_LIMIT) {
    const oldestKey = cache.keys().next().value
    if (oldestKey) cache.delete(oldestKey)
  }
  return request
}

export { EMPTY_RESULT }
