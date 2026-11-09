import { useState } from 'react'
import { Headphones } from 'lucide-react'
import toast from 'react-hot-toast'
import { PlayAllButton } from '@/components/business/PlayAllButton'
import { ResultTable } from '@/components/common/ResultTable'
import { songColumns } from '@/components/search/songColumns'
import { usePlayerStore } from '@/stores/playerStore'
import type { Song } from '@/types'

interface SongCollectionDetailProps {
  coverUrl?: string
  title?: string
  description?: string
  author?: string
  authorAvatar?: string
  tags?: string[]
  details?: string[]
  publishTime?: string
  publishLabel?: string
  playCount?: number
  songs: Song[]
  hasMore: boolean
  loadMoreError: string | null
  sentinelRef: (node: HTMLElement | null) => void
  loadMore: () => void
  loadAllSongs: () => Promise<Song[]>
  loadAllErrorMessage: string
}

function formatPlayCount(count: number): string {
  if (count >= 100000000) return `${(count / 100000000).toFixed(1)}亿`
  if (count >= 10000) return `${(count / 10000).toFixed(1)}w`
  return String(count)
}

function formatPublishDate(value?: string): string {
  const date = value?.trim().split(' ')[0] ?? ''
  return /^\d{4}-\d{2}-\d{2}$/.test(date) ? date : ''
}

/** 歌单与专辑共用的详情布局：封面头图、元信息、播放全部和分页歌曲列表。 */
export function SongCollectionDetail({
  coverUrl,
  title,
  description,
  author,
  authorAvatar,
  tags,
  details,
  publishTime,
  publishLabel,
  playCount = 0,
  songs,
  hasMore,
  loadMoreError,
  sentinelRef,
  loadMore,
  loadAllSongs,
  loadAllErrorMessage,
}: SongCollectionDetailProps) {
  const playAll = usePlayerStore((state) => state.playAll)
  const [isPreparingPlayAll, setIsPreparingPlayAll] = useState(false)
  const publishDate = formatPublishDate(publishTime)

  const handlePlayAll = async () => {
    if (songs.length === 0 || isPreparingPlayAll) return

    let queueSongs = songs
    if (hasMore) {
      setIsPreparingPlayAll(true)
      try {
        queueSongs = await loadAllSongs()
      } catch {
        toast.error(loadAllErrorMessage)
        return
      } finally {
        setIsPreparingPlayAll(false)
      }
    }
    await playAll(queueSongs, { replaceQueue: true })
  }

  return (
    <div className="flex flex-col gap-6">
      {title && (
        <div className="flex gap-6">
          <div className="relative h-[160px] w-[160px] shrink-0 overflow-hidden rounded-lg bg-[var(--color-surface)]">
            {coverUrl ? (
              <img src={coverUrl} alt={title} className="h-full w-full object-cover" loading="lazy" />
            ) : null}
            {playCount > 0 ? (
              <span className="absolute left-1.5 top-1.5 flex items-center gap-1 text-xs font-medium text-white">
                <Headphones size={12} />
                <span>{formatPlayCount(playCount)}</span>
              </span>
            ) : null}
          </div>

          <div className="flex min-w-0 flex-1 flex-col">
            <h1 className="truncate text-xl font-semibold text-[var(--color-text-primary)]">{title}</h1>
            {description ? (
              <p className="mt-2 line-clamp-1 text-sm leading-relaxed text-[var(--color-text-secondary)]">
                {description}
              </p>
            ) : null}
            {author ? (
              <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-[var(--color-text-secondary)]">
                {authorAvatar ? (
                  <img
                    src={authorAvatar}
                    alt={author}
                    className="h-6 w-6 shrink-0 rounded-full object-cover"
                    loading="lazy"
                  />
                ) : null}
                <span className="inline-flex h-6 items-center">{author}</span>
                {tags && tags.length > 0 ? (
                  <span className="inline-flex h-6 items-center">标签：{tags.join(' / ')}</span>
                ) : null}
                {details?.map((detail) => (
                  <span key={detail} className="inline-flex h-6 items-center">
                    {detail}
                  </span>
                ))}
                {publishDate ? (
                  <span className="inline-flex h-6 items-center">
                    <span className="relative top-[0.4px]">{publishDate}</span>
                    {publishLabel ? <span className="ml-1">{publishLabel}</span> : null}
                  </span>
                ) : null}
              </div>
            ) : null}
            <div className="mt-auto pt-4">
              <PlayAllButton
                onClick={() => void handlePlayAll()}
                disabled={songs.length === 0}
                loading={isPreparingPlayAll}
              />
            </div>
          </div>
        </div>
      )}

      <div>
        <ResultTable columns={songColumns} rows={songs} rowKey={(song) => song.id} />

        {hasMore && <div ref={sentinelRef} className="h-px" aria-hidden />}

        <div className="pt-4 text-center text-sm text-[var(--color-text-secondary)]">
          {loadMoreError && (
            <button
              type="button"
              className="cursor-pointer hover:text-[var(--color-text-primary)]"
              onClick={loadMore}
            >
              加载失败，点击重试
            </button>
          )}
          {!loadMoreError && !hasMore && songs.length > 0 && <span>没有更多了</span>}
        </div>
      </div>
    </div>
  )
}
