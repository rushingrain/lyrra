import confetti from 'canvas-confetti'

/** 撒花起点：屏幕下方六成高度，视觉上像是从内容区飘上来 */
const ORIGIN = { y: 0.6 }

/**
 * shapeFromText 会把 emoji 渲染成 canvas 形状，比较费；缓存一次即可复用。
 * （旧版本 canvas-confetti 没有这个 API，调用点会先做存在性判断。）
 */
let cryShape: ReturnType<typeof confetti.shapeFromText> | null = null

function getCryShape() {
  if (!cryShape) {
    cryShape = confetti.shapeFromText({ text: '❄️', scalar: 2 })
  }
  return cryShape
}

/**
 * 彩蛋动画：成功撒花（celebrate），失败撒 ❄️。
 *
 * 这里只是把 canvas-confetti 包成两个无参函数，不持有任何 React 状态：
 * 动画的生命周期由 canvas-confetti 自己管理（会自建、自毁 canvas，不触发重渲染），
 * 所以不需要 useCallback 记忆——两个函数只在事件回调里被调用，不参与依赖数组。
 * 如果将来要传给 memo 组件或放进 useEffect 依赖，再考虑用 useCallback 包一层。
 */
export function useConfetti() {
  const celebrate = () => {
    confetti({ particleCount: 150, spread: 70, origin: ORIGIN })
  }

  const cry = () => {
    if (typeof confetti.shapeFromText !== 'function') return
    confetti({
      particleCount: 8,
      spread: 60,
      origin: ORIGIN,
      shapes: [getCryShape()],
      scalar: 2,
      gravity: 0.8,
      ticks: 120,
    })
  }

  return { celebrate, cry }
}
