import { ArrowLeft01Icon, ArrowRight01Icon } from '@hugeicons/core-free-icons'

import { ProductIcon } from '../../components/ui/product-icon'

type PaginationProps = {
  currentPage: number
  lastPage: number
  pageSize: number
  totalRecords: number
  onPageChange: (page: number) => void
}

export function Pagination({
  currentPage,
  lastPage,
  pageSize,
  totalRecords,
  onPageChange,
}: PaginationProps) {
  if (totalRecords === 0) return null

  const start = (currentPage - 1) * pageSize + 1
  const end = Math.min(currentPage * pageSize, totalRecords)
  const pages = nearbyPages(currentPage, lastPage)

  return (
    <nav className="border-border mt-6 border-t pt-5" aria-label="Discovery pages">
      <p className="font-body text-faint mb-3 text-center text-xs">
        Showing {start.toLocaleString('en-NG')}–{end.toLocaleString('en-NG')} of{' '}
        {totalRecords.toLocaleString('en-NG')}
      </p>

      <div className="flex items-center justify-between gap-3">
        <PageButton
          label="Previous"
          disabled={currentPage <= 1}
          onClick={() => onPageChange(currentPage - 1)}
          icon="previous"
        />

        <div className="hidden items-center gap-1 sm:flex">
          {pages.map((page, index) =>
            page === 'ellipsis' ? (
              <span key={`ellipsis-${index}`} className="text-faint px-2" aria-hidden>
                …
              </span>
            ) : (
              <button
                key={page}
                type="button"
                aria-label={`Page ${page}`}
                aria-current={page === currentPage ? 'page' : undefined}
                className={`focus-ring rounded-control min-h-11 min-w-11 px-2 font-mono text-xs font-semibold ${
                  page === currentPage
                    ? 'bg-foreground text-background'
                    : 'bg-surface text-muted hover:bg-surface-strong'
                }`}
                onClick={() => onPageChange(page)}
              >
                {page}
              </button>
            ),
          )}
        </div>

        <PageButton
          label="Next"
          disabled={currentPage >= lastPage}
          onClick={() => onPageChange(currentPage + 1)}
          icon="next"
        />
      </div>
    </nav>
  )
}

type PageButtonProps = {
  label: string
  disabled: boolean
  onClick: () => void
  icon: 'previous' | 'next'
}

function PageButton({ label, disabled, onClick, icon }: PageButtonProps) {
  return (
    <button
      type="button"
      disabled={disabled}
      className="focus-ring bg-surface text-foreground font-body rounded-control flex min-h-11 items-center gap-2 px-3.5 text-sm font-semibold disabled:cursor-not-allowed disabled:opacity-40"
      onClick={onClick}
    >
      {icon === 'previous' ? <ProductIcon icon={ArrowLeft01Icon} size={16} /> : null}
      {label}
      {icon === 'next' ? <ProductIcon icon={ArrowRight01Icon} size={16} /> : null}
    </button>
  )
}

export function nearbyPages(currentPage: number, lastPage: number) {
  if (lastPage <= 7) {
    return Array.from({ length: lastPage }, (_, index) => index + 1)
  }

  const candidates = new Set([1, lastPage, currentPage - 1, currentPage, currentPage + 1])
  const pages = [...candidates]
    .filter((page) => page >= 1 && page <= lastPage)
    .sort((first, second) => first - second)

  const result: Array<number | 'ellipsis'> = []
  pages.forEach((page, index) => {
    const previous = pages[index - 1]
    if (previous !== undefined && page - previous > 1) {
      result.push('ellipsis')
    }
    result.push(page)
  })
  return result
}
