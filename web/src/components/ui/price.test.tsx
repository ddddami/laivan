import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Price } from './price'

describe('Price', () => {
  it('formats a lowest annual price with honest from language', () => {
    render(<Price amount={350_000} from />)

    expect(screen.getByText('From')).toBeInTheDocument()
    expect(screen.getByText('₦350,000')).toBeInTheDocument()
    expect(screen.getByText('/yr')).toBeInTheDocument()
  })

  it('does not invent a price when none is available', () => {
    render(<Price amount={null} from />)

    expect(screen.getByText('Price unavailable')).toBeInTheDocument()
    expect(screen.queryByText('From')).not.toBeInTheDocument()
  })
})
