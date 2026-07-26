import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { ResponsiveImage } from './responsive-image'

describe('ResponsiveImage', () => {
  it('replaces a failed marketplace image with the accessible fallback', () => {
    render(<ResponsiveImage src="https://media.invalid/room.jpg" alt="Alice Lodge room" />)

    fireEvent.error(screen.getByRole('img', { name: 'Alice Lodge room' }))

    expect(screen.getByRole('img', { name: 'No photos available' })).toBeInTheDocument()
  })
})
