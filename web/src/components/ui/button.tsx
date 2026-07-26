import type { ButtonHTMLAttributes } from 'react'

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: 'primary' | 'secondary' | 'accent'
}

const variants = {
  primary: 'bg-foreground text-background hover:bg-foreground/85',
  secondary: 'bg-surface text-foreground hover:bg-surface-strong',
  accent: 'bg-accent text-foreground hover:bg-accent/80',
} as const

export function Button({
  className = '',
  variant = 'primary',
  type = 'button',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      className={`focus-ring font-display rounded-control min-h-11 px-4 py-2.5 text-sm font-bold transition-colors disabled:cursor-not-allowed disabled:opacity-50 ${variants[variant]} ${className}`}
      {...props}
    />
  )
}
