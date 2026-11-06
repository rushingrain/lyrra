import { useEffect, useState } from 'react'
import toast from 'react-hot-toast'
import { getAlbumDetail } from '@/services/album'
import type { Album } from '@/types'

/** 获取专辑详情；失败时由页面回退到搜索页传入的专辑信息。 */
export function useAlbumDetail(albumId: string): Album | null {
  const [detail, setDetail] = useState<Album | null>(null)

  useEffect(() => {
    if (!albumId) return
    let cancelled = false
    setDetail(null)

    getAlbumDetail(albumId)
      .then((data) => {
        if (!cancelled) setDetail(data)
      })
      .catch(() => {
        if (!cancelled) toast.error('专辑详情加载失败，请稍后重试')
      })

    return () => {
      cancelled = true
    }
  }, [albumId])

  return detail
}
