# Laivan — Product Specification

# Product Goal

Enable students to discover accommodation more efficiently while helping agents manage listing discovery and inquiries in a more structured manner.

The system should reduce fragmentation, improve trust, and simplify coordination without forcing users into unnatural workflows.

---

# Core Product Areas

1. Public Discovery
2. Listing Management
3. Inquiry Coordination
4. Inspection Workflow
5. Trust & Signals
6. Admin Moderation
7. Marketplace Structure

---

# User Roles

## Student

Capabilities:

* browse listings
* filter listings
* view listing details
* save/bookmark listings
* contact agents
* request inspections
* submit reviews

---

## Agent

Capabilities:

* create offers
* upload images
* manage listings
* receive inquiries
* manage availability
* respond to requests
* build reputation

---

## Admin

Capabilities:

* moderate content
* merge duplicates
* remove spam
* manage abuse
* manage trust signals
* curate featured listings

---

# Listing Discovery

## Listing Feed

The platform should provide a structured listing feed.

Listings should include:

* title
* room type
* approximate area
* pricing
* images
* agent information
* trust indicators

The feed should prioritize:

* freshness
* relevance
* trust signals
* availability confidence

---

# Search & Filtering

Users should be able to filter by:

* price range
* room type
* area
* gender preference (if applicable)
* availability

Search should remain lightweight initially.

Complex search infrastructure is unnecessary for MVP.

---

# Listing Detail Page

The detail page should contain:

* images
* room information
* approximate location
* pricing
* amenities
* agent information
* trust indicators
* inquiry actions

Primary actions:

* Contact Agent
* Request Inspection
* Save Listing
* Reservation Interest

---

# Communication Flow

The platform should not implement full messaging infrastructure initially.

Instead:

users initiate structured intents.

Then WhatsApp opens with contextual pre-filled messages.

Examples:

"Hi, I am interested in the self-contained room at Alice Lodge. Is it still available?"

This preserves:

* user familiarity
* operational simplicity
* low adoption friction

---

# Inquiry Types

## General Inquiry

Represents:

"I want more information"

---

## Inspection Request

Represents:

"I want to physically inspect this property"

Fields:

* preferred time
* optional message
* phone number

---

## Reservation Intent

Represents:

"I am seriously interested in securing this room"

This does NOT imply:

* payment
* reservation guarantee
* legal commitment

---

# Agent Dashboard

Agents should have access to:

* active listings
* inquiry overview
* inspection requests
* availability status
* listing performance

The dashboard should remain operationally lightweight.

Avoid enterprise-style complexity.

---

# Media Uploads

Agents should be able to upload:

* images
* optional videos later

The platform should:

* compress media
* moderate abusive uploads
* eventually support duplicate detection

---

# Duplicate Handling

The platform should support:

* manual duplicate merging
* soft duplicate detection
* property grouping

The system should acknowledge that:

multiple agents may advertise the same room.

---

# Trust & Signals

The system should expose lightweight trust indicators.

Examples:

* recently updated
* multiple confirmations
* responsive agent
* repeated submissions
* verified media consistency

The platform should avoid false claims of certainty.

---

# Review System

Students may leave:

* property reviews
* agent reviews

The review system should avoid:

* harassment
* abuse
* fake reputation manipulation

Moderation tooling is important.

---

# Location Handling

Public listings should generally avoid exact coordinates.

Instead:

* approximate area
* landmarks
* district references

Detailed directions may happen later during conversations.

---

# SEO Strategy

The public web application should prioritize discoverability.

Important pages include:

* Campus accommodation landing page (initially FUTA)
* area pages
* listing pages
* room category pages

Rendering strategy:

* ISR for public listing pages
* client-side freshness indicators
* lightweight real-time validation for actions

The platform should avoid expensive SSR for every request.

---

# Performance Requirements

The system should:

* remain mobile-first
* load quickly on low-end devices
* function reasonably on poor connections
* optimize images aggressively

---

# Security Requirements

The system should:

* protect user data
* sanitize uploads
* prevent abuse/spam
* rate limit suspicious behavior

The platform should not overcomplicate security early,
but should remain structurally safe.

---

# Moderation Requirements

Admins should be able to:

* remove fake listings
* merge duplicates
* suspend abusive agents
* manage reported content

Moderation tooling is operationally important.

---

# Initial MVP Scope

The MVP should focus on:

* listing discovery
* structured listings
* inquiry intents
* WhatsApp coordination
* lightweight trust indicators
* basic agent dashboards
* manual moderation

The MVP should NOT initially include:

* chat systems
* payments
* escrow
* advanced AI ranking
* complex recommendation engines
* automated verification systems
* microservices architecture

---

# Future Possibilities

Possible future expansions:

* stronger dedupe systems
* analytics
* richer trust systems
* agent workflow tooling
* cross-campus support
* recommendation engines
* enhanced moderation automation

However,
these should not distract from marketplace fundamentals.

