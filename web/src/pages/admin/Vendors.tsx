import { useState, useEffect, useCallback } from 'react'
import {
  Plus,
  Pencil,
  Trash2,
  Power,
  PowerOff,
  Loader2,
  Radio,
  X,
  Eye,
  EyeOff,
} from 'lucide-react'
import { cn } from '../../lib/utils'
import { encryptWithKey, decryptWithKey } from '../../lib/crypto'
import {
  getVendors,
  createVendor,
  updateVendor,
  updateVendorStatus,
  deleteVendor,
  type Vendor,
} from '../../api/vendor'
import { ConfirmDialog } from '../../components/admin/ConfirmDialog'

const PROTOCOL_OPTIONS = [
  { value: 'openai-chat', label: 'OpenAI Chat Completions' },
  { value: 'openai-responses', label: 'OpenAI Responses' },
  { value: 'anthropic-messages', label: 'Anthropic Messages' },
]

function protocolLabel(type: string): string {
  return PROTOCOL_OPTIONS.find((p) => p.value === type)?.label || type
}

const STATUS_CONFIG: Record<number, { label: string; className: string }> = {
  1: { label: '正常', className: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' },
  2: { label: '已禁用', className: 'bg-red-500/10 text-red-600 dark:text-red-400' },
}

export function AdminVendors() {
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [loading, setLoading] = useState(true)

  // 弹框状态
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editRow, setEditRow] = useState<Vendor | null>(null)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [deleteRow, setDeleteRow] = useState<Vendor | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  // 表单状态
  const [formName, setFormName] = useState('')
  const [formDesc, setFormDesc] = useState('')
  const [formBaseURL, setFormBaseURL] = useState('')
  const [formAPIKey, setFormAPIKey] = useState('')
  const [formProtocol, setFormProtocol] = useState('openai-chat')
  const [showKey, setShowKey] = useState(false)
  const [formError, setFormError] = useState('')
  const [formSaving, setFormSaving] = useState(false)

  const loadVendors = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getVendors()
      if (res.success) setVendors(res.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { loadVendors() }, [loadVendors])

  // ── 新增 ──
  const handleCreate = () => {
    setEditRow(null)
    setFormName('')
    setFormDesc('')
    setFormBaseURL('')
    setFormAPIKey('')
    setFormProtocol('openai-chat')
    setShowKey(false)
    setFormError('')
    setDrawerOpen(true)
  }

  // ── 编辑 ──
  const handleEdit = (v: Vendor) => {
    setEditRow(v)
    setFormName(v.name)
    setFormDesc(v.description)
    setFormBaseURL(v.base_url)
    setFormAPIKey(v.api_key)
    setFormProtocol(v.protocol_type)
    setShowKey(false)
    setFormError('')
    setDrawerOpen(true)
  }

  // ── 保存 ──
  const handleSave = async () => {
    if (!formName.trim()) { setFormError('供应商名称不能为空'); return }
    if (!formBaseURL.trim()) { setFormError('API Base URL 不能为空'); return }
    if (!editRow && !formAPIKey.trim()) { setFormError('API Key 不能为空'); return }

    setFormSaving(true)
    setFormError('')

    try {
      const dataKey = localStorage.getItem('data_key') || ''

      if (editRow) {
        const data: Record<string, string> = {}
        if (formName !== editRow.name) data.name = formName
        if (formDesc !== editRow.description) data.description = formDesc
        if (formBaseURL !== editRow.base_url) data.base_url = formBaseURL
        if (formAPIKey && formAPIKey !== editRow.api_key) {
          data.api_key = await encryptWithKey(formAPIKey, dataKey)
          data.data_key = dataKey
        }
        if (formProtocol !== editRow.protocol_type) data.protocol_type = formProtocol
        const res = await updateVendor(editRow.id, data)
        if (!res.success) { setFormError(res.message || '更新失败'); return }
      } else {
        const res = await createVendor({
          name: formName,
          description: formDesc,
          base_url: formBaseURL,
          api_key: await encryptWithKey(formAPIKey, dataKey),
          protocol_type: formProtocol,
          data_key: dataKey,
        })
        if (!res.success) { setFormError(res.message || '创建失败'); return }
      }
      setDrawerOpen(false)
      loadVendors()
    } catch {
      setFormError('网络错误')
    } finally {
      setFormSaving(false)
    }
  }

  // ── 启用/禁用 ──
  const handleToggle = async (v: Vendor) => {
    setActionLoading(true)
    try {
      await updateVendorStatus(v.id, v.status === 1 ? 2 : 1)
      loadVendors()
    } catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  // ── 删除 ──
  const handleDeleteConfirm = async () => {
    if (!deleteRow) return
    setActionLoading(true)
    try {
      await deleteVendor(deleteRow.id)
      setDeleteDialogOpen(false)
      setDeleteRow(null)
      loadVendors()
    } catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  // ── ESC 关闭 ──
  useEffect(() => {
    if (drawerOpen || deleteDialogOpen) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [drawerOpen, deleteDialogOpen])

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        if (deleteDialogOpen) setDeleteDialogOpen(false)
        else if (drawerOpen) setDrawerOpen(false)
      }
    }
    document.addEventListener('keydown', handleEsc)
    return () => document.removeEventListener('keydown', handleEsc)
  }, [drawerOpen, deleteDialogOpen])

  return (
    <div className='space-y-6'>
      {/* 标题 */}
      <div className='flex items-center justify-between'>
        <div>
          <h2 className='text-2xl font-bold tracking-tight'>供应商管理</h2>
          <p className='text-muted-foreground mt-1 text-sm'>管理 AI 模型供应商及其连接配置</p>
        </div>
        <button
          onClick={handleCreate}
          className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors'
        >
          <Plus className='size-4' />
          添加供应商
        </button>
      </div>

      {/* 表格 */}
      <div className='border-border/60 overflow-hidden rounded-xl border'>
        <div className='overflow-x-auto'>
          <table className='w-full text-sm'>
            <thead>
              <tr className='bg-muted/30 border-border/40 border-b'>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>ID</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>供应商名称</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>API Base URL</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>协议类型</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>状态</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>创建时间</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>操作</th>
              </tr>
            </thead>
            <tbody className='divide-border/40 divide-y'>
              {loading ? (
                <tr>
                  <td colSpan={7} className='px-4 py-12 text-center'>
                    <Loader2 className='text-muted-foreground mx-auto size-6 animate-spin' />
                  </td>
                </tr>
              ) : vendors.length === 0 ? (
                <tr>
                  <td colSpan={7} className='px-4 py-12 text-center'>
                    <p className='text-muted-foreground text-sm'>暂无供应商</p>
                  </td>
                </tr>
              ) : vendors.map((v) => {
                const statusConf = STATUS_CONFIG[v.status] || STATUS_CONFIG[1]
                return (
                  <tr key={v.id} className={cn('hover:bg-muted/20 transition-colors', v.status === 2 && 'opacity-50')}>
                    <td className='px-4 py-3'>
                      <span className='text-muted-foreground font-mono text-xs'>{v.id}</span>
                    </td>
                    <td className='px-4 py-3'>
                      <div className='flex flex-col gap-0.5'>
                        <span className='font-medium'>{v.name}</span>
                        {v.description && (
                          <span className='text-muted-foreground text-xs'>{v.description}</span>
                        )}
                      </div>
                    </td>
                    <td className='px-4 py-3'>
                      <code className='text-muted-foreground font-mono text-xs break-all'>{v.base_url}</code>
                    </td>
                    <td className='px-4 py-3'>
                      <span className='bg-muted inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium'>
                        <Radio className='size-3' />
                        {protocolLabel(v.protocol_type)}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <span className={cn('inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium', statusConf.className)}>
                        <span className={cn('size-1.5 rounded-full', v.status === 1 ? 'bg-emerald-500' : 'bg-red-500')} />
                        {statusConf.label}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <span className='text-muted-foreground text-sm'>{v.created_at}</span>
                    </td>
                    <td className='px-4 py-3'>
                      <div className='flex items-center gap-1'>
                        <button onClick={() => handleEdit(v)} className='hover:bg-muted inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors'>
                          <Pencil className='size-3.5' /> 编辑
                        </button>
                        <button
                          onClick={() => handleToggle(v)}
                          className={cn(
                            'inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors',
                            v.status === 1
                              ? 'hover:bg-amber-500/10 text-amber-600 dark:text-amber-400'
                              : 'hover:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                          )}
                        >
                          {v.status === 1 ? <><PowerOff className='size-3.5' /> 禁用</> : <><Power className='size-3.5' /> 启用</>}
                        </button>
                        <button
                          onClick={() => { setDeleteRow(v); setDeleteDialogOpen(true) }}
                          className='text-destructive hover:bg-destructive/10 inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors'
                        >
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
              <h2 className='text-lg font-semibold'>{editRow ? '编辑供应商' : '添加供应商'}</h2>
              <button onClick={() => setDrawerOpen(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'>
                <X className='size-4' />
              </button>
            </div>

            <div className='p-6 space-y-4' autoComplete='off'>
              {formError && (
                <div className='bg-destructive/10 text-destructive rounded-lg px-4 py-3 text-sm'>{formError}</div>
              )}

              <div className='space-y-2'>
                <label className='text-sm font-medium'>供应商名称</label>
                <input type='text' value={formName} onChange={(e) => setFormName(e.target.value)} placeholder='例如: OpenAI'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>描述</label>
                <input type='text' value={formDesc} onChange={(e) => setFormDesc(e.target.value)} placeholder='可选'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>API Base URL</label>
                <input type='text' value={formBaseURL} onChange={(e) => setFormBaseURL(e.target.value)} placeholder='https://api.openai.com/v1' autoComplete='off'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>API Key</label>
                <div className='relative'>
                  <input
                    type={showKey ? 'text' : 'password'}
                    value={formAPIKey}
                    onChange={(e) => setFormAPIKey(e.target.value)}
                    placeholder={editRow ? '已设置，留空则不修改' : '请输入 API Key'}
                    autoComplete='new-password'
                    className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border pr-10 pl-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                  />
                  <button
                    type='button'
                    onClick={() => setShowKey(!showKey)}
                    className='text-muted-foreground hover:text-foreground absolute right-1 top-1/2 -translate-y-1/2 rounded-md p-1.5 transition-colors'
                  >
                    {showKey ? <EyeOff className='size-4' /> : <Eye className='size-4' />}
                  </button>
                </div>
              </div>

              <div className='space-y-2'>
                <label className='text-sm font-medium'>协议类型</label>
                <select value={formProtocol} onChange={(e) => setFormProtocol(e.target.value)}
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'>
                  {PROTOCOL_OPTIONS.map((p) => (
                    <option key={p.value} value={p.value}>{p.label}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className='flex justify-end gap-3 border-t border-border/40 px-6 py-4'>
              <button onClick={() => setDrawerOpen(false)} disabled={formSaving}
                className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center rounded-lg border px-4 text-sm font-medium transition-colors'>
                取消
              </button>
              <button onClick={handleSave} disabled={formSaving}
                className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors disabled:opacity-50'>
                {formSaving && <Loader2 className='size-4 animate-spin' />}
                {formSaving ? '保存中...' : '保存'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* 删除确认 */}
      <ConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title='确认删除'
        description={<>确定要删除供应商 <span className='text-foreground font-semibold'>{deleteRow?.name}</span> 吗？</>}
        confirmText='删除'
        destructive
        loading={actionLoading}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  )
}
