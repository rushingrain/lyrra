import { useNavigate } from 'react-router-dom'
import { ArrowLeft, Search, User, Settings, Shirt, Crown, Mail } from 'lucide-react'

interface UserInfo {
  avatar?: string
  nickname?: string
  isVip?: boolean
}

interface TopBarProps {
  user?: UserInfo | null
}

export function TopBar({ user = null }: TopBarProps) {
  const navigate = useNavigate()

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      const query = (e.target as HTMLInputElement).value.trim()
      if (query) {
        navigate(`/search?q=${encodeURIComponent(query)}`)
      }
    }
  }

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
        {/* 账号信息：显示头像+VIP信息，未登录时显示未登录 */}
        <button className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1 transition-colors duration-200 hover:bg-gray-100">
          {user ? (
            <>
              <img
                src={user.avatar}
                alt={user.nickname}
                className="h-8 w-8 rounded-md object-cover ring-2 ring-[var(--color-accent)]/20"
              />
              {user.isVip && (
                <span className="flex items-center gap-0.5 rounded-full bg-[var(--color-accent)] px-1.5 py-0.5 text-[10px] font-semibold text-white">
                  <Crown size={10} />
                  VIP
                </span>
              )}
            </>
          ) : (
            <>
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100 text-[var(--color-text-secondary)]">
                <User size={16} />
              </div>
              <span className="text-sm text-[var(--color-text-secondary)]">未登录</span>
            </>
          )}
        </button>

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
    </div>
  )
}
