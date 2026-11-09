import { useEffect, useRef } from 'react'

/** 两次点击之间的最大间隔(ms)：超过这个时间没有后续点击就重新计数 */
const DEFAULT_INTERVAL_MS = 1500

interface UseClickStreakOptions {
  /** 需要连续点击的次数 */
  times: number
  /** 达成连续点击后触发，可以是异步逻辑（返回值会被忽略） */
  onTrigger: () => void
  /** 两次点击的最大间隔(ms)，默认 1500 */
  intervalMs?: number
  /** 入口开关：为 false 时忽略点击并清空已有计数，默认 true */
  enabled?: boolean
}

/**
 * 连续点击检测（隐藏入口用）：在 intervalMs 内累计点击 times 次即触发 onTrigger，
 * 触发后、以及中途超时都会清零重新计数；enabled 为 false 时整体关闭，
 * 忽略点击并清空计数（用于「未登录时不开放入口」这类场景）。
 *
 * 计数与定时器放在 ref 里，不引起重渲染；返回的 handler 直接绑到目标元素的
 * onClick 上即可。组件卸载时清掉定时器，避免已卸载后还在计时。
 */
export function useClickStreak({
  times,
  onTrigger,
  intervalMs = DEFAULT_INTERVAL_MS,
  enabled = true,
}: UseClickStreakOptions) {
  const countRef = useRef(0)
  const timerRef = useRef<number | undefined>(undefined)

  // 入口关闭（例如登出）时清空计数，避免“关着的时候点了 6 次、打开后点 1 次”就触发
  useEffect(() => {
    if (enabled) return
    window.clearTimeout(timerRef.current)
    countRef.current = 0
  }, [enabled])

  useEffect(() => () => window.clearTimeout(timerRef.current), [])

  return () => {
    if (!enabled) return
    window.clearTimeout(timerRef.current)
    countRef.current += 1
    if (countRef.current >= times) {
      countRef.current = 0
      onTrigger()
      return
    }
    timerRef.current = window.setTimeout(() => {
      countRef.current = 0
    }, intervalMs)
  }
}
