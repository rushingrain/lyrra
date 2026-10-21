import { useCallback, useEffect, useRef, useState } from 'react'
import toast from 'react-hot-toast'
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll'
import { PLAYLIST_PAGE_SIZE, getPlaylistSongs } from '@/services/playlist'
import type { Playlist, Song } from '@/types'

export interface PlaylistSongsState {
  /** 歌单元信息（来自后端 list_info，不含播放次数） */
  info: Playlist | null
  songs: Song[]
  hasMore: boolean
  /** 首屏失败原因，非空时页面展示提示 */
  loadError: string | null
  /** 追加失败原因，非空时底部提供重试 */
  loadMoreError: string | null
  /** 挂到列表末尾，进入视口即自动加载下一页 */
  sentinelRef: (node: HTMLElement | null) => void
  loadMore: () => void
}

/**
 * 歌单歌曲分页：首屏取第一页，滚动到底部自动追加下一页。
 *
 * 与搜索分页相同的处理：请求代次丢弃过期响应、在途请求去重、按歌曲 id 去重；
 * 追加失败时保留已加载内容，并停止自动加载，等用户点重试。
 */
export function usePlaylistSongs(globalCollectionId: string): PlaylistSongsState {
  const [info, setInfo] = useState<Playlist | null>(null)
  const [songs, setSongs] = useState<Song[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [page, setPage] = useState(0)
  const [isLoading, setIsLoading] = useState(false)
  const [isLoadingMore, setIsLoadingMore] = useState(false)
  const [loadError, setLoadError] = useState<string | null>(null)
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null)

  const epochRef = useRef(0)
  const pendingRef = useRef(false)
  // 跨页去重：同一首歌可能出现在相邻分页里
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
    if (!globalCollectionId) return

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

    getPlaylistSongs(globalCollectionId, 1, PLAYLIST_PAGE_SIZE)
      .then((data) => {
        if (epoch !== epochRef.current) return
        appendSongs(data.songs)
        setInfo(data.info)
        setPage(1)
        setHasMore(data.hasMore)
      })
      .catch(() => {
        if (epoch !== epochRef.current) return
        const message = '歌单歌曲加载失败，请稍后重试'
        setLoadError(message)
        toast.error(message)
      })
      .finally(() => {
        if (epoch !== epochRef.current) return
        pendingRef.current = false
        setIsLoading(false)
      })

    return () => {
      epochRef.current += 1
    }
  }, [globalCollectionId, appendSongs])

  const loadMore = useCallback(() => {
    if (pendingRef.current || !hasMore || loadMoreError) return

    const epoch = epochRef.current
    const nextPage = page + 1
    pendingRef.current = true
    setIsLoadingMore(true)
    setLoadMoreError(null)

    getPlaylistSongs(globalCollectionId, nextPage, PLAYLIST_PAGE_SIZE)
      .then((data) => {
        if (epoch !== epochRef.current) return
        appendSongs(data.songs)
        setPage(nextPage)
        setHasMore(data.hasMore)
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
  }, [globalCollectionId, page, hasMore, loadMoreError, appendSongs])

  // 加载中或上次追加失败时不再自动触发，避免失败后反复请求
  const canLoadMore = hasMore && !isLoading && !isLoadingMore && !loadMoreError
  const sentinelRef = useInfiniteScroll({ onIntersect: loadMore, enabled: canLoadMore })

  return { info, songs, hasMore, loadError, loadMoreError, sentinelRef, loadMore }
}
