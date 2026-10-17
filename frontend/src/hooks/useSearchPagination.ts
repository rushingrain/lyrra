import { useCallback, useEffect, useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { search } from '../../wailsjs/go/models'
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll'
import {
  EMPTY_RESULT,
  SEARCH_MAX_ITEMS,
  SEARCH_PAGE_SIZE,
  getSearch,
  type SearchResult,
  type SearchTabType,
} from '@/services/search'

/** 前端最多加载的页数：由展示上限换算，用页数判断可避免“累计条数”与去重结果不一致 */
const MAX_PAGES = Math.ceil(SEARCH_MAX_ITEMS / SEARCH_PAGE_SIZE)

/** 综合搜索是分组聚合结果，没有分页语义；其余普通搜索分类可滚动加载 */
export function isPaginatedTab(tab: SearchTabType): boolean {
  return tab !== 'complex'
}

/** 统计一页结果里的条目数：普通搜索只会填充其中一个分类，其余为空数组 */
function countItems(result: SearchResult): number {
  return (
    result.songs.length +
    result.playlists.length +
    result.albums.length +
    result.artists.length +
    result.mvs.length
  )
}

/**
 * 按 id 去重追加：相邻分页可能返回重复条目，重复项既会撞 React key 也会干扰“到底”的判断。
 * id 为 0 表示上游没给标识，无法判定是否重复，因此原样保留。
 */
function appendUnique<T extends { id: number }>(prev: T[], next: T[]): T[] {
  if (next.length === 0) return prev

  const seen = new Set<number>()
  for (const item of prev) {
    if (item.id !== 0) seen.add(item.id)
  }

  const merged = prev.slice()
  for (const item of next) {
    if (item.id !== 0) {
      if (seen.has(item.id)) continue
      seen.add(item.id)
    }
    merged.push(item)
  }
  return merged
}

function mergeResults(prev: SearchResult, next: SearchResult): SearchResult {
  return new search.Result({
    songs: appendUnique(prev.songs, next.songs),
    playlists: appendUnique(prev.playlists, next.playlists),
    albums: appendUnique(prev.albums, next.albums),
    artists: appendUnique(prev.artists, next.artists),
    mvs: appendUnique(prev.mvs, next.mvs),
  })
}

export interface SearchPagination {
  /** 已合并的分页结果；综合搜索始终只有第一页 */
  result: SearchResult
  /** 首屏加载中（此时列表整体让位给加载/错误提示） */
  isLoading: boolean
  /** 首屏失败原因，非空时列表被替换为提示文案 */
  loadError: string | null
  /** 追加失败原因，非空时保留已加载内容并在底部提供重试 */
  loadMoreError: string | null
  hasMore: boolean
  /** 因达到前端展示上限而停止加载 */
  limitReached: boolean
  /** 挂到列表末尾的哨兵元素上，进入视口即自动加载下一页 */
  sentinelRef: (node: HTMLElement | null) => void
  /** 手动加载下一页（追加失败后的重试入口） */
  loadMore: () => void
}

/**
 * 搜索分页：综合搜索只取第一页；普通搜索滚动到底部自动追加下一页，最多展示 SEARCH_MAX_ITEMS 条。
 *
 * 几处关键处理：
 * - 用请求代次（epochRef）丢弃过期响应，切分类或换关键词后旧请求不再写状态；
 * - pendingRef 记录在途请求，避免滚动抖动时重复发起同一页；
 * - 追加失败只提示、不清空已加载内容，并允许手动重试；
 * - 是否还有下一页按「本页返回条数」判断：等于一页说明可能还有，少于则说明已到底。
 */
export function useSearchPagination(query: string, tab: SearchTabType): SearchPagination {
  const [result, setResult] = useState<SearchResult>(EMPTY_RESULT)
  const [page, setPage] = useState(0)
  const [hasMore, setHasMore] = useState(false)
  const [limitReached, setLimitReached] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null)

  const epochRef = useRef(0)
  const pendingRef = useRef(false)

  // 关键词或分类变化：重置为第一页并请求
  useEffect(() => {
    if (!query) return

    const epoch = epochRef.current + 1
    epochRef.current = epoch
    pendingRef.current = true
    setIsLoading(true)
    setIsLoadingMore(false)
    setLoadError(null)
    setLoadMoreError(null)
    setResult(EMPTY_RESULT)
    setPage(0)
    setHasMore(false)
    setLimitReached(false)

    getSearch(query, tab, 1, SEARCH_PAGE_SIZE)
      .then((data) => {
        if (epoch !== epochRef.current) return
        setResult(data)
        setPage(1)
        setHasMore(isPaginatedTab(tab) && countItems(data) >= SEARCH_PAGE_SIZE)
      })
      .catch(() => {
        if (epoch !== epochRef.current) return
        const message = '搜索失败，请稍后重试'
        setLoadError(message)
        toast.error(message)
      })
      .finally(() => {
        if (epoch !== epochRef.current) return
        pendingRef.current = false
        setIsLoading(false)
      })

    // 卸载或依赖变化时作废在途请求，避免过期响应回写状态
    return () => {
      epochRef.current += 1
    }
  }, [query, tab])

  const loadMore = useCallback(() => {
    if (pendingRef.current || !hasMore || limitReached) return

    const epoch = epochRef.current
    const nextPage = page + 1
    pendingRef.current = true
    setIsLoadingMore(true)
    setLoadMoreError(null)

    getSearch(query, tab, nextPage, SEARCH_PAGE_SIZE)
      .then((data) => {
        if (epoch !== epochRef.current) return
        setResult((prev) => mergeResults(prev, data))
        setPage(nextPage)
        if (countItems(data) < SEARCH_PAGE_SIZE) {
          setHasMore(false)
        } else if (nextPage >= MAX_PAGES) {
          // 上游可能还有数据，但已达前端展示上限
          setHasMore(false)
          setLimitReached(true)
        }
      })
      .catch(() => {
        if (epoch !== epochRef.current) return
        const message = '加载更多失败，请重试'
        setLoadMoreError(message)
        toast.error(message)
      })
      .finally(() => {
        if (epoch !== epochRef.current) return
        pendingRef.current = false
        setIsLoadingMore(false)
      })
  }, [query, tab, page, hasMore, limitReached])

  // 加载期间 enabled 为 false：既拦住在途请求的重复触发，也让观察在加载结束后重建，
  // 从而在哨兵仍停留在视口内（首屏未填满）时继续自动加载
  const canLoadMore = isPaginatedTab(tab) && hasMore && !limitReached && !isLoading && !isLoadingMore
  const sentinelRef = useInfiniteScroll({ onIntersect: loadMore, enabled: canLoadMore })

  return {
    result,
    isLoading,
    loadError,
    loadMoreError,
    hasMore,
    limitReached,
    sentinelRef,
    loadMore,
  }
}
