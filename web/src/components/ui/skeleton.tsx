type SkeletonProps = {
  className?: string
}

export function Skeleton({ className = '' }: SkeletonProps) {
  return (
    <div className={`bg-surface-strong rounded-control animate-pulse ${className}`} aria-hidden />
  )
}
