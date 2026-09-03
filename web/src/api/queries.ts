import { keepPreviousData, queryOptions } from '@tanstack/react-query'

import {
  createPublicApiClient,
  createAuthenticatedApiClient,
  normalizeDiscoveryParams,
  type DiscoveryParams,
  type PublicApiClient,
} from './client'
import { publicConfig } from './config'

export const publicApiClient = createPublicApiClient({
  baseUrl: publicConfig.apiBaseUrl,
})

export const authenticatedApiClient = createAuthenticatedApiClient({
  baseUrl: publicConfig.apiBaseUrl,
})

export const publicQueryKeys = {
  campus: (slug: string) => ['public', 'campus', slug.trim().toLowerCase()] as const,
  discovery: (params: DiscoveryParams) =>
    ['public', 'discovery', normalizeDiscoveryParams(params)] as const,
  property: (id: string) => ['public', 'property', id.trim()] as const,
}

export const authQueryKeys = {
  session: ['auth', 'session'] as const,
}

export function sessionQueryOptions(client = authenticatedApiClient) {
  return queryOptions({
    queryKey: authQueryKeys.session,
    queryFn: () => client.getSession(),
    staleTime: 30_000,
  })
}

export function campusQueryOptions(slug: string, client: PublicApiClient = publicApiClient) {
  const normalizedSlug = slug.trim().toLowerCase()
  return queryOptions({
    queryKey: publicQueryKeys.campus(normalizedSlug),
    queryFn: () => client.getCampus(normalizedSlug),
  })
}

export function discoveryQueryOptions(
  params: DiscoveryParams,
  client: PublicApiClient = publicApiClient,
) {
  const normalizedParams = normalizeDiscoveryParams(params)
  return queryOptions({
    queryKey: publicQueryKeys.discovery(normalizedParams),
    queryFn: () => client.discover(normalizedParams),
    placeholderData: keepPreviousData,
  })
}

export function propertyQueryOptions(id: string, client: PublicApiClient = publicApiClient) {
  const normalizedID = id.trim()
  return queryOptions({
    queryKey: publicQueryKeys.property(normalizedID),
    queryFn: () => client.getProperty(normalizedID),
  })
}
