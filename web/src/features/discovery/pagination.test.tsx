import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { Pagination, nearbyPages } from './pagination'

describe('Pagination', () => {
  it('shows the result range and changes pages through semantic controls', () => {
    const onPageChange = vi.fn()
    render(
      <Pagination
        currentPage={2}
        lastPage={4}
        pageSize={20}
        totalRecords={67}
        onPageChange={onPageChange}
      />,
    )

    expect(screen.getByText('Showing 21–40 of 67')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Page 2' })).toHaveAttribute('aria-current', 'page')

    fireEvent.click(screen.getByRole('button', { name: 'Previous' }))
    fireEvent.click(screen.getByRole('button', { name: 'Next' }))

    expect(onPageChange).toHaveBeenNthCalledWith(1, 1)
    expect(onPageChange).toHaveBeenNthCalledWith(2, 3)
  })

  it('keeps first, last, and nearby desktop pages visible', () => {
    expect(nearbyPages(6, 12)).toEqual([1, 'ellipsis', 5, 6, 7, 'ellipsis', 12])
  })
})
