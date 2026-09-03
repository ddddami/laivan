import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import type {
  AgentOfferDetail,
  AuthenticatedApiClient,
  PropertyDetail,
  UnitTypeDetail,
} from '../../api/client'
import { WorkflowActionArea } from './workflow-action-area'

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

describe('WorkflowActionArea', () => {
  it('shows an anonymous visitor a sign-in link with the current return path', async () => {
    const client = fakeWorkflowClient(false)
    renderActionArea(client)

    const link = await screen.findByRole('link', { name: 'Sign in to ask a question' })

    expect(link).toHaveAttribute('href', '/v1/auth/google/start?return_to=%2F')
  })

  it('opens the question sheet for an authenticated visitor', async () => {
    const client = fakeWorkflowClient(true)
    renderActionArea(client)

    fireEvent.click(await screen.findByRole('button', { name: 'Ask a question' }))

    expect(screen.getByRole('dialog')).toBeInTheDocument()
    expect(screen.getByRole('textbox', { name: 'What would you like to ask?' })).toBeInTheDocument()
  })

  it('does not offer a workflow action for an unavailable offer', async () => {
    const client = fakeWorkflowClient(true)
    renderActionArea(client, { ...offer, status: 'paused' })

    expect(await screen.findByText('This offer is not open for new questions.')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Ask a question' })).not.toBeInTheDocument()
    expect(
      screen.queryByRole('link', { name: 'Sign in to ask a question' }),
    ).not.toBeInTheDocument()
  })
})

function renderActionArea(client: AuthenticatedApiClient, selectedOffer = offer) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <WorkflowActionArea
        property={property}
        unit={unit}
        offer={selectedOffer}
        workflowClient={client}
      />
    </QueryClientProvider>,
  )
}

function fakeWorkflowClient(authenticated: boolean): AuthenticatedApiClient {
  return {
    getSession: vi.fn(async () => ({
      authenticated,
      user: authenticated
        ? {
            id: 'user-1',
            email: 'student@example.com',
            display_name: 'Student Example',
            status: 'active' as const,
          }
        : null,
      roles: [],
      agent: null,
      csrf_token: authenticated ? 'csrf-token' : null,
    })),
    createInquiry: vi.fn(async () => {
      throw new Error('not used in action area test')
    }),
  }
}
