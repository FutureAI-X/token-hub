import { useState, useEffect, useRef } from 'react'
import { Search, X, Check } from 'lucide-react'
import { cn } from '../../lib/utils'
import { getUsers, type AdminUser } from '../../api/admin'

interface UserSelectProps {
  value: number
  onChange: (id: number) => void
  placeholder?: string
}

// 可搜索的用户选择器（输入即过滤，配合按用户筛选）
export function UserSelect({ value, onChange, placeholder = '全部用户' }: UserSelectProps) {
  const [users, setUsers] = useState<AdminUser[]>([])
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const wrapRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    getUsers({ page_size: 100 })
      .then((res) => { if (res.success) setUsers(res.data.items || []) })
      .catch(() => {})
  }, [])

  const selected = users.find((u) => u.id === value)

  const filtered = query.trim()
    ? users.filter((u) =>
        `${u.username} ${u.display_name || ''}`.toLowerCase().includes(query.trim().toLowerCase())
      )
    : users

  const handleSelect = (id: number) => {
    onChange(id)
    setOpen(false)
    setQuery('')
  }

  // 点击组件外部时关闭下拉
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (wrapRef.current && !wrapRef.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  return (
    <div ref={wrapRef} className='relative'>
      <div className='relative'>
        <Search className='text-muted-foreground pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2' />
        <input
          value={open ? query : (selected ? selected.username : '')}
          placeholder={placeholder}
          onFocus={() => setOpen(true)}
          onChange={(e) => { setQuery(e.target.value); setOpen(true) }}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          className='border-border/60 bg-background focus-visible:ring-ring focus-visible:ring-ring flex h-9 w-56 rounded-lg border pr-9 pl-9 text-sm focus-visible:ring-2 focus-visible:outline-none'
        />
        {value !== 0 && (
          <button
            type='button'
            onClick={() => onChange(0)}
            className='text-muted-foreground hover:text-foreground absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 transition-colors'
            title='清除筛选'
          >
            <X className='size-4' />
          </button>
        )}
      </div>

      {open && (
        <div className='border-border/60 bg-background absolute z-50 mt-1 max-h-64 w-64 overflow-auto rounded-lg border shadow-lg'>
          <div
            className={cn(
              'hover:bg-muted/50 cursor-pointer px-3 py-2 text-sm transition-colors',
              value === 0 && 'bg-primary/10 text-primary'
            )}
            onMouseDown={() => handleSelect(0)}
          >
            全部用户
          </div>
          {filtered.map((u) => (
            <div
              key={u.id}
              onMouseDown={() => handleSelect(u.id)}
              className={cn(
                'hover:bg-muted/50 flex cursor-pointer items-center justify-between px-3 py-2 text-sm transition-colors',
                u.id === value && 'bg-primary/10 text-primary'
              )}
            >
              <span>
                {u.username}
                {u.display_name && (
                  <span className='text-muted-foreground ml-1.5 text-xs'>{u.display_name}</span>
                )}
              </span>
              {u.id === value && <Check className='text-primary size-4' />}
            </div>
          ))}
          {filtered.length === 0 && (
            <div className='text-muted-foreground px-3 py-2 text-sm'>无匹配用户</div>
          )}
        </div>
      )}
    </div>
  )
}
