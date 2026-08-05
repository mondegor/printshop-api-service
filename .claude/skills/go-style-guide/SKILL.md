---
name: go-style-guide
description: Code style guide for Go library modules following this company style. Use whenever writing, editing, or reviewing Go code in such a repo — or its OpenAPI contracts under contracts/ — so the result matches the established conventions and passes the strict golangci-lint config. Based on the Uber Go Style Guide, adapted with grouped type blocks, English or Russian doc comments, Proto/facade patterns, table-driven _test packages, plus contract file placement/prefix rules.
---

# Go Library Style Guide

Conventions for Go library modules following this company style. Follows the
[Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) as a base;
the deviations and company-specific rules below take precedence. Everything here is enforced
by `.golangci.yaml` (`golangci-lint` runs strict — `make lint` must pass before done).

## Formatting & imports

- `gofumpt` with `extra-rules: true` + `gofmt` + `goimports` + `gci`. Tabs for indentation.
- Line length ≤ **160** (tab-width 4) — `lll`.
- Import groups in this exact order (`gci`/`goimports` local-prefix), blank line between:
  1. standard library
  2. third-party
  3. the module's own packages (its module-path prefix)
- No file/copyright headers (no `goheader` config). No package doc comments required
  (`staticcheck` `-ST1000`, `revive` `package-comments` disabled).
- Use `any`, never `interface{}` (`revive use-any`). Prefer `strconv` over `fmt.Sprintf`
  for simple conversions (`perfsprint`).

## Declarations — company conventions

- **Always wrap type declarations in a grouped `type ( … )` block — even a single type.**
  This is pervasive (the norm here, unlike Uber which groups only related decls):
  ```go
  type (
      // Service - описание типа ...
      Service struct {
          repo Repository
      }
  )
  ```
- Group related `var`/`const` in blocks. Predefined error catalogs are `var ( … )` blocks
  of `ErrXxx` factory protos.
- **Functional options go in their own file.** When a type uses the functional-options
  pattern (`Option` type + `WithXxx` constructors), put the option type and all its
  `WithXxx` functions in a dedicated file next to the type. Name it bare `options.go`
  **only** when the package has a single type (no ambiguity). When the package holds
  other types/files, name the file `<thing>_options.go` after the owning type (e.g.
  `item_list_options.go` next to `item_list.go`), so it's clear which type the
  options belong to — a bare `options.go` would be ambiguous there.
- **Normalize zero-value inputs to defaults in the constructor.** When a constructor takes
  config-like scalar params (durations, counts, names) with a sensible default, replace the
  zero/empty value with a package-level default **inside the constructor**, so callers can pass
  `0`/`""` to mean "use the default". Keep defaults in a private grouped `const ( defaultXxx = … )`
  block next to the type. Two shapes:
  - Plain constructor — guard each param before building the struct:
    ```go
    const defaultTimeout = 5 * time.Second

    func NewService(timeout time.Duration) *Service {
        if timeout == 0 {
            timeout = defaultTimeout
        }

        return &Service{timeout: timeout}
    }
    ```
  - Functional-options constructor — seed defaults in the initial struct literal, apply the
    options, then zero-check the rest:
    ```go
    func newOptions(opts []Option) options {
        o := options{timeout: defaultTimeout}

        for _, opt := range opts {
            opt(&o)
        }

        if o.maxAttempts < 1 {
            o.maxAttempts = defaultMaxAttempts
        }

        return o
    }
    ```
  Validation that *rejects* invalid values (e.g. an upper bound) stays separate — defaults only
  fill in zeros.
- **Flexible/self-validating types live in config; constructors take standard types.** Narrow
  YAML-bound types (`int8`, `uint16`, `uint8`, …) belong to the `wire/<comp>/config` model, where
  the type itself documents and bounds the input. Object **constructors** take the *standard* widened
  type — `int` for sizes/lengths/limits/offsets (which may legitimately be negative), `uint64` for
  identifiers — and the `wire` factory does the conversion (`int(cfg.SampleParam)`). Don't
  push narrow config types into domain signatures.
- **Default an optional interface/callback dependency to a no-op, applied *after* the options
  loop.** For an `// OPTIONAL` collaborator set via `WithXxx` (an interface or func field),
  substitute a no-op default instead of `nil`-checking it at every call site. Do the substitution
  **after** applying the options (not in the initial struct literal), so it covers both "option
  not provided" **and** `WithXxx(nil)` — the latter would overwrite a literal default and panic
  at the call site. Globals are banned, so make the no-op a tiny unexported type, not a package
  `var`:
  ```go
  type defaultAlerter struct{}

  func (defaultAlerter) SendAlert(context.Context, uuid.UUID, int) error { return nil }

  // ... in the constructor:
  for _, opt := range opts {
      opt(&o)
  }

  if o.svc.alerter == nil {
      o.svc.alerter = defaultAlerter{} // covers "not set" and WithAlerter(nil)
  }
  ```
  Then call sites stay clean: `o.svc.alerter.SendAlert(…)` with no `nil` guard.
- No global mutable state (`gochecknoglobals`) — error/sentinel `var`s are the accepted
  exception. No `init()` functions (`gochecknoinits`).
- **Repository/storage methods return simple shapes — slices or entities, never `map`.**
  A data-access method yields `([]entity.X, error)` / `([]uint32, error)`; any derived
  structure (a lookup set, an index, a grouping) is built by the **consuming layer**
  (usecase/service), not by the repository. This keeps the data-access API uniform and
  decoupled from how a caller chooses to index the rows.
- **In Postgres repositories, express "rows lacking a related row" as a `LEFT JOIN … WHERE
  x IS NULL` anti-join — not `NOT EXISTS`/`NOT IN`.** Applies to both `SELECT` and
  `DELETE … USING …` (the `USING` from-list may contain the `LEFT JOIN`). The anti-join is
  the canon in this style (uniform with existing queries) and keeps "find the orphan" and the
  action it feeds atomic in one statement, e.g.:
  ```sql
  DELETE FROM <items> i
  USING UNNEST($1::uuid[], $2::int8[]) as c(owner_id, item_id)
      LEFT JOIN <item_locks> l
          ON  l.owner_id = c.owner_id
          AND l.item_id = c.item_id
          AND l.lock_type = $3 AND l.lock_status = $4 AND l.expires_at > NOW()
  WHERE i.owner_id = c.owner_id AND i.item_id = c.item_id
      AND l.lock_id IS NULL; -- anti-join: no live related row
  ```
- **Inline `LIMIT` (and similar planner-sensitive integer clauses) into the SQL text — don't
  pass them as `$N` bind parameters.** Build the clause from the concrete value
  (`mrstorage.NonZeroLimit(limit)`) and drop the arg from the
  `Query`/`Exec`/`ExecAffected`/`fetchRowsIDs` call. With `LIMIT $N` the planner can't see
  the real value at plan time and may pick a worse plan (poor row-count estimates, wrong
  scan/sort); inlining lets it plan against the actual bound. `limit` is always a
  caller-controlled `int` (a batch-size config, never user input), so this is
  injection-safe — never inline string/user-supplied values this way. `LIMIT` is normally
  the highest-numbered placeholder, so removing it needs no `$N` renumbering; keep the
  `limit` parameter (still used for the inline and for slice-capacity hints). Trade-off
  accepted: distinct limit values produce distinct SQL text (less prepared-statement reuse),
  fine because these limits are fixed configs.
  ```go
  sql := `
      ... ORDER BY
          updated_at ASC
      ` + mrstorage.NonZeroLimit(limit) + `
      FOR UPDATE SKIP LOCKED ...`
  ```
- **Avoid maps in config/input DTOs — use a slice of structs with an explicit key field.**
  Конфиги и входные DTO не используют мапы. Вместо `KindLimits map[string]uint32` — слайс:
  ```go
  type (
      GroupLimits struct {
          Name       string
          KindLimits []KindLimit
      }

      KindLimit struct {
          Kind    string
          ItemMax uint32
      }
  )
  ```
- **Avoid nested maps (`map[K1]map[K2]V`) — use a flat map keyed by a small private struct.**
  Двойные мапы только по согласованию. Внутренний lookup-индекс собирается в конструкторе из
  входного слайса; ключуется приватной составной структурой:
  ```go
  type groupKindKey struct {
      group string
      kind  string
  }

  itemLimits map[groupKindKey]uint32
  ```
- **Name the result parameters of interface methods when the results are native types.**
  In an interface declaration, give the returns names when their types are built-in/native
  (`[]uint32`, `string`, `bool`, `int`, …) so the signature is self-documenting, e.g.
  `FetchActiveItemIDs(ctx context.Context, ownerID uuid.UUID) (itemIDs []uint32, err error)`.
  When a result is already a descriptive named type (`dto.ItemFilter`, `[]entity.Item`), a
  name is optional. When other results are named, name the error `err` too.

## Comments (English or Russian godot-checked)

- Exported symbols **must** have a doc comment, in **English** or **Russian**, format
  `// Name - descript / описание.` (name, space-dash-space, then text). Doc comments
  end with a period (`godot`). Match the existing terse style.
- **Internal comments** (inside function/method bodies) may start with a lowercase
  letter; when they do, they **must not** end with a period. (`godot`'s scope is
  declarations only, so these aren't linter-enforced — follow the convention manually.)
- Document constructor params with a bulleted list when non-trivial:
  ```go
  // NewService - создаёт Service ...
  // Параметры:
  //   - repo - доступ к хранилищу данных;
  //   - handler - функция обработки результата.
  ```

## Naming

- Constructors: `NewXxx`. Receivers: short (1 letter), consistent per type
  (`s *service`, `w *wrapper`). `revive receiver-naming` enforces consistency.
- Sentinel errors prefixed `Err`, error *types* suffixed `Error` (`errname`).
  Note: error *codes* are camelCase string literals (e.g. `"errSomethingFailed"`).
- Initialisms via `revive var-naming`: `HTTP`, `JSON` (not `Http`/`Json`).
- **One canonical spelling per domain abbreviation.** A concept named by an abbreviation gets
  exactly one spelling of it across the whole codebase — no synonyms. Worked example below:
  `2FA` for two-factor auth, where the synonyms to avoid are `secondFactor`, `TwoFA`, `TFA`,
  `MFA`, `two_factor`. Because a Go package/identifier can't start with a digit, the spelling is
  context-bound but the abbreviation never changes:
  - **Exported identifiers** — treat `2FA` as an initialism, uppercase: `Auth2FA`, `User2FA`,
    `Disable2FA`, `Confirm2FA`, `Auth2FAType`.
  - **Sentinel errors** — `Err` must be followed by a letter (`errname` rejects a digit), so put the
    owning concept between them: `ErrAuth2FAIsDisabled`, `ErrAuth2FAMustBeDisabledFirst` — **not**
    `Err2FAIsDisabled` with a `//nolint:errname`. The word to insert is the one the domain already
    uses (here `Auth2FA` — the name the domain entity and its table already carry), not a filler.
  - **Unexported locals/fields** — keep `2fa` lowercase for readability: `auth2faStorage`,
    `auth2faTableName`, `auth2faVerifier` (the compromise — uppercase mid-identifier reads
    heavy here, and `staticcheck -ST1003` is off so it's allowed).
  - **Package/dir names** (lowercase, no leading digit) — prefix with a word: `auth2fa`,
    `auth2fatype` — never a synonym of the abbreviation (`secondfactor`).
  - **Persistence/transport** (SQL columns, JSON, YAML keys, URLs) — `2fa` / `auth_2fa` /
    `disable2fa`; never `two_fa`. An error *code* string literal is the sentinel's own name minus
    the `Err` prefix — `ErrAuth2FAIsDisabled` → `"Auth2FAIsDisabled"`. Keeping them in lockstep
    means one grep finds the Go var, its wire code and its mention in the contract.
  The mechanism (`TOTP`) and supporting concept (`recovery`) are **not** synonyms of `2FA` —
  they keep their own names. Apply the same single-canonical-spelling rule to any future domain
  abbreviation.
- **Layer-based entity & method naming.** In the `usecase`/`service` layers, name domain
  entity values `item` / `items` and give methods **business-intent** names that describe
  the operation in domain terms. In the `repository` layer, name DB-row values
  `row` / `rows` and give methods **technical CRUD** names (`Insert`, `Update`, `Delete`,
  `Fetch`, …). The split keeps domain vocabulary in the upper layers and data-access
  vocabulary at the storage boundary.
- Import aliases lowercase `^[a-z][a-z0-9]*$`; no redundant aliases.
- `staticcheck -ST1003` is off, so some naming rules are relaxed — still follow Go idiom.

## Errors

- Follow the **Proto pattern** where used: a proto is an immutable factory built once;
  derive concrete instances via `New`/`Wrap`/`WithDetails`. Never mutate a proto after
  construction.
- Root packages act as **facades**: expose new behavior via type aliases
  (`X = subpkg.X`) and `var` function aliases (`NewX = subpkg.NewX`); implement in the
  subpackage. Prefer importing the root facade in consumers.
- Wrap errors crossing external/package boundaries — use `%w`, and compare with
  `errors.As`/`errors.Is` (`errorlint`). Boundary-wrapping itself is a **manual
  convention** — `wrapcheck` is currently disabled in `.golangci.yaml`. Never return
  `nil, nil` (`nilnil`). Don't return `nil` after a non-nil error check (`nilerr`).
- Forbidden imports: `crypto/md5`, `crypto/sha1` (`revive imports-blocklist`, `gosec`).
- **Import stdlib `errors` when a file needs nothing but stdlib behaviour.** The core
  `errors` package re-exports `New`, `Is`, `As`, `Unwrap` and `Join` verbatim, so a file
  whose every `errors.*` reference is one of those five must import plain `"errors"` (in
  the stdlib import group) — not the core package. Reach for the core `errors` only when
  the file actually uses something of its own: `NewUserError`/`NewUserProto`,
  `NewInternalProto`, the `ErrInternal*`/`ErrEvent*` sentinels, `Wrapper` and the
  `New*Wrapper` constructors, `WithCustomCode`, `WrapInternalError`, etc. The import then
  says at a glance whether the file participates in the project's error taxonomy or just
  does sentinel plumbing. Applies to declaring kindless sentinels too: a file holding only
  `errors.New(...)` sentinels uses stdlib.

### Storage sentinels: pick the `mrpostgres` call that produces the error you guard on

A caller that branches on `errors.Is(err, errors.ErrEventStorage…)` is only correct if the
repository method underneath can actually emit that sentinel. `go-core/mrpostgres` decides
this purely by **which `DBConn` method you call**:

| Call | 0 rows affected / found → |
|---|---|
| `QueryRow(...).Scan(...)` | `ErrEventStorageNoRecordFound` |
| `ExecRow(...)` | `ErrEventStorageNoRecordFound` (and `ErrInternalStorageQueryFailed` if **>1** row) |
| `Exec(...)` | `ErrEventStorageRecordsNotAffected` |
| `ExecAffected(...)` | **nothing** — returns `(count, nil)` |

Rules:

- Writing a repo method whose absent row is meaningful to the caller (idempotency guard,
  404/"not found" mapping)? Use `ExecRow` when the statement touches **exactly one**
  row — `ExecAffected` silently returns `nil` and the caller's guard becomes dead code.
- Statement touches **many** rows (deactivate every child row of a parent, bulk update)? `ExecRow`
  is wrong — it fails on >1 row. Use plain `Exec`: it already reports "nothing matched" as
  `ErrEventStorageRecordsNotAffected`, so the caller gets its guard for free and the method
  stays a one-liner. Reach for `ExecAffected` **only when the caller genuinely uses the
  number** (batch-full signalling, "there is more" pagination) — a count that only ever gets
  compared to zero means you wanted `Exec`.
- Don't hand a count to a caller that just turns it back into a boolean, and don't emit one
  journal/log record per affected row when the records would be identical — one record for
  the event, not N copies of it.
- `ExecAffected` with a discarded count (`_, err :=`) is correct **only** for genuinely
  idempotent/bulk operations (ack-deletes, enqueue, expired-row cleanup). State the
  idempotency in the doc comment so the choice is not read as an oversight.
- Document the sentinel in the method's doc comment when callers depend on it
  (`// … Если записи нет, возвращает errors.ErrEventStorageNoRecordFound.`).
- **The sentinel survives the repo's `errorWrapper`, but not as the same error.**
  `NewInfraStorageWrapper` only *returns as-is* what matches `ErrEventStorageNoRecordFound`;
  `ErrEventStorageRecordsNotAffected` has no `Kind()`, so it gets wrapped into
  `ErrInternalStorageQueryFailed` — `errors.Is` still finds it through the `Unwrap` chain,
  but `err == errors.ErrEventStorageRecordsNotAffected` and any switch on the error's own
  kind/code will not. Guard with `errors.Is`, never `==`. Service-layer wrappers above it
  (`NewServiceOperationFailedWrapper`) are then identity for `kind.Internal` errors passed
  without attrs, so the sentinel reaches the usecase intact through repo → service → usecase.
- Wrap once, at the boundary that returns to the caller. Wrapping inside a
  `txManager.Do` closure *and* again on the closure's result is redundant — the outer
  wrapper already covers both, and the inner one only obscures where the error came from.
- **Gomock stubs cannot validate this.** A unit test whose mock returns a sentinel the real
  repository never produces will pass over dead code. Any branch keyed on a storage
  sentinel needs an **integration test on the repository** pinning that contract — and
  verify it by reverting the repo change and watching the test fail.

### Normalize "the identifier the client presented doesn't work" in the usecase

When a client presents an opaque identifier (an operation token, a refresh token, a
one-time code) and it turns out not to work, the **usecase** decides which domain sentinel
that is — never the controller. A translation living in an HTTP adapter makes the usecase's
contract depend on the transport, and the next endpoint that forgets to call the adapter
helper silently ships a status nobody declared in the contract. Worked example below: an auth
component whose clients present exactly two kinds of identifier — an operation token and a
session refresh token. The rule is the shape, not the names; a component with other kinds of
identifier names its own sentinels the same way.

- **One event, one sentinel — but one sentinel per *kind* of identifier.** Within a kind,
  empty, unknown, expired and already-consumed are deliberately indistinguishable from outside
  (no oracle), so they all return the *same* error; don't spend a second sentinel on "not
  presented". Across kinds, don't share one: two identifiers that deserve different HTTP
  statuses need different sentinels, or the status becomes unexpressible. A single-use
  credential naming a resource (`ErrOperationInvalid` — a token addressing a pending operation
  → `400` with the code in the body, bindable to the request field that carried it) and a
  session credential (`ErrTokenNotFoundOrExpired` → `401`, never field-bound, since binding to
  a field forces `400`) are two kinds, however alike they look at the storage layer.
- **`errors.ErrRecordNotFound` has no place in such a flow.** It maps to 404, which is the
  wrong answer for a bad credential and is usually not even in the contract. Use
  `NewServiceOperationFailedWrapper` in these usecases, not `NewServiceRecordNotFoundWrapper`.
- **Translate `ErrEventStorageNoRecordFound` at exactly one call site:** the storage call
  that *looks up or consumes* the row by that client-presented identifier. Everywhere else in
  the same flow — in particular any write after a `SELECT … FOR UPDATE` in the same
  transaction — a missing row is an invariant violation and must surface as an internal
  error. Blanket-wrapping every call with a "record not found" wrapper masks data corruption
  as a normal client answer.
- The exception that proves the rule: a `DELETE` keyed by the same token *without* a prior
  lock (optimistic single-use consumption) is a lookup site too — a concurrent request
  legitimately consumed it. Say so in a comment at that call site.
- **Sibling implementations of one interface must speak one error currency.** When two
  adapters back the same port (e.g. a DB-backed and a self-describing/parsing one), each
  adapter converts its own vocabulary into the shared package sentinels. Otherwise the HTTP
  status a client sees depends on which implementation the host happened to wire, and any
  agreement between the two rests on error *code strings* coinciding rather than on
  `errors.Is`.
- Pin the result with tests on both sides: the domain sentinel for the lookup site, and an
  explicit "not the domain sentinel" assertion for the post-lock write.
- **Removing a translating branch from an adapter? Audit every source that fed it.** A wrapper
  left as `NewServiceRecordNotFoundWrapper` in a neighbouring service keeps emitting
  `errors.ErrRecordNotFound`, and `kindlessErrorWrapper` forwards it untouched because it is
  `kind.User` — so it sails past the usecase that no longer expects it and reaches the client as
  a status no contract declares. Grep for the wrapper *and* the sentinel across the whole
  component, not just the files you edited.
- **Idempotency here never rests on "record not found" reaching the caller.** It rests on
  `SELECT … FOR UPDATE` serializing the racers, on explicit `errors.Is(…) → return nil` swallows,
  and on retry branches inside the repository. So switching a service wrapper cannot break a
  retry path — but say which of the three carries it in the doc comment, or the next reader will
  assume the wrapper was load-bearing.

## Control flow (wsl_v5 / whitespace / nlreturn / revive)

- Blank line before `return`/`break`/`continue` (`nlreturn`); no leading/trailing blank
  lines in blocks (`whitespace`). `wsl_v5` governs statement cuddling — keep related
  statements together, separate unrelated ones with a blank line.
- Early return / guard clauses; avoid `else` after a returning `if`
  (`indent-error-flow`, `early-return`, `superfluous-else`, all `preserveScope`).
  Avoid deep nesting (`nestif`).
- **Keep the main positive (happy) path outside `if` blocks.** Branch on the error
  condition `if err != nil { … }` and let execution fall through to the success path —
  do **not** put the success logic inside `if err == nil { … }`. Use `err == nil` as a
  condition only as a last resort (when the negated form is genuinely not expressible or
  would obscure intent), and **always** with a comment explaining why:
  ```go
  // good — error branch guards, happy path continues unindented
  v, err := parse(s)
  if err != nil {
      return err
  }

  use(v)

  // avoid — happy path buried inside the condition
  if v, err := parse(s); err == nil {
      use(v)
  }
  ```
- Prefer `make(...)` to init maps/slices (`enforce-map-style`, `enforce-slice-style`).
  Preallocate slices with a capacity hint when length is known (`prealloc`):
  `make([]string, 0, len(x)*2)`.
- No naked returns in non-trivial funcs (`nakedret`, `bare-return`). Functions return
  ≤ 3 results (`function-result-limit`). Watch `gocyclo`/`gocritic`/`unparam`.
- No `fmt.Print*`/debug forbiddens in library code (`forbidigo`) — allowed in
  `examples/` and `_test.go`.

## Tests

- **Always external test package: `package foo_test` in a file named `foo_test.go`**
  (`testpackage`). This is a hard convention here — there are **no** internal test
  packages in repos following this style. Test through the **public API only**; never reach
  for unexported symbols. Do **not** use the `*_internal_test.go` / `*_export_test.go`
  escape hatch that the `testpackage` linter's default `skip-regexp` would let through —
  the linter allows it, this style does not. To cover an unexported helper, drive it through
  the exported type/constructor that uses it, not directly.
- **Always use `github.com/stretchr/testify` for assertions** (`testifylint`). Use
  `require.*` for **fatal** checks that must abort the test on failure (preconditions,
  setup, `err != nil`, non-nil results you'll dereference); use `assert.*` for
  **non-fatal** checks where the test can keep running and report further failures
  (field-by-field comparisons after a successful operation). Rule of thumb: if continuing
  the test makes no sense (or would panic) once the check fails, use `require`; otherwise
  `assert`.
- For **complex tests with mocks** where the same object/mock initialization repeats across
  most tests in the package, use a **`github.com/stretchr/testify/suite`** test suite: put
  the shared fixtures (mocks, controller, system-under-test) on the suite struct and build
  them in `SetupTest` (or `SetupSubTest`), so each test method starts from a clean,
  consistently-initialized state instead of duplicating setup. Run it via a single
  `func TestXxxSuite(t *testing.T) { suite.Run(t, new(XxxSuite)) }` entry point. Keep using
  `gomock` for the mocks themselves. For simple tests without shared setup, prefer the
  plain table-driven form below.
- For mocks use **only** `go.uber.org/mock/gomock`. Generate mocks with `mockgen`
  (`//go:generate mockgen ...`), drive expectations via `gomock.NewController(t)` and
  `EXPECT()`. Do **not** hand-write mocks or use any other mocking library.
- **Put the `//go:generate mockgen ...` directives in the `_test.go` file that consumes
  the mocks, not in the production source.** Mocks are test-only tooling, so the
  directives belong with the tests; place them right after the import block of the
  package's test file. `go generate` runs per-directory, so `-source=foo.go` (and the
  `-destination=mock/...` paths) still resolve correctly from the test file. When one
  directory has several source files generating mocks, group all their directives in the
  single package test file (e.g. the package's main `<pkg>_test.go`).
- **Always generate mocks into a nested `mock/` directory next to the consuming package**
  (`package mock`, e.g. `service/item/mock/`), one `mock/` per package that owns/consumes
  the interfaces — never into `*_mock_test.go` in the test package. The external `_test`
  package imports `<pkg>/mock` and uses `mock.NewMockXxx(ctrl)`. A mock of an unexported
  interface (generated via `mockgen -source`) is assigned to the constructor's unexported
  interface parameter structurally — the unexported type name is never written in the test.
- `t.Parallel()` at the top of every test and subtest (`tparallel`). Test helpers call
  `t.Helper()` (`thelper`).
- **Exception: integration suites on `PostgresTester` never call `t.Parallel()`.**
  `infra.NewPostgresTester` starts a *fresh* container per suite, so N parallel suites in
  one package means N live Postgres instances — enough of them and the containers stop
  coming up (`wait until ready … context deadline exceeded`), which reads exactly like a
  broken migration. Drop `t.Parallel()` and say why in a comment above the entry point, so
  nobody re-adds it to match the unit tests next door. `paralleltest` is not enabled, so
  nothing forces the call back in. Two corollaries: `-p 1` does **not** help (it bounds
  package parallelism, not top-level tests inside one package), and a suite that fails in a
  full-package run but passes alone (`go test -run TestXxxSuite ./pkg/`) is resource
  contention, not a regression — check that before hunting for a bug.
- Table-driven with a local `type testCase struct`, named cases, `t.Run(tt.name, …)`:
  ```go
  func TestX_Method(t *testing.T) {
      t.Parallel()
      type testCase struct{ name, in, want string }
      tests := []testCase{ {name: "empty", in: "", want: ""} }
      for _, tt := range tests {
          t.Run(tt.name, func(t *testing.T) {
              t.Parallel()
              // ...
          })
      }
  }
  ```
- Relaxed in tests (excluded linters): `dupl`, `gosec`, `forbidigo`, `forcetypeassert`,
  `noctx`, `revive`, `unparam`.
- Benchmarks live in dedicated `*_bench_test.go` files; `go test -bench=. ./pkg/`.

## Contracts (OpenAPI)

The `contracts/<component>/` tree splits by **ownership**, and the filename prefix must match
the directory — the two always move together:

- `contracts/<component>/_shared/components/<kind>/` → prefix **`Api.`** — generic API primitives
  reusable by any component (`Api.Field.DateTimeCreatedAt.yaml`, `Api.ResponseJson.Error401.yaml`).
- `contracts/<component>/components/<kind>/` → prefix **`<Domain>.`** — component-specific
  (`Catalog.Field.DateTimePublishedAt.yaml`, `Catalog.Enum.ItemStatus.yaml`,
  `Catalog.Response.Model.ItemGroup.yaml`). Sub-domains extend the prefix: `Catalog.Check.*`.

`<kind>` is `fields` / `enums` / `models` / `responses` / `parameters` / `headers`.

- **`example` is one valid value, never a description of the alternatives.** OpenAPI 3.0.x has no
  `examples` map inside a schema (it exists only at media-type/parameter level), so `example:
  "A | B"` reaches Swagger UI and the code generators verbatim and clients copy a value that does
  not exist. The variants and the rule for parsing them belong in `description`; `example` carries
  one concrete value, kept consistent with its neighbours (`code: "ValidateError/user_email"` next
  to `detail: "Атрибут не может быть пустым"`). If the field has no `description` yet, add one
  first, then narrow the `example` — otherwise the list of variants is lost, not moved.
- **Don't reuse a shared field whose `description` doesn't describe your semantics.** A `$ref` pulls
  in the description too, so reuse is semantic, not just structural. `published_at` (time the item
  was published) must not `$ref` `Api.Field.DateTimeUpdatedAt.yaml` ("Дата и время обновления
  записи") — that puts a wrong, duplicated description next to the real `updated_at` in the bundled
  spec. Add a component field (`Catalog.Field.DateTimePublishedAt.yaml`) instead. Reuse a shared
  field only when the shared description is the one you want verbatim.
- **An optional request field is a pointer in Go, and it still carries `min` in the contract.**
  Every property of a request model that has a `max…` must have the paired `min…`, whether or not
  it is `required` — `minLength` is only checked when the property is present, so it never makes an
  optional field mandatory. The Go side must be able to tell "absent" from "sent empty", and a plain
  `string` cannot: `encoding/json` writes `""` in both cases, so `{"secret":""}` slips past
  `validate:"omitempty,min=4"` while violating the schema's `minLength: 4`. Model it as:
  ```go
  Token  string  `json:"token" validate:"required,min=64,max=128"`   // required → value type
  Secret *string `json:"secret,omitempty" validate:"omitnil,min=4,max=32"` // optional → pointer
  ```
  `omitnil` skips the field only when it is `nil` (absent or `null`); anything sent — including
  `""` — is validated in full. Don't drop `minLength` from the contract to paper over a mismatch:
  the fix belongs in the Go model. Custom tags work through the pointer (validator dereferences a
  non-nil pointer before the tag runs) and the error attribute keeps the `json` name.
  The same rule holds for an **optional query parameter** (`minLength` in the parameter's `schema`):
  `parser.RawParamString(r, "group")` returns `*string` — `nil` when the key is absent, a pointer to
  `""` for `?group=` — so the handler fills the same pointer-typed struct and validates it with
  `parser.ValidateStruct(ctx, &req)` instead of `parser.Validate` (there is no body to parse).
  Keep the `json` tag on such a struct: it names the attribute of the resulting `400`.
- Match `required` to what the server actually emits: a field the handler omits (Go `omitempty`)
  must not be `required`, or clients get a phantom-mandatory field. Zero-value Go times are the
  usual trap — `time.Time{}` formats to `0001-01-01T00:00:00Z`, which is schema-valid and
  semantically garbage.

## Before finishing

- Run `make lint` (or `golangci-lint run`) and `make test` (or `go test ./...`).
  The lint config is strict — treat any finding as a blocker.
- Files end with a trailing newline.
