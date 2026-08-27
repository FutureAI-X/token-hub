import { useState, useEffect } from 'react'
import { User, Mail, Check, X, Loader2, Coins } from 'lucide-react'
import { getUserAvatarFallback, getUserAvatarStyle } from '../../lib/avatar'

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

export function Profile() {
  const [user, setUser] = useState<UserInfo | null>(null)
  const [loading, setLoading] = useState(true)
  const [email, setEmail] = useState('')
  const [editingEmail, setEditingEmail] = useState(false)
  const [saving, setSaving] = useState(false)
  const [message, setMessage] = useState({ type: '', text: '' })

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
          setEmail(res.data.email || '')
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  const handleSaveEmail = async () => {
    const token = localStorage.getItem('token')
    if (!token) return

    setSaving(true)
    setMessage({ type: '', text: '' })

    try {
      const res = await fetch('/api/user/email', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ email }),
      })
      const data = await res.json()
      if (data.success) {
        setUser((prev) => (prev ? { ...prev, email } : prev))
        setEditingEmail(false)
        setMessage({ type: 'success', text: '邮箱更新成功' })
        setTimeout(() => setMessage({ type: '', text: '' }), 3000)
      } else {
        setMessage({ type: 'error', text: data.message || '更新失败' })
      }
    } catch {
      setMessage({ type: 'error', text: '网络错误' })
    } finally {
      setSaving(false)
    }
  }

  const handleCancelEdit = () => {
    setEmail(user?.email || '')
    setEditingEmail(false)
    setMessage({ type: '', text: '' })
  }

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

  return (
    <div className='space-y-6'>
      {/* 页面标题 */}
      <div>
        <h2 className='text-2xl font-bold tracking-tight'>个人资料</h2>
        <p className='text-muted-foreground mt-1 text-sm'>
          管理您的账户信息
        </p>
      </div>

      {/* 消息提示 */}
      {message.text && (
        <div
          className={`rounded-lg px-4 py-3 text-sm ${
            message.type === 'success'
              ? 'bg-emerald-500/10 text-emerald-600 border border-emerald-500/20'
              : 'bg-destructive/10 text-destructive border border-destructive/20'
          }`}
        >
          {message.text}
        </div>
      )}

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
          <div className='flex-1 space-y-6'>
            {/* 用户名 */}
            <div className='flex items-center gap-3'>
              <div className='bg-muted flex size-10 items-center justify-center rounded-lg'>
                <User className='text-muted-foreground size-5' />
              </div>
              <div>
                <p className='text-muted-foreground text-xs'>用户名</p>
                <p className='text-sm font-medium'>{user.username}</p>
              </div>
            </div>

            {/* 邮箱 */}
            <div className='flex items-center gap-3'>
              <div className='bg-muted flex size-10 items-center justify-center rounded-lg'>
                <Mail className='text-muted-foreground size-5' />
              </div>
              <div className='flex-1'>
                <p className='text-muted-foreground text-xs'>邮箱</p>
                {editingEmail ? (
                  <div className='mt-1 flex items-center gap-2'>
                    <input
                      type='email'
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder='请输入邮箱'
                      className='bg-background border-border/60 focus-visible:ring-ring h-9 flex-1 rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                      autoFocus
                    />
                    <button
                      onClick={handleSaveEmail}
                      disabled={saving}
                      className='bg-emerald-500 hover:bg-emerald-600 inline-flex size-9 items-center justify-center rounded-lg text-white transition-colors disabled:opacity-50'
                    >
                      {saving ? (
                        <Loader2 className='size-4 animate-spin' />
                      ) : (
                        <Check className='size-4' />
                      )}
                    </button>
                    <button
                      onClick={handleCancelEdit}
                      disabled={saving}
                      className='bg-muted hover:bg-muted/80 inline-flex size-9 items-center justify-center rounded-lg transition-colors'
                    >
                      <X className='size-4' />
                    </button>
                  </div>
                ) : (
                  <p
                    className='hover:text-foreground cursor-pointer text-sm font-medium transition-colors'
                    onClick={() => setEditingEmail(true)}
                  >
                    {user.email || '点击设置邮箱'}
                  </p>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* 积分信息 */}
      <div className='border-border/60 rounded-xl border p-6'>
        <div className='flex items-center gap-3'>
          <div className='bg-emerald-500/10 flex size-10 items-center justify-center rounded-lg'>
            <Coins className='size-5 text-emerald-600' />
          </div>
          <div>
            <p className='text-muted-foreground text-xs'>当前积分</p>
            <p className='text-2xl font-bold tabular-nums'>{formatPoints(user.quota - user.used_quota)}</p>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── 格式化积分 ──
function formatPoints(points: number): string {
  if (points <= 0) return '0.00'
  return points.toFixed(2)
}
