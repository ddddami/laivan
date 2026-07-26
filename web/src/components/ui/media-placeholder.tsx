import { ImageNotFound01Icon } from '@hugeicons/core-free-icons'

import { ProductIcon } from './product-icon'

type MediaPlaceholderProps = {
  label?: string
  className?: string
}

export function MediaPlaceholder({
  label = 'No photos available',
  className = '',
}: MediaPlaceholderProps) {
  return (
    <div
      className={`bg-surface-strong text-faint flex min-h-40 flex-col items-center justify-center gap-2 ${className}`}
      role="img"
      aria-label={label}
    >
      <ProductIcon icon={ImageNotFound01Icon} size={28} />
      <span className="font-body text-xs">{label}</span>
    </div>
  )
}
