import { describe, expect, it, vi } from 'vitest'

import {
  ApiError,
  createAuthenticatedApiClient,
  createPublicApiClient,
  type PropertyDetail,
} from './client'

describe('public API client', () => {
  it('resolves a campus by its public slug', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(
        JSON.stringify({
          campus: {
            id: '550e8400-e29b-41d4-a716-446655440002',
            slug: 'futa',
            name: 'Federal University of Technology, Akure',
            short_name: 'FUTA',
          },
        }),
        { headers: { 'Content-Type': 'application/json' } },
      )
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test/',
      fetch: fetcher,
    })

    const campus = await client.getCampus('futa')

    expect(campus.short_name).toBe('FUTA')
    expect(fetcher).toHaveBeenCalledWith(
      'https://api.laivan.test/v1/campuses/futa',
      expect.objectContaining({ headers: { Accept: 'application/json' } }),
    )
  })

  it('normalizes discovery parameters into a stable request URL', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(
        JSON.stringify({
          results: [],
          metadata: {
            current_page: 0,
            page_size: 0,
            first_page: 0,
            last_page: 0,
            total_records: 0,
          },
        }),
        { headers: { 'Content-Type': 'application/json' } },
      )
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    const response = await client.discover({
      campus_id: '550e8400-e29b-41d4-a716-446655440002',
      category: 'single_room, self_contained,single_room',
      q: '  Alice Lodge  ',
      area: '  South Gate  ',
      has_parlour: false,
      page: 2,
      page_size: 20,
      sort: 'recommended',
    })

    expect(response.results).toEqual([])
    expect(fetcher).toHaveBeenCalledWith(
      'https://api.laivan.test/v1/discovery?campus_id=550e8400-e29b-41d4-a716-446655440002&category=self_contained%2Csingle_room&q=Alice+Lodge&area=South+Gate&has_parlour=false&page=2&page_size=20&sort=recommended',
      expect.any(Object),
    )
  })

  it('exposes API validation errors with their field details', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(
        JSON.stringify({
          error: {
            code: 'validation_failed',
            message: 'One or more fields failed validation',
            fields: {
              min_price: 'must be greater than zero',
            },
          },
        }),
        {
          status: 422,
          headers: { 'Content-Type': 'application/json' },
        },
      )
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    const request = client.discover({
      campus_id: '550e8400-e29b-41d4-a716-446655440002',
      min_price: -1,
    })

    await expect(request).rejects.toBeInstanceOf(ApiError)
    await expect(request).rejects.toMatchObject({
      name: 'ApiError',
      kind: 'response',
      status: 422,
      code: 'validation_failed',
      fields: {
        min_price: 'must be greater than zero',
      },
    })
  })

  it('translates a network failure into a retryable API error', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      throw new TypeError('Failed to fetch')
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    const request = client.getProperty('550e8400-e29b-41d4-a716-446655440010')

    await expect(request).rejects.toBeInstanceOf(ApiError)
    await expect(request).rejects.toMatchObject({
      kind: 'network',
      code: 'network_error',
      status: undefined,
    })
  })

  it.each([
    { status: 404, code: 'not_found' },
    { status: 500, code: 'internal_server_error' },
  ])('preserves a $status API failure', async ({ status, code }) => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(
        JSON.stringify({
          error: {
            code,
            message: 'Request failed',
          },
        }),
        {
          status,
          headers: { 'Content-Type': 'application/json' },
        },
      )
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    const request = client.getCampus('futa')

    await expect(request).rejects.toMatchObject({
      kind: 'response',
      status,
      code,
    })
  })

  it('falls back to a structured error when an error response is not JSON', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response('Bad gateway', { status: 502 })
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    await expect(client.getCampus('futa')).rejects.toMatchObject({
      kind: 'response',
      status: 502,
      code: 'request_failed',
    })
  })

  it('reports malformed successful JSON as an invalid response', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response('{', {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      })
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    await expect(client.getCampus('futa')).rejects.toMatchObject({
      kind: 'response',
      status: 200,
      code: 'invalid_response',
    })
  })

  it('returns a property detail from its response envelope', async () => {
    const property: PropertyDetail = {
      id: '550e8400-e29b-41d4-a716-446655440010',
      campus_id: '550e8400-e29b-41d4-a716-446655440002',
      name: 'Alice Lodge',
      area: 'Obanla',
      landmark: 'Near South Gate',
      description: 'Gated student lodge.',
      version: 1,
      unit_types: [],
      media: [],
      created_at: '2026-05-01T10:00:00Z',
      updated_at: '2026-05-02T10:00:00Z',
    }
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(JSON.stringify({ property }), {
        headers: { 'Content-Type': 'application/json' },
      })
    })
    const client = createPublicApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    const result = await client.getProperty(property.id)

    expect(result).toEqual(property)
    expect(fetcher).toHaveBeenCalledWith(
      `https://api.laivan.test/v1/properties/${property.id}`,
      expect.any(Object),
    )
  })
})

describe('authenticated API client', () => {
  it('includes the browser session when loading session state', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(
        JSON.stringify({
          authenticated: false,
          user: null,
          roles: [],
          agent: null,
          csrf_token: null,
        }),
        { headers: { 'Content-Type': 'application/json' } },
      )
    })
    const client = createAuthenticatedApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })

    await client.getSession()

    expect(fetcher).toHaveBeenCalledWith(
      'https://api.laivan.test/v1/auth/session',
      expect.objectContaining({ credentials: 'include' }),
    )
  })

  it('sends CSRF and credentials for inquiry mutations', async () => {
    const fetcher = vi.fn<typeof fetch>(async () => {
      return new Response(
        JSON.stringify({
          inquiry: {
            id: '550e8400-e29b-41d4-a716-446655440060',
            agent_offer_id: '550e8400-e29b-41d4-a716-446655440010',
            message: 'Is the kitchen private?',
            status: 'open',
            created_at: '2026-05-01T10:00:00Z',
          },
          handoff: {
            channel: 'whatsapp',
            url: 'https://wa.me/2348000000000',
          },
        }),
        { headers: { 'Content-Type': 'application/json' } },
      )
    })
    const client = createAuthenticatedApiClient({
      baseUrl: 'https://api.laivan.test',
      fetch: fetcher,
    })
    const input = {
      message: 'Is the kitchen private?',
      submission_id: '550e8400-e29b-41d4-a716-446655440061',
    }

    await client.createInquiry('550e8400-e29b-41d4-a716-446655440010', input, 'csrf-token')

    expect(fetcher).toHaveBeenCalledWith(
      'https://api.laivan.test/v1/agent-offers/550e8400-e29b-41d4-a716-446655440010/inquiries',
      expect.objectContaining({
        method: 'POST',
        credentials: 'include',
        headers: {
          Accept: 'application/json',
          'Content-Type': 'application/json',
          'X-CSRF-Token': 'csrf-token',
        },
        body: JSON.stringify(input),
      }),
    )
  })
})
