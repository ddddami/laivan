import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { DiscoveryFilters } from './discovery-filters'
import { defaultDiscoverySearch } from './search'

describe('DiscoveryFilters', () => {
  it('applies one form model as validated route search state', () => {
    const onApply = vi.fn()
    render(<DiscoveryFilters search={defaultDiscoverySearch} onApply={onApply} />)

    fireEvent.click(screen.getByRole('button', { name: 'Self-contained' }))
    fireEvent.change(screen.getByLabelText('Area'), { target: { value: ' South Gate ' } })
    fireEvent.change(screen.getByLabelText('Minimum (₦)'), {
      target: { value: '200000' },
    })
    fireEvent.change(screen.getByLabelText('Maximum (₦)'), {
      target: { value: '500000' },
    })
    fireEvent.change(screen.getByLabelText('Bathroom'), { target: { value: 'private' } })
    fireEvent.change(screen.getByLabelText('Kitchen'), { target: { value: 'shared' } })
    fireEvent.change(screen.getByLabelText('Parlour'), { target: { value: 'false' } })
    fireEvent.change(screen.getByLabelText('Availability'), { target: { value: 'all' } })
    fireEvent.click(screen.getByRole('button', { name: 'Apply filters' }))

    expect(onApply).toHaveBeenCalledWith({
      category: 'self_contained',
      area: 'South Gate',
      min_price: 200_000,
      max_price: 500_000,
      bathroom_type: 'private',
      kitchen_type: 'shared',
      has_parlour: false,
      availability: 'all',
      sort: 'recommended',
      page: 1,
    })
  })

  it('clears filters while preserving the current sort', () => {
    const onApply = vi.fn()
    render(
      <DiscoveryFilters
        search={{
          ...defaultDiscoverySearch,
          category: 'single_room',
          area: 'Obanla',
          sort: '-created_at',
        }}
        onApply={onApply}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Clear' }))

    expect(onApply).toHaveBeenCalledWith({
      availability: 'available',
      sort: '-created_at',
      page: 1,
    })
  })
})
