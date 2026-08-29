import { useState, useEffect } from 'react'
import { X, Loader2, Check, Globe } from 'lucide-react'
import { cn } from '../../lib/utils'
import { getModelEndpoints, syncModelEndpoints } from '../../api/admin-model'
import { getEndpoints, type Endpoint } from '../../api/endpoint'

interface ModelEndpointManagerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  modelId: number
  modelName: string
}

export function ModelEndpointManager({ open, onOpenChange, modelId, modelName }: ModelEndpointManagerProps) {
  const [allEndpoints, setAllEndpoints] = useState<Endpoint[]>([])
  const [selectedIds, setSelectedIds] = useState<number[]>([])
  const [loading, setLoading] = useState(false)
  const [saving, setSaving] = useState(false)
  const [msg, setMsg] = useState('')

  useEffect(() => {
    if (!open) return
    setLoading(true)
    setMsg('')
    Promise.all([
      getEndpoints(),
      getModelEndpoints(modelId),
    ]).then(([epRes, bindRes]) => {
      if (epRes.success) setAllEndpoints((epRes.data || []).filter(e => e.status === 1))
      if (bindRes.success) setSelectedIds((bindRes.data || []).map(b => b.endpoint_id))
    }).catch(() => {}).finally(() => setLoading(false))
  }, [open, modelId])

  useEffect(() => {
    if (open) document.body.style.overflow = 'hidden'
    else document.body.style.overflow = ''
    return () => { document.body.style.overflow = '' }
  }, [open])

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && open) onOpenChange(false)
    }
    document.addEventListener('keydown', handleEsc)
    return () => document.removeEventListener('keydown', handleEsc)
  }, [open, onOpenChange])

  const toggleEndpoint = (id: number) => {
    setSelectedIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])
  }

  const handleSave = async () => {
    setSaving(true); setMsg('')
    try {
      const res = await syncModelEndpoints(modelId, selectedIds)
      if (res.success) {
        setMsg('保存成功')
        setTimeout(() => setMsg(''), 3000)
      } else {
        setMsg(res.message || '保存失败')
      }
    } catch { setMsg('网络错误') }
    finally { setSaving(false) }
  }

  if (!open) return null

  return (
    <div className='fixed inset-0 z-[100] flex items-center justify-center'>
      <div className='bg-background/80 fixed inset-0 backdrop-blur-sm' onClick={() => onOpenChange(false)} />
      <div className='bg-background border-border/60 relative z-10 flex max-h-[70vh] w-full max-w-lg flex-col rounded-xl border shadow-lg'>
        {/* Header */}
        <div className='flex items-center justify-between border-b border-border/40 px-6 py-4'>
          <div>
            <h2 className='text-lg font-semibold'>端点绑定</h2>
            <p className='text-muted-foreground mt-0.5 text-sm'>模型：<span className='text-foreground font-medium'>{modelName}</span></p>
          </div>
          <button onClick={() => onOpenChange(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'>
            <X className='size-4' />
          </button>
        </div>

        {/* 列表 */}
        <div className='flex-1 overflow-y-auto p-6'>
          {loading ? (
            <div className='flex items-center justify-center py-8'>
              <Loader2 className='text-muted-foreground size-5 animate-spin' />
            </div>
          ) : allEndpoints.length === 0 ? (
            <p className='text-muted-foreground py-8 text-center text-sm'>暂无可用端点，请先在端点管理中添加</p>
          ) : (
            <div className='space-y-2'>
              {allEndpoints.map((ep) => {
                const checked = selectedIds.includes(ep.id)
                return (
                  <label
                    key={ep.id}
                    className={cn(
                      'flex cursor-pointer items-center gap-3 rounded-lg border p-3 transition-colors',
                      checked ? 'border-primary/40 bg-primary/5' : 'border-border/40 hover:bg-muted/20'
                    )}
                  >
                    <input
                      type='checkbox'
                      checked={checked}
                      onChange={() => toggleEndpoint(ep.id)}
                      className='border-border/60 size-4 rounded'
                    />
                    <Globe className='text-muted-foreground size-4 shrink-0' />
                    <div className='min-w-0 flex-1'>
                      <div className='flex items-center gap-2'>
                        <code className='text-sm font-medium'>{ep.path}</code>
                        <span className='text-muted-foreground text-xs'>·</span>
                        <span className='text-sm'>{ep.name}</span>
                      </div>
                      {ep.description && <p className='text-muted-foreground mt-0.5 text-xs'>{ep.description}</p>}
                    </div>
                    {checked && <Check className='text-primary size-4 shrink-0' />}
                  </label>
                )
              })}
            </div>
          )}
        </div>

        {/* 底部 */}
        <div className='border-border/40 flex items-center justify-between border-t px-6 py-4'>
          <div className='flex items-center gap-3'>
            <span className='text-muted-foreground text-sm'>已选 <span className='text-foreground font-medium'>{selectedIds.length}</span> 个端点</span>
            {msg && <span className='text-sm text-emerald-600 dark:text-emerald-400'>{msg}</span>}
          </div>
          <div className='flex gap-3'>
            <button onClick={() => onOpenChange(false)}
              className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center rounded-lg border px-4 text-sm font-medium transition-colors'>取消</button>
            <button onClick={handleSave} disabled={saving}
              className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors disabled:opacity-50'>
              {saving && <Loader2 className='size-4 animate-spin' />}
              {saving ? '保存中...' : '保存'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
