import { useState, useEffect, useMemo, useCallback } from 'react'
import { Link } from 'react-router-dom'
import { Search, Copy, ChevronRight, LayoutGrid, List } from 'lucide-react'
import { cn } from '../lib/utils'
import { Header } from '../components/Header'
import { PricingSidebar } from '../components/PricingSidebar'
import type { PricingModel, PricingVendor, PricingData } from '../types/pricing'

// ── 筛选常量 ──
const FILTER_ALL = '__all__'

// ── 常量 ──
const TOKEN_UNITS = ['K', 'M'] as const
type TokenUnit = (typeof TOKEN_UNITS)[number]

// ── 工具函数 ──
function formatPrice(
  model: PricingModel,
  type: 'input' | 'output',
  tokenUnit: TokenUnit
): string {
  if (model.quota_type === 1) {
    return `¥${model.model_price.toFixed(4)}`
  }
  const base = model.model_ratio * 0.002 // 基础价格：0.002元/1K tokens
  const multiplier = type === 'output' ? model.completion_ratio : 1
  const unitDivisor = tokenUnit === 'M' ? 1 : 1000
  return `¥${(base * multiplier * unitDivisor).toFixed(4)}`
}

function formatNumber(n: number): string {
  if (n >= 1000000) return `${(n / 1000000).toFixed(0)}M`
  if (n >= 1000) return `${(n / 1000).toFixed(0)}K`
  return n.toString()
}

function parseTags(tags?: string): string[] {
  if (!tags) return []
  return tags.split(',').map((t) => t.trim()).filter(Boolean)
}

// ── 复制到剪贴板 ──
function useCopyToClipboard() {
  const [copied, setCopied] = useState(false)
  const copy = useCallback((text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }, [])
  return { copied, copy }
}

// ── 模型卡片 ──
function ModelCard({
  model,
  tokenUnit,
  onClick,
}: {
  model: PricingModel
  tokenUnit: TokenUnit
  onClick: () => void
}) {
  const { copy } = useCopyToClipboard()
  const tags = parseTags(model.tags)
  const initial = model.name?.charAt(0).toUpperCase() || '?'
  const isTokenBased = model.quota_type === 0
  const tokenUnitLabel = tokenUnit === 'K' ? '1K' : '1M'

  return (
    <div className='group relative flex flex-col rounded-xl border p-3 transition-colors hover:bg-muted/20 sm:p-5'>
      {/* 头部：图标 + 名称 + 操作 */}
      <div className='flex items-start justify-between gap-2.5 sm:gap-3'>
        <div className='flex min-w-0 items-start gap-2.5 sm:gap-3'>
          <div className='bg-muted/40 flex size-9 shrink-0 items-center justify-center rounded-lg sm:size-10 sm:rounded-xl'>
            <span className='text-muted-foreground text-sm font-bold'>
              {initial}
            </span>
          </div>
          <div className='min-w-0'>
            <h3 className='text-foreground truncate font-mono text-[15px] leading-tight font-bold'>
              {model.name}
            </h3>
            <div className='mt-0.5 flex flex-wrap items-baseline gap-x-2 gap-y-0.5 text-sm sm:mt-1 sm:gap-x-3'>
              {isTokenBased ? (
                <>
                  <span className='text-muted-foreground whitespace-nowrap'>
                    输入{' '}
                    <span className='text-foreground font-mono font-semibold'>
                      {formatPrice(model, 'input', tokenUnit)}
                    </span>
                  </span>
                  <span className='text-muted-foreground whitespace-nowrap'>
                    输出{' '}
                    <span className='text-foreground font-mono font-semibold'>
                      {formatPrice(model, 'output', tokenUnit)}
                    </span>
                  </span>
                </>
              ) : (
                <span className='text-muted-foreground whitespace-nowrap'>
                  <span className='text-foreground font-mono font-semibold'>
                    ¥{model.model_price.toFixed(4)}
                  </span>{' '}
                  / 次
                </span>
              )}
            </div>
          </div>
        </div>

        <div className='flex shrink-0 items-center gap-1.5'>
          <button
            type='button'
            onClick={onClick}
            className='text-muted-foreground hover:text-foreground hover:bg-muted inline-flex items-center gap-1 rounded-md border px-2 py-1 text-xs transition-colors sm:px-2.5 sm:py-1.5'
          >
            详情
            <ChevronRight className='size-3.5' />
          </button>
          <button
            type='button'
            onClick={() => copy(model.name)}
            className='text-muted-foreground hover:text-foreground hover:bg-muted rounded-md border p-1.5 transition-colors'
            title='复制模型名称'
          >
            <Copy className='size-3.5' />
          </button>
        </div>
      </div>

      {/* 描述 */}
      <p className='text-muted-foreground mt-2 line-clamp-1 flex-1 text-[13px] leading-relaxed sm:mt-4 sm:line-clamp-2 sm:min-h-[2.5rem]'>
        {model.description || '暂无描述'}
      </p>

      {/* 底部信息 */}
      <div className='mt-2 flex min-w-0 flex-wrap items-center gap-x-2.5 gap-y-1 sm:mt-4 sm:gap-x-3'>
        {model.vendor_name && (
          <span className='text-muted-foreground text-sm font-medium'>
            {model.vendor_name}
          </span>
        )}
        {model.context_length > 0 && (
          <span className='text-muted-foreground/70 text-xs'>
            {formatNumber(model.context_length)} 上下文
          </span>
        )}
        {tags.slice(0, 3).map((tag) => (
          <span key={tag} className='text-muted-foreground/70 text-xs'>
            {tag}
          </span>
        ))}
        {isTokenBased && (
          <span className='text-muted-foreground/50 text-xs'>
            {tokenUnitLabel} tokens
          </span>
        )}
      </div>
    </div>
  )
}

// ── 主页面 ──
export function Pricing() {
  const [data, setData] = useState<PricingData | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [tokenUnit, setTokenUnit] = useState<TokenUnit>('M')
  const [vendorFilter, setVendorFilter] = useState(FILTER_ALL)
  const [tagFilter, setTagFilter] = useState(FILTER_ALL)
  const [quotaTypeFilter, setQuotaTypeFilter] = useState(FILTER_ALL)
  const [viewMode, setViewMode] = useState<'card' | 'list'>('card')

  // 获取定价数据
  useEffect(() => {
    fetch('/api/pricing')
      .then((res) => res.json())
      .then((res) => {
        if (res.success) {
          setData(res.data)
        } else {
          setError(res.message || '获取数据失败')
        }
      })
      .catch(() => setError('网络错误'))
      .finally(() => setLoading(false))
  }, [])

  // 过滤模型
  const filteredModels = useMemo(() => {
    if (!data?.models) return []
    let models = data.models

    // 供应商过滤
    if (vendorFilter !== FILTER_ALL) {
      models = models.filter((m) => m.vendor_name === vendorFilter)
    }

    // 标签过滤
    if (tagFilter !== FILTER_ALL) {
      models = models.filter((m) =>
        parseTags(m.tags)
          .map((t) => t.toLowerCase())
          .includes(tagFilter.toLowerCase())
      )
    }

    // 计费类型过滤
    if (quotaTypeFilter !== FILTER_ALL) {
      models = models.filter((m) => m.quota_type === Number(quotaTypeFilter))
    }

    // 搜索过滤
    if (search) {
      const q = search.toLowerCase()
      models = models.filter(
        (m) =>
          m.name.toLowerCase().includes(q) ||
          m.vendor_name?.toLowerCase().includes(q) ||
          m.tags?.toLowerCase().includes(q) ||
          m.description?.toLowerCase().includes(q)
      )
    }

    return models
  }, [data?.models, search, vendorFilter, tagFilter, quotaTypeFilter])

  // 可用供应商列表
  const vendors = data?.vendors || []

  // 清除所有筛选
  const clearFilters = useCallback(() => {
    setSearch('')
    setVendorFilter(FILTER_ALL)
    setTagFilter(FILTER_ALL)
    setQuotaTypeFilter(FILTER_ALL)
  }, [])

  const hasActiveFilters =
    search !== '' ||
    vendorFilter !== FILTER_ALL ||
    tagFilter !== FILTER_ALL ||
    quotaTypeFilter !== FILTER_ALL

  // 加载状态
  if (loading) {
    return (
      <div className='bg-background text-foreground relative min-h-svh'>
        <div className='mx-auto w-full max-w-[1400px] px-3 pt-20 pb-8 sm:px-6 sm:pt-24 sm:pb-10'>
          <div className='animate-pulse space-y-4'>
            <div className='bg-muted/30 mx-auto h-10 w-64 rounded-lg' />
            <div className='bg-muted/30 mx-auto h-6 w-96 rounded-lg' />
            <div className='mt-8 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3'>
              {Array.from({ length: 6 }).map((_, i) => (
                <div key={i} className='bg-muted/20 h-48 rounded-xl' />
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }

  // 错误状态
  if (error) {
    return (
      <div className='bg-background text-foreground flex min-h-svh items-center justify-center'>
        <div className='text-center'>
          <p className='text-destructive text-lg font-medium'>{error}</p>
          <button
            onClick={() => window.location.reload()}
            className='bg-primary text-primary-foreground mt-4 rounded-lg px-4 py-2 text-sm'
          >
            重试
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className='bg-background text-foreground relative min-h-svh overflow-x-clip'>
      <Header />

      {/* ── 主内容 ── */}
      <div className='relative'>
        {/* 背景渐变 */}
        <div
          aria-hidden
          className='pointer-events-none absolute inset-x-0 top-0 h-[600px] opacity-20 dark:opacity-[0.10]'
          style={{
            background: [
              'radial-gradient(ellipse 60% 50% at 20% 20%, oklch(0.72 0.18 250 / 80%) 0%, transparent 70%)',
              'radial-gradient(ellipse 50% 40% at 80% 15%, oklch(0.65 0.15 200 / 60%) 0%, transparent 70%)',
              'radial-gradient(ellipse 40% 35% at 50% 70%, oklch(0.70 0.12 280 / 40%) 0%, transparent 70%)',
            ].join(', '),
            maskImage: 'linear-gradient(to bottom, black 40%, transparent 100%)',
            WebkitMaskImage: 'linear-gradient(to bottom, black 40%, transparent 100%)',
          }}
        />

        <div className='relative mx-auto w-full max-w-[1800px] px-3 pt-16 pb-8 sm:px-6 sm:pt-20 sm:pb-10 xl:px-8'>
          {/* 页头 */}
          <header className='mx-auto mb-5 max-w-3xl pt-5 text-center sm:mb-10 sm:pt-10'>
            <h1 className='text-[clamp(2rem,5.5vw,3.5rem)] leading-[1.15] font-bold tracking-tight'>
              模型广场
            </h1>
            <p className='text-muted-foreground/80 mt-3 text-sm sm:mt-4 sm:text-base'>
              当前共有 <span className='text-foreground font-semibold'>{data?.models.length || 0}</span> 个模型可用
            </p>
            <p className='text-muted-foreground/60 mx-auto mt-2 max-w-2xl text-xs leading-relaxed sm:text-sm'>
              探索精选 AI 模型，比较价格和能力，为每个场景选择合适的模型。
            </p>

            {/* 搜索框 */}
            <div className='relative mx-auto mt-4 max-w-2xl sm:mt-6'>
              <Search className='text-muted-foreground absolute top-1/2 left-3.5 size-4 -translate-y-1/2' />
              <input
                type='text'
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder='搜索模型名称、供应商、标签...'
                className='bg-background border-border/60 ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-full rounded-xl border pr-4 pl-10 text-sm focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none'
              />
              {search && (
                <button
                  onClick={() => setSearch('')}
                  className='text-muted-foreground hover:text-foreground absolute top-1/2 right-3 -translate-y-1/2 text-xs'
                >
                  清除
                </button>
              )}
            </div>
          </header>

          {/* 侧边栏 + 内容区 */}
          <div className='grid gap-4 xl:grid-cols-[330px_minmax(0,1fr)]'>
            {/* 侧边栏（桌面端显示） */}
            <PricingSidebar
              vendorFilter={vendorFilter}
              tagFilter={tagFilter}
              quotaTypeFilter={quotaTypeFilter}
              onVendorChange={setVendorFilter}
              onTagChange={setTagFilter}
              onQuotaTypeChange={setQuotaTypeFilter}
              vendors={vendors}
              models={data?.models || []}
              hasActiveFilters={hasActiveFilters}
              onClearFilters={clearFilters}
              className='hover-scrollbar sticky top-4 hidden max-h-[calc(100dvh-2rem)] self-start overflow-y-auto xl:block'
            />

            {/* 主内容区 */}
            <main className='min-w-0'>
              {/* 工具栏 */}
              <div className='mb-4 flex flex-wrap items-center justify-between gap-3'>
                <div className='flex items-center gap-2'>
                  <span className='text-muted-foreground text-sm'>
                    共 <span className='text-foreground font-medium'>{filteredModels.length}</span> 个模型
                    {filteredModels.length !== (data?.models.length || 0) && (
                      <span className='text-muted-foreground/60'> / {data?.models.length || 0}</span>
                    )}
                  </span>
                  {hasActiveFilters && (
                    <button
                      onClick={clearFilters}
                      className='text-muted-foreground hover:text-foreground text-xs underline underline-offset-2'
                    >
                      清除筛选
                    </button>
                  )}
                </div>

                <div className='flex items-center gap-2'>
                  {/* Token 单位切换 */}
                  <div className='bg-muted/50 flex items-center rounded-lg p-0.5'>
                    {TOKEN_UNITS.map((unit) => (
                      <button
                        key={unit}
                        onClick={() => setTokenUnit(unit)}
                        className={cn(
                          'rounded-md px-2.5 py-1 text-xs font-medium transition-colors',
                          tokenUnit === unit
                            ? 'bg-background text-foreground shadow-sm'
                            : 'text-muted-foreground hover:text-foreground'
                        )}
                      >
                        {unit}
                      </button>
                    ))}
                  </div>

                  {/* 视图切换 */}
                  <div className='bg-muted/50 flex items-center rounded-lg p-0.5'>
                    <button
                      onClick={() => setViewMode('card')}
                      className={cn(
                        'rounded-md p-1.5 transition-colors',
                        viewMode === 'card'
                          ? 'bg-background text-foreground shadow-sm'
                          : 'text-muted-foreground hover:text-foreground'
                      )}
                    >
                      <LayoutGrid className='size-3.5' />
                    </button>
                    <button
                      onClick={() => setViewMode('list')}
                      className={cn(
                        'rounded-md p-1.5 transition-colors',
                        viewMode === 'list'
                          ? 'bg-background text-foreground shadow-sm'
                          : 'text-muted-foreground hover:text-foreground'
                      )}
                    >
                      <List className='size-3.5' />
                    </button>
                  </div>
                </div>
              </div>

          {/* 模型列表 */}
          {filteredModels.length === 0 ? (
            <div className='flex flex-col items-center justify-center py-20'>
              <p className='text-muted-foreground text-lg'>没有找到匹配的模型</p>
              {hasActiveFilters && (
                <button
                  onClick={clearFilters}
                  className='text-primary mt-2 text-sm underline underline-offset-2'
                >
                  清除筛选条件
                </button>
              )}
            </div>
          ) : viewMode === 'card' ? (
            <div className='grid grid-cols-1 gap-3 sm:gap-4 md:grid-cols-2 lg:grid-cols-3'>
              {filteredModels.map((model) => (
                <ModelCard
                  key={model.id}
                  model={model}
                  tokenUnit={tokenUnit}
                  onClick={() => {}}
                />
              ))}
            </div>
          ) : (
            <div className='border-border/40 overflow-hidden rounded-xl border'>
              <table className='w-full'>
                <thead>
                  <tr className='bg-muted/30 border-border/40 border-b'>
                    <th className='px-4 py-3 text-left text-xs font-medium'>模型</th>
                    <th className='px-4 py-3 text-left text-xs font-medium'>供应商</th>
                    <th className='px-4 py-3 text-left text-xs font-medium'>输入价格</th>
                    <th className='px-4 py-3 text-left text-xs font-medium'>输出价格</th>
                    <th className='px-4 py-3 text-left text-xs font-medium'>上下文</th>
                    <th className='px-4 py-3 text-left text-xs font-medium'>标签</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredModels.map((model) => (
                    <tr
                      key={model.id}
                      className='border-border/20 hover:bg-muted/20 border-b transition-colors'
                    >
                      <td className='px-4 py-3'>
                        <span className='font-mono text-sm font-medium'>{model.name}</span>
                      </td>
                      <td className='text-muted-foreground px-4 py-3 text-sm'>
                        {model.vendor_name || '-'}
                      </td>
                      <td className='px-4 py-3 font-mono text-sm'>
                        {formatPrice(model, 'input', tokenUnit)}
                      </td>
                      <td className='px-4 py-3 font-mono text-sm'>
                        {formatPrice(model, 'output', tokenUnit)}
                      </td>
                      <td className='text-muted-foreground px-4 py-3 text-sm'>
                        {model.context_length > 0 ? formatNumber(model.context_length) : '-'}
                      </td>
                      <td className='px-4 py-3'>
                        <div className='flex flex-wrap gap-1'>
                          {parseTags(model.tags)
                            .slice(0, 2)
                            .map((tag) => (
                              <span
                                key={tag}
                                className='bg-muted/50 rounded px-1.5 py-0.5 text-[11px]'
                              >
                                {tag}
                              </span>
                            ))}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}

          {/* 底部信息 */}
          <div className='text-muted-foreground/60 mt-8 text-center text-xs'>
            价格单位：每 {tokenUnit === 'K' ? '千' : '百万'} tokens（人民币）
          </div>
            </main>
          </div>
        </div>
      </div>

      {/* ── Footer ── */}
      <footer className='border-border/40 relative z-10 border-t px-6 py-10'>
        <div className='mx-auto flex max-w-6xl flex-col items-center gap-4 text-center'>
          <div className='flex items-center gap-2'>
            <span>⚡</span>
            <span className='text-sm font-semibold'>Token Hub</span>
          </div>
          <p className='text-muted-foreground text-xs'>下一代 LLM 网关和 AI 资产管理系统</p>
          <p className='text-muted-foreground/60 text-xs'>© 2026 Token Hub. 基于 Go + Gin + React 构建</p>
        </div>
      </footer>
    </div>
  )
}
