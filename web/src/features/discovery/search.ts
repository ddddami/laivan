import type { DiscoveryParams } from '../../api/client'
import type { components } from '../../api/schema.gen'

export const unitCategories = [
  'single_room',
  'self_contained',
  'room_and_parlour',
  'one_bedroom_flat',
  'two_bedroom_flat',
  'three_bedroom_flat',
  'other',
] as const satisfies readonly components['schemas']['UnitCategory'][]

export const discoverySorts = [
  'recommended',
  '-created_at',
  'lowest_price_naira',
  '-lowest_price_naira',
] as const satisfies readonly NonNullable<DiscoveryParams['sort']>[]

type UnitCategory = (typeof unitCategories)[number]
export type DiscoverySort = (typeof discoverySorts)[number]

export type DiscoverySearch = {
  category?: string
  q?: string
  area?: string
  min_price?: number
  max_price?: number
  bathroom_type?: 'private' | 'shared'
  kitchen_type?: 'private' | 'shared' | 'none'
  has_parlour?: boolean
  availability: 'available' | 'all'
  sort: DiscoverySort
  page: number
}

export const defaultDiscoverySearch: DiscoverySearch = {
  availability: 'available',
  sort: 'recommended',
  page: 1,
}

export function normalizeDiscoverySearch(raw: Record<string, unknown>): DiscoverySearch {
  return {
    category: normalizeCategories(raw.category),
    q: normalizeSearchText(raw.q),
    area: normalizeText(raw.area),
    min_price: positiveInteger(raw.min_price),
    max_price: positiveInteger(raw.max_price),
    bathroom_type: enumValue(raw.bathroom_type, ['private', 'shared']),
    kitchen_type: enumValue(raw.kitchen_type, ['private', 'shared', 'none']),
    has_parlour: booleanValue(raw.has_parlour),
    availability: enumValue(raw.availability, ['available', 'all']) ?? 'available',
    sort: enumValue(raw.sort, discoverySorts) ?? 'recommended',
    page: positiveInteger(raw.page) ?? 1,
  }
}

export function discoverySearchError(search: DiscoverySearch) {
  if (
    search.min_price !== undefined &&
    search.max_price !== undefined &&
    search.min_price > search.max_price
  ) {
    return 'Minimum price cannot be higher than maximum price.'
  }
  return undefined
}

export function activeFilterCount(search: DiscoverySearch) {
  return [
    search.category,
    search.area,
    search.min_price,
    search.max_price,
    search.bathroom_type,
    search.kitchen_type,
    search.has_parlour,
    search.availability === 'all' ? 'all' : undefined,
  ].filter((value) => value !== undefined && value !== '').length
}

function normalizeCategories(value: unknown) {
  if (typeof value !== 'string') return undefined
  const allowed = new Set<UnitCategory>(unitCategories)
  const categories = [
    ...new Set(
      value
        .split(',')
        .map((category) => category.trim())
        .filter((category): category is UnitCategory => allowed.has(category as UnitCategory)),
    ),
  ].sort()
  return categories.length > 0 ? categories.join(',') : undefined
}

function normalizeText(value: unknown) {
  if (typeof value !== 'string') return undefined
  const normalized = value.trim()
  return normalized || undefined
}

function normalizeSearchText(value: unknown) {
  const normalized = normalizeText(value)
  return normalized ? [...normalized].slice(0, 100).join('') : undefined
}

function positiveInteger(value: unknown) {
  const parsed =
    typeof value === 'number'
      ? value
      : typeof value === 'string' && value.trim() !== ''
        ? Number(value)
        : Number.NaN
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : undefined
}

function booleanValue(value: unknown) {
  if (value === true || value === 'true') return true
  if (value === false || value === 'false') return false
  return undefined
}

function enumValue<const TValue extends string>(
  value: unknown,
  allowed: readonly TValue[],
): TValue | undefined {
  return typeof value === 'string' && allowed.includes(value as TValue)
    ? (value as TValue)
    : undefined
}
