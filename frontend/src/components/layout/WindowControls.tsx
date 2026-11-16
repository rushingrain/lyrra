import { useEffect, useState } from 'react'
import { Copy, Minus, Square, X } from 'lucide-react'
import { Quit, WindowIsMaximised, WindowMinimise, WindowToggleMaximise } from '../../../wailsjs/runtime/runtime'

const BUTTON_CLASS = 'titlebar-control flex h-full w-[46px] cursor-pointer items-center justify-center text-[var(--color-on-accent)] transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-inset focus-visible:ring-[var(--color-on-accent)]'

export function WindowControls() {
  const [isMaximised, setIsMaximised] = useState(false)

  useEffect(() => {
    let active = true
    const updateMaximisedState = () => {
      WindowIsMaximised()
        .then((maximised) => {
          if (active) setIsMaximised(maximised)
        })
        .catch(() => {})
    }
    updateMaximisedState()
    window.addEventListener('resize', updateMaximisedState)
    return () => {
      active = false
      window.removeEventListener('resize', updateMaximisedState)
    }
  }, [])

  return (
    <div className="flex h-full shrink-0" role="group" aria-label="窗口控制" onDoubleClick={(event) => event.stopPropagation()}>
      <button type="button" className={`${BUTTON_CLASS} titlebar-control-neutral`} onClick={WindowMinimise} title="最小化" aria-label="最小化">
        <Minus size={14} aria-hidden="true" />
      </button>
      <button
        type="button"
        className={`${BUTTON_CLASS} titlebar-control-neutral`}
        onClick={WindowToggleMaximise}
        title={isMaximised ? '还原' : '最大化'}
        aria-label={isMaximised ? '还原' : '最大化'}
      >
        {isMaximised ? <Copy size={12} aria-hidden="true" /> : <Square size={12} aria-hidden="true" />}
      </button>
      <button
        type="button"
        className={[BUTTON_CLASS, 'hover:bg-[#e81123]', 'hover:text-white'].join(' ')}
        onClick={Quit}
        title="关闭"
        aria-label="关闭"
      >
        <X size={16} aria-hidden="true" />
      </button>
    </div>
  )
}
