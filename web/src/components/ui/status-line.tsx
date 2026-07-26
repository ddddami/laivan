import { Alert02Icon, RefreshIcon } from '@hugeicons/core-free-icons'
import type { ReactNode } from 'react'

import { ProductIcon } from './product-icon'

type StatusLineProps = {
  children: ReactNode
  tone?: 'neutral' | 'warning'
  actionLabel?: string
  onAction?: () => void
}

export function StatusLine({ children, tone = 'neutral', actionLabel, onAction }: StatusLineProps) {
  return (
    <div
      className={`font-body rounded-control mb-4 flex min-h-11 items-center gap-2 px-3 text-xs ${
        tone === 'warning' ? 'bg-surface-strong text-foreground' : 'bg-surface text-muted'
      }`}
      role="status"
    >
      <ProductIcon icon={tone === 'warning' ? Alert02Icon : RefreshIcon} size={16} />
      <span className="flex-1">{children}</span>
      {actionLabel && onAction ? (
        <button
          type="button"
          className="focus-ring text-foreground rounded-control min-h-11 px-2 font-semibold underline underline-offset-2"
          onClick={onAction}
        >
          {actionLabel}
        </button>
      ) : null}
    </div>
  )
}
