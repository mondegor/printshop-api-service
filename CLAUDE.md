# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Print Shop Back — a Go web service for calculating the cost and production time of print products. It is a **modular monolith** structured along DDD / clean-architecture lines. All Go code lives under `app/` (module path `print-shop-back`, Go 1.26).

The repository is operated through the external **Mrcmd** CLI (https://github.com/mondegor/mrcmd), which wraps Docker Compose, migrations, codegen, linting and tests. The root `Makefile` provides short aliases over the most common `mrcmd` calls.

## Common commands

Run from the repository root. Each `make` target is a thin wrapper over `mrcmd`:

| Task | Make | Underlying mrcmd |
|------|------|------------------|
| Install / first build | `make build` | `mrcmd install` |
| Download deps | `make deps` | `mrcmd go deps` |
| Apply DB migrations | `make migrate` | `mrcmd go-migrate up` |
| Codegen (localization, mocks) | `make generate` | `mrcmd go generate` |
| Format + lint | `make lint` | `mrcmd go fmt && fmti && fmti2 && lint` |
| Run all tests | `make test` | `mrcmd go test` |
| Tests + HTML coverage | `make test-report` | `mrcmd go test-report` |
| Start / stop stack | `make app-start` / `make app-stop` | `mrcmd start` / `stop` |
| Container status / logs | `make app-state` / `make app-logs` | |

`Makefile.mk` adds `make full` (deps + generate + fmt + lint + test + plantuml) used as the pre-commit check.

**Lint** is governed by `app/.golangci.yaml` — a strict config (`default: none` with a large explicit enable list, including `gochecknoglobals`, `gochecknoinits`, `wsl`, `gosec`, `lll`). New code must pass it; `make lint` auto-fixes formatting and imports first (`goimports` local prefix is `print-shop-back`).

**Running a single test**: tests live under `app/tests/integration/` and spin up Postgres + Redis via **testcontainers** (see `app/config/config_tests.yaml`, `db_host: testcontainer`), so Docker must be running. From `app/`:
```
go test -run TestAlgoRectCutting ./tests/integration/calculations/...
```
Integration tests use `tests/integration.HttpHandlerTester` to exercise real HTTP handlers through the full stack.

## Architecture

### Startup and composition root
- `app/cmd/main.go` boots the app, then runs a chain of processes via `oklog/run`: background services (task schedulers, Postgres LISTEN/NOTIFY listener, mailer/notifier processors) start first, then the two HTTP servers.
- `app/cmd/factory/` is the **composition root** — manual dependency injection. `factory.InitApp` → `InitAppEnvironment` builds shared infrastructure (Postgres, Redis, S3/MinIO, locker, locale pool, perms, request parsers/response senders) and then wires each domain module.
- Two HTTP servers run: the public API (exposed via Traefik at `api.print-shop.local`) and a monitoring server (`/health`, `/metrics`, `/v1/system-info` at `print-shop.internal`).

### Domain modules
Each business domain is a directory under `app/internal/`: `calculations`, `catalog`, `controls`, `dictionaries`, `filestation`, `provideraccounts`, `warehousing`. The matching `app/internal/factory/<domain>/` holds that module's wiring (`InitHttpModule`, per-controller `init*Controller` constructors).

A module is further split by **audience/realm** and a shared backend:
- `usr` (UserAPI), `adm` (AdminAPI), `prov` (ProviderAPI), `pub` (PublicAPI) — audience-facing API sections.
- `back` — shared backend domain logic reused across sections (e.g. `warehousing/actiongroup/back/usecase` contains refresh logic invoked from the `usr` controllers).
- `module/` — module-level constants: `Name`, `Permission`, DB schema and table names, URL param names, error definitions.
- `enum/` — typed enumerations.

### Layering inside a module section
Within e.g. `internal/warehousing/actiongroup/usr/`:
- **root `*.go` files** (e.g. `container.go`) declare the **ports** — `Service` and `Storage` interfaces consumed by that section.
- `entity/` — domain entities; `dto/` — input/output data structures.
- `repository/*_postgres.go` — storage implementations (Postgres via `go-storage`).
- `service/` — service implementations.
- `usecase/` — use cases, each a struct with an `Execute(ctx, dto) (result, error)` method; this is where transactional orchestration and business rules live (see `usecase/create_stock_container.go`).
- `transport/httpv1/` — HTTP controllers. Each exposes `Handlers() []mrserver.HttpHandler` mapping method+URL to handler funcs; `transport/model/` holds request/response models.

Controllers depend on narrow inline interfaces (e.g. `createContainerUseCase`) rather than concrete types — dependencies are interfaces defined at the consumer side.

### REST routing & access control
Routes are registered per realm in `app/cmd/factory/service/rest/rest_router_{usr,adm,prov,pub,auth}_register.go`. Each registration builds the section's HTTP modules, wraps them in a check-access middleware (`actionGroup`, `userProvider`, `PermsProvider`), and mounts them. Access control is config-driven: roles in `app/roles/*.yaml`, with realms, permissions and privileges configured under `config.AccessControl`.

### Cross-cutting infrastructure
Built heavily on the **mondegor** library family: `go-webcore` (HTTP server, routing, access, runner), `go-storage` (Postgres/Redis/locks, S3/MinIO via `mrminio`), `go-core` (errors, logging, async processes via `mrprocess`), `go-components` (auth, mailer, notifier, settings). When adding features, prefer these libraries' abstractions over rolling your own. `app/go.mod` carries commented-out `replace` directives pointing these libs at sibling checkouts (`../../go-*`) for local development.

Thin local wrappers over these libs live in `app/internal/adapter/`: `log` (logger), `trace` (Sentry), `workflow`. Use them instead of importing the underlying clients directly.

- **Config**: `ilyakaznacheev/cleanenv` over layered `app/config/*.yaml` + `.env`; env vars (mostly `APPX_*`) override YAML. `config_tests.yaml` is used by integration tests.
- **Migrations**: `golang-migrate` SQL files in `app/migrations/` (timestamped `*.up.sql`/`*.down.sql`), auto-applied on startup when `db_migrations_dir` is set. DB is split into Postgres schemas per domain (e.g. `warehousing`, `printshop_catalog`).
- **Events / async**: `mrevent.Emitter` for domain events; Postgres `LISTEN/NOTIFY` drives task schedulers for mailer, notifier and settings reload. The async process toolkit lives in `go-core/mrprocess` (formerly `mrworker`): `schedule` (task schedulers), `collect`/`consume` (message collectors/processors), `job/task`, plus startup/signal processes and period strategies (`NewDoubleDelayedStartStrategy`, `NewQuadQuickStartStrategy`). These are wired in `app/cmd/factory/service/*_service.go`.
- **Errors**: `go-core/errors` — sentinel errors prefixed `Err`, wrapped through an `errorWrapper` (`errors.NewServiceOperationFailedWrapper()`) to distinguish system from user errors.
- **Localization**: `golang.org/x/text` message catalogs under `internal/localization/dict/`, generated via `go generate` (sources in `localization/`).

### API contracts & docs
OpenAPI contracts live in `contracts/published/<realm>/` (admin, auth, provider, public, user, monitoring), with a `template/` skeleton and per-realm `_shared/` fragments. Each realm dir carries `lint.sh`, `bundle.sh` and `build.sh` scripts that run `redocly/cli` in Docker to lint, bundle (`bundled-openapi.yaml`) and render docs from its `openapi.yaml`. Project documentation, PlantUML diagrams and sequence diagrams are in `docs/`.

## Conventions
- Code comments and documentation are written in **English** or **Russian**; follow the surrounding style.
- Constructors are `New*` (returning concrete types) for components and `Init*` for factory wiring functions.
- No global variables and no `init()` functions (enforced by lint) — everything is wired explicitly through the factory.
