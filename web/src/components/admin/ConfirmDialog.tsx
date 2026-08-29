import { useEffect, useRef } from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '../../lib/utils'

interface ConfirmDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: React.ReactNode
  confirmText?: string
  cancelText?: string
  destructive?: boolean
  loading?: boolean
  onConfirm: () => void
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  destructive = false,
  loading = false,
  onConfirm,
}: ConfirmDialogProps) {
  const overlayRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => { document.body.style.overflow = '' }
  }, [open])

  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && open) onOpenChange(false)
    }
    document.addEventListener('keydown', handleEsc)
    return () => document.removeEventListener('keydown', handleEsc)
  }, [open, onOpenChange])

  if (!open) return null

  return (
    <div className='fixed inset-0 z-[100] flex items-center justify-center'>
      {/* Overlay */}
      <div
        ref={overlayRef}
        className='bg-background/80 fixed inset-0 backdrop-blur-sm'
        onClick={() => onOpenChange(false)}
      />
      {/* Dialog */}
      <div className='bg-background border-border/60 relative z-10 w-full max-w-md rounded-xl border p-6 shadow-lg'>
        <h2 className='text-lg font-semibold'>{title}</h2>
        <p className='text-muted-foreground mt-2 text-sm'>{description}</p>
        <div className='mt-6 flex justify-end gap-3'>
          <button
            onClick={() => onOpenChange(false)}
            disabled={loading}
            className='border-border/60 hover:bg-muted inline-flex h-9 items-center justify-center rounded-lg border px-4 text-sm font-medium transition-colors'
          >
            {cancelText}
          </button>
          <button
            onClick={onConfirm}
            disabled={loading}
            className={cn(
              'inline-flex h-9 items-center justify-center gap-2 rounded-lg px-4 text-sm font-medium text-white transition-colors disabled:opacity-50',
              destructive
                ? 'bg-destructive hover:bg-destructive/90'
                : 'bg-primary hover:bg-primary/90'
            )}
          >
            {loading && <Loader2 className='size-4 animate-spin' />}
            {loading ? '处理中...' : confirmText}
          </button>
        </div>
      </div>
    </div>
  )
}
