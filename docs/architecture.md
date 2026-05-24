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

**Initial decision:** Start with React + Vite + TanStack Router + TanStack Query + PWA support without introducing SSR infrastructure immediately.

**Why:** Laivan is primarily interaction-heavy, workflow-heavy, dashboard-heavy, mobile-first, authenticated. Most critical flows are client-side operational flows: inquiries, inspections, reservation intents, uploads, dashboard activity, listing management. Behaves more like an operational platform than a content website.

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

**Phase 1:** React + Vite + TanStack Router SPA + PWA

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
Property -> Room Type -> Agent Offer
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

**Decision:** pnpm workspaces + mise + Taskfile

**Why:** Provides lightweight orchestration, pinned tooling versions, clean local setup, modern developer experience without introducing heavy monorepo infrastructure. Want to avoid excessive orchestration tooling, enterprise-style build systems, unnecessary complexity.

---

# Mobile Strategy

**Decision:** Use React Native later if native mobile becomes necessary.

**Why:** React Native has stronger ecosystem maturity, better long-term support, larger hiring pool, more stable tooling. Rejected NativeScript mainly because ecosystem maturity matters more than theoretical elegance.

---

# Complexity Philosophy

**Decision:** Prefer conceptual sophistication over infrastructure sophistication.

**Meaning:** The product itself contains nuanced marketplace reasoning, trust reasoning, workflow reasoning, operational edge cases. Don't need to add infrastructure complexity on top of that.

---

# Final Principle

Continuously optimize for: maintainability, operational simplicity, marketplace liquidity, low friction, gradual evolution, explicit systems, realistic workflows.

Resist becoming: infrastructure-heavy, framework-obsessed, over-abstracted, enterprise-styled, hype-driven.

Goal isn't architectural impressiveness. Goal is building a marketplace that actually works in the real FUTA housing environment.
