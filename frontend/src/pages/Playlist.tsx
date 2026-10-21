import { useState } from 'react'
import { Headphones } from 'lucide-react'
import { useLocation, useParams } from 'react-router-dom'
import toast from 'react-hot-toast'
import { PlayAllButton } from '@/components/business/PlayAllButton'
import { ResultTable } from '@/components/common/ResultTable'
import { songColumns } from '@/components/search/songColumns'
import { usePlaylistSongs } from '@/hooks/usePlaylistSongs'
import { getAllPlaylistSongs } from '@/services/playlist'
import { usePlayerStore } from '@/stores/playerStore'
import type { Playlist as PlaylistType } from '@/types'

function formatPlayCount(count: number): string {
  if (count >= 100000000) return `${(count / 100000000).toFixed(1)}亿`
  if (count >= 10000) return `${(count / 10000).toFixed(1)}w`
  return String(count)
}

/** 取上游时间里的日期部分（yyyy-mm-dd）；格式不符时返回空串 */
function formatPublishDate(value?: string): string {
  const date = value?.trim().split(' ')[0] ?? ''
  return /^\d{4}-\d{2}-\d{2}$/.test(date) ? date : ''
}

/**
 * 歌单详情页：左侧封面（含播放次数），右侧标题/描述/作者信息与播放全部，
 * 下方是支持滚动分页的歌曲列表。
 */
export default function Playlist() {
  const { globalId = '' } = useParams<{ globalId: string }>()
  const location = useLocation()
  const statePlaylist = (location.state as { playlist?: PlaylistType } | null)?.playlist
  const { info, songs, hasMore, loadError, loadMoreError, sentinelRef, loadMore } =
    usePlaylistSongs(globalId)
  const playAll = usePlayerStore((state) => state.playAll)
  const [isPreparingPlayAll, setIsPreparingPlayAll] = useState(false)

  // 接口返回的 list_info 为准；拿不到有效信息时回退首页传入的歌单
  const playlist = info?.name ? info : statePlaylist
  // 歌曲接口不返回播放次数，用首页带入的数据补齐（直接打开链接时没有角标）
  const playCount = statePlaylist?.playCount ?? 0
  const publishDate = formatPublishDate(playlist?.publishTime)

  /** 播放全部：歌单尚未加载完时先按需拉取全量，再用歌单歌曲替换整个播放队列 */
  const handlePlayAll = async () => {
    if (songs.length === 0 || isPreparingPlayAll) return

    let queueSongs = songs
    if (hasMore) {
      setIsPreparingPlayAll(true)
      try {
        queueSongs = await getAllPlaylistSongs(globalId)
      } catch {
        toast.error('歌单歌曲加载失败，请稍后重试')
        return
      } finally {
        setIsPreparingPlayAll(false)
      }
    }
    await playAll(queueSongs, { replaceQueue: true })
  }

  if (loadError) {
    return <p className="py-8 text-center text-sm text-[var(--color-text-secondary)]">{loadError}</p>
  }

  return (
    <div className="flex flex-col gap-6">
      {playlist && (
        <div className="flex gap-6">
          <div className="relative h-[160px] w-[160px] shrink-0 overflow-hidden rounded-lg bg-[var(--color-surface)]">
            {playlist.coverUrl ? (
              <img
                src={playlist.coverUrl}
                alt={playlist.name}
                className="h-full w-full object-cover"
                loading="lazy"
              />
            ) : null}
            {playCount > 0 ? (
              <span className="absolute left-1.5 top-1.5 flex items-center gap-1 text-xs font-medium text-white">
                <Headphones size={12} />
                <span>{formatPlayCount(playCount)}</span>
              </span>
            ) : null}
          </div>

          <div className="flex min-w-0 flex-1 flex-col">
            <h1 className="truncate text-xl font-semibold text-[var(--color-text-primary)]">
              {playlist.name}
            </h1>
            {playlist.description ? (
              <p className="mt-2 line-clamp-1 text-sm leading-relaxed text-[var(--color-text-secondary)]">
                {playlist.description}
              </p>
            ) : null}
            {playlist.creator ? (
              <div className="mt-2 flex flex-wrap items-center gap-2 text-sm text-[var(--color-text-secondary)]">
                {playlist.creatorAvatar ? (
                  <img
                    src={playlist.creatorAvatar}
                    alt={playlist.creator}
                    className="h-6 w-6 shrink-0 rounded-full object-cover"
                    loading="lazy"
                  />
                ) : null}
                <span className="inline-flex h-6 items-center">{playlist.creator}</span>
                {playlist.tags && playlist.tags.length > 0 ? (
                  <span className="inline-flex h-6 items-center">标签：{playlist.tags.join(' / ')}</span>
                ) : null}
                {publishDate ? (
                  <span className="inline-flex h-6 items-center">
                    <span className="relative top-[0.4px]">{publishDate.trim()}</span>
                    <span className="ml-1">创建</span>
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

        {/* 无限滚动哨兵：进入视口即追加下一页 */}
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
