import { Search, SearchComplex } from '../../wailsjs/go/search/Service'
import { search } from '../../wailsjs/go/models'

export type SearchResult = search.Result

export type NormalSearchType = 'song' | 'special' | 'album' | 'author' | 'mv' | 'lyric'
export type SearchTabType = 'complex' | NormalSearchType

const EMPTY_RESULT = new search.Result({
  songs: [],
  playlists: [],
  albums: [],
  artists: [],
  mvs: [],
})

/** 分页大小：后端已支持任意 pageSize，前后端统一用 10 */
export const SEARCH_PAGE_SIZE = 10
/** 无限滚动时前端最多展示的条数：到了就停止加载，避免无节制地占用内存 */
export const SEARCH_MAX_ITEMS = 100

/**
 * 在途请求表：同一请求并发时复用同一个 Promise，请求结束立即移除。
 *
 * 这里只做「同一请求的并发合并」（例如开发模式下 StrictMode 会重复触发 effect），
 * 不保留任何已完成的搜索结果：切换分类或关键词后重新请求即可，避免多个分类的
 * 结果常驻内存。
 */
const inFlight = new Map<string, Promise<SearchResult>>()

function requestOnce(key: string, run: () => Promise<SearchResult>): Promise<SearchResult> {
  const existing = inFlight.get(key)
  if (existing) return existing

  const request = run().finally(() => inFlight.delete(key))
  inFlight.set(key, request)
  return request
}

function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.search?.Service)
}

/** 获取综合搜索结果，后端已返回稳定 DTO。 */
export function getComplexSearch(keyword: string, pageNum = 1, pageSize = SEARCH_PAGE_SIZE): Promise<SearchResult> {
  if (!isBackendAvailable()) return Promise.reject(new Error('搜索服务不可用'))

  const key = `complex:${keyword}:${pageNum}:${pageSize}`
  return requestOnce(key, () => SearchComplex(keyword, pageNum, pageSize))
}

/** 根据搜索分类获取结果，后端已返回稳定 DTO。 */
export function getSearch(
  keyword: string,
  searchType: SearchTabType,
  pageNum = 1,
  pageSize = SEARCH_PAGE_SIZE,
): Promise<SearchResult> {
  if (searchType === 'complex') return getComplexSearch(keyword, pageNum, pageSize)
  if (!isBackendAvailable()) return Promise.reject(new Error('搜索服务不可用'))

  const key = `${searchType}:${keyword}:${pageNum}:${pageSize}`
  return requestOnce(key, () => Search(keyword, pageNum, pageSize, searchType))
}

export { EMPTY_RESULT }
