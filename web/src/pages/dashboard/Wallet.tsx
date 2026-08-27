import { useState, useEffect } from 'react'
import { Wallet as WalletIcon, TrendingUp, Activity } from 'lucide-react'
import { cn } from '../../lib/utils'

interface UserInfo {
  id: number
  username: string
  display_name?: string
  role: string
  quota: number
  used_quota: number
}

export function Wallet() {
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
          <div className='bg-muted/30 h-24 rounded-xl' />
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

  const balance = user.quota - user.used_quota
  const usagePercent = user.quota > 0 ? (user.used_quota / user.quota) * 100 : 0

  return (
    <div className='space-y-6'>
      {/* 页面标题 */}
      <div>
        <h2 className='text-2xl font-bold tracking-tight'>钱包</h2>
        <p className='text-muted-foreground mt-1 text-sm'>
          查看您的账户余额和使用情况
        </p>
      </div>

      {/* 统计卡片 */}
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-3'>
        <StatsCard
          icon={WalletIcon}
          iconColor='text-emerald-600'
          iconBg='bg-emerald-500/10'
          label='当前余额'
          value={formatQuota(balance)}
          description='可用配额'
        />
        <StatsCard
          icon={TrendingUp}
          iconColor='text-blue-600'
          iconBg='bg-blue-500/10'
          label='总配额'
          value={formatQuota(user.quota)}
          description='账户总额度'
        />
        <StatsCard
          icon={Activity}
          iconColor='text-violet-600'
          iconBg='bg-violet-500/10'
          label='已使用'
          value={formatQuota(user.used_quota)}
          description={`使用率 ${usagePercent.toFixed(1)}%`}
        />
      </div>

      {/* 使用进度 */}
      <div className='border-border/60 rounded-xl border p-6'>
        <h3 className='mb-4 text-sm font-semibold'>配额使用情况</h3>
        <div className='space-y-3'>
          <div className='flex items-center justify-between text-sm'>
            <span className='text-muted-foreground'>已使用</span>
            <span className='font-mono font-medium'>
              {formatQuota(user.used_quota)} / {formatQuota(user.quota)}
            </span>
          </div>
          <div className='bg-muted h-3 overflow-hidden rounded-full'>
            <div
              className={cn(
                'h-full rounded-full transition-all duration-500',
                usagePercent > 80
                  ? 'bg-destructive'
                  : usagePercent > 50
                    ? 'bg-amber-500'
                    : 'bg-emerald-500'
              )}
              style={{ width: `${Math.min(usagePercent, 100)}%` }}
            />
          </div>
          <div className='flex items-center justify-between text-xs'>
            <span className='text-muted-foreground'>
              剩余 {formatQuota(balance)}
            </span>
            <span
              className={cn(
                'font-medium',
                usagePercent > 80
                  ? 'text-destructive'
                  : usagePercent > 50
                    ? 'text-amber-600'
                    : 'text-emerald-600'
              )}
            >
              {usagePercent.toFixed(1)}%
            </span>
          </div>
        </div>
      </div>

      {/* 提示信息 */}
      <div className='bg-muted/30 rounded-xl p-5'>
        <h4 className='mb-2 text-sm font-medium'>关于配额</h4>
        <ul className='text-muted-foreground space-y-1.5 text-xs leading-relaxed'>
          <li>• 配额用于 API 调用计费，不同模型消耗不同额度</li>
          <li>• 按量计费模型根据 token 数量消耗配额</li>
          <li>• 按次计费模型每次调用消耗固定配额</li>
          <li>• 具体价格请参考模型广场</li>
        </ul>
      </div>
    </div>
  )
}

// ── 统计卡片 ──
function StatsCard({
  icon: Icon,
  iconColor,
  iconBg,
  label,
  value,
  description,
}: {
  icon: React.ElementType
  iconColor: string
  iconBg: string
  label: string
  value: string
  description: string
}) {
  return (
    <div className='border-border/60 rounded-xl border p-5'>
      <div className='flex items-center gap-3'>
        <div className={cn('flex size-10 items-center justify-center rounded-xl', iconBg)}>
          <Icon className={cn('size-5', iconColor)} />
        </div>
        <div>
          <p className='text-muted-foreground text-xs font-medium uppercase tracking-wider'>
            {label}
          </p>
          <p className='text-2xl font-bold tabular-nums'>{value}</p>
        </div>
      </div>
      <p className='text-muted-foreground mt-3 text-xs'>{description}</p>
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
