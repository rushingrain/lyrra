import { useEffect, useRef, useState } from 'react'

interface UseInfiniteScrollOptions {
  /** 哨兵进入视口时触发；内部始终调用最新一次传入的回调，无需用 useCallback 包裹 */
  onIntersect: () => void
  /** 是否观察哨兵：加载中、已到末尾等场景传 false */
  enabled: boolean
  /** 提前触发的距离，默认提前 200px，滚到底之前请求已在路上 */
  rootMargin?: string
}

/**
 * 无限滚动哨兵：用 IntersectionObserver 监听底部元素是否进入视口。
 *
 * 相比监听 scroll 事件，IntersectionObserver 由浏览器异步派发交叉结果，
 * 不会在滚动过程中每帧回调 JS，滚动时的 CPU 占用更低。
 *
 * 返回 ref 回调，挂到列表末尾的哨兵元素上即可。
 * 注意：IntersectionObserver 只在交叉状态「变化」时回调，若哨兵一直停留在视口内
 * 就不会再次触发；因此调用方应在加载期间把 enabled 置为 false、加载结束后恢复 true，
 * 本 hook 会在重新观察时补一次当前交叉状态，从而支持「首屏没填满时继续加载」。
 */
export function useInfiniteScroll({ onIntersect, enabled, rootMargin = '200px 0px' }: UseInfiniteScrollOptions) {
  const [sentinel, setSentinel] = useState<HTMLElement | null>(null)
  const onIntersectRef = useRef(onIntersect)

  useEffect(() => {
    onIntersectRef.current = onIntersect
  }, [onIntersect])

  useEffect(() => {
    if (!enabled || !sentinel) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) onIntersectRef.current()
      },
      { rootMargin },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [enabled, sentinel, rootMargin])

  return setSentinel
}
