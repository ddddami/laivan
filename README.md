# Laivan

[![API](https://github.com/ddddami/laivan/actions/workflows/api.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/api.yml)
[![Go Lint](https://github.com/ddddami/laivan/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/golangci-lint.yml)
[![Contract](https://github.com/ddddami/laivan/actions/workflows/contract.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/contract.yml)
[![Web](https://github.com/ddddami/laivan/actions/workflows/web.yml/badge.svg)](https://github.com/ddddami/laivan/actions/workflows/web.yml)

Laivan, a coordination, trust, and workflow layer for the fragmented FUTA housing market.

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
mise run app                                # start API and web together
```

The API listens on `:4000` and the Vite development server listens on `:3000`.

### Work Across Devices

Use one HTTPS address for Laivan in a desktop browser and on a mobile device. It
makes Google sign-in work on mobile and keeps every device on the same development
surface. The Vite proxy stays in place: requests to `/v1` and `/__imgproxy` are
forwarded to the local API and image proxy.

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

Every day, run:

```sh
mise run app
```

Open the HTTPS ngrok address in a desktop browser and on a mobile device. Do not use
the localhost or LAN addresses for normal development: they cannot complete the
Google OAuth callback on mobile. The ngrok command runs with the app and prints the
address when it starts.

### Common commands

| Command | Description |
|---|---|
| `mise run app` | Run Laivan across devices |
| `mise run api` | Run only the API with hot reload |
| `mise run web` | Run only the web app with Vite |
| `mise run test` | Run fast backend and frontend tests |
| `mise run verify` | Run complete local CI verification |
| `mise run infra:up` | Start Docker services |
| `mise run infra:ps` | Show Docker service status |
| `mise run infra:logs` | Follow Docker service logs |
| `mise run db:up` | Run database migrations |
| `mise run db:seed:dev` | Seed dev data |
| `mise run db:reset` | Reset the local development database |
