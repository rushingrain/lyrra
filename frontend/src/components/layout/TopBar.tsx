import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Search, User, Settings, Shirt, Mail, ChevronDown, Power } from 'lucide-react'
import { LoginModal } from '@/components/business/LoginModal'
import type { LoginResult } from '@/types'

interface UserInfo {
  avatar?: string
  nickname?: string
  isVip?: boolean
  friends?: number
  follows?: number
  fans?: number
}

interface TopBarProps {
  user?: UserInfo | null
  /** 登录成功回调：由 MainLayout 更新登录态 */
  onLoginSuccess?: (user: LoginResult['user']) => void
  /** 退出登录回调：由 MainLayout 调用后端并清空登录态 */
  onLogout?: () => void
}

export function TopBar({ user = null, onLoginSuccess, onLogout }: TopBarProps) {
  const navigate = useNavigate()
  const [loginModalOpen, setLoginModalOpen] = useState(false)
  const [cardOpen, setCardOpen] = useState(false)
  const accountRef = useRef<HTMLDivElement>(null)
  const closeLoginModal = useCallback(() => setLoginModalOpen(false), [])

  // 登录态切换（登录/登出）时关闭弹出的个人卡片；
  // 仅比较登录态本身，用户信息刷新导致对象引用变化时不应关闭卡片
  const prevUserRef = useRef(user)
  useEffect(() => {
    const wasLoggedIn = Boolean(prevUserRef.current)
    const isLoggedIn = Boolean(user)
    if (wasLoggedIn !== isLoggedIn) {
      setCardOpen(false)
    }
    prevUserRef.current = user
  }, [user])

  // 点击卡片外部 / ESC 关闭卡片
  useEffect(() => {
    if (!cardOpen) return
    const handlePointerDown = (e: MouseEvent | TouchEvent) => {
      if (accountRef.current && !accountRef.current.contains(e.target as Node)) {
        setCardOpen(false)
      }
    }
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setCardOpen(false)
    }
    document.addEventListener('mousedown', handlePointerDown)
    document.addEventListener('touchstart', handlePointerDown)
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('mousedown', handlePointerDown)
      document.removeEventListener('touchstart', handlePointerDown)
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [cardOpen])

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      const query = (e.target as HTMLInputElement).value.trim()
      if (query) {
        navigate(`/search?q=${encodeURIComponent(query)}`)
      }
    }
  }

  const handleLogout = () => {
    setCardOpen(false)
    onLogout?.()
  }

  // user_detail 接口返回了好友/关注/粉丝统计时才展示
  const hasStats = [user?.friends, user?.follows, user?.fans].some((n) => typeof n === 'number')

  // 统计数字格式化：万以上显示为「1.2万」
  const formatCount = (n: number) => (n >= 10000 ? `${(n / 10000).toFixed(1).replace(/\.0$/, '')}万` : String(n))

  return (
    <div className="flex w-full items-center justify-between">
      {/* 左侧：回退按钮 + 搜索框 */}
      <div className="flex items-center gap-3">
        <button
          className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-md border border-[var(--color-border)] bg-[var(--color-background)] text-[var(--color-text-secondary)] transition-colors duration-200 hover:bg-gray-100"
          aria-label="回退"
          onClick={() => navigate(-1)}
        >
          <ArrowLeft size={16} />
        </button>

        <div className="flex w-[280px] items-center gap-3 rounded-md border border-[var(--color-border)] bg-[var(--color-background)] px-3 py-2 transition-colors duration-200 focus-within:border-[var(--color-accent)]">
          <Search size={16} className="shrink-0 text-[var(--color-text-secondary)]" />
          <input
            type="text"
            placeholder="在这里输入你要搜索的内容"
            onKeyDown={handleKeyDown}
            className="flex-1 border-none bg-transparent text-sm outline-none placeholder:text-[var(--color-text-secondary)]/50 text-[var(--color-text-primary)]"
          />
        </div>
      </div>

      {/* 右侧按钮组 */}
      <div className="flex items-center">
        {/* 账号信息：头像在按钮外，只有用户名和 VIP 图标作为按钮；点击用户名弹出个人卡片 */}
        <div ref={accountRef} className="relative flex items-center">
          {user ? (
            <>
              <img
                src={user.avatar}
                alt={user.nickname}
                className="h-8 w-8 rounded-full object-cover"
              />

              <button
                type="button"
                aria-haspopup="menu"
                aria-expanded={cardOpen}
                onClick={() => setCardOpen((o) => !o)}
                className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm text-[var(--color-text-secondary)]"
              >
                <span className="max-w-[100px] truncate">{user.nickname}</span>

                <span className="flex h-8 w-8 shrink-0 items-center justify-center" title="VIP">
                  <svg viewBox="0 0 1024 1024" className="h-full w-full">
                    {/* 徽章背景：颜色随 VIP 状态切换 */}
                    <path
                      d="M768 307.2H256a153.6 153.6 0 0 0-153.6 153.6v102.4a153.6 153.6 0 0 0 153.6 153.6h512a153.6 153.6 0 0 0 153.6-153.6v-102.4a153.6 153.6 0 0 0-153.6-153.6z"
                      fill={user.isVip ? 'var(--color-accent)' : '#9CA3AF'}
                    />
                    {/* VIP 线条/文字：原 svg 的单个 path 拆分成 V / I / P 三部分 */}
                    <path
                      d="M353.536 640h-2.816a25.6 25.6 0 0 1-22.016-19.712l-46.592-204.8a25.6 25.6 0 1 1 51.2-11.264L362.496 537.6l67.84-139.264a25.6 25.6 0 1 1 45.824 22.528l-99.584 204.8a25.6 25.6 0 0 1-23.04 14.336z"
                      fill="#FFFFFF"
                    />
                    <path
                      d="M519.168 618.752a25.6 25.6 0 0 1-25.6 21.248 14.08 14.08 0 0 1-4.608 0 25.6 25.6 0 0 1-20.736-29.44l36.096-204.8a25.6 25.6 0 0 1 51.2 8.704z"
                      fill="#FFFFFF"
                    />
                    <path
                      d="M751.104 465.152A89.344 89.344 0 0 1 665.6 537.6h-29.696l-14.336 81.152a25.6 25.6 0 0 1-25.6 21.248 14.08 14.08 0 0 1-4.608 0 25.6 25.6 0 0 1-20.736-29.44l18.176-102.4a25.6 25.6 0 0 1 25.6-21.76h51.2a38.4 38.4 0 0 0 35.072-29.952 18.432 18.432 0 0 0-3.328-15.616 17.152 17.152 0 0 0-13.568-5.632h-51.2a25.6 25.6 0 0 1 0-51.2h51.2a68.352 68.352 0 0 1 52.736 23.808 69.888 69.888 0 0 1 14.592 57.344z"
                      fill="#FFFFFF"
                    />
                  </svg>
                </span>

                <ChevronDown
                  size={14}
                  className={`shrink-0 text-[var(--color-text-secondary)] transition-transform duration-200 ${cardOpen ? 'rotate-180' : ''
                    }`}
                />
              </button>

              {/* 个人卡片：好友/关注/粉丝统计 + 退出登录 */}
              {cardOpen && (
                <div
                  role="menu"
                  className="p-3 absolute right-0 top-full z-50 mt-2 w-60 overflow-hidden rounded-xl border border-[var(--color-border)] bg-[var(--color-background)] shadow-xl shadow-black/5"
                >

                  {/* 好友 / 关注 / 粉丝统计 */}
                  {hasStats && (
                    <div className="grid grid-cols-3 border-b border-[var(--color-border)]">
                      {[
                        { label: '好友', value: user.friends },
                        { label: '关注', value: user.follows },
                        { label: '粉丝', value: user.fans },
                      ].map((item) => (
                        <div key={item.label} className="flex flex-col items-center gap-0.5 py-3">
                          <span className="text-xl font-semibold text-[var(--color-text-primary)]">
                            {formatCount(item.value ?? 0)}
                          </span>
                          <span className="text-xs text-[var(--color-text-secondary)]">{item.label}</span>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* 退出登录 */}
                  <button
                    type="button"
                    role="menuitem"
                    onClick={handleLogout}
                    className="flex w-full cursor-pointer items-center gap-2 px-4 py-3 text-sm text-red-400 transition-colors duration-200 hover:bg-red-400/10"
                  >
                    <Power size={14} />
                    退出登录
                  </button>
                </div>
              )}
            </>
          ) : (
            <button
              className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1 transition-colors duration-200 hover:bg-gray-100"
              onClick={() => setLoginModalOpen(true)}
            >
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100 text-[var(--color-text-secondary)]">
                <User size={16} />
              </div>
              <span className="text-sm text-[var(--color-text-secondary)]">未登录</span>
            </button>
          )}
        </div>

        {/* 消息按钮 */}
        <button className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-md text-[var(--color-text-secondary)] transition-colors duration-200 hover:bg-gray-100">
          <Mail size={18} />
        </button>

        {/* 皮肤按钮 */}
        <button className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-md text-[var(--color-text-secondary)] transition-colors duration-200 hover:bg-gray-100">
          <Shirt size={18} />
        </button>

        {/* 设置按钮 */}
        <button className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-md text-[var(--color-text-secondary)] transition-colors duration-200 hover:bg-gray-100">
          <Settings size={18} />
        </button>
      </div>

      {/* 未登录时点击账号按钮弹出的扫码登录弹窗 */}
      <LoginModal
        open={loginModalOpen}
        onClose={closeLoginModal}
        onLoginSuccess={onLoginSuccess}
      />
    </div>
  )
}
