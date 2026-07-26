import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { MediaPlaceholder } from './media-placeholder'

describe('MediaPlaceholder', () => {
  it('provides an accessible no-image description', () => {
    render(<MediaPlaceholder />)

    expect(screen.getByRole('img', { name: 'No photos available' })).toBeInTheDocument()
  })
})
