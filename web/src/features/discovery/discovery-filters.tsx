import { useState, type FormEvent } from 'react'

import { unitCategoryLabel } from '../../components/marketplace/labels'
import { Button } from '../../components/ui/button'
import { FilterChip } from '../../components/ui/filter-chip'
import { TextField } from '../../components/ui/text-field'
import { defaultDiscoverySearch, unitCategories, type DiscoverySearch } from './search'

type FilterDraft = {
  categories: string[]
  area: string
  minPrice: string
  maxPrice: string
  bathroom: '' | 'private' | 'shared'
  kitchen: '' | 'private' | 'shared' | 'none'
  parlour: '' | 'true' | 'false'
  availability: 'available' | 'all'
}

type DiscoveryFiltersProps = {
  search: DiscoverySearch
  onApply: (search: DiscoverySearch) => void
  onClose?: () => void
}

export function DiscoveryFilters({ search, onApply, onClose }: DiscoveryFiltersProps) {
  const [draft, setDraft] = useState<FilterDraft>(() => draftFromSearch(search))

  function toggleCategory(category: string) {
    setDraft((current) => ({
      ...current,
      categories: current.categories.includes(category)
        ? current.categories.filter((value) => value !== category)
        : [...current.categories, category],
    }))
  }

  function apply(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    onApply({
      category: draft.categories.length > 0 ? [...draft.categories].sort().join(',') : undefined,
      q: search.q,
      area: draft.area.trim() || undefined,
      min_price: inputPrice(draft.minPrice),
      max_price: inputPrice(draft.maxPrice),
      bathroom_type: draft.bathroom || undefined,
      kitchen_type: draft.kitchen || undefined,
      has_parlour: draft.parlour === '' ? undefined : draft.parlour === 'true',
      availability: draft.availability,
      sort: search.sort,
      page: 1,
    })
    onClose?.()
  }

  function clear() {
    const cleared = { ...defaultDiscoverySearch, q: search.q, sort: search.sort }
    setDraft(draftFromSearch(cleared))
    onApply(cleared)
    onClose?.()
  }

  return (
    <form className="flex min-h-0 flex-1 flex-col" onSubmit={apply}>
      <div className="flex-1 space-y-6 overflow-y-auto px-5 py-5">
        <fieldset>
          <legend className="section-label">Category</legend>
          <div className="flex flex-wrap gap-2">
            {unitCategories.map((category) => (
              <FilterChip
                key={category}
                selected={draft.categories.includes(category)}
                onClick={() => toggleCategory(category)}
              >
                {unitCategoryLabel(category)}
              </FilterChip>
            ))}
          </div>
        </fieldset>

        <TextField
          label="Area"
          value={draft.area}
          placeholder="e.g. South Gate"
          autoComplete="off"
          onChange={(event) => setDraft((current) => ({ ...current, area: event.target.value }))}
        />

        <fieldset>
          <legend className="section-label">Annual price</legend>
          <div className="grid grid-cols-2 gap-3">
            <TextField
              label="Minimum (₦)"
              type="number"
              min={1}
              step={1}
              inputMode="numeric"
              value={draft.minPrice}
              placeholder="200000"
              onChange={(event) =>
                setDraft((current) => ({ ...current, minPrice: event.target.value }))
              }
            />
            <TextField
              label="Maximum (₦)"
              type="number"
              min={1}
              step={1}
              inputMode="numeric"
              value={draft.maxPrice}
              placeholder="600000"
              onChange={(event) =>
                setDraft((current) => ({ ...current, maxPrice: event.target.value }))
              }
            />
          </div>
        </fieldset>

        <FilterSelect
          label="Bathroom"
          value={draft.bathroom}
          onChange={(value) =>
            setDraft((current) => ({
              ...current,
              bathroom: value as FilterDraft['bathroom'],
            }))
          }
          options={[
            ['', 'Any bathroom'],
            ['private', 'Private'],
            ['shared', 'Shared'],
          ]}
        />

        <FilterSelect
          label="Kitchen"
          value={draft.kitchen}
          onChange={(value) =>
            setDraft((current) => ({
              ...current,
              kitchen: value as FilterDraft['kitchen'],
            }))
          }
          options={[
            ['', 'Any kitchen'],
            ['private', 'Private'],
            ['shared', 'Shared'],
            ['none', 'No kitchen'],
          ]}
        />

        <FilterSelect
          label="Parlour"
          value={draft.parlour}
          onChange={(value) =>
            setDraft((current) => ({
              ...current,
              parlour: value as FilterDraft['parlour'],
            }))
          }
          options={[
            ['', 'Any'],
            ['true', 'Has parlour'],
            ['false', 'No parlour'],
          ]}
        />

        <FilterSelect
          label="Availability"
          value={draft.availability}
          onChange={(value) =>
            setDraft((current) => ({
              ...current,
              availability: value as FilterDraft['availability'],
            }))
          }
          options={[
            ['available', 'Available offers only'],
            ['all', 'All unit types'],
          ]}
        />
      </div>

      <div className="border-border bg-background flex shrink-0 gap-3 border-t px-5 pt-3 pb-[max(1.25rem,env(safe-area-inset-bottom))]">
        <Button variant="secondary" className="flex-1" onClick={clear}>
          Clear
        </Button>
        <Button className="flex-1" type="submit">
          Apply filters
        </Button>
      </div>
    </form>
  )
}

type FilterSelectProps = {
  label: string
  value: string
  onChange: (value: string) => void
  options: readonly (readonly [string, string])[]
}

function FilterSelect({ label, value, onChange, options }: FilterSelectProps) {
  return (
    <label className="font-body block">
      <span className="section-label">{label}</span>
      <select
        value={value}
        className="focus-ring bg-surface text-foreground rounded-control min-h-12 w-full border border-transparent px-3.5 text-sm"
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map(([optionValue, optionLabel]) => (
          <option key={optionValue || 'any'} value={optionValue}>
            {optionLabel}
          </option>
        ))}
      </select>
    </label>
  )
}

function draftFromSearch(search: DiscoverySearch): FilterDraft {
  return {
    categories: search.category?.split(',') ?? [],
    area: search.area ?? '',
    minPrice: search.min_price?.toString() ?? '',
    maxPrice: search.max_price?.toString() ?? '',
    bathroom: search.bathroom_type ?? '',
    kitchen: search.kitchen_type ?? '',
    parlour: search.has_parlour === undefined ? '' : search.has_parlour ? 'true' : 'false',
    availability: search.availability,
  }
}

function inputPrice(value: string) {
  const parsed = Number(value)
  return Number.isSafeInteger(parsed) && parsed > 0 ? parsed : undefined
}
