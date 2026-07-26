import { useState } from 'react'

import type { AgentOfferDetail, PropertyDetail, UnitTypeDetail } from '../../api/client'
import { unitCategoryLabel } from '../../components/marketplace/labels'
import { Button } from '../../components/ui/button'
import { Price } from '../../components/ui/price'
import { Sheet } from '../../components/ui/sheet'

type WorkflowKind = 'availability' | 'inspection' | 'question' | 'reservation' | 'save'

const workflows: ReadonlyArray<{
  value: WorkflowKind
  label: string
  description: string
  preview: string
}> = [
  {
    value: 'availability',
    label: 'Ask about availability',
    description: 'Request a current update for this offer.',
    preview:
      'Availability checks will be structured and recorded before agent coordination begins.',
  },
  {
    value: 'inspection',
    label: 'Request an inspection',
    description: 'Share that you want to view this unit.',
    preview:
      'Inspection requests will collect a preferred time and preserve their workflow status.',
  },
  {
    value: 'question',
    label: 'Ask a question',
    description: 'Ask about terms, facilities, or other details.',
    preview: 'Questions will keep this property, unit, and offer context attached to the request.',
  },
  {
    value: 'reservation',
    label: 'Express reservation interest',
    description: 'Indicate serious interest without claiming a booking.',
    preview:
      'Reservation interest will remain an intent, not a payment, booking, or availability guarantee.',
  },
  {
    value: 'save',
    label: 'Save property',
    description: 'Keep this property for later comparison.',
    preview:
      'Saved properties will become available after student identity and persistence are introduced.',
  },
]

type WorkflowPreviewSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  property: Pick<PropertyDetail, 'id' | 'name' | 'area' | 'landmark'>
  unit: Pick<UnitTypeDetail, 'id' | 'name' | 'category'>
  offer: AgentOfferDetail
  requestsAvailable?: boolean
}

export function WorkflowPreviewSheet({
  open,
  onOpenChange,
  property,
  unit,
  offer,
  requestsAvailable = true,
}: WorkflowPreviewSheetProps) {
  const [selected, setSelected] = useState<WorkflowKind | null>(null)
  const selectedWorkflow = workflows.find((workflow) => workflow.value === selected)
  const unitName = unit.name || unitCategoryLabel(unit.category)

  function handleOpenChange(nextOpen: boolean) {
    onOpenChange(nextOpen)
    if (!nextOpen) setSelected(null)
  }

  return (
    <Sheet
      open={open}
      onOpenChange={handleOpenChange}
      title="Student workflows"
      description="Preview only. Structured requests are coming next."
    >
      <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
        <section className="bg-surface rounded-card p-4" aria-labelledby="workflow-context">
          <p className="font-body text-muted text-xs">Selected offer</p>
          <h2
            id="workflow-context"
            className="font-display text-foreground mt-1 text-base font-bold tracking-[-0.02em]"
          >
            {property.name} · {unitName}
          </h2>
          <p className="font-body text-muted mt-1 text-xs">
            {[property.area, property.landmark].filter(Boolean).join(' · ')}
          </p>
          <div className="mt-3 flex items-end justify-between gap-3">
            <Price amount={offer.price_naira} />
            <p className="font-body text-muted text-right text-xs">
              via {offer.agent.display_name}
            </p>
          </div>
        </section>

        <fieldset className="mt-6">
          <legend className="section-label">What would you like to do?</legend>
          <div className="space-y-2">
            {workflows
              .filter((workflow) => requestsAvailable || workflow.value === 'save')
              .map((workflow) => (
                <button
                  key={workflow.value}
                  type="button"
                  aria-pressed={selected === workflow.value}
                  className={`focus-ring rounded-control w-full px-4 py-3.5 text-left transition-colors ${
                    selected === workflow.value
                      ? 'bg-foreground text-background'
                      : 'bg-surface text-foreground hover:bg-surface-strong'
                  }`}
                  onClick={() => setSelected(workflow.value)}
                >
                  <span className="font-body block text-sm font-semibold">{workflow.label}</span>
                  <span
                    className={`font-body mt-1 block text-xs leading-5 ${
                      selected === workflow.value ? 'text-background/65' : 'text-muted'
                    }`}
                  >
                    {workflow.description}
                  </span>
                </button>
              ))}
          </div>
        </fieldset>

        {selectedWorkflow ? (
          <section className="border-border mt-5 border-t pt-5" aria-live="polite">
            <p className="section-label">Coming next</p>
            <p className="font-body text-muted text-sm leading-6">{selectedWorkflow.preview}</p>
          </section>
        ) : null}

        <p className="font-body text-muted mt-5 text-xs leading-5">
          Nothing in this preview is submitted or saved. No inspection, reservation, availability
          confirmation, or agent conversation has started.
        </p>
      </div>

      <div className="border-border bg-background border-t px-5 pt-3 pb-[max(1.25rem,env(safe-area-inset-bottom))]">
        <Button className="w-full" onClick={() => handleOpenChange(false)}>
          Close preview
        </Button>
      </div>
    </Sheet>
  )
}
