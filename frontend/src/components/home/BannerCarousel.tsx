import { useState, useEffect, useCallback, useRef } from 'react'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import type { Banner } from '../../types'

interface BannerCarouselProps {
  banners: Banner[]
}

export function BannerCarousel({ banners }: BannerCarouselProps) {
  const len = banners.length
  const containerRef = useRef<HTMLDivElement>(null)
  // 一次展示的图片数量：默认 2 张，容器足够宽时切换为 3 张
  const [visibleCount, setVisibleCount] = useState(2)
  const [page, setPage] = useState(0)
  const [loaded, setLoaded] = useState<Record<number, boolean>>({})

  // 展示数量不超过 banner 总数
  const vc = Math.min(visibleCount, len)
  const numPages = Math.max(1, Math.ceil(len / vc))
  const startIndex = page * vc

  // 根据容器实际宽度决定一次展示 2 张还是 3 张
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const update = () => {
      setVisibleCount(el.clientWidth > 1200 ? 3 : 2)
    }
    update()
    const ro = new ResizeObserver(update)
    ro.observe(el)
    return () => ro.disconnect()
  }, [])

  // 展示数量变化后纠正越界的页码
  useEffect(() => {
    setPage((p) => Math.min(p, numPages - 1))
  }, [numPages])

  const next = useCallback(() => {
    setPage((p) => (p + 1) % numPages)
  }, [numPages])

  const prev = useCallback(() => {
    setPage((p) => (p - 1 + numPages) % numPages)
  }, [numPages])

  // 自动播放：整行轮换
  useEffect(() => {
    if (numPages <= 1) return
    const timer = setInterval(next, 4000)
    return () => clearInterval(timer)
  }, [numPages, next])

  if (len === 0) return null

  return (
    <div>
      <div ref={containerRef} className="group relative">
        {/* 图片行：一次展示 vc 张，随容器宽度自动放大 */}
        <div className="flex gap-4">
          {Array.from({ length: vc }).map((_, slot) => (
            <div
              key={slot}
              className="relative aspect-[800/240] flex-1 overflow-hidden rounded-xl"
            >
              {banners.map((banner, index) => {
                const activeIndex = (startIndex + slot) % len
                return (
                  <div
                    key={banner.id}
                    className={`absolute inset-0 transition-opacity duration-500 ${index === activeIndex ? 'opacity-100' : 'opacity-0'
                      }`}
                  >
                    {!loaded[index] && (
                      <div className="absolute inset-0 animate-pulse bg-[var(--color-border)]" />
                    )}
                    <img
                      src={banner.imageUrl}
                      alt={banner.title}
                      className="h-full w-full object-cover"
                      onLoad={() =>
                        setLoaded((prev) => ({ ...prev, [index]: true }))
                      }
                    />
                  </div>
                )
              })}
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
              className={`cursor-pointer rounded-full transition-all ${i === page
                ? 'h-2.5 w-2.5 bg-[var(--color-accent)]'
                : 'h-2 w-2 bg-gray-300'
                }`}
              onClick={() => setPage(i)}
              aria-label={`第 ${i + 1} 组`}
            />
          ))}
        </div>
      )}
    </div>
  )
}
