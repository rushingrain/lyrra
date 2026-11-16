import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import type { FormEvent, ReactNode } from 'react'
import { useLocation, useSearchParams } from 'react-router-dom'
import { Check } from 'lucide-react'
import { DEFAULT_THEME_COLOR, normalizeHexColor, PRESET_THEMES, useThemeStore } from '@/stores/themeStore'
import { usePlayerStore } from '@/stores/playerStore'

const QUALITY_OPTIONS = [
  { value: 'standard', label: '标准' },
  { value: 'hq', label: 'HQ（高品）' },
  { value: 'sq', label: 'SQ（无损）' },
  { value: 'res', label: 'HI-RES（高解析）' },
] as const

const SECTIONS = [
  { id: 'skin', label: '皮肤' },
  { id: 'playback', label: '播放' },
  { id: 'desktop-lyrics', label: '桌面歌词' },
  { id: 'about', label: '关于' },
] as const

type SectionId = typeof SECTIONS[number]['id']

function SettingsSection({ id, title, children }: { id: SectionId; title: string; children: ReactNode }) {
  return (
    <section
      id={`settings-${id}`}
      aria-labelledby={`settings-${id}-title`}
      className="grid items-start grid-cols-[96px_minmax(0,1fr)] gap-8 border-b border-[var(--color-border)] py-7 last:border-b-0"
    >
      <h2 id={`settings-${id}-title`} className="relative top-0.5 text-base font-semibold leading-7">{title}</h2>
      <div className="min-w-0">{children}</div>
    </section>
  )
}

function ThemeSettings() {
  const { color, ready, saving, setColor } = useThemeStore()
  const [customHex, setCustomHex] = useState('')
  const [error, setError] = useState('')
  const customColor = normalizeHexColor(customHex)
  const currentName = color === DEFAULT_THEME_COLOR ? '默认' : PRESET_THEMES.find((theme) => theme.color === color)?.name ?? '自定义'
  const disabled = !ready || saving

  const restoreDefaultColor = async () => {
    if (await setColor(DEFAULT_THEME_COLOR)) {
      setCustomHex('')
      setError('')
    }
  }

  const applyCustomColor = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (!customColor) {
      setError('请输入六位 HEX 色值，如 #3D8E86')
      return
    }
    setError('')
    if (await setColor(customColor)) setCustomHex(customColor)
  }

  return (
    <div>
      <div className="mb-5 flex flex-wrap items-center justify-between gap-2">
        <h3 className="text-base font-medium leading-7">主题色</h3>
        <div className="flex flex-wrap items-center gap-4">
          <span className="flex items-center gap-2 text-sm text-[var(--color-text-secondary)]" aria-live="polite">
            <span className="h-2 w-2 rounded-full border border-black/10" style={{ backgroundColor: color }} />
            {currentName} · {color}
          </span>
          <button
            type="button"
            disabled={disabled}
            onClick={() => void restoreDefaultColor()}
            className="cursor-pointer rounded-md border border-[var(--color-border)] px-3 py-1.5 text-sm text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-surface)] disabled:cursor-default disabled:opacity-60"
          >
            恢复默认
          </button>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3" role="group" aria-label="预设主题色" aria-busy={saving}>
        {PRESET_THEMES.map((theme) => {
          const selected = theme.color === color
          return (
            <button
              key={theme.color}
              type="button"
              aria-pressed={selected}
              disabled={disabled}
              onClick={() => void setColor(theme.color)}
              className={`relative flex min-w-0 cursor-pointer items-center gap-3 rounded-xl border px-3 py-4 text-left transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-accent)] disabled:cursor-default ${selected
                ? 'border-[var(--color-accent)] bg-[var(--color-accent-bg)]'
                : 'border-[var(--color-border)] hover:bg-[var(--color-surface)]'
                }`}
            >
              <span
                className="h-8 w-8 shrink-0 rounded-full border border-black/10"
                style={{ backgroundColor: theme.color }}
              />
              <span className="min-w-0">
                <span className="block text-[15px] font-medium">{theme.name}</span>
                <span className="mt-1 block font-mono text-xs text-[var(--color-text-secondary)]">{theme.color}</span>
              </span>
              {selected && <Check size={12} aria-hidden className="absolute right-2 top-2 text-[var(--color-text-primary)]" />}
            </button>
          )
        })}
      </div>

      <form onSubmit={(event) => void applyCustomColor(event)} className="mt-6 border-t border-[var(--color-border)] pt-5" noValidate>
        <label htmlFor="theme-hex" className="text-base font-medium">自定义颜色</label>
        <p id="theme-hex-hint" className="mb-3 mt-1 text-sm leading-5 text-[var(--color-text-secondary)]">
          输入六位 HEX 色值
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <div className={`flex items-center gap-3 rounded-lg border bg-[var(--color-background)] px-3 focus-within:border-[var(--color-accent)] ${error ? 'border-red-500' : 'border-[var(--color-border)]'}`}>
            <span
              className="h-5 w-5 rounded border border-black/10"
              style={{ backgroundColor: customColor ?? color }}
              aria-label={`颜色预览 ${customColor ?? color}`}
              role="img"
            />
            <input
              id="theme-hex"
              type="text"
              value={customHex}
              placeholder={color}
              maxLength={7}
              autoComplete="off"
              spellCheck={false}
              disabled={disabled}
              aria-invalid={Boolean(error)}
              aria-describedby={`theme-hex-hint${error ? ' theme-hex-error' : ''}`}
              onChange={(event) => {
                setCustomHex(event.target.value)
                setError('')
              }}
              className="h-10 w-28 bg-transparent font-mono text-[15px] uppercase outline-none placeholder:text-[var(--color-text-secondary)]/60"
            />
          </div>
          <button
            type="submit"
            disabled={disabled}
            className="h-10 cursor-pointer rounded-lg bg-[var(--color-accent)] px-5 text-[15px] font-medium text-[var(--color-on-accent)] transition-colors hover:bg-[var(--color-accent-hover)] disabled:cursor-default disabled:opacity-60"
          >
            应用颜色
          </button>
        </div>
        {error && <p id="theme-hex-error" role="alert" className="mt-2 text-sm text-red-500">{error}</p>}
        <p className="mt-3 text-sm text-[var(--color-text-secondary)]" role="status">
          {!ready ? '正在读取主题设置…' : saving ? '正在保存主题色…' : '主题色自动保存，重启后保留'}
        </p>
      </form>
    </div>
  )
}

export default function Settings() {
  const pageRef = useRef<HTMLDivElement>(null)
  const stickyHeaderRef = useRef<HTMLDivElement>(null)
  const qualityPreference = usePlayerStore((state) => state.qualityPreference)
  const deviceSettingsStatus = usePlayerStore((state) => state.deviceSettingsStatus)
  const setQualityPreference = usePlayerStore((state) => state.setQualityPreference)
  const [searchParams, setSearchParams] = useSearchParams()
  const location = useLocation()
  const requestedSection = searchParams.get('section')
  const target = SECTIONS.find((section) => section.id === requestedSection)?.id ?? 'skin'
  const [activeSection, setActiveSection] = useState<SectionId>(target)

  useLayoutEffect(() => {
    // 直接滚动 SimpleBar 内容区，避免 scrollIntoView 连带滚动外层窗口。
    const scrollRoot = pageRef.current?.closest('.simplebar-content-wrapper')
    if (!scrollRoot) return

    if (target === 'skin') {
      scrollRoot.scrollTo({ top: 0, behavior: 'instant' })
      return
    }

    const section = pageRef.current?.querySelector<HTMLElement>(`#settings-${target}`)
    if (!section) return

    const rootTop = scrollRoot.getBoundingClientRect().top
    const sectionTop = section.getBoundingClientRect().top
    const stickyHeight = stickyHeaderRef.current?.getBoundingClientRect().height ?? 0
    const top = scrollRoot.scrollTop + sectionTop - rootTop - stickyHeight
    scrollRoot.scrollTo({ top: Math.max(0, top), behavior: 'instant' })
  }, [target, location.key])

  useEffect(() => {
    // 页面复用 MainLayout 的 SimpleBar，只滚动主内容区。
    const scrollRoot = pageRef.current?.closest('.simplebar-content-wrapper')
    if (!scrollRoot) return
    const updateActiveSection = () => {
      const rootRect = scrollRoot.getBoundingClientRect()
      const stickyHeight = stickyHeaderRef.current?.getBoundingClientRect().height ?? 0
      const threshold = rootRect.top + stickyHeight + 8
      const atBottom = scrollRoot.scrollTop + scrollRoot.clientHeight >= scrollRoot.scrollHeight - 2
      if (atBottom) {
        setActiveSection('about')
        return
      }

      let active: SectionId = 'skin'
      for (const section of SECTIONS) {
        const element = pageRef.current?.querySelector(`#settings-${section.id}`)
        if (element && element.getBoundingClientRect().top <= threshold) active = section.id
      }
      setActiveSection(active)
    }
    updateActiveSection()
    scrollRoot.addEventListener('scroll', updateActiveSection, { passive: true })
    window.addEventListener('resize', updateActiveSection)
    return () => {
      scrollRoot.removeEventListener('scroll', updateActiveSection)
      window.removeEventListener('resize', updateActiveSection)
    }
  }, [])

  return (
    <div ref={pageRef} className="text-[var(--color-text-primary)]">
      <div ref={stickyHeaderRef} className="sticky top-0 z-20 -mx-10 bg-[var(--color-background)] px-10">
        <h1 className="pb-3 text-2xl font-semibold">设置</h1>
        <nav aria-label="设置分类" className="flex gap-8 border-b border-[var(--color-border)]">
          {SECTIONS.map((section) => (
            <button
              key={section.id}
              type="button"
              aria-current={activeSection === section.id ? 'location' : undefined}
              aria-controls={`settings-${section.id}`}
              onClick={() => {
                setActiveSection(section.id)
                setSearchParams({ section: section.id }, { replace: true })
              }}
              className={`relative cursor-pointer whitespace-nowrap py-3 text-base transition-colors after:absolute after:inset-x-0 after:bottom-0 after:h-0.5 after:rounded-full ${activeSection === section.id
                ? 'font-semibold text-[var(--color-accent)] after:bg-[var(--color-accent)]'
                : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'
                }`}
            >
              {section.label}
            </button>
          ))}
        </nav>
      </div>

      <SettingsSection id="skin" title="皮肤">
        <ThemeSettings />
      </SettingsSection>

      <SettingsSection id="playback" title="播放">
        <fieldset disabled={deviceSettingsStatus !== 'ready'}>
          <legend className="relative top-0.5 text-base font-medium leading-7">音质播放设置</legend>
          <p className="mb-4 mt-1 text-sm text-[var(--color-text-secondary)]">标准以上的品质需要 VIP 才生效</p>
          <div className="flex flex-col gap-3">
            {QUALITY_OPTIONS.map((quality) => (
              <label
                key={quality.value}
                className={`flex items-center gap-2 text-[15px] ${deviceSettingsStatus === 'ready' ? 'cursor-pointer' : 'cursor-default'}`}
              >
                <input
                  type="radio"
                  name="quality"
                  value={quality.value}
                  checked={qualityPreference === quality.value}
                  onChange={() => setQualityPreference(quality.value)}
                  className="h-4 w-4 appearance-none rounded-full border border-gray-400 bg-white checked:border-[5px] checked:border-[var(--color-accent)] disabled:opacity-100"
                />
                <span>{quality.label}</span>
              </label>
            ))}
          </div>
          {deviceSettingsStatus === 'loading' && (
            <p role="status" className="mt-3 text-sm text-[var(--color-text-secondary)]">正在读取音质设置…</p>
          )}
          {deviceSettingsStatus === 'error' && (
            <p role="alert" className="mt-3 text-sm text-red-500">设备设置读取失败，请重启应用后重试</p>
          )}
        </fieldset>
      </SettingsSection>

      <SettingsSection id="desktop-lyrics" title="桌面歌词">
        <label className="relative top-0.5 flex min-h-7 cursor-default items-center gap-3">
          <input
            type="checkbox"
            disabled
            className="h-4 w-4 shrink-0 appearance-none rounded border border-gray-400 bg-white checked:border-[var(--color-accent)] checked:bg-[var(--color-accent)] disabled:opacity-100"
          />
          <div>
            <h3 className="text-[15px] font-medium leading-7">开启桌面歌词</h3>
          </div>
        </label>
      </SettingsSection>

      <SettingsSection id="about" title="关于">
        <div className="mb-5">
          <span className="relative top-0.5 text-xl font-medium leading-7 tracking-tight text-[var(--color-text-primary)]">Lyrra</span>
          <p className="mt-2 text-sm text-[var(--color-text-secondary)]">第三方酷狗音乐桌面客户端</p>
        </div>
        <div>
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 py-4 text-base">
            <span>当前版本</span>
            <span className="text-[var(--color-text-secondary)]">{__APP_VERSION__ === '开发版' ? __APP_VERSION__ : `v${__APP_VERSION__}`}</span>
          </div>
          <div className="border-t border-[var(--color-border)] py-4">
            <h3 className="text-base">免责声明</h3>
            <p className="mt-3 text-sm leading-6 text-[var(--color-text-secondary)]">
              本项目为个人学习和技术研究项目，非酷狗音乐官方客户端，未获得酷狗音乐或其关联公司的授权、认可或支持。
              音乐、封面及歌词等内容的相关权利归原权利人所有。请遵守相关法律法规及平台服务条款，尊重版权，支持正版音乐。
            </p>
          </div>
          <div className="flex flex-wrap gap-x-6 gap-y-2 pt-4 text-sm">
            <a className="text-[var(--color-text-secondary)] underline underline-offset-4 decoration-[var(--color-border)] hover:text-[var(--color-accent)]">《意见反馈》</a>
            <a className="text-[var(--color-text-secondary)] underline underline-offset-4 decoration-[var(--color-border)] hover:text-[var(--color-accent)]">《酷狗音乐官网》</a>
          </div>
        </div>
      </SettingsSection>
    </div>
  )
}
