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
mise install                                # installs Go, Node, pnpm, dev tools
cp .env.example .env                        # configure environment
mise run infra:up                           # start postgres, minio, imgproxy
mise run db:up                              # run database migrations
mise run db:seed:dev                        # seed database (optional)
mise run api:dev                            # start API server on :4000
# in another terminal:
mise run web:dev                            # start web dev server on :3000
```

### Common commands

| Command | Description |
|---|---|
| `mise run api:dev` | API server with hot reload |
| `mise run api:test` | Run API tests |
| `mise run api:lint` | Lint API code |
| `mise run api:check` | Format, tidy, lint, and test API |
| `mise run api:verify` | Non-mutating checks (CI) |
| `mise run web:dev` | Web dev server |
| `mise run web:test` | Run web tests |
| `mise run test` | All tests |
| `mise run infra:up` | Start Docker services |
| `mise run db:up` | Run database migrations |
| `mise run db:seed:dev` | Seed dev data |

