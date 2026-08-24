# Laivan — Domain Model

## Why The Domain Model Matters

One of the most important realizations during planning was that the accommodation market does not map cleanly to a simple listing model.

A naive marketplace model usually assumes:

Agent -> Listing -> User

But the actual market behaves differently.

The same property may:

* appear under multiple agents
* have multiple unit types
* use different titles
* use different images
* have different pricing depending on agent
* exist in multiple states simultaneously

This means:

Property != Listing
Listing != Agent

That realization fundamentally changes the architecture.

---

# Core Domain Relationships

The domain should roughly evolve around:

Property
-> Property Unit Types
-> Agent Offers
-> Media
-> Signals
-> Inquiries
-> Inspection Requests
-> Reservation Intents

This structure allows the platform to represent the market more realistically.

---

# Core Entities

## Campus

A campus represents a distinct physical university location.

Examples:

* Federal University of Technology Akure (FUTA)
* University of Lagos, Akoka Campus
* University of Lagos, Idi-Araba Campus
* University of Ibadan
* Obafemi Awolowo University, Ile-Ife

Campus is the current marketplace boundary. All properties, agents, and workflows are scoped to a campus.

Campus is intentionally the top-level entity for now. There is no University table or hierarchy above Campus. A student at one campus does not see listings from another campus by default.

When a university has multiple campuses (e.g. UNILAG: Akoka and Idi-Araba), each campus is modeled as a separate Campus row. A future University entity can be introduced later if needed for admin, analytics, or partnership workflows. Do not add it now.

Properties belong to a campus. Property unit types belong to a property. Agent offers belong to property unit types. This hierarchy prevents cross-campus contamination.

---

## Property

A property represents the real-world accommodation building or location.

Examples:

* Alice Lodge
* A self-contained building near South Gate
* A hostel building near FUTA Junction

Properties are the highest-level real-world entities.

Properties may contain:

* multiple unit types
* multiple agents
* multiple media submissions
* multiple trust signals

Properties should eventually become canonicalized entities.

However, this process may initially be partially manual.

---

## Property Unit Type

A property may contain different rentable unit types.

Examples:

* self-contained
* single room
* two-bedroom flat
* one-bedroom apartment
* shared apartment

Property unit types exist independently from agents.

This distinction matters because:

multiple agents may advertise the same property unit type differently.

Property unit types should capture both market language and light physical structure.

### Category IS the structure

In Nigerian student housing, category names carry well-defined structural meaning. These are not arbitrary labels.

| Category | Bedrooms | Bathroom | Kitchen | Parlour |
|----------|----------|----------|---------|---------|
| single_room | 1 | shared | shared | no |
| self_contained | 1 | private | private | no |
| room_and_parlour | 1 | private | private | yes |
| one_bedroom_flat | 1 | private | private | yes |
| two_bedroom_flat | 2 | private | private | yes |
| three_bedroom_flat | 3 | private | private | yes |

This is what agents and students actually mean when they use these terms. A "self-contained" without a private kitchen is not a self-contained -- it is a single room miscategorized. The system treats category as the primary structural definition.

### Optional overrides

The explicit structure fields (`bedroom_count`, `bathroom_type`, `kitchen_type`, `has_parlour`) exist for edge cases and for the `other` category. When provided, they override the category default. When omitted, the system derives them from category.

This preserves flexibility without requiring agents to fill out a form when the standard term already says everything.

### Name is optional

The `name` field is a custom label for marketing flair ("Premium Self-con", "Executive Room and Parlour"). When omitted, it defaults to the category display name (e.g. "Self-contained"). Most unit types will not need a custom name.

### Two kinds of notes

Notes are split by meaning. This separation is critical.

**Property/unit-level notes** (structural truth):

* Describe what the unit physically is
* Examples: "Kitchen is outside but private", "Top floor corner unit"
* Belong to: `PropertyUnitType`
* These are slow-changing facts about the building

**Agent/offer-level notes** (operational/market context):

* Describe how the unit is being sold or managed right now
* Examples: "2 left", "Inspection tomorrow only", "Landlord prefers students"
* Belong to: `AgentOffer`
* These change quickly and are specific to one agent's situation

This separation preserves the architectural boundary:

* Property layer = structural reality
* Offer layer = marketplace/operational state

---

## Agent Offer

An agent offer represents:

"this particular agent offering this particular property unit type"

This is extremely important.

The agent does not necessarily own the listing.

Instead, the agent participates in distribution.

This distinction affects:

* ranking
* boosts
* monetization
* trust
* inquiry routing

Agent offers may contain:

* custom titles
* custom descriptions
* custom images
* pricing variations
* response quality differences
* reputation differences

---

# Why This Matters

Without this separation, the platform creates major problems.

Example:

If one agent boosts a listing,
and another agent can capture the booking,
then monetization becomes broken.

Therefore boosts should likely apply to:

Agent Offer

not:

Property globally.

---

# Media Model

Media should not necessarily belong exclusively to a single listing.

Images may:

* represent the same room
* originate from different agents
* vary in quality
* become stale over time

The platform should eventually support:

* media confidence
* duplicate media detection
* property-level galleries
* best-image selection

However, sophisticated automation should not be an MVP requirement.

Initially:

manual moderation and lightweight heuristics are sufficient.

---

# Duplicate Listing Reality

Duplicate listings are not edge cases.

They are normal behavior.

Examples:

* multiple agents advertising same room
* slightly different titles
* different price presentations
* reused images
* inconsistent location descriptions

The platform should therefore think in terms of:

"entity consolidation"

rather than:

"every listing is unique"

---

# Canonicalization Strategy

The system should gradually evolve toward canonical property entities.

Examples:

Instead of:

* "Cheap Selfcon Near South Gate"
* "Affordable Room Close To FUTA"
* "Alice Lodge Self Contain"

the system should eventually infer:

These likely refer to:

Property: Alice Lodge
Property Unit Type: Self-Contained

However:

this should not rely heavily on AI complexity early.

Initially:

* admin moderation
* soft suggestions
* manual merges
* lightweight heuristics

are sufficient.

---

# Inquiry Model

An inquiry is not necessarily a conversation.

An inquiry represents:

structured intent before free-form communication.

Examples:

* availability check
* inspection request
* reservation interest
* general inquiry

After the intent is captured,
communication may continue externally through WhatsApp.

This distinction allows the platform to:

* preserve operational simplicity
* avoid building chat infrastructure
* still capture marketplace signals

---

# Inspection Request Model

An inspection request represents:

"I want to physically see this property"

Important fields may include:

* preferred time
* contact number
* student identity
* status
* agent response state

The actual coordination may still happen on WhatsApp.

The platform tracks workflow state,
not necessarily the entire conversation.

---

# Reservation Intent Model

The platform should avoid pretending that reservations are fully digital initially.

A reservation intent means:

"this user is seriously interested in securing this room"

This is distinct from:

* payment completion
* legal agreement
* move-in confirmation

The Nigerian accommodation workflow is often conversational and flexible.

The system should acknowledge that reality.

---

# Signals System

Signals are probabilistic trust indicators.

The platform cannot guarantee perfect verification.

Instead, signals accumulate confidence.

Examples:

* repeated listing consistency
* multiple agent agreement
* review history
* media consistency
* stale activity detection
* inquiry volume
* agent response reliability
* successful inspection history

Signals should influence:

* ranking
* trust display
* moderation priority
* recommendation quality

Signals should not falsely imply absolute certainty.

---

# Location Model

Exact public coordinates should generally not be exposed.

Reasons:

* privacy
* agent concerns
* theft/security concerns
* bypass risks

Instead, listings should expose:

* approximate area
* landmarks
* district-level positioning

Detailed directions may happen later in the workflow.

---

# Identity And Sessions

`User` is the application identity created from a verified external identity. The current provider is Google, but provider identity is stored separately from the user record so another provider can be introduced without changing the user model. A user has a normalized email, display name, account status, and timestamps.

An `ExternalIdentity` is the pair of provider name and provider subject linked to one user. The pair is unique, and a user has at most one identity per provider in this slice.

The API stores only hashes of opaque session and CSRF tokens. The browser keeps the session in an HttpOnly cookie and receives the CSRF token through the session endpoint and a non-HttpOnly cookie. A missing, expired, revoked, or suspended session is anonymous.

Google sign-in does not automatically claim a legacy agent. Agent applications, explicit operator activation, and role scope are later identity workflows.

## Agent Applications And Ownership

An `AgentApplication` records a signed-in user's request to participate as an agent for one campus. The submitted name and normalized Nigerian phone number are application data until an operator decides the application. A user may have one pending application per campus.

Activation has two explicit paths. The operator can create a new active agent linked to the applicant, or provide a selected legacy agent ID to link an existing active agent. Phone-number matches are review signals only; they never claim or link a legacy agent automatically. A legacy link does not overwrite the legacy agent's public fields.

Applications move through `pending`, `active`, `declined`, or `suspended` lifecycle states. Activation, explicit legacy linking, and decline are audited with the operator as actor and are committed transactionally with the application transition.

Effective access is derived from explicit global-admin roles, campus-operator assignments, and a linked agent profile. A linked agent's `active` or `suspended` status is returned separately from the role so callers do not mistake linkage for write permission.

## Agent Campus Scope And Lifecycle

`AgentCampus` is an explicit association between an agent and a campus. It is not inferred from the agent's current offers. Migration backfills associations from existing offers through their property unit types and properties, while activation idempotently associates the resulting or explicitly linked agent with the application's campus. An agent may later be associated with more than one campus without changing the offer model.

Campus operators may suspend or reinstate agents only when `agent_campuses` places the target in one of their assigned campuses. Global admins may act across campuses. Suspension is a transactional lifecycle change: the agent becomes `suspended`, linked active applications become `suspended`, all non-revoked sessions for the linked user are revoked, and an audit event records the actor, transition, and note. Unlinked legacy agents have no user sessions to revoke. Reinstatement restores linked suspended applications and writes an audit event, but never creates a session automatically.

Invalid lifecycle transitions are conflicts rather than idempotent successes. Suspension does not currently protect the public property, offer, or media write routes; those ownership and protected-write rules are a later slice. Public records and their visibility remain controlled by their own lifecycle status.

## User Roles

## Student

Students primarily:

* browse listings
* request inspections
* contact agents
* compare options
* save listings
* leave reviews

---

## Agent

Agents primarily:

* upload/manage offers
* respond to inquiries
* coordinate inspections
* maintain media
* build reputation

---

## Admin

Admins primarily:

* moderate listings
* merge duplicates
* handle trust issues
* manage abuse
* curate signals

---

# Key Architectural Insight

The platform should not model:

"who owns the listing"

as aggressively as typical marketplaces.

Instead, the platform models:

"who participates in distributing access to the property"

That is a more accurate reflection of market reality.
