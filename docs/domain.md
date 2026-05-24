# Laivan — Domain Model

## Why The Domain Model Matters

One of the most important realizations during planning was that the accommodation market does not map cleanly to a simple listing model.

A naive marketplace model usually assumes:

Agent -> Listing -> User

But the actual market behaves differently.

The same property may:

* appear under multiple agents
* have multiple room types
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
-> Room Types
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

Properties belong to a campus. Room types belong to a property. Agent offers belong to room types. This hierarchy prevents cross-campus contamination.

---

## Property

A property represents the real-world accommodation building or location.

Examples:

* Alice Lodge
* A self-contained building near South Gate
* A hostel building near FUTA Junction

Properties are the highest-level real-world entities.

Properties may contain:

* multiple room types
* multiple agents
* multiple media submissions
* multiple trust signals

Properties should eventually become canonicalized entities.

However, this process may initially be partially manual.

---

## Room Type

A property may contain different room categories.

Examples:

* self-contained
* single room
* two-bedroom flat
* one-bedroom apartment
* shared apartment

Room types exist independently from agents.

This distinction matters because:

multiple agents may advertise the same room type differently.

---

## Agent Offer

An agent offer represents:

"this particular agent offering this particular room"

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
Room Type: Self-Contained

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

# User Roles

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

