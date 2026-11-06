import { useCallback, useEffect, useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll'
import type { Song } from '@/types'

export interface PaginatedSongsPage<TInfo = unknown> {
  info?: TInfo
  songs: Song[]
  total: number
  hasMore: boolean
}

export interface PaginatedSongsState<TInfo = unknown> {
  info: TInfo | null
  songs: Song[]
  hasMore: boolean
  loadError: string | null
  loadMoreError: string | null
  sentinelRef: (node: HTMLElement | null) => void
  loadMore: () => void
}

interface UsePaginatedSongsOptions<TInfo> {
  collectionId: string
  pageSize: number
  fetchPage: (id: string, page: number, pageSize: number) => Promise<PaginatedSongsPage<TInfo>>
  firstPageError: string
  nextPageError: string
}

/**
 * 歌曲集合分页：首屏取第一页，滚动到底部自动追加下一页。
 * 歌单和专辑共用相同的请求代次、去重和错误恢复逻辑。
 */
export function usePaginatedSongs<TInfo = unknown>({
  collectionId,
  pageSize,
  fetchPage,
  firstPageError,
  nextPageError,
}: UsePaginatedSongsOptions<TInfo>): PaginatedSongsState<TInfo> {
  const [info, setInfo] = useState<TInfo | null>(null)
  const [songs, setSongs] = useState<Song[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [page, setPage] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null)

  const epochRef = useRef(0)
  const pendingRef = useRef(false)
  const seenIdsRef = useRef(new Set<number>())

  const appendSongs = useCallback((next: Song[]) => {
    const fresh: Song[] = []
    for (const song of next) {
      if (song.id !== 0) {
        if (seenIdsRef.current.has(song.id)) continue
        seenIdsRef.current.add(song.id)
      }
      fresh.push(song)
    }
    if (fresh.length > 0) setSongs((prev) => [...prev, ...fresh])
  }, [])

  useEffect(() => {
    if (!collectionId) return

    const epoch = epochRef.current + 1
    epochRef.current = epoch
    pendingRef.current = true
    setIsLoading(true)
    setLoadError(null)
    setLoadMoreError(null)
    setInfo(null)
    setSongs([])
    setPage(0)
    setHasMore(false)
    seenIdsRef.current = new Set()

    fetchPage(collectionId, 1, pageSize)
      .then((data) => {
        if (epoch !== epochRef.current) return
        appendSongs(data.songs)
        setInfo(data.info ?? null)
        setPage(1)
        setHasMore(data.hasMore)
      })
      .catch(() => {
        if (epoch !== epochRef.current) return
        setLoadError(firstPageError)
        toast.error(firstPageError)
      })
      .finally(() => {
        if (epoch !== epochRef.current) return
        pendingRef.current = false
        setIsLoading(false)
      })

    return () => {
      epochRef.current += 1
    }
  }, [collectionId, pageSize, fetchPage, firstPageError, appendSongs])

  const loadMore = useCallback(() => {
    if (pendingRef.current || !hasMore || loadMoreError) return

    const epoch = epochRef.current
    const nextPage = page + 1
    pendingRef.current = true
    setIsLoadingMore(true)
    setLoadMoreError(null)

    fetchPage(collectionId, nextPage, pageSize)
      .then((data) => {
        if (epoch !== epochRef.current) return
        appendSongs(data.songs)
        setPage(nextPage)
        setHasMore(data.hasMore)
      })
      .catch(() => {
        if (epoch !== epochRef.current) return
        setLoadMoreError(nextPageError)
        toast.error(nextPageError)
      })
      .finally(() => {
        if (epoch !== epochRef.current) return
        pendingRef.current = false
        setIsLoadingMore(false)
      })
  }, [collectionId, pageSize, fetchPage, nextPageError, page, hasMore, loadMoreError, appendSongs])

  const canLoadMore = hasMore && !isLoading && !isLoadingMore && !loadMoreError
  const sentinelRef = useInfiniteScroll({ onIntersect: loadMore, enabled: canLoadMore })

  return { info, songs, hasMore, loadError, loadMoreError, sentinelRef, loadMore }
}
