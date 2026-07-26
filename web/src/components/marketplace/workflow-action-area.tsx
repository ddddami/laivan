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

  if (offer.status !== 'available') {
    return (
      <p className="font-body text-faint border-border mt-4 border-t pt-4 text-xs leading-5">
        This offer is not currently open for student requests.
      </p>
    )
  }

  return (
    <div className="border-border mt-4 border-t pt-4">
      <Button variant="accent" className="w-full" onClick={() => setOpen(true)}>
        Explore request options
      </Button>
      <p className="font-body text-faint mt-2 text-center text-xs">
        Preview only. Nothing will be sent.
      </p>
      <WorkflowPreviewSheet
        open={open}
        onOpenChange={setOpen}
        property={property}
        unit={unit}
        offer={offer}
      />
    </div>
  )
}
