import { useState, useEffect } from 'react'
import { Link, useLocation, Outlet } from 'react-router-dom'
import { User, Wallet, PanelLeft } from 'lucide-react'
import { cn } from '../../lib/utils'
import { ProfileDropdown } from '../../components/ProfileDropdown'

// ── 侧边栏导航项 ──
const sidebarNav = [
  {
    group: '个人',
    items: [
      { title: '个人资料', href: '/dashboard/profile', icon: User },
      { title: '钱包', href: '/dashboard/wallet', icon: Wallet },
    ],
  },
]

// ── 主布局 ──
export function DashboardLayout() {
  const location = useLocation()
  const [user, setUser] = useState<{ username: string; display_name?: string; role?: string } | null>(null)
  const [scrolled, setScrolled] = useState(false)
  const [sidebarOpen, setSidebarOpen] = useState(true)

  // 滚动检测
  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  // 读取用户信息
  useEffect(() => {
    const stored = localStorage.getItem('user')
    if (stored) {
      try {
        setUser(JSON.parse(stored))
      } catch {
        // ignore
      }
    }
  }, [])

  return (
    <div className='bg-background text-foreground min-h-svh'>
      {/* ── 顶部 Header（固定，不受侧边栏影响） ── */}
      <header className='pointer-events-none fixed inset-x-0 top-0 z-50'>
        <div
          className={cn(
            'pointer-events-auto mx-auto transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]',
            scrolled ? 'max-w-[52rem] px-3 pt-3' : 'max-w-7xl px-4 pt-0 md:px-6'
          )}
        >
          <nav
            className={cn(
              'flex items-center justify-between transition-all duration-700 ease-[cubic-bezier(0.16,1,0.3,1)]',
              scrolled
                ? 'bg-background/60 ring-border/50 h-12 rounded-2xl pr-1.5 pl-4 shadow-[0_2px_16px_-6px_rgba(0,0,0,0.08),0_0_0_0.5px_rgba(0,0,0,0.02)] ring-[0.5px] backdrop-blur-2xl dark:shadow-[0_2px_16px_-6px_rgba(0,0,0,0.4)]'
                : 'h-16 px-2'
            )}
          >
            {/* 左侧：收起按钮 + Logo */}
            <div className='flex items-center gap-1.5'>
              <button
                onClick={() => setSidebarOpen(!sidebarOpen)}
                className='hover:bg-muted size-8 rounded-lg transition-colors'
              >
                <PanelLeft className='mx-auto size-4' />
              </button>
              <a href='/' className='group flex shrink-0 items-center gap-2.5'>
                <div className='flex size-7 shrink-0 items-center justify-center transition-all duration-300 group-hover:scale-105'>
                  <span className='text-lg'>⚡</span>
                </div>
                <span className='text-sm font-semibold tracking-tight'>Token Hub</span>
              </a>
            </div>

            {/* 右侧：导航菜单 */}
            <div className='hidden items-center gap-0.5 sm:flex'>
              {[
                { title: '主页', href: '/' },
                { title: '控制台', href: '/dashboard/profile' },
                { title: '模型广场', href: '/pricing' },
              ].map((link) => {
                const isActive =
                  link.href === '/'
                    ? location.pathname === '/'
                    : location.pathname.startsWith(link.href)
                return (
                  <Link
                    key={link.href}
                    to={link.href}
                    className={cn(
                      'rounded-lg px-3 py-1.5 text-sm font-medium transition-colors',
                      isActive
                        ? 'text-foreground'
                        : 'text-muted-foreground hover:text-foreground'
                    )}
                  >
                    {link.title}
                  </Link>
                )
              })}
              <a
                href='https://future-ai.feishu.cn/wiki/Gs3Ow9fSfinMpwkCNuschCcYnab'
                target='_blank'
                rel='noopener noreferrer'
                className='text-muted-foreground hover:text-foreground rounded-lg px-3 py-1.5 text-sm font-medium transition-colors'
              >
                文档
              </a>
              <div className='bg-border/40 mx-2 h-4 w-px' />
              {user ? (
                <ProfileDropdown user={user} />
              ) : (
                <Link
                  to='/login'
                  className='bg-foreground text-background inline-flex h-8 items-center justify-center rounded-lg px-3.5 text-xs font-medium transition-opacity hover:opacity-90'
                >
                  登录
                </Link>
              )}
            </div>
          </nav>
        </div>
      </header>

      {/* ── 主体：侧边栏 + 内容 ── */}
      <div className='flex pt-16'>
        {/* 侧边栏 */}
        <aside
          className={cn(
            'border-border/40 sticky top-16 z-30 h-[calc(100svh-4rem)] shrink-0 border-r transition-all duration-300',
            sidebarOpen ? 'w-60' : 'w-12'
          )}
        >
          {/* 展开状态：完整菜单 */}
          {sidebarOpen && (
            <nav className='flex h-full flex-col space-y-6 overflow-y-auto p-4'>
              {sidebarNav.map((group) => (
                <div key={group.group}>
                  <h3 className='text-muted-foreground mb-2 px-2 text-[11px] font-semibold tracking-wider uppercase'>
                    {group.group}
                  </h3>
                  <div className='space-y-0.5'>
                    {group.items.map((item) => {
                      const Icon = item.icon
                      const isActive = location.pathname === item.href
                      return (
                        <Link
                          key={item.href}
                          to={item.href}
                          className={cn(
                            'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                            isActive
                              ? 'bg-accent text-accent-foreground'
                              : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                          )}
                        >
                          <Icon className='size-4' />
                          {item.title}
                        </Link>
                      )
                    })}
                  </div>
                </div>
              ))}
            </nav>
          )}

          {/* 收起状态：只显示图标 */}
          {!sidebarOpen && (
            <nav className='flex flex-col items-center gap-1 px-2 pt-2'>
              {sidebarNav.flatMap((group) =>
                group.items.map((item) => {
                  const Icon = item.icon
                  const isActive = location.pathname === item.href
                  return (
                    <Link
                      key={item.href}
                      to={item.href}
                      className={cn(
                        'flex size-8 items-center justify-center rounded-lg transition-colors',
                        isActive
                          ? 'bg-accent text-accent-foreground'
                          : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground'
                      )}
                      title={item.title}
                    >
                      <Icon className='size-4' />
                    </Link>
                  )
                })
              )}
            </nav>
          )}
        </aside>

        {/* 内容区 */}
        <main className='min-w-0 flex-1'>
          <div className='mx-auto max-w-5xl px-4 py-6 sm:px-6 lg:px-8'>
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
