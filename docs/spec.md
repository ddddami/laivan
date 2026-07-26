# Laivan — Product Specification and Implementation Plan

## Purpose

This document defines the product scope and the order in which it should be implemented.

Laivan is a coordination, trust, and workflow layer for the fragmented FUTA accommodation market. It should help students discover accommodation, compare real marketplace options, express structured intent, and continue flexible coordination through WhatsApp. It should help agents distribute offers and manage demand without forcing the market into unfamiliar behavior.

The product should reduce fragmentation, improve trust, and simplify coordination. It should not become a generic listing board, chat app, payment product, landlord tool, or property management system.

## Product Principles

These constraints apply to every implementation phase:

- Preserve the core marketplace model: `Property -> PropertyUnitType -> AgentOffer`.
- Treat Campus as the marketplace boundary. Do not introduce a University hierarchy yet.
- Treat duplicate inventory and multiple agent offers as normal market behavior.
- Structure intent before WhatsApp; do not attempt to replace WhatsApp.
- Represent trust through observable, probabilistic signals rather than absolute verification claims.
- Keep public location approximate by default.
- Prioritize mobile-first usability, poor-network resilience, marketplace liquidity, and operational simplicity.
- Keep the initial web product a React and Vite installable PWA. Defer SSR or ISR infrastructure until public search acquisition justifies it.

## Users and Responsibilities

### Student

Students should eventually be able to:

- browse, search, filter, and compare accommodation
- view property, unit type, agent offer, media, and trust information
- save properties
- submit general inquiries and availability checks
- request inspections
- express reservation intent
- continue coordination with an agent through contextual WhatsApp handoff
- leave property and agent reviews after eligible interactions

### Agent

Agents should eventually be able to:

- create and manage offers for property unit types
- upload and maintain offer media
- manage price, availability, and operational notes
- receive and respond to structured student intents
- coordinate inspection and reservation-intent states
- view lightweight offer performance information
- build reputation from observable marketplace behavior

### Admin

Admins should eventually be able to:

- moderate properties, offers, media, reviews, and reports
- group or merge probable duplicate properties
- suspend abusive accounts and remove spam
- manage manual trust and moderation signals
- curate public marketplace presentation where necessary

## Current Implementation Snapshot

This is a planning aid, not a substitute for acceptance criteria.

| Product area | Status | Current position |
| --- | --- | --- |
| Marketplace foundation and listing management API | Built | Core schema, repositories, endpoints, media, validation, and local infrastructure exist. |
| Public discovery API | Built | Discovery, marketplace search, filtering, sorting, pagination, and property detail data exist. |
| Public discovery UI | Built | Phase 1 provides accommodation-first discovery, shareable search and filters, property context, competing offers, responsive media, and honest workflow previews. |
| Identity and authorization | Planned | Authentication, sessions, roles, ownership, and protected writes are not implemented. |
| Inquiry, inspection, and reservation workflows | Planned | Domain types exist, but persistence and API workflows do not. |
| Agent operations UI | Planned | Offer management and request queues are not available in the frontend. |
| Trust and moderation | Planned | Concepts and domain types exist, but operational workflows do not. |
| Launch readiness | In progress | The PWA shell and local phone-preview path exist; production operations, monitoring, security hardening, and pilot readiness remain. |

## Delivery Strategy

Each phase should produce a coherent, testable product capability. Later phases may begin only when their required dependencies are stable. A phase is complete when its exit criteria are met, not merely when its screens or tables exist.

### Phase 0 — Re-establish the Marketplace Foundation

**Outcome:** The existing backend foundation is green, reproducible, and safe to build on.

**Scope:**

- confirm the `Property -> PropertyUnitType -> AgentOffer` schema and API behavior
- verify campus scoping, money handling, filters, pagination, and media delivery
- run the PostgreSQL integration suite against local infrastructure
- verify a real MinIO upload and imgproxy delivery path
- resolve known partial-success behavior in multi-file media persistence
- confirm OpenAPI output and generated frontend contract readiness
- keep configuration examples safe, complete, and startup-validated

**Exit criteria:**

- canonical backend checks and integration tests pass
- local setup reproduces discovery and media behavior
- API documentation matches implemented behavior
- no known data-integrity issue blocks public discovery work

### Phase 1 — Public Discovery

**Outcome:** A student can move from opening Laivan to understanding and comparing available accommodation.

**Dependencies:** Phase 0.

**Status:** Complete.

**Scope:**

- build a mobile-first discovery feed backed by `/v1/discovery`
- support validated marketplace search, price, category, area, availability, sorting, and pagination controls
- build a property detail page that preserves the domain hierarchy
- show property facts, unit types, competing agent offers, approximate location, pricing, media, and available trust indicators
- provide clear paths toward inquiry, inspection, saving, and reservation intent, even where later workflows are not yet enabled
- handle loading, empty, error, stale-data, and poor-network states
- keep desktop layouts intentionally designed rather than stretched from mobile
- optimize responsive images and baseline accessibility

**Discovery ranking should initially favor:**

- freshness
- filter relevance
- availability confidence
- listing completeness
- trustworthy observable signals when available

Do not introduce advanced recommendations, AI ranking, or separate search infrastructure during this phase.

**Exit criteria:**

- a student can browse, filter, open, and compare real seeded marketplace inventory on mobile and desktop
- the UI distinguishes property facts, unit types, and agent-specific offers
- public location remains approximate
- user-visible loading, empty, error, and retry behavior is verified

Phase 1 closes with discovery presented in the order students use it:
accommodation type first, property context second, and competing agent offers
third. Property, accommodation, and offer media remain attached to their source
layers, with unit media preferred over property fallback in discovery.

### Phase 2 — Identity, Ownership, and Protected Writes

**Outcome:** Public reads remain low-friction while mutations are attributable and permission-controlled.

**Dependencies:** Phase 0. Complete before public write workflows ship.

**Scope:**

- add student and agent identity persistence
- implement authentication and session handling
- enforce student, agent, and admin roles at route or middleware boundaries
- define ownership and permission rules for properties, offers, and media
- remove caller-supplied identity where authenticated context should be authoritative
- add rate limiting for suspicious public and authenticated behavior
- add upload abuse controls and safe media validation
- establish account suspension behavior and protected-route error responses

**Exit criteria:**

- every write is associated with an authenticated identity or an explicitly documented administrative path
- agents cannot mutate another agent's protected resources
- students cannot access agent or admin operations
- authentication, authorization, validation, and abuse paths have focused behavior tests

### Phase 3 — Structured Intent and WhatsApp Handoff

**Outcome:** Laivan captures what a student wants before flexible coordination continues through WhatsApp.

**Dependencies:** Phases 1 and 2.

**Workflow types:**

- **General inquiry:** the student wants more information
- **Availability check:** the student wants current availability confirmed
- **Inspection request:** the student wants to physically inspect the property
- **Reservation intent:** the student is seriously interested in securing the room

A reservation intent does not imply payment, a legal agreement, guaranteed availability, or a completed reservation.

**Scope:**

- persist each intent against the student, agent offer, and relevant property context
- capture only the fields needed for the selected workflow
- support preferred inspection time, phone number, and an optional message where relevant
- define explicit workflow statuses and valid state transitions
- preserve workflow history needed for operations and future trust signals
- generate a contextual WhatsApp handoff containing the property, unit type, offer, and expressed intent
- avoid storing or recreating the subsequent WhatsApp conversation
- give students clear submission, failure, duplicate-submission, and next-step feedback

**Exit criteria:**

- a signed-in student can submit every supported intent from a real offer
- the intent is persisted before WhatsApp opens
- the receiving agent and offer are unambiguous
- retries do not create uncontrolled duplicate workflows
- copy does not imply guaranteed availability, reservation, payment, or verification

### Phase 4 — Lightweight Agent Operations

**Outcome:** Agents can keep offers current and act on structured demand without an enterprise-style dashboard.

**Dependencies:** Phases 2 and 3.

**Scope:**

- provide an agent overview of active and inactive offers
- allow agents to manage price, availability, operational notes, and media
- provide queues for inquiries, availability checks, inspection requests, and reservation intents
- support the minimum response and status actions defined by each workflow
- surface stale offers and encourage availability updates
- show lightweight performance information such as views, intents, response behavior, and successful workflow outcomes
- keep structural property facts separate from agent-specific offer information

**Exit criteria:**

- an agent can maintain an offer without administrative assistance
- an agent can identify and act on pending student intents
- availability and workflow status changes are visible to the relevant student experience
- the dashboard remains usable on a phone and does not introduce chat, CRM, or property-management scope

### Phase 5 — Trust, Reviews, Duplicates, and Moderation

**Outcome:** The marketplace becomes more understandable and safer through explainable signals and practical admin tools.

**Dependencies:** Phases 2 through 4, because trust should be based on real behavior.

**Initial trust signals may include:**

- recently updated
- listing completeness
- agent response reliability
- successful inspection history
- media consistency
- repeated offer consistency
- account and activity history
- reports and moderation outcomes

**Scope:**

- expose only explainable, evidence-backed trust indicators
- avoid a single opaque score until sufficient marketplace behavior exists
- allow eligible students to review properties and agents
- establish review eligibility, reporting, and abuse controls
- add content and account reporting
- give admins queues for reports, abusive content, suspicious offers, and stale inventory
- support manual duplicate grouping and carefully controlled property merges
- preserve distinct agent offers when grouping the same underlying property
- record moderation actions and their reasons
- add saved properties if evidence shows they improve comparison and return usage

Automated duplicate detection should begin with soft suggestions and lightweight heuristics. Sophisticated media matching or AI canonicalization is not required.

**Exit criteria:**

- every public trust claim maps to a known signal
- admins can resolve common marketplace abuse and duplicate cases
- reviews cannot be submitted without the defined eligibility signal
- merging duplicate properties does not collapse or transfer agent ownership incorrectly
- moderation actions are attributable and reviewable

### Phase 6 — Pilot and Launch Hardening

**Outcome:** Laivan is safe, observable, installable, and operationally ready for a small FUTA pilot.

**Dependencies:** The launch-critical parts of Phases 1 through 5.

**Scope:**

- complete and verify the installable PWA experience on iOS and Android
- test core workflows on low-end devices and poor connections
- configure production PostgreSQL, object storage, media delivery, and backups
- add monitoring, alerting, structured operational logs, and recovery procedures
- validate security headers, rate limits, upload controls, and privacy behavior
- define moderation ownership and response procedures
- seed enough current inventory for useful discovery
- run a small pilot with real students and agents
- measure discovery success, intent conversion, response reliability, stale inventory, and repeat usage

**Exit criteria:**

- core student and agent workflows pass production smoke tests
- backups and recovery procedures are verified
- operational owners can identify and respond to failures and abuse
- the marketplace has enough current supply to make discovery useful
- pilot feedback can be traced to concrete product and operational changes

## Cross-Phase Requirements

### Media

- support images first; defer video until there is demonstrated need
- compress and resize uploads for mobile delivery
- moderate abusive uploads
- preserve media provenance across property, unit type, and agent offer contexts
- avoid partial database persistence for a logically atomic upload operation
- treat automated duplicate-media detection as a later enhancement

### Location and Privacy

- expose area, landmarks, district references, or relative positioning publicly
- do not expose exact public coordinates or directions by default
- share more precise directions only when the workflow and permissions justify it
- collect and retain only the personal information needed for the active workflow

### Performance and Resilience

- remain fast and usable on low-end phones and poor networks
- optimize images aggressively
- keep request and payload sizes proportionate
- make retry and stale-data behavior understandable
- avoid making native app distribution a launch dependency

### Security and Abuse Prevention

- validate request bodies, route parameters, query parameters, uploads, and configuration
- reject unknown JSON fields at API boundaries
- protect personal data and avoid leaking internal errors
- rate limit suspicious behavior
- make authentication and permission requirements explicit
- sanitize and moderate user-controlled content
- keep audit history for consequential administrative actions

### Rendering and SEO

During the initial phases:

- use clean URLs, semantic HTML, descriptive metadata, structured content, and strong mobile performance
- keep the React and Vite application client-rendered
- do not add SSR, ISR, or TanStack Start solely because listing pages are public

After marketplace workflows are validated and organic search becomes a meaningful acquisition channel:

- evaluate campus, area, category, and property pages for prerendering or cached server rendering
- migrate selectively rather than converting operational dashboards unnecessarily

## Explicitly Out of MVP Scope

- in-app chat
- payments or escrow
- guaranteed reservations
- guaranteed listing availability
- absolute property or agent verification claims
- advanced AI ranking or recommendation systems
- fully automated duplicate detection
- video-first media workflows
- landlord property-management tooling
- cross-campus discovery
- native mobile applications
- microservices or premature event-driven infrastructure

## Post-MVP Decision Gates

Future work should be triggered by evidence rather than added automatically:

- **SEO rendering:** when indexed public discovery becomes a material acquisition channel
- **Cross-campus support:** when the FUTA marketplace works and another campus has an operational launch path
- **Native mobile:** when PWA limitations block validated user behavior
- **Payments:** when trust, inspection, and reservation workflows demonstrate a safe and valuable transaction boundary
- **Advanced deduplication:** when manual moderation volume becomes a measurable bottleneck
- **Recommendations:** when inventory and interaction data are sufficient to outperform explicit filters
- **Agent monetization:** when Laivan already creates measurable agent value without compromising trust
