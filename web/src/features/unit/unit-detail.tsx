import { Link } from '@tanstack/react-router'

import type { AuthenticatedApiClient, PropertyDetail, UnitTypeDetail } from '../../api/client'
import { AgentOfferCard } from '../../components/marketplace/agent-offer-card'
import { unitCategoryLabel } from '../../components/marketplace/labels'
import { MediaGallery } from '../../components/marketplace/media-gallery'
import { selectAccommodationMedia } from '../../components/marketplace/media-policy'
import { UnitStructure } from '../../components/marketplace/unit-structure'
import { navigationEntryPoint } from '../discovery/navigation-state'

type UnitDetailProps = {
  property: PropertyDetail
  unit: UnitTypeDetail
  workflowClient?: AuthenticatedApiClient
}

export function UnitDetail({ property, unit, workflowClient }: UnitDetailProps) {
  const category = unitCategoryLabel(unit.category)
  const mediaSelection = selectAccommodationMedia(unit.media, property.media)
  const offers = [...unit.agent_offers].sort((first, second) => {
    const statusDifference = offerStatusOrder(first.status) - offerStatusOrder(second.status)
    if (statusDifference !== 0) return statusDifference
    if (first.price_naira !== second.price_naira) return first.price_naira - second.price_naira
    return first.agent.display_name.localeCompare(second.agent.display_name)
  })
  const availableOfferCount = offers.filter((offer) => offer.status === 'available').length

  return (
    <article>
      <header className="max-w-reading">
        <p className="section-label">{category}</p>
        <h1 className="font-display text-foreground text-3xl leading-none font-bold tracking-[-0.045em] sm:text-4xl">
          {unit.name || category}
        </h1>
        <p className="font-body text-muted mt-2 text-sm">
          <Link
            to="/properties/$propertyId"
            params={{ propertyId: property.id }}
            state={(current) => ({ ...current, entryPoint: navigationEntryPoint.property })}
            className="focus-ring text-foreground rounded-control -my-3 inline-flex min-h-11 items-center font-semibold underline-offset-4 hover:underline"
          >
            {property.name}
          </Link>
          {' · '}
          {[property.area, property.landmark].filter(Boolean).join(' · ')}
        </p>
      </header>

      <div className="mt-6 grid gap-8 lg:grid-cols-[minmax(20rem,0.8fr)_minmax(0,1.2fr)] lg:items-start lg:gap-10">
        <div className="space-y-5 lg:sticky lg:top-5">
          <MediaGallery
            media={mediaSelection.items}
            emptyLabel={`No unit photos available for ${unit.name || category}`}
            ariaLabel={`${unit.name || category} ${mediaSelection.source === 'property' ? 'property context' : 'unit'} photos`}
            altFallback={`${unit.name || category} unit`}
            heroSizes="(max-width: 1023px) 100vw, 40vw"
            thumbnailSizes="160px"
            notice={
              mediaSelection.source === 'property'
                ? {
                    label: 'Property context photos',
                    description: `These show ${property.name}; unit-specific photos are not available yet.`,
                  }
                : undefined
            }
          />
          <UnitStructure unit={unit} />

          {unit.description || unit.notes ? (
            <section className="border-border border-t pt-5" aria-labelledby="unit-details-heading">
              <h2
                id="unit-details-heading"
                className="font-display text-foreground text-base font-bold"
              >
                Unit details
              </h2>
              {unit.description ? (
                <p className="font-body text-muted mt-2 text-sm leading-6">{unit.description}</p>
              ) : null}
              {unit.notes ? (
                <div className="bg-surface rounded-control mt-3 px-3.5 py-3">
                  <p className="font-body text-muted text-xs font-semibold tracking-[0.08em] uppercase">
                    Structural notes
                  </p>
                  <p className="font-body text-muted mt-1.5 text-sm leading-5">{unit.notes}</p>
                </div>
              ) : null}
            </section>
          ) : null}
        </div>

        <section aria-labelledby="agent-offers-heading">
          <div className="mb-4">
            <p className="section-label">Compare market options</p>
            <div className="flex items-end justify-between gap-4">
              <h2
                id="agent-offers-heading"
                className="font-display text-foreground text-2xl font-bold tracking-[-0.035em]"
              >
                Agent offers
              </h2>
              <p className="text-muted font-mono text-xs">{availableOfferCount} available</p>
            </div>
            <p className="font-body text-muted mt-2 text-sm leading-6">
              These are independent offers for the same unit type. Compare their supported price and
              operational details.
            </p>
          </div>

          {offers.length > 0 ? (
            <div className="bg-surface rounded-sheet space-y-2 p-2">
              {offers.map((offer) => (
                <AgentOfferCard
                  key={offer.id}
                  offer={offer}
                  property={property}
                  unit={unit}
                  workflowClient={workflowClient}
                />
              ))}
            </div>
          ) : (
            <div className="bg-surface rounded-card px-5 py-9 text-center">
              <p className="font-display text-foreground text-base font-bold">
                No agent offers yet
              </p>
              <p className="font-body text-muted mt-2 text-sm">
                Offers will appear here when an agent represents this unit type.
              </p>
            </div>
          )}
        </section>
      </div>
    </article>
  )
}

function offerStatusOrder(status: UnitTypeDetail['agent_offers'][number]['status']) {
  if (status === 'available') return 0
  if (status === 'paused') return 1
  return 2
}
