import { useState, useEffect, useCallback } from 'react'
import {
  Plus,
  Pencil,
  Trash2,
  Power,
  PowerOff,
  Loader2,
  X,
  Box,
} from 'lucide-react'
import { cn } from '../../lib/utils'
import {
  getModels,
  createModel,
  updateModel,
  updateModelStatus,
  deleteModel,
  type AdminModel,
} from '../../api/admin-model'
import { ConfirmDialog } from '../../components/admin/ConfirmDialog'

const TYPE_OPTIONS = [
  { value: 'text', label: '文本' },
  { value: 'image', label: '图像' },
  { value: 'video', label: '视频' },
  { value: 'audio', label: '音频' },
]

const TYPE_BADGE: Record<string, string> = {
  text: 'bg-blue-500/10 text-blue-600 dark:text-blue-400',
  image: 'bg-purple-500/10 text-purple-600 dark:text-purple-400',
  video: 'bg-amber-500/10 text-amber-600 dark:text-amber-400',
  audio: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400',
}

const STATUS_CONFIG: Record<number, { label: string; className: string }> = {
  1: { label: '正常', className: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' },
  2: { label: '已禁用', className: 'bg-red-500/10 text-red-600 dark:text-red-400' },
}

export function AdminModels() {
  const [models, setModels] = useState<AdminModel[]>([])
  const [loading, setLoading] = useState(true)

  // 弹框
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editRow, setEditRow] = useState<AdminModel | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteRow, setDeleteRow] = useState<AdminModel | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  // 表单
  const [formName, setFormName] = useState('')
  const [formOwner, setFormOwner] = useState('')
  const [formType, setFormType] = useState('text')
  const [formPath, setFormPath] = useState('')
  const [formDesc, setFormDesc] = useState('')
  const [formError, setFormError] = useState('')
  const [formSaving, setFormSaving] = useState(false)

  const loadModels = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getModels()
      if (res.success) setModels(res.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { loadModels() }, [loadModels])

  const openCreate = () => {
    setEditRow(null)
    setFormName(''); setFormOwner(''); setFormType('text'); setFormPath('')
    setFormDesc(''); setFormError('')
    setDrawerOpen(true)
  }

  const openEdit = (m: AdminModel) => {
    setEditRow(m)
    setFormName(m.name); setFormOwner(m.owner); setFormType(m.model_type); setFormPath(m.request_path)
    setFormDesc(m.description); setFormError('')
    setDrawerOpen(true)
  }

  const handleSave = async () => {
    if (!formName.trim()) { setFormError('模型ID不能为空'); return }
    if (!formOwner.trim()) { setFormError('模型开发者不能为空'); return }

    setFormSaving(true); setFormError('')
    try {
      if (editRow) {
        const data: Record<string, string | number> = {}
        if (formName !== editRow.name) data.name = formName
        if (formOwner !== editRow.owner) data.owner = formOwner
        if (formType !== editRow.model_type) data.model_type = formType
        if (formPath !== editRow.request_path) data.request_path = formPath
        if (formDesc !== editRow.description) data.description = formDesc
        const res = await updateModel(editRow.id, data)
        if (!res.success) { setFormError(res.message || '更新失败'); return }
      } else {
        const res = await createModel({
          name: formName, owner: formOwner, model_type: formType,
          request_path: formPath, description: formDesc,
        })
        if (!res.success) { setFormError(res.message || '创建失败'); return }
      }
      setDrawerOpen(false); loadModels()
    } catch { setFormError('网络错误') }
    finally { setFormSaving(false) }
  }

  const handleToggle = async (m: AdminModel) => {
    setActionLoading(true)
    try {
      await updateModelStatus(m.id, m.status === 1 ? 2 : 1)
      loadModels()
    } catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteRow) return
    setActionLoading(true)
    try {
      await deleteModel(deleteRow.id)
      setDeleteOpen(false); setDeleteRow(null); loadModels()
    } catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  // ESC
  useEffect(() => {
    if (drawerOpen || deleteOpen) document.body.style.overflow = 'hidden'
    else document.body.style.overflow = ''
    return () => { document.body.style.overflow = '' }
  }, [drawerOpen, deleteOpen])

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return
      if (deleteOpen) setDeleteOpen(false)
      else if (drawerOpen) setDrawerOpen(false)
    }
    document.addEventListener('keydown', handleEsc)
    return () => document.removeEventListener('keydown', handleEsc)
  }, [drawerOpen, deleteOpen])

  return (
    <div className='space-y-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h2 className='text-2xl font-bold tracking-tight'>模型管理</h2>
          <p className='text-muted-foreground mt-1 text-sm'>管理系统可用的 AI 模型</p>
        </div>
        <button onClick={openCreate}
          className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors'>
          <Plus className='size-4' /> 添加模型
        </button>
      </div>

      <div className='border-border/60 overflow-hidden rounded-xl border'>
        <div className='overflow-x-auto'>
          <table className='w-full text-sm'>
            <thead>
              <tr className='bg-muted/30 border-border/40 border-b'>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>ID</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>模型开发者</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>模型ID</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>类型</th>
                <th className='text-muted-foreground min-w-[200px] px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>请求路径</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>描述</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>状态</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>操作</th>
              </tr>
            </thead>
            <tbody className='divide-border/40 divide-y'>
              {loading ? (
                <tr><td colSpan={8} className='px-4 py-12 text-center'><Loader2 className='text-muted-foreground mx-auto size-6 animate-spin' /></td></tr>
              ) : models.length === 0 ? (
                <tr><td colSpan={8} className='px-4 py-12 text-center'><p className='text-muted-foreground text-sm'>暂无模型</p></td></tr>
              ) : models.map((m) => {
                const statusConf = STATUS_CONFIG[m.status] || STATUS_CONFIG[1]
                return (
                  <tr key={m.id} className={cn('hover:bg-muted/20 transition-colors', m.status === 2 && 'opacity-50')}>
                    <td className='px-4 py-3'><span className='text-muted-foreground font-mono text-xs'>{m.id}</span></td>
                    <td className='px-4 py-3'><span className='text-sm'>{m.owner}</span></td>
                    <td className='px-4 py-3'><span className='font-medium'>{m.name}</span></td>
                    <td className='px-4 py-3'>
                      <span className={cn('inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium', TYPE_BADGE[m.model_type] || TYPE_BADGE.text)}>
                        <Box className='size-3' />
                        {TYPE_OPTIONS.find(t => t.value === m.model_type)?.label || m.model_type}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <code className='text-muted-foreground font-mono text-xs break-all'>{m.request_path || '-'}</code>
                    </td>
                    <td className='px-4 py-3'>
                      <span className='text-muted-foreground text-sm'>{m.description || '-'}</span>
                    </td>
                    <td className='px-4 py-3'>
                      <span className={cn('inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium', statusConf.className)}>
                        <span className={cn('size-1.5 rounded-full', m.status === 1 ? 'bg-emerald-500' : 'bg-red-500')} />
                        {statusConf.label}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <div className='flex items-center gap-1'>
                        <button onClick={() => openEdit(m)} className='hover:bg-muted inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors'>
                          <Pencil className='size-3.5' /> 编辑
                        </button>
                        <button onClick={() => handleToggle(m)}
                          className={cn('inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors',
                            m.status === 1 ? 'hover:bg-amber-500/10 text-amber-600 dark:text-amber-400' : 'hover:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400')}>
                          {m.status === 1 ? <><PowerOff className='size-3.5' /> 禁用</> : <><Power className='size-3.5' /> 启用</>}
                        </button>
                        <button onClick={() => { setDeleteRow(m); setDeleteOpen(true) }}
                          className='text-destructive hover:bg-destructive/10 inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors'>
                          <Trash2 className='size-3.5' /> 删除
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>

      {/* 新增/编辑弹框 */}
      {drawerOpen && (
        <div className='fixed inset-0 z-[100] flex items-center justify-center'>
          <div className='bg-background/80 fixed inset-0 backdrop-blur-sm' onClick={() => setDrawerOpen(false)} />
          <div className='bg-background border-border/60 relative z-10 w-full max-w-lg rounded-xl border shadow-lg'>
            <div className='flex items-center justify-between border-b border-border/40 px-6 py-4'>
              <h2 className='text-lg font-semibold'>{editRow ? '编辑模型' : '添加模型'}</h2>
              <button onClick={() => setDrawerOpen(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'><X className='size-4' /></button>
            </div>
            <div className='p-6 space-y-4'>
              {formError && <div className='bg-destructive/10 text-destructive rounded-lg px-4 py-3 text-sm'>{formError}</div>}

              <div className='space-y-2'>
                <label className='text-sm font-medium'>模型开发者</label>
                <input type='text' value={formOwner} onChange={(e) => setFormOwner(e.target.value)} placeholder='例如: OpenAI'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>模型ID</label>
                <input type='text' value={formName} onChange={(e) => setFormName(e.target.value)} placeholder='例如: gpt-4o'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>模型类型</label>
                <select value={formType} onChange={(e) => setFormType(e.target.value)}
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'>
                  {TYPE_OPTIONS.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
                </select>
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>请求路径</label>
                <input type='text' value={formPath} onChange={(e) => setFormPath(e.target.value)} placeholder='例如: /v1/chat/completions'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>描述</label>
                <input type='text' value={formDesc} onChange={(e) => setFormDesc(e.target.value)} placeholder='可选'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

            </div>
            <div className='flex justify-end gap-3 border-t border-border/40 px-6 py-4'>
              <button onClick={() => setDrawerOpen(false)} disabled={formSaving}
                className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center rounded-lg border px-4 text-sm font-medium transition-colors'>取消</button>
              <button onClick={handleSave} disabled={formSaving}
                className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors disabled:opacity-50'>
                {formSaving && <Loader2 className='size-4 animate-spin' />}
                {formSaving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}

      <ConfirmDialog open={deleteOpen} onOpenChange={setDeleteOpen} title='确认删除'
        description={<>确定要删除模型 <span className='text-foreground font-semibold'>{deleteRow?.name}</span> 吗？</>}
        confirmText='删除' destructive loading={actionLoading} onConfirm={handleDeleteConfirm} />
    </div>
  )
}
