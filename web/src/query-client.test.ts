import { describe, expect, it, vi } from 'vitest'

import { ApiError } from './api/client'
import { createAppQueryClient } from './query-client'

describe('application query client', () => {
  it('does not retry a validation failure', async () => {
    const queryClient = createAppQueryClient()
    const queryFn = vi.fn(async () => {
      throw new ApiError({
        kind: 'response',
        status: 422,
        code: 'validation_failed',
        message: 'Invalid request',
      })
    })

    await expect(
      queryClient.fetchQuery({
        queryKey: ['validation-failure'],
        queryFn,
      }),
    ).rejects.toMatchObject({ status: 422 })
    expect(queryFn).toHaveBeenCalledTimes(1)
  })

  it('retries a server failure once and then stops', async () => {
    const queryClient = createAppQueryClient()
    const queryFn = vi.fn(async () => {
      throw new ApiError({
        kind: 'response',
        status: 500,
        code: 'internal_server_error',
        message: 'Request failed',
      })
    })

    await expect(
      queryClient.fetchQuery({
        queryKey: ['server-failure'],
        queryFn,
      }),
    ).rejects.toMatchObject({ status: 500 })
    expect(queryFn).toHaveBeenCalledTimes(2)
  })

  it('retries a network failure once and then stops', async () => {
    const queryClient = createAppQueryClient()
    const queryFn = vi.fn(async () => {
      throw new ApiError({
        kind: 'network',
        code: 'network_error',
        message: 'Unable to reach Laivan',
      })
    })

    await expect(
      queryClient.fetchQuery({
        queryKey: ['network-failure'],
        queryFn,
      }),
    ).rejects.toMatchObject({ kind: 'network' })
    expect(queryFn).toHaveBeenCalledTimes(2)
  })

  it('reuses fresh public data without another request', async () => {
    const queryClient = createAppQueryClient()
    const queryFn = vi.fn(async () => 'Alice Lodge')
    const query = {
      queryKey: ['property', 'alice-lodge'],
      queryFn,
    } as const

    const first = await queryClient.fetchQuery(query)
    const second = await queryClient.fetchQuery(query)

    expect(first).toBe('Alice Lodge')
    expect(second).toBe('Alice Lodge')
    expect(queryFn).toHaveBeenCalledTimes(1)
  })
})
