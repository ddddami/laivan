import type { ButtonHTMLAttributes } from 'react'

type FilterChipProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  selected?: boolean
}

export function FilterChip({
  selected = false,
  className = '',
  type = 'button',
  ...props
}: FilterChipProps) {
  return (
    <button
      type={type}
      aria-pressed={selected}
      className={`focus-ring font-body min-h-11 shrink-0 rounded-full px-4 py-2 text-xs font-semibold transition-colors ${
        selected ? 'bg-accent text-foreground' : 'bg-surface text-muted hover:bg-surface-strong'
      } ${className}`}
      {...props}
    />
  )
}
