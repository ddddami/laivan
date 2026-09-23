# Laivan

[![API](https://github.com/ddddami/laivan/actions/workflows/api.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/api.yml)
[![Go Lint](https://github.com/ddddami/laivan/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/golangci-lint.yml)
[![Contract](https://github.com/ddddami/laivan/actions/workflows/contract.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/contract.yml)
[![Web](https://github.com/ddddami/laivan/actions/workflows/web.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/web.yml)

Laivan, a coordination, trust, and workflow layer for the fragmented FUTA housing market.

See [current state and priorities](docs/roadmap.md) for the shipped product
boundary, remaining acceptance work, and the next marketplace steps.

## Setup

Requires [mise](https://mise.jdx.dev) for toolchain management.

```sh
curl https://mise.run | sh
eval "$(~/.local/bin/mise activate zsh)"   # add to ~/.zshrc
cd laivan
mise trust                                  # trust this mise.toml on first use
mise install                                # installs Go, Node, pnpm, dev tools
cp .env.example .env                        # configure environment
mise run infra:up                           # start postgres, minio, imgproxy
mise run db:up                              # run database migrations
mise run db:seed:dev                        # seed database (optional, after migrations)
mise run dev                                # start the local development app
```

The API listens on `:4000` and the Vite development server listens on `:3000`.

### Development Modes

Run one full-stack mode at a time. They use the same local API and Vite ports.

| Command | Use |
|---|---|
| `mise run dev` | Everyday desktop development with local Google sign-in |
| `mise run dev:preview` | Fast public-flow preview from desktop and mobile over the local network |
| `mise run dev:shared` | Full authenticated development on desktop and mobile through HTTPS |

#### Local Development

`mise run dev` is the default. It runs the complete app on `http://localhost:3000`
and configures the local OAuth callback automatically:

```text
http://localhost:3000/v1/auth/google/callback
```

Register that callback with the Google OAuth client if local sign-in is enabled.
The task keeps the origin and callback overrides out of `.env`, so switching modes
does not require editing local configuration.

#### Mobile Preview

`mise run dev:preview` exposes Vite on the local network. Open `localhost:3000` on
the desktop and the network URL printed by Vite on a mobile device. Use this mode
for fast marketplace browsing, layout, and other public flows. Use
`mise run dev:shared` for authenticated flows.

#### Shared HTTPS Development

Use one HTTPS address on a desktop browser and mobile device when testing Google
sign-in, authenticated inquiries, or the complete cross-device workflow.

Set this up once:

1. Authenticate the installed ngrok CLI with your ngrok account.
2. In the ngrok dashboard, open **Domains** and copy the account's free Dev Domain.
3. Add the following values to your local environment, replacing `<ngrok-host>`:

```text
LAIVAN_WEB_ORIGIN=https://<ngrok-host>
LAIVAN_ALLOWED_ORIGINS=${LAIVAN_WEB_ORIGIN}
LAIVAN_GOOGLE_REDIRECT_URL=${LAIVAN_WEB_ORIGIN}/v1/auth/google/callback
```

4. Add `https://<ngrok-host>/v1/auth/google/callback` to the authorized redirect
   URIs for the Google OAuth client.

Then run:

```sh
mise run dev:shared
```

Open the HTTPS ngrok address in a desktop browser and on a mobile device. The Vite
proxy stays in place: requests to `/v1` and `/__imgproxy` are forwarded to the local
API and image proxy.

### Common commands

| Command | Description |
|---|---|
| `mise run dev` | Run the local development app |
| `mise run dev:preview` | Run a fast desktop and mobile preview over the local network |
| `mise run dev:shared` | Run the app through the shared HTTPS origin |
| `mise run api` | Run only the API with hot reload |
| `mise run web` | Run only the web app with Vite |
| `mise run test` | Run fast backend and frontend tests |
| `mise run verify` | Run complete local CI verification |
| `mise run infra:up` | Start Docker services |
| `mise run infra:ps` | Show Docker service status |
| `mise run infra:logs` | Follow Docker service logs |
| `mise run db:up` | Run database migrations |
| `mise run db:seed:dev` | Seed dev data |
| `mise run db:reset` | Reset the local database to the latest migration |
