import { describe, expect, it } from 'vitest'

import { createPublicConfig } from './config'

describe('public API configuration', () => {
  it('uses same-origin requests when no API base URL is configured', () => {
    expect(createPublicConfig(undefined)).toEqual({ apiBaseUrl: '' })
  })

  it('normalizes an absolute API base URL', () => {
    expect(createPublicConfig('  https://api.laivan.test/  ')).toEqual({
      apiBaseUrl: 'https://api.laivan.test',
    })
  })

  it('rejects unsupported API base URLs', () => {
    expect(() => createPublicConfig('javascript:alert(1)')).toThrow(
      'VITE_API_BASE_URL must use http or https',
    )
  })

  it.each(['https://api.laivan.test?region=futa', 'https://api.laivan.test#discovery'])(
    'rejects an API base URL with request-specific parts',
    (value) => {
      expect(() => createPublicConfig(value)).toThrow(
        'VITE_API_BASE_URL must not include a query string or fragment',
      )
    },
  )
})
