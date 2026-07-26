import { useState } from 'react'

import type { AgentOfferDetail, PropertyDetail, UnitTypeDetail } from '../../api/client'
import { Button } from '../ui/button'
import { WorkflowPreviewSheet } from '../../features/workflow/workflow-preview-sheet'

type WorkflowActionAreaProps = {
  property: Pick<PropertyDetail, 'id' | 'name' | 'area' | 'landmark'>
  unit: Pick<UnitTypeDetail, 'id' | 'name' | 'category'>
  offer: AgentOfferDetail
}

export function WorkflowActionArea({ property, unit, offer }: WorkflowActionAreaProps) {
  const [open, setOpen] = useState(false)
  const requestsAvailable = offer.status === 'available'

  return (
    <div className="border-border mt-4 border-t pt-4">
      <Button
        variant={requestsAvailable ? 'accent' : 'secondary'}
        className="w-full"
        onClick={() => setOpen(true)}
      >
        {requestsAvailable ? 'Explore request options' : 'Preview save option'}
      </Button>
      <p className="font-body text-muted mt-2 text-center text-xs">
        {requestsAvailable
          ? 'Preview only. Nothing will be sent.'
          : 'This offer is not open for requests.'}
      </p>
      <WorkflowPreviewSheet
        open={open}
        onOpenChange={setOpen}
        property={property}
        unit={unit}
        offer={offer}
        requestsAvailable={requestsAvailable}
      />
    </div>
  )
}
