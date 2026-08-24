# Laivan Public UI

This guide defines the Phase 1 public student interface. It is a thin application-level design system, not a standalone component package.

## Brand and type

The application uses white (`#ffffff`), near-black (`#0f0f0f`), quiet surfaces (`#f4f5f1` and `#ebebea`), and Laivan green (`#4ade80`). Positive and destructive state colours are `#22c55e` and `#e53935`.

The prototype’s muted reference is `#737373`. Production readable muted text uses `#707070` so small text remains WCAG AA on both white and the quiet surface. Faint `#a8a8a8` is reserved for non-text decoration and must not carry essential information.

- Bricolage Grotesque: wordmark, headings, and important actions.
- Geist: body copy, navigation, labels, and controls.
- Geist Mono: prices, counts, pagination, and compact numeric emphasis.

Use the semantic Tailwind tokens in `styles.css`. Do not spread raw colour, radius, shadow, font, or motion values through components.

## Icons

Use named static imports from the Hugeicons Stroke Rounded free set and render them through `ProductIcon`.

- Use `currentColor`, normally at 20–24 px and 1.7 stroke width.
- Give every icon-only control an accessible name through `IconButton`.
- Decorative icons stay hidden from assistive technology.
- Do not add Lucide, MUI, emoji icons, or a second icon renderer.

## Spacing, shape, and motion

Controls, cards, media, and sheets use their semantic radii. Interactive targets are at least 44 px. Public content is bounded by the `content` container, while long prose uses `reading`.

Motion uses the standard short duration and easing. Movement communicates state rather than decorating it. All transitions collapse under `prefers-reduced-motion`. Sheets respect safe-area insets and use the semantic sheet elevation.

## Components and accessibility

Generic primitives live under `web/src/components/ui`; marketplace-aware components live under `web/src/components/marketplace`.

- Use `Button`, `IconButton`, `TextField`, and `FilterChip` instead of restyling controls locally.
- Use `Sheet` for focus-trapped modal sheets with Escape, backdrop, and explicit-close behavior.
- Use `ResponsiveImage` for stable media layout and an accessible failure fallback.
- Use `FeedbackState` for blocking empty/error states and `StatusLine` for refresh or cached-data notices.
- Use semantic links for navigation and buttons for actions. Never nest interactive controls.
- Preserve visible keyboard focus and label every field and dialog.
- Announce status changes without moving focus unexpectedly.

## Marketplace hierarchy and language

The public hierarchy is always:

`Property -> PropertyUnitType -> AgentOffer`

“Accommodation type” is the student-facing label for `PropertyUnitType`. Discovery
starts with the accommodation a student is seeking, then presents the property as
context and the competing agent offers as market options. A discovery card opens
that accommodation type directly; the property remains available through its
context link.

Property facts and property media stay at the property layer. Structural facts, notes, and unit media stay at the unit layer. Agent title, description, operational notes, price, status, media, identity, and update time stay at the offer layer.

Discovery imagery prefers media attached to the selected accommodation type and
falls back to property media. Do not promote an individual agent offer’s media onto
an aggregated accommodation card.

Unit detail imagery follows the same preference. When a unit has no unit-level
photos, it may show clearly labelled property-context photos instead. This fallback
must not imply that a property photo depicts the selected unit.

Marketplace search uses the shareable `q` URL parameter for property names, areas,
landmarks, and accommodation-type language. It composes with the explicit area,
category, price, structure, and availability filters and resets pagination to the
first page. Applying or clearing filters must preserve the current search; clearing
search must preserve the current filters.

Public location is approximate: area and landmark only. Do not expose exact coordinates, directions, phone numbers, or direct contact details.

Use factual status language such as “2 available offers.” Do not claim verification, guaranteed availability, confirmed reservations, recommended agents, response times, ratings, inspections, or trust signals without contract-backed evidence. Workflow previews must say that nothing is submitted or saved.

## Responsive composition

Mobile is the starting point:

- one discovery column
- horizontally scrollable category chips
- near-full-height filter sheet
- stacked property and unit content
- compact page controls

Desktop is deliberately recomposed:

- persistent discovery filters with a bounded two-column result grid
- property media/facts beside unit comparison
- unit facts beside competing offers

Do not wrap the application in a phone frame or stretch a single mobile column across desktop.

## Anti-patterns

- Generic agent-owned listing cards
- Unsupported trust or freshness copy
- Exact public location
- Fake workflow submission or generic WhatsApp handoff
- Clickable `div` elements or nested buttons
- Raw colour values inside product components
- External image hotlinks
- Unbounded desktop content
- New component frameworks, animation libraries, or speculative abstractions
