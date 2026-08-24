import type { Media } from '../../api/client'

export type AccommodationMediaSource = 'unit_type' | 'property' | 'none'

export type AccommodationMediaSelection = {
  items: readonly Media[]
  source: AccommodationMediaSource
}

export function selectAccommodationMedia(
  unitMedia: readonly Media[],
  propertyMedia: readonly Media[],
): AccommodationMediaSelection {
  const unitImages = imageMedia(unitMedia, 'unit_type')
  if (unitImages.length > 0) {
    return { items: unitImages, source: 'unit_type' }
  }

  const propertyImages = imageMedia(propertyMedia, 'property')
  if (propertyImages.length > 0) {
    return { items: propertyImages, source: 'property' }
  }

  return { items: [], source: 'none' }
}

function imageMedia(media: readonly Media[], source: 'unit_type' | 'property'): Media[] {
  return media.filter((item) => {
    if (item.kind !== 'image') return false

    if (source === 'unit_type') {
      return (
        item.property_unit_type_id != null &&
        item.property_id == null &&
        item.agent_offer_id == null
      )
    }

    return (
      item.property_id != null && item.property_unit_type_id == null && item.agent_offer_id == null
    )
  })
}
