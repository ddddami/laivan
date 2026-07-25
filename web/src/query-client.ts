import { QueryClient } from '@tanstack/react-query'

import { ApiError } from './api/client'

const publicDataStaleTime = 30_000
const publicDataCacheTime = 5 * 60_000

export function createAppQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: publicDataStaleTime,
        gcTime: publicDataCacheTime,
        retry: (failureCount, error) => {
          if (failureCount >= 1 || !(error instanceof ApiError)) {
            return false
          }
          return error.kind === 'network' || (error.status !== undefined && error.status >= 500)
        },
        retryDelay: 500,
      },
    },
  })
}
