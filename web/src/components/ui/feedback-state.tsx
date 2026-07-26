import { Alert02Icon } from '@hugeicons/core-free-icons'

import { Button } from './button'
import { ProductIcon } from './product-icon'

type FeedbackStateProps = {
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
}

export function FeedbackState({ title, description, actionLabel, onAction }: FeedbackStateProps) {
  return (
    <div className="border-border bg-surface rounded-card border px-5 py-10 text-center">
      <span className="bg-background text-faint rounded-control mx-auto mb-4 flex size-12 items-center justify-center">
        <ProductIcon icon={Alert02Icon} size={23} />
      </span>
      <h2 className="font-display text-foreground text-lg font-bold tracking-[-0.02em]">{title}</h2>
      <p className="font-body text-muted mx-auto mt-2 max-w-sm text-sm leading-6">{description}</p>
      {actionLabel && onAction ? (
        <Button className="mt-5" onClick={onAction}>
          {actionLabel}
        </Button>
      ) : null}
    </div>
  )
}
