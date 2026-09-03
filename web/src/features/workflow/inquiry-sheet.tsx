import { useState, type FormEvent } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import {
  ApiError,
  type AuthenticatedApiClient,
  type AgentOfferDetail,
  type PropertyDetail,
  type UnitTypeDetail,
} from '../../api/client'
import { publicQueryKeys, sessionQueryOptions } from '../../api/queries'
import { unitCategoryLabel } from '../../components/marketplace/labels'
import { Button } from '../../components/ui/button'
import { Price } from '../../components/ui/price'
import { Sheet } from '../../components/ui/sheet'

type InquirySheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  property: Pick<PropertyDetail, 'id' | 'name' | 'area' | 'landmark'>
  unit: Pick<UnitTypeDetail, 'id' | 'name' | 'category'>
  offer: AgentOfferDetail
  workflowClient: AuthenticatedApiClient
}

export function InquirySheet({
  open,
  onOpenChange,
  property,
  unit,
  offer,
  workflowClient,
}: InquirySheetProps) {
  const queryClient = useQueryClient()
  const sessionQuery = useQuery(sessionQueryOptions(workflowClient))
  const [message, setMessage] = useState('')
  const [submissionID, setSubmissionID] = useState(() => globalThis.crypto.randomUUID())
  const mutation = useMutation({
    mutationFn: async (question: string) => {
      const csrfToken = sessionQuery.data?.csrf_token
      if (!sessionQuery.data?.authenticated || typeof csrfToken !== 'string') {
        throw new ApiError({
          kind: 'response',
          status: 401,
          code: 'unauthenticated',
          message: 'Authentication is required',
        })
      }
      return workflowClient.createInquiry(
        offer.id,
        { message: question, submission_id: submissionID },
        csrfToken,
      )
    },
    retry: false,
    onError: (error) => {
      if (error instanceof ApiError && error.code === 'offer_unavailable') {
        void queryClient.invalidateQueries({ queryKey: publicQueryKeys.property(property.id) })
      }
    },
  })
  const unitName = unit.name || unitCategoryLabel(unit.category)
  const question = message.trim()
  const sessionReady =
    sessionQuery.data?.authenticated === true && typeof sessionQuery.data.csrf_token === 'string'
  const errorMessage = inquiryErrorMessage(mutation.error)

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      mutation.reset()
      setMessage('')
      setSubmissionID(globalThis.crypto.randomUUID())
    }
    onOpenChange(nextOpen)
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!question || mutation.isPending) return
    mutation.mutate(question)
  }

  return (
    <Sheet
      open={open}
      onOpenChange={handleOpenChange}
      title="Ask a question"
      description="Your question is recorded before you continue in WhatsApp."
    >
      <div className="min-h-0 flex-1 overflow-y-auto px-5 py-5">
        <section className="bg-surface rounded-card p-4" aria-labelledby="inquiry-context">
          <p className="font-body text-muted text-xs">Selected offer</p>
          <h2
            id="inquiry-context"
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

        {mutation.data ? (
          <section className="mt-6" aria-live="polite">
            <p className="section-label">Question recorded</p>
            <p className="font-body text-muted mt-2 text-sm leading-6">
              Your question is attached to this offer. Continue in WhatsApp when you are ready.
            </p>
            <a
              href={mutation.data.handoff.url}
              target="_blank"
              rel="noreferrer"
              className="focus-ring bg-accent text-foreground font-display rounded-control hover:bg-accent/80 mt-5 flex min-h-11 items-center justify-center px-4 py-2.5 text-sm font-bold transition-colors"
            >
              Continue in WhatsApp
            </a>
          </section>
        ) : (
          <form className="mt-6" onSubmit={handleSubmit}>
            <label htmlFor="inquiry-message" className="font-body block">
              <span className="text-muted mb-1.5 block text-xs font-medium">
                What would you like to ask?
              </span>
              <textarea
                id="inquiry-message"
                value={message}
                onChange={(event) => setMessage(event.target.value)}
                maxLength={1000}
                rows={5}
                aria-describedby="inquiry-message-hint"
                className="focus-ring bg-surface text-foreground placeholder:text-muted rounded-control w-full resize-y border border-transparent px-3.5 py-3 text-sm transition-colors"
                placeholder="Ask about facilities, terms, or timing."
              />
            </label>
            <div
              id="inquiry-message-hint"
              className="font-body text-muted mt-1.5 flex justify-between text-xs"
            >
              <span>
                Your question is recorded in Laivan. You choose whether to continue in WhatsApp.
              </span>
              <span className="shrink-0 pl-3">{message.length}/1,000</span>
            </div>
            {errorMessage ? (
              <p className="font-body text-destructive mt-4 text-sm" role="alert">
                {errorMessage}
              </p>
            ) : null}
            <Button
              className="mt-5 w-full"
              type="submit"
              disabled={!question || mutation.isPending || !sessionReady}
            >
              {mutation.isPending ? 'Recording question...' : 'Record question'}
            </Button>
          </form>
        )}
      </div>

      <div className="border-border bg-background border-t px-5 pt-3 pb-[max(1.25rem,env(safe-area-inset-bottom))]">
        <Button className="w-full" variant="secondary" onClick={() => handleOpenChange(false)}>
          Close
        </Button>
      </div>
    </Sheet>
  )
}

function inquiryErrorMessage(error: unknown) {
  if (!(error instanceof ApiError))
    return error ? 'We could not send your question. Check your connection and try again.' : null
  if (error.code === 'validation_failed') return error.fields?.message ?? error.message
  if (error.code === 'offer_unavailable')
    return 'This offer is no longer accepting questions. Refresh and choose another offer.'
  if (error.code === 'unauthenticated')
    return 'Your session has ended. Sign in again to send this question.'
  return 'We could not send your question. Check your connection and try again.'
}
