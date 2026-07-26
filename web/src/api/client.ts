import type { components, operations } from './schema.gen'

export type Campus = components['schemas']['Campus']
export type DiscoveryResponse = components['schemas']['DiscoveryResponse']
export type DiscoveryResult = components['schemas']['DiscoveryResult']
export type PropertyDetail = components['schemas']['PropertyDetail']
export type UnitTypeDetail = components['schemas']['UnitTypeDetail']
export type AgentOfferDetail = components['schemas']['AgentOfferDetail']
export type Media = components['schemas']['Media']
export type DiscoveryParams = operations['discover']['parameters']['query']

type ClientOptions = {
  baseUrl: string
  fetch?: typeof fetch
}

type ApiErrorOptions = {
  kind: 'network' | 'response'
  code: string
  message: string
  status?: number
  fields?: Readonly<Record<string, string>>
  cause?: unknown
}

export class ApiError extends Error {
  readonly name = 'ApiError'
  readonly kind: ApiErrorOptions['kind']
  readonly code: string
  readonly status: number | undefined
  readonly fields: Readonly<Record<string, string>> | undefined

  constructor({ kind, code, message, status, fields, cause }: ApiErrorOptions) {
    super(message, { cause })
    this.kind = kind
    this.code = code
    this.status = status
    this.fields = fields
  }
}

export function createPublicApiClient({
  baseUrl,
  fetch: fetcher = globalThis.fetch.bind(globalThis),
}: ClientOptions) {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')

  async function getJson<T>(path: string): Promise<T> {
    let response: Response
    try {
      response = await fetcher(`${normalizedBaseUrl}${path}`, {
        headers: { Accept: 'application/json' },
      })
    } catch (cause) {
      throw new ApiError({
        kind: 'network',
        code: 'network_error',
        message: 'Unable to reach Laivan',
        cause,
      })
    }

    if (!response.ok) {
      throw await responseError(response)
    }

    try {
      // The generated OpenAPI contract supplies the response type; JSON has no runtime type metadata.
      return (await response.json()) as T
    } catch (cause) {
      throw new ApiError({
        kind: 'response',
        status: response.status,
        code: 'invalid_response',
        message: 'Laivan returned an invalid response',
        cause,
      })
    }
  }

  return {
    async getCampus(slug: string): Promise<Campus> {
      const response = await getJson<components['schemas']['CampusResponse']>(
        `/v1/campuses/${encodeURIComponent(slug)}`,
      )
      return response.campus
    },

    async discover(params: DiscoveryParams): Promise<DiscoveryResponse> {
      return getJson<DiscoveryResponse>(
        `/v1/discovery?${discoverySearchParams(normalizeDiscoveryParams(params))}`,
      )
    },

    async getProperty(id: string): Promise<PropertyDetail> {
      const response = await getJson<components['schemas']['PropertyDetailResponse']>(
        `/v1/properties/${encodeURIComponent(id)}`,
      )
      return response.property
    },
  }
}

export type PublicApiClient = ReturnType<typeof createPublicApiClient>

export function normalizeDiscoveryParams(params: DiscoveryParams): DiscoveryParams {
  const categories = [
    ...new Set(
      params.category
        ?.split(',')
        .map((category) => category.trim())
        .filter((category) => category !== ''),
    ),
  ].sort()

  return {
    campus_id: params.campus_id.trim(),
    category: categories.length > 0 ? categories.join(',') : undefined,
    q: normalizedString(params.q),
    area: normalizedString(params.area),
    bathroom_type: params.bathroom_type,
    kitchen_type: params.kitchen_type,
    has_parlour: params.has_parlour,
    min_price: params.min_price,
    max_price: params.max_price,
    availability: params.availability,
    page: params.page,
    page_size: params.page_size,
    sort: params.sort,
  }
}

function discoverySearchParams(params: DiscoveryParams): string {
  const search = new URLSearchParams()
  appendString(search, 'campus_id', params.campus_id)
  appendString(search, 'category', params.category)
  appendString(search, 'q', params.q)
  appendString(search, 'area', params.area)
  appendString(search, 'bathroom_type', params.bathroom_type)
  appendString(search, 'kitchen_type', params.kitchen_type)
  appendValue(search, 'has_parlour', params.has_parlour)
  appendValue(search, 'min_price', params.min_price)
  appendValue(search, 'max_price', params.max_price)
  appendString(search, 'availability', params.availability)
  appendValue(search, 'page', params.page)
  appendValue(search, 'page_size', params.page_size)
  appendString(search, 'sort', params.sort)
  return search.toString()
}

function appendString(search: URLSearchParams, name: string, value: string | undefined) {
  const normalized = normalizedString(value)
  if (normalized) {
    search.set(name, normalized)
  }
}

function normalizedString(value: string | undefined) {
  const normalized = value?.trim()
  return normalized || undefined
}

function appendValue(search: URLSearchParams, name: string, value: boolean | number | undefined) {
  if (value !== undefined) {
    search.set(name, String(value))
  }
}

async function responseError(response: Response): Promise<ApiError> {
  let body: unknown
  try {
    body = await response.json()
  } catch {
    body = undefined
  }

  const error = recordValue(recordValue(body, 'error'))
  return new ApiError({
    kind: 'response',
    status: response.status,
    code: stringValue(error, 'code') ?? 'request_failed',
    message: stringValue(error, 'message') ?? `Request failed with status ${response.status}`,
    fields: stringRecordValue(error, 'fields'),
  })
}

function recordValue(value: unknown, key?: string): Readonly<Record<string, unknown>> | undefined {
  const candidate = key && isRecord(value) ? value[key] : value
  return isRecord(candidate) ? candidate : undefined
}

function isRecord(value: unknown): value is Readonly<Record<string, unknown>> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function stringValue(value: Readonly<Record<string, unknown>> | undefined, key: string) {
  const candidate = value?.[key]
  return typeof candidate === 'string' ? candidate : undefined
}

function stringRecordValue(
  value: Readonly<Record<string, unknown>> | undefined,
  key: string,
): Readonly<Record<string, string>> | undefined {
  const record = recordValue(value?.[key])
  if (!record) {
    return undefined
  }

  const entries = Object.entries(record)
  if (!entries.every((entry): entry is [string, string] => typeof entry[1] === 'string')) {
    return undefined
  }
  return Object.fromEntries(entries)
}
