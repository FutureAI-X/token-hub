import { useState, useEffect, useCallback } from 'react'
import {
  Plus,
  Trash2,
  Loader2,
  X,
  Coins,
  Settings,
  CheckCircle2,
} from 'lucide-react'
import {
  getQuotaRule,
  saveQuotaRule,
  deleteModelQuotaRule,
  type QuotaRule,
  type QuotaRuleItem,
} from '../../api/admin-model'
import { ConfirmDialog } from './ConfirmDialog'

const RULE_TYPE_OPTIONS = [
  { value: 'per_request', label: '按次计费' },
]

// 表单项接口（price 用字符串处理输入）
interface FormQuotaRuleItem {
  param_path: string
  param_value: string
  price: string
}

interface QuotaRuleManagerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  modelId: number
  modelName: string
}

export function QuotaRuleManager({ open, onOpenChange, modelId, modelName }: QuotaRuleManagerProps) {
  const [rule, setRule] = useState<QuotaRule | null>(null)
  const [loading, setLoading] = useState(true)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)

  // 表单
  const [formRuleType, setFormRuleType] = useState('per_request')
  const [formBasePrice, setFormBasePrice] = useState('')
  const [formDesc, setFormDesc] = useState('')
  const [formItems, setFormItems] = useState<FormQuotaRuleItem[]>([])
  const [formError, setFormError] = useState('')
  const [formSaving, setFormSaving] = useState(false)
  const [showSuccess, setShowSuccess] = useState(false)

  // 格式化价格为两位小数
  const formatPrice = (price: number): string => {
    return price.toFixed(2)
  }

  const loadRule = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getQuotaRule(modelId)
      if (res.success && res.data) {
        setRule(res.data)
        setFormRuleType(res.data.rule_type)
        setFormBasePrice(formatPrice(res.data.base_price))
        setFormDesc(res.data.description || '')
        // 将后端数据转换为表单格式
        setFormItems((res.data.items || []).map(item => ({
          param_path: item.param_path,
          param_value: item.param_value,
          price: formatPrice(item.price),
        })))
      } else {
        setRule(null)
        setFormRuleType('per_request')
        setFormBasePrice('')
        setFormDesc('')
        setFormItems([])
      }
    } catch {
      setRule(null)
    } finally {
      setLoading(false)
    }
  }, [modelId])

  useEffect(() => {
    if (open) loadRule()
  }, [open, loadRule])

  // 验证小数位数不超过2位
  const validateDecimalPlaces = (value: number): boolean => {
    const str = value.toString()
    const decimalPart = str.split('.')[1]
    return !decimalPart || decimalPart.length <= 2
  }

  // 添加一行参数映射
  const addItem = () => {
    setFormItems([...formItems, { param_path: '', param_value: '', price: '' }])
  }

  // 删除一行参数映射
  const removeItem = (index: number) => {
    setFormItems(formItems.filter((_, i) => i !== index))
  }

  // 更新参数映射
  const updateItem = (index: number, field: keyof FormQuotaRuleItem, value: string) => {
    const newItems = [...formItems]
    newItems[index] = { ...newItems[index], [field]: value }
    setFormItems(newItems)
  }

  const handleSave = async () => {
    if (!formBasePrice || parseFloat(formBasePrice) <= 0) {
      setFormError('基础价格必须大于 0')
      return
    }

    if (!validateDecimalPlaces(parseFloat(formBasePrice))) {
      setFormError('基础价格最多支持2位小数')
      return
    }

    // 验证参数映射
    for (let i = 0; i < formItems.length; i++) {
      const item = formItems[i]
      if (!item.param_path.trim()) {
        setFormError(`第 ${i + 1} 项的参数路径不能为空`)
        return
      }
      if (!item.param_value.trim()) {
        setFormError(`第 ${i + 1} 项的参数值不能为空`)
        return
      }
      const price = parseFloat(item.price)
      if (isNaN(price) || price <= 0) {
        setFormError(`第 ${i + 1} 项的价格必须大于 0`)
        return
      }
      if (!validateDecimalPlaces(price)) {
        setFormError(`第 ${i + 1} 项的价格最多支持2位小数`)
        return
      }
    }

    setFormSaving(true)
    setFormError('')
    try {
      const res = await saveQuotaRule(modelId, {
        rule_type: formRuleType,
        base_price: parseFloat(formBasePrice),
        description: formDesc || undefined,
        items: formItems.length > 0 ? formItems.map(item => ({
          param_path: item.param_path,
          param_value: item.param_value,
          price: parseFloat(item.price),
        })) : undefined,
      })
      if (!res.success) {
        setFormError(res.message || '保存失败')
        return
      }
      setShowSuccess(true)
      setTimeout(() => setShowSuccess(false), 2000)
      loadRule()
    } catch {
      setFormError('网络错误')
    } finally {
      setFormSaving(false)
    }
  }

  const handleDeleteConfirm = async () => {
    if (!rule) return
    setActionLoading(true)
    try {
      await deleteModelQuotaRule(modelId)
      setDeleteOpen(false)
      setRule(null)
      setFormRuleType('per_request')
      setFormBasePrice('')
      setFormDesc('')
      setFormItems([])
    } catch { /* ignore */ }
    finally { setActionLoading(false) }
  }

  if (!open) return null

  return (
    <div className='fixed inset-0 z-[100] flex items-center justify-center p-4'>
      <div className='bg-background/80 fixed inset-0 backdrop-blur-sm' onClick={() => onOpenChange(false)} />
      <div className='bg-background border-border/60 relative z-10 flex h-[85vh] w-full max-w-4xl flex-col rounded-xl border shadow-lg'>
        <div className='flex items-center justify-between border-b border-border/40 px-6 py-4'>
          <div className='flex items-center gap-3'>
            <Coins className='size-5 text-amber-500' />
            <div>
              <h2 className='text-lg font-semibold'>积分规则配置</h2>
              <p className='text-muted-foreground text-sm'>
                配置模型 <span className='font-medium text-foreground'>{modelName}</span> 的积分扣除规则
              </p>
            </div>
          </div>
          <button onClick={() => onOpenChange(false)} className='hover:bg-muted rounded-lg p-1.5 transition-colors'>
            <X className='size-4' />
          </button>
        </div>

        <div className='flex-1 overflow-auto p-6'>
          {loading ? (
            <div className='flex h-full items-center justify-center'>
              <Loader2 className='text-muted-foreground size-6 animate-spin' />
            </div>
          ) : (
            <div className='space-y-6'>
              {/* 基本规则配置 */}
              <div className='rounded-lg border border-border/60 p-5'>
                <div className='mb-5 flex items-center gap-2'>
                  <Settings className='size-4 text-muted-foreground' />
                  <h3 className='text-sm font-medium'>基本规则</h3>
                </div>

                <div className='grid grid-cols-2 gap-6'>
                  <div className='space-y-2'>
                    <label className='text-sm font-medium'>规则类型</label>
                    <select
                      value={formRuleType}
                      onChange={(e) => setFormRuleType(e.target.value)}
                      className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                    >
                      {RULE_TYPE_OPTIONS.map((t) => (
                        <option key={t.value} value={t.value}>{t.label}</option>
                      ))}
                    </select>
                    <p className='text-muted-foreground text-xs'>目前仅支持按次计费</p>
                  </div>

                  <div className='space-y-2'>
                    <label className='text-sm font-medium'>基础积分价格</label>
                    <div className='flex items-center gap-2'>
                      <input
                        type='number'
                        value={formBasePrice}
                        onChange={(e) => setFormBasePrice(e.target.value)}
                        placeholder='例如: 10'
                        min='0'
                        step='0.01'
                        className='border-border/60 bg-background focus-visible:ring-ring flex h-9 flex-1 rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                      />
                      <span className='text-muted-foreground text-sm whitespace-nowrap'>积分/次</span>
                    </div>
                    <p className='text-muted-foreground text-xs'>每次 API 调用扣除的基础积分数量</p>
                  </div>
                </div>

                <div className='mt-5 space-y-2'>
                  <label className='text-sm font-medium'>描述 <span className='text-muted-foreground font-normal'>（可选）</span></label>
                  <input
                    type='text'
                    value={formDesc}
                    onChange={(e) => setFormDesc(e.target.value)}
                    placeholder='规则说明'
                    className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                  />
                </div>
              </div>

              {/* 参数价格映射 */}
              <div className='rounded-lg border border-border/60 p-5'>
                <div className='mb-4 flex items-center justify-between'>
                  <div className='flex items-center gap-2'>
                    <Coins className='size-4 text-muted-foreground' />
                    <h3 className='text-sm font-medium'>参数价格映射 <span className='text-muted-foreground font-normal'>（可选）</span></h3>
                  </div>
                  <button
                    type='button'
                    onClick={addItem}
                    className='text-primary hover:bg-primary/10 inline-flex h-8 items-center gap-1 rounded-lg px-3 text-xs font-medium transition-colors'
                  >
                    <Plus className='size-3.5' /> 添加映射
                  </button>
                </div>

                <p className='text-muted-foreground mb-4 text-xs'>
                  为不同的请求参数值设置差异化价格，匹配时将覆盖基础价格
                </p>

                {formItems.length === 0 ? (
                  <div className='rounded-lg border border-dashed border-border/60 p-6 text-center'>
                    <p className='text-muted-foreground text-sm'>暂无参数映射</p>
                    <p className='text-muted-foreground text-xs mt-1'>点击上方按钮添加差异化计费规则</p>
                  </div>
                ) : (
                  <div className='space-y-3'>
                    {/* 表头 */}
                    <div className='grid grid-cols-[1fr_1.5fr_1fr_40px] gap-4 text-xs font-medium text-muted-foreground'>
                      <div>参数路径</div>
                      <div>参数值</div>
                      <div>积分价格</div>
                      <div></div>
                    </div>

                    {/* 表单行 */}
                    {formItems.map((item, index) => (
                      <div key={index} className='grid grid-cols-[1fr_1.5fr_1fr_40px] items-center gap-4'>
                        <div>
                          <input
                            type='text'
                            value={item.param_path}
                            onChange={(e) => updateItem(index, 'param_path', e.target.value)}
                            placeholder='如: size'
                            className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                          />
                        </div>
                        <div>
                          <input
                            type='text'
                            value={item.param_value}
                            onChange={(e) => updateItem(index, 'param_value', e.target.value)}
                            placeholder='如: 1024x1024'
                            className='border-border/60 bg-background focus-visible:ring-ring flex h-9 w-full rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                          />
                        </div>
                        <div>
                          <div className='flex items-center gap-2'>
                            <input
                              type='text'
                              inputMode='decimal'
                              value={item.price}
                              onChange={(e) => updateItem(index, 'price', e.target.value)}
                              placeholder='价格'
                              className='border-border/60 bg-background focus-visible:ring-ring flex h-9 flex-1 rounded-lg border px-3 text-sm focus-visible:ring-2 focus-visible:outline-none'
                            />
                            <span className='text-muted-foreground text-xs whitespace-nowrap'>积分/次</span>
                          </div>
                        </div>
                        <div className='flex justify-center'>
                          <button
                            type='button'
                            onClick={() => removeItem(index)}
                            className='hover:bg-destructive/10 text-destructive inline-flex h-9 w-9 items-center justify-center rounded-lg transition-colors'
                          >
                            <Trash2 className='size-4' />
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* 错误提示 */}
              {formError && (
                <div className='bg-destructive/10 text-destructive rounded-lg px-4 py-3 text-sm'>{formError}</div>
              )}

              {/* 成功提示 */}
              {showSuccess && (
                <div className='bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center gap-2 rounded-lg px-4 py-3 text-sm'>
                  <CheckCircle2 className='size-4' />
                  积分规则保存成功
                </div>
              )}
            </div>
          )}
        </div>

        {/* 底部按钮 */}
        <div className='flex items-center justify-between border-t border-border/40 px-6 py-4'>
          <div>
            {rule && (
              <button
                onClick={() => setDeleteOpen(true)}
                className='text-destructive hover:bg-destructive/10 inline-flex h-9 items-center gap-1 rounded-lg px-4 text-sm font-medium transition-colors'
              >
                删除规则
              </button>
            )}
          </div>
          <div className='flex items-center gap-3'>
            <button
              onClick={() => onOpenChange(false)}
              disabled={formSaving}
              className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center rounded-lg border px-4 text-sm font-medium transition-colors'
            >
              取消
            </button>
            <button
              onClick={handleSave}
              disabled={formSaving}
              className='bg-primary hover:bg-primary/90 inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors disabled:opacity-50'
            >
              {formSaving && <Loader2 className='size-4 animate-spin' />}
              {formSaving ? '保存中...' : '保存'}
            </button>
          </div>
        </div>

        <ConfirmDialog
          open={deleteOpen}
          onOpenChange={setDeleteOpen}
          title='确认删除'
          description={<>确定要删除该模型的积分规则吗？</>}
          confirmText='删除'
          destructive
          loading={actionLoading}
          onConfirm={handleDeleteConfirm}
        />
      </div>
    </div>
  )
}
