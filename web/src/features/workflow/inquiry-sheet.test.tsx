import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import {
  ApiError,
  type AgentOfferDetail,
  type AuthenticatedApiClient,
  type InquiryResponse,
  type PropertyDetail,
  type UnitTypeDetail,
} from '../../api/client'
import { InquirySheet } from './inquiry-sheet'

const property: Pick<PropertyDetail, 'id' | 'name' | 'area' | 'landmark'> = {
  id: 'property-1',
  name: 'Alice Lodge',
  area: 'Obanla',
  landmark: 'Near South Gate',
}

const unit: Pick<UnitTypeDetail, 'id' | 'name' | 'category'> = {
  id: 'unit-1',
  name: 'Premium self-contained',
  category: 'self_contained',
}

const offer: AgentOfferDetail = {
  id: 'offer-1',
  property_unit_type_id: 'unit-1',
  agent_id: 'agent-1',
  title: 'Fresh self-contained room',
  description: 'Private room with its own facilities.',
  notes: 'Inspection on weekdays.',
  price_naira: 350_000,
  status: 'available',
  version: 1,
  archived_at: null,
  agent: { id: 'agent-1', display_name: 'Bisi Housing Connect' },
  media: [],
  created_at: '2026-05-01T10:00:00Z',
  updated_at: '2026-05-02T10:00:00Z',
}

const inquiryResponse: InquiryResponse = {
  inquiry: {
    id: 'inquiry-1',
    agent_offer_id: offer.id,
    message: 'Is the kitchen private?',
    status: 'open',
    created_at: '2026-09-03T14:00:00Z',
  },
  handoff: {
    channel: 'whatsapp',
    url: 'https://wa.me/2348012345678?text=question',
  },
}

describe('InquirySheet', () => {
  it('requires a message and submits a trimmed question with CSRF and stable ID', async () => {
    const createInquiry = vi.fn<AuthenticatedApiClient['createInquiry']>(
      async () => inquiryResponse,
    )
    const client = fakeWorkflowClient(createInquiry)
    renderSheet(client)

    const submit = screen.getByRole('button', { name: /record question/i })
    expect(submit).toBeDisabled()

    fireEvent.change(screen.getByRole('textbox', { name: /what would you like to ask/i }), {
      target: { value: '  Is the kitchen private?  ' },
    })
    await waitFor(() => expect(submit).toBeEnabled())

    fireEvent.click(submit)
    expect(await screen.findByRole('link', { name: /continue in whatsapp/i })).toHaveAttribute(
      'href',
      inquiryResponse.handoff.url,
    )

    expect(createInquiry).toHaveBeenCalledTimes(1)
    const [offerID, input, csrfToken] = createInquiry.mock.calls[0]!
    expect(offerID).toBe(offer.id)
    expect(input.message).toBe('Is the kitchen private?')
    expect(input.submission_id).toMatch(/^[0-9a-f-]{36}$/)
    expect(csrfToken).toBe('csrf-token')
  })

  it('disables submission while pending', async () => {
    let resolveRequest: (response: InquiryResponse) => void = () => undefined
    const request = new Promise<InquiryResponse>((resolve) => {
      resolveRequest = resolve
    })
    const createInquiry = vi.fn<AuthenticatedApiClient['createInquiry']>(() => request)
    renderSheet(fakeWorkflowClient(createInquiry))

    fireEvent.change(screen.getByRole('textbox', { name: /what would you like to ask/i }), {
      target: { value: 'When can I inspect it?' },
    })
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /record question/i })).toBeEnabled(),
    )
    fireEvent.click(screen.getByRole('button', { name: /record question/i }))

    await waitFor(() =>
      expect(screen.getByRole('button', { name: /recording question/i })).toBeDisabled(),
    )
    resolveRequest(inquiryResponse)
  })

  it('preserves the question and submission ID for a retry', async () => {
    const createInquiry = vi
      .fn<AuthenticatedApiClient['createInquiry']>()
      .mockRejectedValueOnce(
        new ApiError({
          kind: 'network',
          code: 'network_error',
          message: 'Unable to reach Laivan',
        }),
      )
      .mockResolvedValueOnce(inquiryResponse)
    renderSheet(fakeWorkflowClient(createInquiry))

    const input = screen.getByRole('textbox', { name: /what would you like to ask/i })
    fireEvent.change(input, { target: { value: 'When can I inspect it?' } })
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /record question/i })).toBeEnabled(),
    )
    fireEvent.click(screen.getByRole('button', { name: /record question/i }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'We could not send your question. Check your connection and try again.',
    )
    fireEvent.click(screen.getByRole('button', { name: /record question/i }))
    await screen.findByRole('link', { name: /continue in whatsapp/i })

    expect(input).toHaveValue('When can I inspect it?')
    expect(createInquiry.mock.calls[0]![1].submission_id).toBe(
      createInquiry.mock.calls[1]![1].submission_id,
    )
  })

  it('shows factual offer-unavailable feedback and invalidates the property query', async () => {
    const createInquiry = vi.fn<AuthenticatedApiClient['createInquiry']>(async () => {
      throw new ApiError({
        kind: 'response',
        code: 'offer_unavailable',
        status: 409,
        message: 'This offer is no longer accepting questions',
      })
    })
    const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')
    renderSheet(fakeWorkflowClient(createInquiry), queryClient)

    fireEvent.change(screen.getByRole('textbox', { name: /what would you like to ask/i }), {
      target: { value: 'Is this still available?' },
    })
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /record question/i })).toBeEnabled(),
    )
    fireEvent.click(screen.getByRole('button', { name: /record question/i }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'This offer is no longer accepting questions. Refresh and choose another offer.',
      )
    })
    expect(invalidateQueries).toHaveBeenCalledWith({
      queryKey: ['public', 'property', property.id],
    })
  })
})

function renderSheet(
  client: AuthenticatedApiClient,
  queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } }),
) {
  return render(
    <QueryClientProvider client={queryClient}>
      <InquirySheet
        open
        onOpenChange={vi.fn()}
        property={property}
        unit={unit}
        offer={offer}
        workflowClient={client}
      />
    </QueryClientProvider>,
  )
}

function fakeWorkflowClient(
  createInquiry: AuthenticatedApiClient['createInquiry'],
): AuthenticatedApiClient {
  return {
    getSession: vi.fn(async () => ({
      authenticated: true,
      user: {
        id: 'user-1',
        email: 'student@example.com',
        display_name: 'Student Example',
        status: 'active' as const,
      },
      roles: [],
      agent: null,
      csrf_token: 'csrf-token',
    })),
    createInquiry,
  }
}
