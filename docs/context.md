# Context

## What this is

Context doc for anyone working on Laivan. Written because too many housing platforms get built without understanding how the FUTA housing market actually works.

## What Laivan isn't

- Airbnb clone
- Property management SaaS
- Landlord dashboard
- Generic classifieds

## What Laivan actually is

A layer that makes the chaos of off-platform housing activity slightly less chaotic. The platform works if it reduces friction, increases trust, creates structure around intent, and helps both students and agents operate better. It fails if it tries to replace WhatsApp, forces everything on-platform, or adds unnecessary overhead.

## How the FUTA housing market works

- Listings get duplicated constantly
- Agents share inventory informally (ownership is ambiguous)
- WhatsApp is where everything happens
- Students don't trust anyone
- Landlords don't care about software
- Speed > perfection
- Verification is social, not institutional

Build around these realities, not against them.

## Product principles

### WhatsApp stays

Don't build features assuming users will communicate inside the app. WhatsApp remains the negotiation layer, the trust layer, the coordination layer. Laivan structures intent *before* WhatsApp.

The app should:
- Collect context
- Organize requests
- Preserve state
- Reduce chaos
- Provide structured entry points

The app should NOT:
- Try to become a chat app
- Force messaging lock-in
- Block external communication

### Workflows > generic "Contact Agent"

Students want one of three things:
1. Ask questions
2. Inspect a room
3. Secure a room

These are distinct workflows. Design for:
- Inquiry flow
- Inspection request flow
- Reservation intent flow

This creates structured operational data instead of dumping everything into unstructured chat.

## Trust system

Can't physically verify everything, so trust emerges from signals:

- Responsiveness
- Consistency
- Successful inspections
- Completed reservations
- Media quality
- Listing completeness
- Repeat interactions
- Historical activity
- Account age
- Reports
- Duplicate consistency
- Behavioral patterns

Verification is probabilistic, not absolute.

## Duplicate listings

Expect duplicates. Multiple agents may represent the same property, upload the same room with different titles, different media, different prices.

The system should:
- Detect probable duplicates
- Group them internally
- Merge public presentation carefully
- Avoid punishing agents aggressively

Goal: reduce marketplace confusion, not achieve perfect deduplication.

## UX direction

Users are busy, distracted, impatient, mobile-first, bandwidth-constrained.

Keep it:
- Fast
- Obvious
- Low friction
- Familiar
- Lightweight

Avoid:
- Excessive onboarding
- Dashboard complexity
- Enterprise workflows
- Too many required fields
- Forcing perfect data

## Agent philosophy

Agents aren't enemies. Platform dies if agents feel threatened.

Design to:
- Help agents earn more
- Help agents operate faster
- Help agents appear more trustworthy
- Reduce repetitive explanation work
- Reduce chaos

Never: "We're replacing agents"
Always: "We make your hustle cleaner and more efficient"

## Technical approach

Prioritize:
- Operational simplicity
- Maintainability
- Observability
- Explicit architecture
- Modularity
- Boring technology

Avoid:
- Unnecessary microservices
- Premature event-driven systems
- Over-abstraction
- Trendy complexity
- Infrastructure obsession

First goal is product-market fit, not architecture flexing.
