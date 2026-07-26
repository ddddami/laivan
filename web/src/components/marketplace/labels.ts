import type { components } from '../../api/schema.gen'

type UnitCategory = components['schemas']['UnitCategory']
type BathroomType = components['schemas']['BathroomType']
type KitchenType = components['schemas']['KitchenType']

const categoryLabels: Record<UnitCategory, string> = {
  single_room: 'Single room',
  self_contained: 'Self-contained',
  room_and_parlour: 'Room and parlour',
  one_bedroom_flat: 'One-bedroom flat',
  two_bedroom_flat: 'Two-bedroom flat',
  three_bedroom_flat: 'Three-bedroom flat',
  other: 'Other',
}

export function unitCategoryLabel(category: UnitCategory) {
  return categoryLabels[category]
}

export function bathroomLabel(value: BathroomType | null | undefined) {
  return value && value !== 'unknown' ? `${sentenceCase(value)} bathroom` : 'Bathroom unknown'
}

export function kitchenLabel(value: KitchenType | null | undefined) {
  if (!value || value === 'unknown') return 'Kitchen unknown'
  if (value === 'none') return 'No kitchen'
  return `${sentenceCase(value)} kitchen`
}

function sentenceCase(value: string) {
  return value.charAt(0).toUpperCase() + value.slice(1).replaceAll('_', ' ')
}
