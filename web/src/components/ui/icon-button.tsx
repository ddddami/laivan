import type { ButtonHTMLAttributes } from 'react'
import type { IconSvgElement } from '@hugeicons/react'

import { ProductIcon } from './product-icon'

type IconButtonProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'children'> & {
  icon: IconSvgElement
  label: string
}

export function IconButton({
  icon,
  label,
  className = '',
  type = 'button',
  ...props
}: IconButtonProps) {
  return (
    <button
      type={type}
      className={`focus-ring bg-surface text-muted rounded-control hover:bg-surface-strong flex size-11 shrink-0 items-center justify-center transition-colors ${className}`}
      aria-label={label}
      {...props}
    >
      <ProductIcon icon={icon} />
    </button>
  )
}
