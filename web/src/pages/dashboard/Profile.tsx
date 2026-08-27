import { useState, useEffect } from 'react'
import { User, Mail, Shield, Calendar } from 'lucide-react'
import { getUserAvatarFallback, getUserAvatarStyle } from '../../lib/avatar'
import { cn } from '../../lib/utils'

interface UserInfo {
  id: number
  username: string
  display_name?: string
  role: string
  status: number
  email?: string
  quota: number
  used_quota: number
}

const roleLabels: Record<string, string> = {
  root: '超级管理员',
  admin: '管理员',
  user: '普通用户',
}

export function Profile() {
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      setLoading(false)
      return
    }

    fetch('/api/user/info', {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((res) => res.json())
      .then((res) => {
        if (res.success) {
          setUser(res.data)
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className='space-y-6'>
        <div className='animate-pulse space-y-4'>
          <div className='bg-muted/30 h-32 rounded-xl' />
          <div className='bg-muted/30 h-48 rounded-xl' />
        </div>
      </div>
    )
  }

  if (!user) {
    return (
      <div className='flex flex-col items-center justify-center py-20'>
        <p className='text-muted-foreground text-lg'>请先登录</p>
      </div>
    )
  }

  const avatarFallback = getUserAvatarFallback(user.username)
  const avatarStyle = getUserAvatarStyle(user.username)
  const displayName = user.display_name || user.username

  return (
    <div className='space-y-6'>
      {/* 页面标题 */}
      <div>
        <h2 className='text-2xl font-bold tracking-tight'>个人资料</h2>
        <p className='text-muted-foreground mt-1 text-sm'>
          管理您的账户信息
        </p>
      </div>

      {/* 用户信息卡片 */}
      <div className='border-border/60 rounded-xl border p-6'>
        <div className='flex flex-col items-start gap-6 sm:flex-row'>
          {/* 头像 */}
          <div className='shrink-0'>
            <span
              className='flex size-20 items-center justify-center rounded-2xl text-2xl font-bold text-white'
              style={avatarStyle}
            >
              {avatarFallback}
            </span>
          </div>

          {/* 基本信息 */}
          <div className='flex-1 space-y-4'>
            <div>
              <h3 className='text-xl font-semibold'>{displayName}</h3>
              <p className='text-muted-foreground mt-0.5 text-sm'>
                @{user.username}
              </p>
            </div>

            <div className='grid grid-cols-1 gap-4 sm:grid-cols-2'>
              <InfoItem
                icon={User}
                label='用户名'
                value={user.username}
              />
              <InfoItem
                icon={Mail}
                label='邮箱'
                value={user.email || '未设置'}
              />
              <InfoItem
                icon={Shield}
                label='角色'
                value={roleLabels[user.role] || user.role}
              />
              <InfoItem
                icon={Calendar}
                label='状态'
                value={
                  <span
                    className={cn(
                      'inline-flex items-center gap-1.5 text-sm',
                      user.status === 1 ? 'text-emerald-600' : 'text-destructive'
                    )}
                  >
                    <span
                      className={cn(
                        'size-1.5 rounded-full',
                        user.status === 1 ? 'bg-emerald-500' : 'bg-destructive'
                      )}
                    />
                    {user.status === 1 ? '正常' : '已禁用'}
                  </span>
                }
              />
            </div>
          </div>
        </div>
      </div>

      {/* 账户统计 */}
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-3'>
        <StatCard
          label='总配额'
          value={formatQuota(user.quota)}
          description='账户总配额额度'
        />
        <StatCard
          label='已使用'
          value={formatQuota(user.used_quota)}
          description='已消耗的配额'
        />
        <StatCard
          label='剩余配额'
          value={formatQuota(user.quota - user.used_quota)}
          description='可用配额余额'
          highlight
        />
      </div>
    </div>
  )
}

// ── 信息项 ──
function InfoItem({
  icon: Icon,
  label,
  value,
}: {
  icon: React.ElementType
  label: string
  value: React.ReactNode
}) {
  return (
    <div className='flex items-center gap-3'>
      <div className='bg-muted flex size-9 items-center justify-center rounded-lg'>
        <Icon className='text-muted-foreground size-4' />
      </div>
      <div>
        <p className='text-muted-foreground text-xs'>{label}</p>
        <p className='text-sm font-medium'>{value}</p>
      </div>
    </div>
  )
}

// ── 统计卡片 ──
function StatCard({
  label,
  value,
  description,
  highlight = false,
}: {
  label: string
  value: string
  description: string
  highlight?: boolean
}) {
  return (
    <div
      className={cn(
        'rounded-xl border p-5',
        highlight
          ? 'border-emerald-500/30 bg-emerald-500/5'
          : 'border-border/60'
      )}
    >
      <p className='text-muted-foreground text-xs font-medium uppercase tracking-wider'>
        {label}
      </p>
      <p
        className={cn(
          'mt-2 text-2xl font-bold tabular-nums',
          highlight ? 'text-emerald-600' : ''
        )}
      >
        {value}
      </p>
      <p className='text-muted-foreground mt-1 text-xs'>{description}</p>
    </div>
  )
}

// ── 格式化配额 ──
function formatQuota(quota: number): string {
  if (quota <= 0) return '0'
  if (quota >= 1000000) return `${(quota / 1000000).toFixed(2)}M`
  if (quota >= 1000) return `${(quota / 1000).toFixed(2)}K`
  return quota.toString()
}
