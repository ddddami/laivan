import { HugeiconsIcon, type IconSvgElement } from '@hugeicons/react'

type ProductIconProps = {
  icon: IconSvgElement
  size?: number
  className?: string
  label?: string
}

export function ProductIcon({ icon, size = 20, className, label }: ProductIconProps) {
  return (
    <HugeiconsIcon
      icon={icon}
      size={size}
      strokeWidth={1.7}
      className={className}
      aria-hidden={label ? undefined : true}
      aria-label={label}
    />
  )
}
