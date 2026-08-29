import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
import type { Banner } from '@/types'

interface BannerCarouselProps {
  banners: Banner[]
}

/** 自动播放间隔(ms) */
const AUTOPLAY_MS = 4000
/** 切换动画时长(ms)，与轨道 transition 保持一致 */
const TRANSITION_MS = 500
/** 轮播固定为 5 页 */
const PAGE_COUNT = 5
/** 窄窗口至少展示 3 列 */
const MIN_COLUMNS = 3
/** 每列理想宽度(px)，超过此宽度时增加一列 */
const TARGET_CARD_WIDTH = 300

/**
 * 将 banner 分配到固定页数中。每页优先选取尚未出现过的 banner；
 * 当不同 banner 的数量不足一页列数时，才允许在当前页重复。
 */
function buildPages(banners: Banner[], columnCount: number): Banner[][] {
  if (banners.length === 0) return []

  const pages: Banner[][] = []
  let cursor = 0

  for (let pageIndex = 0; pageIndex < PAGE_COUNT; pageIndex += 1) {
    const page: Banner[] = []
    const usedIds = new Set<string>()

    while (page.length < columnCount) {
      let nextIndex = -1
      for (let offset = 0; offset < banners.length; offset += 1) {
        const candidateIndex = (cursor + offset) % banners.length
        if (!usedIds.has(banners[candidateIndex].id)) {
          nextIndex = candidateIndex
          break
        }
      }

      if (nextIndex === -1) {
        // 数据中的不同 banner 不足一页，按用户确认的规则允许重复补齐。
        nextIndex = cursor % banners.length
      }

      const banner = banners[nextIndex]
      page.push(banner)
      usedIds.add(banner.id)
      cursor = (nextIndex + 1) % banners.length
    }

    pages.push(page)
  }

  return pages
}

/**
 * 首页轮播。
 *
 * 轨道滑动方案：banner 固定分为 5 页，每页填满且尽量不出现重复 banner。
 * 轨道头尾各追加一页副本，
 * 向前/向后回卷都先滑到副本再瞬间跳回，两个方向均无缝。
 * 切页通过 translateX 过渡（GPU 合成，流畅且无 DOM 结构变化）；
 * 手动翻页/点指示器会重置自动播放计时，避免与自动轮播重叠。
 */
export function BannerCarousel({ banners }: BannerCarouselProps) {
  const len = banners.length
  const containerRef = useRef<HTMLDivElement>(null)
  // 每页图片数量随容器宽度变化，且必须能在固定 5 页内放下全部 banner。
  const [visibleCount, setVisibleCount] = useState(MIN_COLUMNS)
  // 轨道平移位置（单位：一页 = 100%），范围 -1..numPages：
  // -1 / numPages 对应复制出的回卷页，动画结束后瞬间跳回逻辑页
  const [visual, setVisual] = useState(0)
  // 回卷跳回时禁用过渡，避免瞬间产生反向动画
  const [noTransition, setNoTransition] = useState(false)
  // 手动操作后 +1，触发自动播放 effect 重新计时
  const [autoplayKey, setAutoplayKey] = useState(0)

  const vc = visibleCount
  const numPages = PAGE_COUNT
  // 逻辑页码（指示器用）：-1 视为最后一页，numPages 视为第 0 页
  const page = ((visual % numPages) + numPages) % numPages

  // 以理想卡片宽度决定列数：窗口更宽时，列数和卡片宽度都会随之增大。
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const update = () => {
      const responsiveColumns = Math.max(MIN_COLUMNS, Math.floor(el.clientWidth / TARGET_CARD_WIDTH))
      const columnsForAllBanners = Math.ceil(len / PAGE_COUNT)
      setVisibleCount(Math.max(responsiveColumns, columnsForAllBanners))
    }
    update()
    const ro = new ResizeObserver(update)
    ro.observe(el)
    return () => ro.disconnect()
    // 依赖 len：数据变化时重新计算固定 5 页所需的容量。
  }, [len])

  // 保证回卷结束后始终落在有效逻辑页。
  useEffect(() => {
    setVisual((v) => Math.min(Math.max(v, 0), numPages - 1))
  }, [numPages])

  const next = useCallback(() => {
    setVisual((v) => (v >= numPages - 1 ? numPages : v + 1))
    setAutoplayKey((k) => k + 1)
  }, [numPages])

  const prev = useCallback(() => {
    setVisual((v) => (v <= 0 ? -1 : v - 1))
    setAutoplayKey((k) => k + 1)
  }, [numPages])

  // 无缝回卷：滑到复制页（numPages / -1）后，动画结束瞬间跳回逻辑页。
  // 用定时器替代 transitionend（更确定）；跳回置 noTransition，由下方独立 effect
  // 用双 rAF 恢复过渡，避免“恢复过渡”与“跳回”同帧生效而产生反向动画。
  // 回卷完成后重置自动播放计时。
  useEffect(() => {
    if (visual !== numPages && visual !== -1) return
    const target = visual === numPages ? 0 : numPages - 1
    const timer = window.setTimeout(() => {
      setNoTransition(true)
      setVisual(target)
      // 从回卷结束重新计时，避免“刚回卷完又立即滑动”
      setAutoplayKey((k) => k + 1)
    }, TRANSITION_MS)
    return () => window.clearTimeout(timer)
  }, [visual, numPages])

  // noTransition 置位后，用双 rAF 恢复过渡：
  // 第一帧确保“无过渡跳回”真正渲染，第二帧再恢复，避免反向动画。
  // 独立于回卷 effect，防止其 cleanup 误取消恢复任务；卸载时取消两段 rAF。
  useEffect(() => {
    if (!noTransition) return
    let raf2 = 0
    const raf1 = requestAnimationFrame(() => {
      raf2 = requestAnimationFrame(() => setNoTransition(false))
    })
    return () => {
      cancelAnimationFrame(raf1)
      cancelAnimationFrame(raf2)
    }
  }, [noTransition])

  // 自动播放：手动操作（autoplayKey 变化）或页数变化后重新计时，避免与点击重叠
  useEffect(() => {
    if (numPages <= 1) return
    const timer = window.setInterval(next, AUTOPLAY_MS)
    return () => window.clearInterval(timer)
  }, [numPages, next, autoplayKey])

  const pages = useMemo(() => buildPages(banners, vc), [banners, vc])

  if (len === 0) return null

  // 轨道：[最后一页副本, ...各页, 第一页副本]
  // 头尾各加一页副本，向前/向后回卷都先滑到副本再瞬间跳回，两个方向均无缝。
  const track = [pages[numPages - 1], ...pages, pages[0]]
  // 真正的第一页前面有一页副本，故 visual=0 对应轨道第 1 页
  const offset = visual + 1

  return (
    <div className="mb-10">
      <div ref={containerRef} className="group relative overflow-hidden rounded-xl">
        {/* 轨道：transform 过渡由 GPU 合成，切换流畅 */}
        <div
          className="flex w-full"
          style={{
            transform: `translateX(-${offset * 100}%)`,
            transition: noTransition ? 'none' : `transform ${TRANSITION_MS}ms ease-out`,
            willChange: 'transform',
          }}
        >
          {track.map((pageItems, pi) => (
            <div
              key={pi}
              className="grid w-full shrink-0 gap-4"
              style={{ gridTemplateColumns: `repeat(${vc}, minmax(0, 1fr))` }}
            >
              {pageItems.map((banner, bi) => (
                <div
                  key={`${pi}-${bi}`}
                  className="min-w-0"
                >
                  <button
                    type="button"
                    className={`block w-full ${banner.linkUrl ? 'cursor-pointer' : ''}`}
                    onClick={banner.linkUrl ? () => BrowserOpenURL(banner.linkUrl) : undefined}
                    aria-label={`打开：${banner.title}`}
                  >
                    <img
                      src={banner.imageUrl}
                      alt={banner.title}
                      className="aspect-[11/6] w-full rounded-lg object-cover"
                      loading="lazy"
                    />
                  </button>
                </div>
              ))}
            </div>
          ))}
        </div>

        {/* 左右箭头 */}
        {numPages > 1 && (
          <>
            <button
              className="absolute left-3 top-1/2 flex h-8 w-8 cursor-pointer -translate-y-1/2 items-center justify-center rounded-full bg-black/30 text-white opacity-0 transition-opacity group-hover:opacity-100"
              onClick={prev}
              aria-label="上一组"
            >
              <ChevronLeft size={18} />
            </button>
            <button
              className="absolute right-3 top-1/2 flex h-8 w-8 cursor-pointer -translate-y-1/2 items-center justify-center rounded-full bg-black/30 text-white opacity-0 transition-opacity group-hover:opacity-100"
              onClick={next}
              aria-label="下一组"
            >
              <ChevronRight size={18} />
            </button>
          </>
        )}
      </div>

      {/* 指示器（圆形）：放在图片下方空白处 */}
      {numPages > 1 && (
        <div className="mt-3 flex items-center justify-center gap-1.5">
          {Array.from({ length: numPages }).map((_, i) => (
            <button
              key={i}
              className={`cursor-pointer rounded-full transition-all ${i === page
                ? 'h-2 w-2 bg-[var(--color-accent)]'
                : 'h-1.5 w-1.5 bg-gray-300'
                }`}
              onClick={() => {
                setVisual(i)
                setAutoplayKey((k) => k + 1)
              }}
              aria-label={`第 ${i + 1} 组`}
            />
          ))}
        </div>
      )}
    </div>
  )
}
