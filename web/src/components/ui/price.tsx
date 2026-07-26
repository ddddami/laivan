type PriceProps = {
  amount: number | null
  from?: boolean
  className?: string
}

const naira = new Intl.NumberFormat('en-NG', {
  style: 'currency',
  currency: 'NGN',
  maximumFractionDigits: 0,
})

export function formatPrice(amount: number) {
  return naira.format(amount).replace('NGN', '₦')
}

export function Price({ amount, from = false, className = '' }: PriceProps) {
  if (amount === null) {
    return <span className={`font-body text-muted text-sm ${className}`}>Price unavailable</span>
  }

  return (
    <span className={className}>
      {from ? <span className="font-body text-muted mb-0.5 block text-xs">From</span> : null}
      <span className="text-foreground font-mono text-[17px] leading-none font-semibold tracking-[-0.025em]">
        {formatPrice(amount)}
        <span className="text-muted ml-0.5 text-[11px] font-normal">/yr</span>
      </span>
    </span>
  )
}
