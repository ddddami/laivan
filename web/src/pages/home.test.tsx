import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Home } from './home'

describe('Home', () => {
  it('renders the Laivan homepage', () => {
    render(<Home />)

    expect(screen.getByRole('link', { name: 'Laivan' })).toBeInTheDocument()
  })
})
