import { type ReactNode, useState } from 'react'
import { Sidebar } from './Sidebar'
import { MiniPlayer } from './MiniPlayer'
import { TopBar } from '../home/TopBar'
import { mockPlayHistory } from '../../data/mock'
import SimpleBar from 'simplebar-react'
import 'simplebar-react/dist/simplebar.min.css'

interface MainLayoutProps {
  children: ReactNode
}

// 主体内容最大宽度：窗口很宽时限制并居中内容，避免内容被拉得过宽；
// 窗口较窄时 max-width 不生效，内容自然铺满。如需调整可改此值。
const CONTENT_WRAPPER = 'mx-auto w-full max-w-[1400px]'

export function MainLayout({ children }: MainLayoutProps) {
  const [playHistory] = useState(() => mockPlayHistory.length > 0 ? mockPlayHistory : [])
  const showPlayer = playHistory.length > 0

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--color-background)]">
      <Sidebar showPlayer={showPlayer} />

      <div className="flex flex-1 flex-col overflow-hidden">
        <header
          className="flex h-[72px] min-h-[72px] items-center bg-[var(--color-background)] px-10"
        >
          <div className={CONTENT_WRAPPER}>
            <TopBar />
          </div>
        </header>

        <SimpleBar
          className="h-full scrollbar"
        >
          <main
            className={`py-8 px-10 ${showPlayer ? 'pb-40' : 'pb-16'}`}
          >
            <div className={CONTENT_WRAPPER}>
              {children}
            </div>
          </main>
        </SimpleBar>
      </div>

      {showPlayer && (
        <div className="fixed bottom-0 left-0 right-0 z-50 h-[82px] bg-[var(--color-background)]">
          <MiniPlayer />
        </div>
      )}
    </div>
  )
}
