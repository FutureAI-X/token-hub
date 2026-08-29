import { useState, useEffect } from 'react'
import { Plus, Pencil, Trash2, X, Loader2 } from 'lucide-react'
import { cn } from '../../lib/utils'
import {
  getModelFields,
  createModelField,
  updateModelField,
  deleteModelField,
  type ModelField,
} from '../../api/model-field'

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

interface FieldManagerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  modelId: number
  modelName: string
}

export function FieldManager({ open, onOpenChange, modelId, modelName }: FieldManagerProps) {
  const [fields, setFields] = useState<ModelField[]>([])
  const [loading, setLoading] = useState(false)
  const [section, setSection] = useState<'request' | 'response'>('request')

  const [editingId, setEditingId] = useState<number | null>(null)
  const [adding, setAdding] = useState(false)

  const [fKey, setFKey] = useState('')
  const [fName, setFName] = useState('')
  const [fType, setFType] = useState('string')
  const [fRequired, setFRequired] = useState(false)
  const [fDesc, setFDesc] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')

  const loadFields = async () => {
    setLoading(true)
    try {
      const res = await getModelFields(modelId, section)
      if (res.success) setFields(res.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }

  useEffect(() => { if (open) loadFields() }, [open, modelId, section])

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
    setFKey(''); setFName(''); setFType('string'); setFRequired(false); setFDesc(''); setError('')
  }

  const startAdd = () => { resetForm(); setAdding(true); setEditingId(null) }

  const startEdit = (f: ModelField) => {
    setFKey(f.field_key); setFName(f.field_name); setFType(f.field_type)
    setFRequired(f.required); setFDesc(f.description); setError('')
    setEditingId(f.id); setAdding(false)
  }

  const cancelEdit = () => { setEditingId(null); setAdding(false); resetForm() }

  const handleSave = async () => {
    if (!fKey.trim()) { setError('字段ID不能为空'); return }
    if (!fName.trim()) { setError('字段名称不能为空'); return }

    setSaving(true); setError('')
    try {
      if (adding) {
        const res = await createModelField(modelId, {
          field_key: fKey, field_name: fName, field_type: fType,
          required: fRequired, description: fDesc, section,
        })
        if (!res.success) { setError(res.message || '创建失败'); return }
      } else if (editingId) {
        const res = await updateModelField(modelId, editingId, {
          field_key: fKey, field_name: fName, field_type: fType,
          required: fRequired, description: fDesc,
        })
        if (!res.success) { setError(res.message || '更新失败'); return }
      }
      cancelEdit(); loadFields()
    } catch { setError('网络错误') }
    finally { setSaving(false) }
  }

  const handleDelete = async (id: number) => {
    try { await deleteModelField(modelId, id); loadFields() } catch { /* ignore */ }
  }

  if (!open) return null

  const isEditing = (id: number) => editingId === id

  // 渲染编辑行
  const renderEditRow = (isAdding: boolean) => (
    <tr>
      <td colSpan={6} className='px-4 py-3'>
        <div className='space-y-3'>
          {error && <div className='bg-destructive/10 text-destructive rounded-lg px-3 py-2 text-xs'>{error}</div>}
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
              <input type='checkbox' checked={fRequired} onChange={(e) => setFRequired(e.target.checked)}
                className='border-border/60 size-4 rounded' />
              必填
            </label>
            <div className='flex-1' />
            <button onClick={cancelEdit} disabled={saving}
              className='border-border/60 hover:bg-muted inline-flex h-7 items-center justify-center rounded-md border px-3 text-xs font-medium transition-colors'>取消</button>
            <button onClick={handleSave} disabled={saving}
              className='bg-primary hover:bg-primary/90 inline-flex h-7 items-center justify-center gap-1 rounded-md px-3 text-xs font-medium text-white transition-colors disabled:opacity-50'>
              {saving && <Loader2 className='size-3 animate-spin' />}
              {isAdding ? '添加' : '保存'}
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
        {/* Header */}
        <div className='flex items-center justify-between border-b border-border/40 px-6 py-4'>
          <div>
            <h2 className='text-lg font-semibold'>字段管理</h2>
            <p className='text-muted-foreground mt-0.5 text-sm'>模型：<span className='text-foreground font-medium'>{modelName}</span></p>
          </div>
          <button onClick={() => onOpenChange(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'>
            <X className='size-4' />
          </button>
        </div>

        {/* Tab 切换 */}
        <div className='border-border/40 flex gap-1 border-b px-6'>
          {(['request', 'response'] as const).map((s) => (
            <button
              key={s}
              onClick={() => { setSection(s); cancelEdit() }}
              className={cn(
                'relative px-4 py-2.5 text-sm font-medium transition-colors',
                section === s
                  ? 'text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
            >
              {s === 'request' ? '请求体字段' : '响应体字段'}
              {section === s && (
                <span className='bg-primary absolute inset-x-0 -bottom-px h-0.5' />
              )}
            </button>
          ))}
        </div>

        {/* 表格 */}
        <div className='flex-1 overflow-y-auto'>
          <table className='w-full text-sm'>
            <thead className='sticky top-0 z-10'>
              <tr className='bg-muted/30 border-border/40 border-b'>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>字段ID</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>字段名称</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>类型</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>必填</th>
                <th className='text-muted-foreground min-w-[200px] px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>说明</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>操作</th>
              </tr>
            </thead>
            <tbody className='divide-border/40 divide-y'>
              {loading ? (
                <tr><td colSpan={6} className='px-4 py-12 text-center'><Loader2 className='text-muted-foreground mx-auto size-5 animate-spin' /></td></tr>
              ) : fields.length === 0 && !adding ? (
                <tr><td colSpan={6} className='text-muted-foreground px-4 py-12 text-center text-sm'>暂无字段定义</td></tr>
              ) : (
                fields.map((f) => (
                  isEditing(f.id) ? renderEditRow(false) : (
                    <tr key={f.id} className='hover:bg-muted/20 transition-colors'>
                      <td className='px-4 py-3'><code className='font-mono text-xs'>{f.field_key}</code></td>
                      <td className='px-4 py-3'><span className='font-medium'>{f.field_name}</span></td>
                      <td className='px-4 py-3'>
                        <span className={cn('rounded px-1.5 py-0.5 text-[11px] font-medium', TYPE_BADGE[f.field_type] || TYPE_BADGE.string)}>
                          {TYPE_OPTIONS.find(t => t.value === f.field_type)?.label || f.field_type}
                        </span>
                      </td>
                      <td className='px-4 py-3'>
                        {f.required ? (
                          <span className='text-destructive text-xs font-medium'>必填</span>
                        ) : (
                          <span className='text-muted-foreground text-xs'>-</span>
                        )}
                      </td>
                      <td className='px-4 py-3'>
                        <span className='text-muted-foreground text-xs'>{f.description || '-'}</span>
                      </td>
                      <td className='px-4 py-3'>
                        <div className='flex items-center gap-1'>
                          <button onClick={() => startEdit(f)} className='hover:bg-muted rounded-md p-1.5 transition-colors'>
                            <Pencil className='size-3.5' />
                          </button>
                          <button onClick={() => handleDelete(f.id)} className='text-destructive hover:bg-destructive/10 rounded-md p-1.5 transition-colors'>
                            <Trash2 className='size-3.5' />
                          </button>
                        </div>
                      </td>
                    </tr>
                  )
                ))
              )}
              {adding && renderEditRow(true)}
            </tbody>
          </table>
        </div>

        {/* 底部 */}
        <div className='border-border/40 flex justify-between border-t px-6 py-4'>
          <button onClick={startAdd} disabled={adding}
            className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center gap-2 rounded-lg border px-4 text-sm font-medium transition-colors disabled:opacity-50'>
            <Plus className='size-4' /> 添加字段
          </button>
          <button onClick={() => onOpenChange(false)}
            className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center rounded-lg px-4 text-sm font-medium text-white transition-colors'>
            完成
          </button>
        </div>
      </div>
    </div>
  )
}
