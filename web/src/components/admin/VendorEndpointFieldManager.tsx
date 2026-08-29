import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, X, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '../../lib/utils'
import {
  getVendorEndpointFields,
  syncVEFieldsFromEndpoint,
  createVEField,
  updateVEField,
  deleteVEField,
  type VendorEndpointField,
} from '../../api/vendor-endpoint'
import { getEndpointFields, type EndpointField } from '../../api/endpoint'

const TYPE_OPTIONS = [
  { value: 'string', label: '字符串' },
  { value: 'number', label: '数值' },
  { value: 'boolean', label: '布尔' },
  { value: 'array', label: '数组' },
  { value: 'object', label: '对象' },
]

const TYPE_BADGE: Record<string, string> = {
  string: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  number: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
  boolean: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
  array: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
  object: 'bg-pink-500/10 text-pink-600 dark:text-pink-400',
}

interface VendorEndpointFieldManagerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  vendorEndpointId: number
  endpointId: number
  vendorName: string
  endpointName: string
}

export function VendorEndpointFieldManager({ open, onOpenChange, vendorEndpointId, endpointId, vendorName, endpointName }: VendorEndpointFieldManagerProps) {
  const [fields, setFields] = useState<VendorEndpointField[]>([])
  const [loading, setLoading] = useState(false)
  const [syncing, setSyncing] = useState(false)
  const [endpointFields, setEndpointFields] = useState<EndpointField[]>([])

  const [editingId, setEditingId] = useState<number | null>(null)
  const [adding, setAdding] = useState(false)

  const [fKey, setFKey] = useState('')
  const [fName, setFName] = useState('')
  const [fType, setFType] = useState('string')
  const [fRequired, setFRequired] = useState(false)
  const [fDesc, setFDesc] = useState('')
  const [selectedEndpointFieldId, setSelectedEndpointFieldId] = useState<number | null>(null)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [syncMsg, setSyncMsg] = useState('')

  const loadFields = async () => {
    setLoading(true)
    try {
      const res = await getVendorEndpointFields(vendorEndpointId)
      if (res.success) setFields(res.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  useEffect(() => {
    if (open) {
      loadFields()
      setSyncMsg('')
      getEndpointFields(endpointId).then(res => {
        if (res.success) setEndpointFields(res.data || [])
      }).catch(() => {})
    }
  }, [open, vendorEndpointId, endpointId])

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

  const resetForm = () => {
    setFKey(''); setFName(''); setFType('string'); setFRequired(false); setFDesc(''); setError(''); setSelectedEndpointFieldId(null)
  }

  const startAdd = () => { resetForm(); setAdding(true); setEditingId(null) }

  const handleEndpointFieldSelect = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const val = e.target.value
    if (!val) { setSelectedEndpointFieldId(-1); return }
    const ef = endpointFields.find(f => f.id === Number(val))
    if (!ef) return
    setSelectedEndpointFieldId(ef.id)
    if (adding) {
      setFKey(ef.field_key); setFName(ef.field_name); setFType(ef.field_type)
      setFRequired(ef.required); setFDesc(ef.description)
    }
  }

  const startEdit = (f: VendorEndpointField) => {
    setFKey(f.field_key); setFName(f.field_name); setFType(f.field_type)
    setFRequired(f.required); setFDesc(f.description); setError('')
    setSelectedEndpointFieldId(f.endpoint_field_id ?? null)
    setEditingId(f.id); setAdding(false)
  }
  const cancelEdit = () => { setEditingId(null); setAdding(false); resetForm() }

  const handleSave = async () => {
    if (!fKey.trim()) { setError('字段ID不能为空'); return }
    if (!fName.trim()) { setError('字段名称不能为空'); return }
    setSaving(true); setError('')
    try {
      if (adding) {
        const res = await createVEField(vendorEndpointId, {
          field_key: fKey, field_name: fName, field_type: fType,
          required: fRequired, description: fDesc,
          endpoint_field_id: selectedEndpointFieldId && selectedEndpointFieldId > 0 ? selectedEndpointFieldId : null,
        })
        if (!res.success) { setError(res.message || '创建失败'); return }
      } else if (editingId) {
        const res = await updateVEField(vendorEndpointId, editingId, {
          field_key: fKey, field_name: fName, field_type: fType,
          required: fRequired, description: fDesc,
          endpoint_field_id: selectedEndpointFieldId ?? -1,
        })
        if (!res.success) { setError(res.message || '更新失败'); return }
      }
      cancelEdit(); loadFields()
    } catch { setError('网络错误') }
    finally { setSaving(false) }
  }

  const handleDelete = async (id: number) => {
    try { await deleteVEField(vendorEndpointId, id); loadFields() } catch { /* ignore */ }
  }

  const handleSync = async () => {
    setSyncing(true); setSyncMsg('')
    try {
      const res = await syncVEFieldsFromEndpoint(vendorEndpointId)
      if (res.success) { setSyncMsg('同步成功'); loadFields(); setTimeout(() => setSyncMsg(''), 3000) }
      else { setSyncMsg(res.message || '同步失败') }
    } catch { setSyncMsg('网络错误') }
    finally { setSyncing(false) }
  }

  if (!open) return null

  const isEditing = (id: number) => editingId === id

  const renderEditRow = () => (
    <tr>
      <td colSpan={7} className='px-4 py-3'>
        <div className='space-y-3'>
          {error && <div className='bg-destructive/10 text-destructive rounded-lg px-3 py-2 text-xs'>{error}</div>}
          {endpointFields.length > 0 && (
            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>绑定端点字段（可选）</label>
              <select value={selectedEndpointFieldId ?? ''} onChange={handleEndpointFieldSelect}
                className='border-border/60 bg-background focus-visible:ring-ring flex h-8 w-full rounded-md border px-2.5 text-sm focus-visible:ring-1 focus-visible:outline-none'>
                <option value=''>-- 不绑定 --</option>
                {endpointFields.map(ef => (
                  <option key={ef.id} value={ef.id}>{ef.field_key} · {ef.field_name}</option>
                ))}
              </select>
            </div>
          )}
          <div className='grid grid-cols-4 gap-3'>
            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>字段ID</label>
              <input value={fKey} onChange={(e) => setFKey(e.target.value)} placeholder='例如: model'
                className='border-border/60 bg-background focus-visible:ring-ring flex h-8 w-full rounded-md border px-2.5 text-sm focus-visible:ring-1 focus-visible:outline-none' />
            </div>
            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>字段名称</label>
              <input value={fName} onChange={(e) => setFName(e.target.value)} placeholder='例如: 模型名称'
                className='border-border/60 bg-background focus-visible:ring-ring flex h-8 w-full rounded-md border px-2.5 text-sm focus-visible:ring-1 focus-visible:outline-none' />
            </div>
            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>类型</label>
              <select value={fType} onChange={(e) => setFType(e.target.value)}
                className='border-border/60 bg-background focus-visible:ring-ring flex h-8 w-full rounded-md border px-2.5 text-sm focus-visible:ring-1 focus-visible:outline-none'>
                {TYPE_OPTIONS.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </div>
            <div className='space-y-1'>
              <label className='text-muted-foreground text-xs'>说明</label>
              <textarea value={fDesc} onChange={(e) => setFDesc(e.target.value)} placeholder='可选' rows={2}
                className='border-border/60 bg-background focus-visible:ring-ring w-full rounded-md border px-2.5 py-1.5 text-sm focus-visible:ring-1 focus-visible:outline-none resize-none' />
            </div>
          </div>
          <div className='flex items-center gap-3'>
            <label className='flex items-center gap-2 text-sm'>
              <input type='checkbox' checked={fRequired} onChange={(e) => setFRequired(e.target.checked)} className='border-border/60 size-4 rounded' />
              必填
            </label>
            <div className='flex-1' />
            <button onClick={cancelEdit} disabled={saving}
              className='border-border/60 hover:bg-muted inline-flex h-7 items-center justify-center rounded-md border px-3 text-xs font-medium transition-colors'>取消</button>
            <button onClick={handleSave} disabled={saving}
              className='bg-primary hover:bg-primary/90 inline-flex h-7 items-center justify-center gap-1 rounded-md px-3 text-xs font-medium text-white transition-colors disabled:opacity-50'>
              {saving && <Loader2 className='size-3 animate-spin' />}
              {adding ? '添加' : '保存'}
            </button>
          </div>
        </div>
      </td>
    </tr>
  )

  return (
    <div className='fixed inset-0 z-[100] flex items-center justify-center'>
      <div className='bg-background/80 fixed inset-0 backdrop-blur-sm' onClick={() => onOpenChange(false)} />
      <div className='bg-background border-border/60 relative z-10 flex max-h-[85vh] w-full max-w-5xl flex-col rounded-xl border shadow-lg'>
        <div className='flex items-center justify-between border-b border-border/40 px-6 py-4'>
          <div>
            <h2 className='text-lg font-semibold'>字段管理</h2>
            <p className='text-muted-foreground mt-0.5 text-sm'>
              供应商：<span className='text-foreground font-medium'>{vendorName}</span>
              {' · '}
              端点：<span className='text-foreground font-medium'>{endpointName}</span>
            </p>
          </div>
          <button onClick={() => onOpenChange(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'><X className='size-4' /></button>
        </div>

        <div className='flex-1 overflow-y-auto'>
          <table className='w-full text-sm'>
            <thead className='sticky top-0 z-10'>
              <tr className='bg-muted/30 border-border/40 border-b'>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>字段ID</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>字段名称</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>类型</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>必填</th>
                <th className='text-muted-foreground min-w-[200px] px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>说明</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>绑定字段</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>操作</th>
              </tr>
            </thead>
            <tbody className='divide-border/40 divide-y'>
              {loading ? (
                <tr><td colSpan={7} className='px-4 py-12 text-center'><Loader2 className='text-muted-foreground mx-auto size-5 animate-spin' /></td></tr>
              ) : fields.length === 0 && !adding ? (
                <tr><td colSpan={7} className='text-muted-foreground px-4 py-12 text-center text-sm'>暂无字段定义，可从端点同步或手动添加</td></tr>
              ) : (
                fields.map((f) => (
                  isEditing(f.id) ? renderEditRow() : (
                    <tr key={f.id} className='hover:bg-muted/20 transition-colors'>
                      <td className='px-4 py-3'><code className='font-mono text-xs'>{f.field_key}</code></td>
                      <td className='px-4 py-3'><span className='font-medium'>{f.field_name}</span></td>
                      <td className='px-4 py-3'>
                        <span className={cn('rounded px-1.5 py-0.5 text-[11px] font-medium', TYPE_BADGE[f.field_type] || TYPE_BADGE.string)}>
                          {TYPE_OPTIONS.find(t => t.value === f.field_type)?.label || f.field_type}
                        </span>
                      </td>
                      <td className='px-4 py-3'>
                        {f.required ? <span className='text-destructive text-xs font-medium'>必填</span> : <span className='text-muted-foreground text-xs'>-</span>}
                      </td>
                      <td className='px-4 py-3'><span className='text-muted-foreground text-xs'>{f.description || '-'}</span></td>
                      <td className='px-4 py-3'>
                        {f.endpoint_field_id ? (() => {
                          const bound = endpointFields.find(ef => ef.id === f.endpoint_field_id)
                          return bound ? (
                            <span className='rounded bg-blue-500/10 px-1.5 py-0.5 text-[11px] font-medium text-blue-600 dark:text-blue-400'>{bound.field_key}</span>
                          ) : <span className='text-muted-foreground text-xs'>已删除</span>
                        })() : <span className='text-muted-foreground text-xs'>未绑定</span>}
                      </td>
                      <td className='px-4 py-3'>
                        <div className='flex items-center gap-1'>
                          <button onClick={() => startEdit(f)} className='hover:bg-muted rounded-md p-1.5 transition-colors'><Pencil className='size-3.5' /></button>
                          <button onClick={() => handleDelete(f.id)} className='text-destructive hover:bg-destructive/10 rounded-md p-1.5 transition-colors'><Trash2 className='size-3.5' /></button>
                        </div>
                      </td>
                    </tr>
                  )
                ))
              )}
              {adding && renderEditRow()}
            </tbody>
          </table>
        </div>

        <div className='border-border/40 flex items-center justify-between border-t px-6 py-4'>
          <div className='flex items-center gap-3'>
            <button onClick={startAdd} disabled={adding}
              className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center gap-2 rounded-lg border px-4 text-sm font-medium transition-colors disabled:opacity-50'>
              <Plus className='size-4' /> 添加字段
            </button>
            <button onClick={handleSync} disabled={syncing}
              className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center gap-2 rounded-lg border px-4 text-sm font-medium transition-colors disabled:opacity-50'>
              {syncing ? <Loader2 className='size-4 animate-spin' /> : <RefreshCw className='size-4' />}
              从端点同步
            </button>
            {syncMsg && <span className='text-sm text-emerald-600 dark:text-emerald-400'>{syncMsg}</span>}
          </div>
          <button onClick={() => onOpenChange(false)}
            className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center rounded-lg px-4 text-sm font-medium text-white transition-colors'>完成</button>
        </div>
      </div>
    </div>
  )
}
