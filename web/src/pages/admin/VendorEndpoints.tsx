import { useState, useEffect, useCallback } from 'react'
import {
  Plus,
  Pencil,
  Trash2,
  Power,
  PowerOff,
  Loader2,
  X,
  Link2,
  Settings,
  Globe,
} from 'lucide-react'
import { cn } from '../../lib/utils'
import {
  getVendorEndpoints,
  createVendorEndpoint,
  updateVendorEndpoint,
  updateVendorEndpointStatus,
  deleteVendorEndpoint,
  type VendorEndpoint,
} from '../../api/vendor-endpoint'
import { getVendors, type Vendor } from '../../api/vendor'
import { getEndpoints, type Endpoint } from '../../api/endpoint'
import { ConfirmDialog } from '../../components/admin/ConfirmDialog'
import { VendorEndpointFieldManager } from '../../components/admin/VendorEndpointFieldManager'

const STATUS_CONFIG: Record<number, { label: string; className: string }> = {
  1: { label: '正常', className: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' },
  2: { label: '已禁用', className: 'bg-red-500/10 text-red-600 dark:text-red-400' },
}

export function AdminVendorEndpoints() {
  const [items, setItems] = useState<VendorEndpoint[]>([])
  const [vendors, setVendors] = useState<Vendor[]>([])
  const [endpoints, setEndpoints] = useState<Endpoint[]>([])
  const [loading, setLoading] = useState(true)

  const [drawerOpen, setDrawerOpen] = useState(false)
  const [editRow, setEditRow] = useState<VendorEndpoint | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [deleteRow, setDeleteRow] = useState<VendorEndpoint | null>(null)
  const [fieldOpen, setFieldOpen] = useState(false)
  const [fieldRow, setFieldRow] = useState<VendorEndpoint | null>(null)
  const [actionLoading, setActionLoading] = useState(false)

  const [formVendor, setFormVendor] = useState('')
  const [formEndpoint, setFormEndpoint] = useState('')
  const [formPath, setFormPath] = useState('')
  const [formName, setFormName] = useState('')
  const [formDesc, setFormDesc] = useState('')
  const [formAsync, setFormAsync] = useState(false)
  const [formError, setFormError] = useState('')
  const [formSaving, setFormSaving] = useState(false)

  const loadAll = useCallback(async () => {
    setLoading(true)
    try {
      const [veRes, vRes, epRes] = await Promise.all([getVendorEndpoints(), getVendors(), getEndpoints()])
      if (veRes.success) setItems(veRes.data || [])
      if (vRes.success) setVendors(vRes.data || [])
      if (epRes.success) setEndpoints(epRes.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [])

  useEffect(() => { loadAll() }, [loadAll])

  const openCreate = () => {
    setEditRow(null)
    setFormVendor(''); setFormEndpoint(''); setFormPath(''); setFormName(''); setFormDesc(''); setFormAsync(false); setFormError('')
    setDrawerOpen(true)
  }

  const openEdit = (ve: VendorEndpoint) => {
    setEditRow(ve)
    setFormVendor(String(ve.vendor_id)); setFormEndpoint(String(ve.endpoint_id))
    setFormPath(ve.path); setFormName(ve.name); setFormDesc(ve.description); setFormAsync(ve.is_async); setFormError('')
    setDrawerOpen(true)
  }

  const handleSave = async () => {
    if (!formVendor) { setFormError('请选择供应商'); return }
    if (!formEndpoint) { setFormError('请选择端点'); return }
    setFormSaving(true); setFormError('')
    try {
      if (editRow) {
        const data: Record<string, string | number | boolean> = {}
        if (Number(formVendor) !== editRow.vendor_id) data.vendor_id = Number(formVendor)
        if (Number(formEndpoint) !== editRow.endpoint_id) data.endpoint_id = Number(formEndpoint)
        if (formPath !== editRow.path) data.path = formPath
        if (formName !== editRow.name) data.name = formName
        if (formDesc !== editRow.description) data.description = formDesc
        if (formAsync !== editRow.is_async) data.is_async = formAsync
        const res = await updateVendorEndpoint(editRow.id, data)
        if (!res.success) { setFormError(res.message || '更新失败'); return }
      } else {
        const res = await createVendorEndpoint({
          vendor_id: Number(formVendor), endpoint_id: Number(formEndpoint),
          path: formPath, name: formName, description: formDesc, is_async: formAsync,
        })
        if (!res.success) { setFormError(res.message || '创建失败'); return }
      }
      setDrawerOpen(false); loadAll()
    } catch { setFormError('网络错误') }
    finally { setFormSaving(false) }
  }

  const handleToggle = async (ve: VendorEndpoint) => {
    setActionLoading(true)
    try { await updateVendorEndpointStatus(ve.id, ve.status === 1 ? 2 : 1); loadAll() }
    catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  const handleDeleteConfirm = async () => {
    if (!deleteRow) return
    setActionLoading(true)
    try { await deleteVendorEndpoint(deleteRow.id); setDeleteOpen(false); setDeleteRow(null); loadAll() }
    catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  useEffect(() => {
    if (drawerOpen || deleteOpen || fieldOpen) document.body.style.overflow = 'hidden'
    else document.body.style.overflow = ''
    return () => { document.body.style.overflow = '' }
  }, [drawerOpen, deleteOpen, fieldOpen])

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return
      if (fieldOpen) setFieldOpen(false)
      else if (deleteOpen) setDeleteOpen(false)
      else if (drawerOpen) setDrawerOpen(false)
    }
    document.addEventListener('keydown', handleEsc)
    return () => document.removeEventListener('keydown', handleEsc)
  }, [drawerOpen, deleteOpen, fieldOpen])

  return (
    <div className='space-y-6'>
      <div className='flex items-center justify-between'>
        <div>
          <h2 className='text-2xl font-bold tracking-tight'>供应商端点</h2>
          <p className='text-muted-foreground mt-1 text-sm'>管理供应商的端点配置及字段</p>
        </div>
        <button onClick={openCreate}
          className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors'>
          <Plus className='size-4' /> 添加供应商端点
        </button>
      </div>

      <div className='border-border/60 overflow-hidden rounded-xl border'>
        <div className='overflow-x-auto'>
          <table className='w-full text-sm'>
            <thead>
              <tr className='bg-muted/30 border-border/40 border-b'>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>ID</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>供应商</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>绑定端点</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>端点名称</th>
                <th className='text-muted-foreground min-w-[200px] px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>端点路径</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>描述</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>异步</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>状态</th>
                <th className='text-muted-foreground px-4 py-3 text-left text-xs font-semibold tracking-wider uppercase'>操作</th>
              </tr>
            </thead>
            <tbody className='divide-border/40 divide-y'>
              {loading ? (
                <tr><td colSpan={9} className='px-4 py-12 text-center'><Loader2 className='text-muted-foreground mx-auto size-6 animate-spin' /></td></tr>
              ) : items.length === 0 ? (
                <tr><td colSpan={9} className='px-4 py-12 text-center'><p className='text-muted-foreground text-sm'>暂无供应商端点</p></td></tr>
              ) : items.map((ve) => {
                const statusConf = STATUS_CONFIG[ve.status] || STATUS_CONFIG[1]
                return (
                  <tr key={ve.id} className={cn('hover:bg-muted/20 transition-colors', ve.status === 2 && 'opacity-50')}>
                    <td className='px-4 py-3'><span className='text-muted-foreground font-mono text-xs'>{ve.id}</span></td>
                    <td className='px-4 py-3'>
                      <span className='inline-flex items-center gap-1.5 rounded-md bg-blue-500/10 px-2 py-0.5 text-xs font-medium text-blue-600 dark:text-blue-400'>
                        <Link2 className='size-3' />
                        {ve.vendor_name || `#${ve.vendor_id}`}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <span className='inline-flex items-center gap-1.5 rounded-md bg-purple-500/10 px-2 py-0.5 text-xs font-medium text-purple-600 dark:text-purple-400'>
                        <Globe className='size-3' />
                        {ve.endpoint_name || `#${ve.endpoint_id}`}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <span className='text-sm'>{ve.name || '-'}</span>
                    </td>
                    <td className='px-4 py-3'>
                      <code className='font-mono text-xs break-all'>{ve.path || ve.endpoint_path || '-'}</code>
                    </td>
                    <td className='px-4 py-3'>
                      <span className='text-muted-foreground text-xs'>{ve.description || '-'}</span>
                    </td>
                    <td className='px-4 py-3'>
                      {ve.is_async ? (
                        <span className='rounded bg-amber-500/10 px-1.5 py-0.5 text-[11px] font-medium text-amber-600 dark:text-amber-400'>异步</span>
                      ) : <span className='text-muted-foreground text-xs'>同步</span>}
                    </td>
                    <td className='px-4 py-3'>
                      <span className={cn('inline-flex items-center gap-1.5 rounded-md px-2 py-0.5 text-xs font-medium', statusConf.className)}>
                        <span className={cn('size-1.5 rounded-full', ve.status === 1 ? 'bg-emerald-500' : 'bg-red-500')} />
                        {statusConf.label}
                      </span>
                    </td>
                    <td className='px-4 py-3'>
                      <div className='flex items-center gap-1'>
                        <button onClick={() => openEdit(ve)} className='hover:bg-muted inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors'>
                          <Pencil className='size-3.5' /> 编辑
                        </button>
                        <button onClick={() => { setFieldRow(ve); setFieldOpen(true) }} className='hover:bg-muted inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors'>
                          <Settings className='size-3.5' /> 字段
                        </button>
                        <button onClick={() => handleToggle(ve)}
                          className={cn('inline-flex items-center gap-1 rounded-lg px-2 py-1.5 text-xs font-medium transition-colors',
                            ve.status === 1 ? 'hover:bg-amber-500/10 text-amber-600 dark:text-amber-400' : 'hover:bg-emerald-500/10 text-emerald-600 dark:text-emerald-400')}>
                          {ve.status === 1 ? <><PowerOff className='size-3.5' /> 禁用</> : <><Power className='size-3.5' /> 启用</>}
                        </button>
                        <button onClick={() => { setDeleteRow(ve); setDeleteOpen(true) }}
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
              <h2 className='text-lg font-semibold'>{editRow ? '编辑供应商端点' : '添加供应商端点'}</h2>
              <button onClick={() => setDrawerOpen(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'><X className='size-4' /></button>
            </div>
            <div className='p-6 space-y-4'>
              {formError && <div className='bg-destructive/10 text-destructive rounded-lg px-4 py-3 text-sm'>{formError}</div>}
              <div className='space-y-2'>
                <label className='text-sm font-medium'>供应商</label>
                <select value={formVendor} onChange={(e) => setFormVendor(e.target.value)}
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'>
                  <option value=''>请选择供应商</option>
                  {vendors.filter(v => v.status === 1).map(v => <option key={v.id} value={v.id}>{v.name}</option>)}
                </select>
              </div>
              <div className='space-y-2'>
                <label className='text-sm font-medium'>绑定端点</label>
                <select value={formEndpoint} onChange={(e) => {
                  setFormEndpoint(e.target.value)
                  const ep = endpoints.find(ep => ep.id === Number(e.target.value))
                  if (ep) {
                    if (!formPath) setFormPath(ep.path)
                    if (!formName) setFormName(ep.name)
                    if (!formDesc) setFormDesc(ep.description)
                  }
                }}
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'>
                  <option value=''>请选择端点</option>
                  {endpoints.filter(e => e.status === 1).map(e => <option key={e.id} value={e.id}>{e.name} ({e.path})</option>)}
                </select>
              </div>
              <div className='space-y-2'>
                <label className='text-sm font-medium'>端点路径</label>
                <input type='text' value={formPath} onChange={(e) => setFormPath(e.target.value)} placeholder='供应商实际调用的路径'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>
              <div className='space-y-2'>
                <label className='text-sm font-medium'>端点名称</label>
                <input type='text' value={formName} onChange={(e) => setFormName(e.target.value)} placeholder='供应商侧的端点名称'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>
              <div className='space-y-2'>
                <label className='text-sm font-medium'>描述</label>
                <input type='text' value={formDesc} onChange={(e) => setFormDesc(e.target.value)} placeholder='可选'
                  className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none' />
              </div>
              <label className='flex items-center gap-2 text-sm'>
                <input type='checkbox' checked={formAsync} onChange={(e) => setFormAsync(e.target.checked)} className='border-border/60 size-4 rounded' />
                异步模式
              </label>
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
        description={<>确定要删除此供应商端点吗？</>}
        confirmText='删除' destructive loading={actionLoading} onConfirm={handleDeleteConfirm} />

      {fieldRow && (
        <VendorEndpointFieldManager
          open={fieldOpen}
          onOpenChange={setFieldOpen}
          vendorEndpointId={fieldRow.id}
          endpointId={fieldRow.endpoint_id}
          vendorName={fieldRow.vendor_name}
          endpointName={fieldRow.endpoint_name || fieldRow.name}
        />
      )}
    </div>
  )
}
