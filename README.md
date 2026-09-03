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

The API listens on `:4000` and the Vite development server listens on `:5173`.
The `app` task starts infrastructure and applies database migrations before
starting both development servers.

### Preview on a phone

Connect the phone and development machine to the same local network. After the
one-time setup above, run the mobile app task:

```sh
mise run app:mobile
```

Open the network URL printed by Vite, such as `http://192.168.0.101:5173`, on the
phone. The mobile task exposes Vite on the local network and routes image delivery
through the development proxy. Keep `VITE_API_BASE_URL` empty or unset for this
workflow so the browser uses the relative `/v1` proxy instead of trying to reach
`localhost` on the phone. This is an HTTP development preview; installing the PWA
and testing offline behaviour still require a secure context.

### Common commands

| Command | Description |
|---|---|
| `mise run app` | Run the full app on this computer |
| `mise run app:mobile` | Run the full app for phone preview |
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
