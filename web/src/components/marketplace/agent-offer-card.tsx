import type { AgentOfferDetail } from '../../api/client'
import { Price } from '../ui/price'
import { ResponsiveImage } from '../ui/responsive-image'

type AgentOfferCardProps = {
  offer: AgentOfferDetail
}

const statusLabels = {
  available: 'Available',
  unavailable: 'Unavailable',
  paused: 'Paused',
} as const

export function AgentOfferCard({ offer }: AgentOfferCardProps) {
  const images = offer.media.filter((media) => media.kind === 'image')
  const initials = agentInitials(offer.agent.display_name)
  const isAvailable = offer.status === 'available'

  return (
    <article
      className={`rounded-card border p-4 sm:p-5 ${
        isAvailable ? 'border-border bg-background' : 'border-border bg-surface text-muted'
      }`}
      aria-label={`${offer.agent.display_name} offer`}
    >
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 items-center gap-3">
          <span className="bg-surface-strong text-foreground flex size-10 shrink-0 items-center justify-center rounded-full font-mono text-xs font-semibold">
            {initials}
          </span>
          <div className="min-w-0">
            <p className="font-body text-faint text-xs">Offered by</p>
            <p className="font-body text-foreground mt-0.5 truncate text-sm font-semibold">
              {offer.agent.display_name}
            </p>
          </div>
        </div>
        <span
          className={`font-body rounded-full px-2.5 py-1 text-xs font-semibold ${
            isAvailable ? 'bg-accent text-foreground' : 'bg-surface-strong text-muted'
          }`}
        >
          {statusLabels[offer.status]}
        </span>
      </div>

      <div className="border-border mt-4 border-t pt-4">
        <h3 className="font-display text-foreground text-lg leading-tight font-bold tracking-[-0.025em]">
          {offer.title}
        </h3>
        {offer.description ? (
          <p className="font-body text-muted mt-2 text-sm leading-6">{offer.description}</p>
        ) : null}
        {offer.notes ? (
          <div className="bg-surface rounded-control mt-3 px-3.5 py-3">
            <p className="font-body text-faint text-xs font-semibold tracking-[0.08em] uppercase">
              Offer notes
            </p>
            <p className="font-body text-muted mt-1.5 text-sm leading-5">{offer.notes}</p>
          </div>
        ) : null}
      </div>

      {images.length > 0 ? (
        <div className="mt-4 grid grid-cols-3 gap-2" aria-label={`${offer.title} offer photos`}>
          {images.slice(0, 3).map((media) => (
            <ResponsiveImage
              key={media.id}
              src={media.thumbnail_url}
              alt={media.caption || `${offer.title} from ${offer.agent.display_name}`}
              className="bg-surface-strong rounded-control aspect-square w-full object-cover"
              sizes="160px"
            />
          ))}
        </div>
      ) : null}

      <div className="border-border mt-4 flex items-end justify-between gap-3 border-t pt-4">
        <Price amount={offer.price_naira} />
        <p className="font-body text-faint text-right text-xs">
          Updated {formatUpdatedDate(offer.updated_at)}
        </p>
      </div>
    </article>
  )
}

function agentInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part.charAt(0))
    .join('')
    .toUpperCase()
}

function formatUpdatedDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return 'date unknown'
  return new Intl.DateTimeFormat('en-NG', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  }).format(date)
}
