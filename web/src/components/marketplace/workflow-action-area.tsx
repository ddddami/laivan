import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'

import {
  type AgentOfferDetail,
  type AuthenticatedApiClient,
  type PropertyDetail,
  type UnitTypeDetail,
} from '../../api/client'
import { authenticatedApiClient, sessionQueryOptions } from '../../api/queries'
import { InquirySheet } from '../../features/workflow/inquiry-sheet'
import { Button } from '../ui/button'

type WorkflowActionAreaProps = {
  property: Pick<PropertyDetail, 'id' | 'name' | 'area' | 'landmark'>
  unit: Pick<UnitTypeDetail, 'id' | 'name' | 'category'>
  offer: AgentOfferDetail
  workflowClient?: AuthenticatedApiClient
}

export function WorkflowActionArea({
  property,
  unit,
  offer,
  workflowClient = authenticatedApiClient,
}: WorkflowActionAreaProps) {
  const [open, setOpen] = useState(false)
  const requestsAvailable = offer.status === 'available'
  const sessionQuery = useQuery(sessionQueryOptions(workflowClient))
  const authenticated = sessionQuery.data?.authenticated === true

  return (
    <div className="border-border mt-4 border-t pt-4">
      {requestsAvailable && authenticated ? (
        <Button variant="accent" className="w-full" onClick={() => setOpen(true)}>
          Ask a question
        </Button>
      ) : requestsAvailable ? (
        <a
          href={signInHref()}
          className="focus-ring bg-accent text-foreground font-display rounded-control hover:bg-accent/80 flex min-h-11 w-full items-center justify-center px-4 py-2.5 text-sm font-bold transition-colors"
        >
          Sign in to ask a question
        </a>
      ) : null}
      <p className="font-body text-muted mt-2 text-center text-xs">
        {requestsAvailable
          ? 'Your question is recorded before you continue in WhatsApp.'
          : 'This offer is not open for new questions.'}
      </p>
      {requestsAvailable && authenticated ? (
        <InquirySheet
          open={open}
          onOpenChange={setOpen}
          property={property}
          unit={unit}
          offer={offer}
          workflowClient={workflowClient}
        />
      ) : null}
    </div>
  )
}

function signInHref() {
  const returnTo = `${window.location.pathname}${window.location.search}`
  return `/v1/auth/google/start?${new URLSearchParams({ return_to: returnTo })}`
}
