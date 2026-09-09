import { useState } from 'react'
import { Copy, Check } from 'lucide-react'
import { cn } from '../lib/utils'

interface CopyButtonProps {
  text: string
  className?: string
}

// 复制按钮（点击复制，短暂显示对勾）
export function CopyButton({ text, className }: CopyButtonProps) {
  const [copied, setCopied] = useState(false)

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      // 兼容非安全上下文（如 http）
      const input = document.createElement('input')
      input.value = text
      document.body.appendChild(input)
      input.select()
      document.execCommand('copy')
      document.body.removeChild(input)
    }
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <button
      type='button'
      onClick={copy}
      title='复制'
      className={cn('hover:bg-muted shrink-0 rounded p-1 transition-colors', className)}
    >
      {copied ? <Check className='text-emerald-500 size-3.5' /> : <Copy className='text-muted-foreground size-3.5' />}
    </button>
  )
}
