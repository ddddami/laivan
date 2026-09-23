# Laivan — Architecture Decisions

## What this is

Notes on why I made certain technical and product decisions. Mostly to avoid forgetting the tradeoffs later, prevent architectural drift, and stop myself from rewriting things for no reason.

A lot of the complexity in Laivan isn't infrastructure complexity. It's marketplace complexity. That distinction matters.

---

# Repository Structure

**Decision:** Flat structure

```
repo/
  api/
  web/
  packages/
  openapi/
```

**Why:** The backend is the primary operational system. Didn't want the repo to feel like "a frontend monorepo with Go inside it." This treats backend, frontend, and shared tooling as equal systems.

---

# Backend

**Decision:** Go

**Why:** Fits the kind of system this actually is. Need operational simplicity, concurrency, predictable deployments, straightforward infrastructure. Want the backend to be explicit, boring, maintainable, observable, pragmatic.

---

# Router

**Decision:** chi

**Why:** Lightweight routing, strong middleware support, composability, minimal abstraction, idiomatic Go patterns. Don't need heavyweight framework behavior or deeply opinionated architecture. Backend should stay understandable.

---

# Database

**Decision:** PostgreSQL

**Why:** The domain is heavily relational.

---

# Query Strategy

**Decision:** sqlc

**Why:** Want to stay SQL-first, not ORM-first. Marketplace logic becomes nuanced over time. Don't want to hide query behavior behind excessive ORM abstraction.

---

# API Contract Strategy

**Decision:** OpenAPI. Generate TypeScript clients/types from the backend contract.

**Why:** Frontend and backend will evolve independently. OpenAPI helps maintain shared contracts, type safety, endpoint discoverability, API discipline, frontend/backend synchronization. Especially useful because React Query, generated clients, and future mobile clients all benefit from consistent contracts.

**Constraint:** OpenAPI exists to support development, not dominate it. Avoid over-generated architectures and excessive codegen complexity. Backend remains the source of truth.

---

# Frontend Architecture

**Initial decision:** Start with React + Vite + TanStack Router + TanStack Query + installable PWA support without introducing SSR infrastructure immediately.

**Why:** Laivan is primarily interaction-heavy, workflow-heavy, dashboard-heavy, mobile-first, authenticated. Most critical flows are client-side operational flows: inquiries, inspections, reservation intents, uploads, dashboard activity, listing management. Behaves more like an operational platform than a content website.

The first web product should feel like a strong mobile app experience on iOS and Android while still being excellent on desktop web. Most early users are phone-first and many will not have practical desktop access, so mobile is the starting point for prioritization, interaction design, and performance. Desktop must still be deliberately designed, not treated as stretched mobile UI. The PWA path keeps iteration fast and avoids App Store and Play Store review cycles before product validation.

---

# Why TanStack Router

**Decision:** Use TanStack Router from the beginning.

**Why:** Provides type-safe routing, route organization, loader patterns, search param validation, strong React Query integration, future compatibility with TanStack Start without forcing SSR complexity early. Preserves future flexibility.

---

# Why NOT TanStack Start Initially

At the beginning, need shipping speed, reduced complexity, operational simplicity, lower frontend infrastructure burden more than advanced rendering infrastructure.

The early-stage product problem is marketplace execution, trust, liquidity, workflow structure — not advanced rendering optimization.

Introducing SSR too early increases deployment complexity, caching complexity, hydration issues, service worker complexity, operational burden before product-market fit exists.

---

# Planned Frontend Evolution

**Expected path:**

**Phase 1:** React + Vite + TanStack Router SPA + mobile-first installable PWA

**Phase 2:** Migrate to TanStack Start when SEO becomes important, public discovery matters significantly, listing indexing becomes growth-critical

**Phase 3:** Use selective rendering strategies (SSG, SSR, ISR-like caching) only where useful

---

# Rendering Philosophy

Most of the application should remain client-side even after migrating to TanStack Start.

Example future structure:

```
/dashboard              -> client-heavy SPA
/agent                  -> SPA
/messages               -> SPA
/                        -> prerendered
/accommodation/futa     -> SSG
/area/south-gate        -> SSG
/listing/alice-lodge    -> SSR or cached rendering
```

This matches the actual shape of the product.

---

# SEO Philosophy

**Decision:** Don't prematurely optimize rendering infrastructure.

**Why:** Early SEO wins will come more from clean URLs, semantic pages, mobile performance, listing freshness, structured content, location pages, fast UX than from sophisticated SSR infrastructure.

Not competing against world-class SEO systems initially. Competing against weak local listing sites, WhatsApp groups, Facebook posts, low-quality classifieds. Architecture should reflect that reality.

---

# State Management

**Decision:** TanStack Query

**Why:** Most frontend state is actually server state, cached data, async workflows, mutations, freshness coordination. TanStack Query handles caching, background refresh, optimistic updates, mutation coordination, stale handling without unnecessary custom state infrastructure. Don't want to turn server state into global app state.

---

# Communication Architecture

**Decision:** Don't build a full internal chat system initially. Use WhatsApp-assisted workflows.

**Why:** The actual market already runs on WhatsApp. Trying to replace WhatsApp introduces adoption friction, notification complexity, moderation complexity, operational overhead, behavior mismatch. Laivan should structure intent before WhatsApp, not replace WhatsApp itself.

---

# Identity And Sessions

**Decision:** Use Google OpenID Connect for the initial sign-in flow and opaque, revocable browser sessions stored server-side.

The auth service depends on a provider interface for authorization URLs, code exchange, and verified identity claims. The Google OIDC adapter is wired in `cmd/api`; the callback validates the provider-issued ID token, verified email, nonce, and PKCE exchange before upserting the application user. PostgreSQL stores provider/subject identities separately from users, and stores SHA-256 hashes of session and CSRF tokens, never usable browser tokens.

The session endpoint is anonymous-safe and returns effective identity state without making public browsing depend on authentication. Logout requires the session, CSRF token, and configured same-origin request.

Agent linkage, roles, and protected marketplace writes are separate delivery slices. The strict agent-identity slice is explicit: a Google identity never claims an existing or seeded agent automatically, and every active agent has a non-null linked user.

The first ownership slice stores non-null `agents.user_id` and agent lifecycle status, explicit global-admin and campus-operator assignments, campus-scoped agent applications, explicit `agent_campuses` associations, and audit events. Agent-campus scope is independent of current offers: migration backfills legacy associations through `Property -> PropertyUnitType -> AgentOffer`, and activation idempotently associates a newly created agent with the application campus. Application decisions live in a small repository workflow so operator scope, agent creation, application state changes, and audit writes share one PostgreSQL transaction. Phone conflicts remain review errors rather than automatic matches.

The legacy-link model was intentionally removed after confirming that the unlinked agents were dummy data, not valid user identities. Migration `000016` therefore deletes those agents and their dependent offers, campus associations, and application references before enforcing `agents.user_id NOT NULL`. That data disposition cannot be reconstructed by a down migration, so the migration is operationally irreversible even though its down section can relax the constraint for local tooling. User deletion is restrictive after linkage; deleting a user requires removing the agent relationship explicitly rather than silently violating the non-null identity invariant.

Canonical property and unit-type contributions record the authenticated creating user when the contribution is new. This is provenance for review, not ownership of the shared canonical record; historical rows remain nullable. Canonical media removal is a soft state change that retains the row and original uploader provenance while recording the removing user and timestamp. Public queries filter removed media in the same way they filter other inactive marketplace records.

Security-sensitive mutations write their audit event in the same PostgreSQL transaction as the state change. A failed audit write rolls back the mutation rather than producing an unaudited success. Audit metadata is bounded and excludes credentials, tokens, cookies, secrets, and raw phone numbers.

Agent suspension and reinstatement use the same repository transaction boundary. A campus operator or global admin locks and authorizes the target agent through `agent_campuses`, changes the agent status, updates linked application statuses, and writes an audit event together. Suspension also revokes every non-revoked session for the linked user. The session lookup already treats revoked sessions as anonymous and effective access omits `active_agent` when the linked agent is suspended. Reinstatement does not create sessions. Agent-offer creation requires an authenticated active agent, derives the agent from the session, and enforces the target property's campus through `agent_campuses`. Canonical property and unit contributions are scoped to active agents, campus operators, and global admins. Canonical corrections and media removal are scoped to campus operators or global admins; offer media may also be removed by its active agent owner.

---

# Workflow Philosophy

"Contact Agent" is too generic.

Users usually want one of:
1. Ask a question
2. Request inspection
3. Attempt reservation

These should become structured workflows. Platform should capture intent, context, workflow state before communication moves external.

---

# Payments

**Decision:** Don't aggressively prioritize payments initially.

**Why:** The accommodation workflow is highly conversational. Users often negotiate first, inspect first, ask questions first, discuss logistics externally. Forcing rigid payment flows too early likely increases friction. Trust and workflow quality matter first.

---

# Trust System

**Decision:** Signal-based trust architecture, not strict verification systems.

**Why:** Can't realistically inspect every room, verify every claim, guarantee every listing, manually moderate everything. Trust must emerge probabilistically.

Signals include: responsiveness, consistency, successful inspections, successful reservations, media quality, profile history, duplicate consistency, behavioral patterns, account age, activity history.

Trust is behavioral, not binary.

---

# Duplicate Listing Philosophy

**Decision:** Model the marketplace around:

```
Property -> Property Unit Type -> Agent Offer
```

not:

```
Agent -> Listing
```

**Why:** The actual market naturally contains duplicate inventory, shared inventory, inconsistent titles, inconsistent media, multiple agents representing same property. System should reflect reality instead of pretending ownership is always clean.

Goal is reduce marketplace chaos, reduce duplicate confusion, preserve operational flexibility — not achieve perfect deduplication.

---

# Media Storage

**Decision:** Cloudflare R2 in production, MinIO locally

**Why:** R2 provides S3 compatibility, low operational cost, no egress fees, scalable object storage. MinIO enables local S3-compatible development, environment consistency, easier local workflows.

---

# Tooling

**Decision:** pnpm workspaces + mise (tools, env, tasks)

**Why:** Provides lightweight orchestration, pinned tooling versions, clean local setup, modern developer experience without introducing heavy monorepo infrastructure. Want to avoid excessive orchestration tooling, enterprise-style build systems, unnecessary complexity.

---

# Mobile Strategy

**Decision:** Launch with a mobile-first installable PWA. Use React Native later if native mobile becomes necessary after validation.

**Why:** The current go-to-market need is fast iteration, low distribution friction, and rapid learning from FUTA students and agents. A PWA can be opened from links, shared through WhatsApp, installed on phones, and improved without waiting on App Store or Play Store review. Native mobile can wait until the core marketplace workflows are validated.

If native mobile becomes necessary, React Native is the preferred direction because it has stronger ecosystem maturity, better long-term support, larger hiring pool, and more stable tooling. Rejected NativeScript mainly because ecosystem maturity matters more than theoretical elegance.

The web app must remain great on desktop, but design, performance, navigation, and interaction decisions should start from the mobile user. Every meaningful screen should be checked at mobile and desktop breakpoints before considering it complete.

---

# Complexity Philosophy

**Decision:** Prefer conceptual sophistication over infrastructure sophistication.

**Meaning:** The product itself contains nuanced marketplace reasoning, trust reasoning, workflow reasoning, operational edge cases. Don't need to add infrastructure complexity on top of that.

---

# Final Principle

Continuously optimize for: maintainability, operational simplicity, marketplace liquidity, low friction, gradual evolution, explicit systems, realistic workflows.

Resist becoming: infrastructure-heavy, framework-obsessed, over-abstracted, enterprise-styled, hype-driven.

Goal isn't architectural impressiveness. Goal is building a marketplace that actually works in the real FUTA housing environment.
