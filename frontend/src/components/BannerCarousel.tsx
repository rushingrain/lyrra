import { useCallback, useEffect, useRef, useState } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'
import type { Banner } from '@/types'

interface BannerCarouselProps {
  banners: Banner[]
}

/** 自动播放间隔(ms) */
const AUTOPLAY_MS = 4000
/** 切换动画时长(ms)，与轨道 transition 保持一致 */
const TRANSITION_MS = 500

/**
 * 首页轮播。
 *
 * 轨道滑动方案：banner 按每页 vc 张分页，最后一页不足 vc 张时用空位占位
 * （每页等宽、且不出现重复 banner）。轨道头尾各追加一页副本，
 * 向前/向后回卷都先滑到副本再瞬间跳回，两个方向均无缝。
 * 切页通过 translateX 过渡（GPU 合成，流畅且无 DOM 结构变化）；
 * 手动翻页/点指示器会重置自动播放计时，避免与自动轮播重叠。
 */
export function BannerCarousel({ banners }: BannerCarouselProps) {
  const len = banners.length
  const containerRef = useRef<HTMLDivElement>(null)
  // 一次展示的图片数量：默认 2 张，容器足够宽时切换为 3 张
  const [visibleCount, setVisibleCount] = useState(2)
  // 轨道平移位置（单位：一页 = 100%），范围 -1..numPages：
  // -1 / numPages 对应复制出的回卷页，动画结束后瞬间跳回逻辑页
  const [visual, setVisual] = useState(0)
  // 回卷跳回时禁用过渡，避免瞬间产生反向动画
  const [noTransition, setNoTransition] = useState(false)
  // 手动操作后 +1，触发自动播放 effect 重新计时
  const [autoplayKey, setAutoplayKey] = useState(0)

  const vc = Math.min(visibleCount, Math.max(len, 1))
  const numPages = len === 0 ? 1 : Math.max(1, Math.ceil(len / vc))
  // 逻辑页码（指示器用）：-1 视为最后一页，numPages 视为第 0 页
  const page = ((visual % numPages) + numPages) % numPages

  // 根据容器实际宽度决定一次展示 2 张还是 3 张
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const update = () => setVisibleCount(el.clientWidth > 1200 ? 3 : 2)
    update()
    const ro = new ResizeObserver(update)
    ro.observe(el)
    return () => ro.disconnect()
    // 依赖 len：首次挂载 banner 为空时容器未渲染，数据到达后重新挂载观察器
  }, [len])

  // 容器变窄/数据变化导致页数变化时，把越界的平移位置拉回有效范围
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

  if (len === 0) return null

  // 分页：每页 vc 张，最后一页不足 vc 张时用 null 占位，保证每页等宽且不重复
  const pages: (Banner | null)[][] = []
  for (let i = 0; i < numPages; i++) {
    const slice: (Banner | null)[] = banners.slice(i * vc, i * vc + vc)
    while (slice.length < vc) slice.push(null)
    pages.push(slice)
  }
  // 轨道：[最后一页副本, ...各页, 第一页副本]
  // 头尾各加一页副本，向前/向后回卷都先滑到副本再瞬间跳回，两个方向均无缝。
  const track = [pages[numPages - 1], ...pages, pages[0]]
  // 真正的第一页前面有一页副本，故 visual=0 对应轨道第 1 页
  const offset = visual + 1

  return (
    <div>
      {/* 板块标题：与“官方歌单”等标题同款样式，但仅为标题，无箭头、不可点击 */}
      <div className="mb-4 flex items-center justify-between">
        <h2 className="font-bold text-lg text-[var(--color-text-primary)]">精选活动</h2>
      </div>

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
            <div key={pi} className="flex w-full shrink-0">
              {pageItems.map((banner, bi) => (
                <div
                  key={`${pi}-${bi}`}
                  className="shrink-0 px-2"
                  style={{ width: `${100 / vc}%` }}
                >
                  {banner && (
                    <button
                      type="button"
                      className={`block w-full ${banner.linkUrl ? 'cursor-pointer' : ''}`}
                      onClick={banner.linkUrl ? () => BrowserOpenURL(banner.linkUrl) : undefined}
                      aria-label={`打开：${banner.title}`}
                    >
                      <img
                        src={banner.imageUrl}
                        alt={banner.title}
                        className="aspect-[800/360] w-full rounded-lg object-cover"
                        loading="lazy"
                      />
                    </button>
                  )}
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
        <div className="mt-3 flex items-center justify-center gap-2">
          {Array.from({ length: numPages }).map((_, i) => (
            <button
              key={i}
              className={`cursor-pointer rounded-full transition-all ${
                i === page
                  ? 'h-2.5 w-2.5 bg-[var(--color-accent)]'
                  : 'h-2 w-2 bg-gray-300'
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
