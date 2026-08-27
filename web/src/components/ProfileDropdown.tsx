import { useState, useRef, useEffect, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { User, LogOut } from 'lucide-react'
import { getUserAvatarFallback, getUserAvatarStyle } from '../lib/avatar'

const avatarFallbackClassName = 'font-semibold text-white'

interface ProfileDropdownProps {
  user: { username: string; display_name?: string; role?: string }
}

const roleLabels: Record<string, string> = {
  admin: '管理员',
  user: '普通用户',
  root: '超级管理员',
}

export function ProfileDropdown({ user }: ProfileDropdownProps) {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  const displayName = user.display_name || user.username
  const roleLabel = roleLabels[user.role || ''] || '用户'
  const avatarName = user.username || displayName
  const avatarFallback = getUserAvatarFallback(avatarName)
  const avatarFallbackStyle = useMemo(() => getUserAvatarStyle(avatarName), [avatarName])

  // 点击外部关闭
  useEffect(() => {
    if (!open) return
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  const handleLogout = () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    navigate('/login')
  }

  return (
    <div ref={ref} className='relative'>
      {/* 头像触发器 */}
      <button
        onClick={() => setOpen(!open)}
        className='relative flex size-7 items-center justify-center rounded-full transition-opacity hover:opacity-80'
      >
        <span
          className={`${avatarFallbackClassName} flex size-7 items-center justify-center rounded-full text-[11px]`}
          style={avatarFallbackStyle}
        >
          {avatarFallback}
        </span>
      </button>

      {/* 下拉菜单 */}
      {open && (
        <div className='bg-popover text-popover-foreground ring-border/50 absolute right-0 z-50 mt-2 w-56 overflow-hidden rounded-xl border shadow-[0_4px_24px_-4px_rgba(0,0,0,0.12)] ring-[0.5px] backdrop-blur-xl dark:shadow-[0_4px_24px_-4px_rgba(0,0,0,0.4)]'>
          {/* 用户信息区 */}
          <div className='flex items-center gap-2.5 px-3 py-3'>
            <span
              className={`${avatarFallbackClassName} flex size-9 items-center justify-center rounded-full text-xs`}
              style={avatarFallbackStyle}
            >
              {avatarFallback}
            </span>
            <div className='flex flex-1 flex-col gap-0.5 overflow-hidden'>
              <p className='text-foreground truncate text-sm font-medium'>
                {displayName}
              </p>
              <span className='text-muted-foreground text-xs'>
                {roleLabel}
              </span>
            </div>
          </div>

          {/* 分割线 */}
          <div className='bg-border mx-2 h-px' />

          {/* 菜单项 */}
          <div className='p-1'>
            <button
              onClick={() => { setOpen(false); navigate('/profile') }}
              className='hover:bg-accent flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-sm transition-colors'
            >
              <User className='size-4 text-muted-foreground' />
              个人中心
            </button>
            <button
              onClick={handleLogout}
              className='text-destructive hover:bg-destructive/10 flex w-full items-center gap-2 rounded-lg px-2.5 py-2 text-sm transition-colors'
            >
              <LogOut className='size-4' />
              退出登录
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
