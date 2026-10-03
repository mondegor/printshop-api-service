# Per-agent audit checklist (Go)

Per-dimension checklist for the dispatched subagents. The orchestrator pastes the
block(s) for an agent's assigned dimensions (see the Agent-assignment table in `SKILL.md`)
into that agent's brief. Sections are ordered by agent (A → B → C → D) so each agent's
block is contiguous. Each item is a concrete trap to look for; cite `file:line`,
assign severity (see `SKILL.md` → Severity scale), and do **not** modify files.

## 1. Security — Agent A (always applies)

For crypto/SQL/auth packages, also use the **Auth / crypto / SQL** block below.

- All external/untrusted input validated before use; size/length bounds on slices, strings, reads.
- Integer overflow on numeric conversions and arithmetic (`int`↔`int32/uint`, length math, `len()`
  used as a signed delta); narrowing/`unsafe` casts justified and bounds-checked.
- No `panic` reachable from untrusted input; out-of-range index/slice and nil-map writes guarded.
- No secrets, tokens, or PII baked into the binary, logs, or error texts.

## 4. Fault-tolerance — Agent A

- Every goroutine running user-supplied code has `defer recover()`; recovery is per-iteration/
  per-item, not per-goroutine (a panic must not end the worker loop).
- `Shutdown`/`Stop`/`Close` is idempotent — `close(ch)` wrapped in `sync.Once` (a 2nd `close`
  panics). Remember `Shutdown` runs on the *caller's* goroutine — worker recover does NOT cover it.
- `errorHandler.Handle` / user callbacks: are *they* protected from panicking?
- `ctx.Done()` branches: is data flushed (drain + `context.WithoutCancel` + handler timeout) or lost?
- Does `Start` return `ctx.Err()` or silently `nil` (masking the cause from a supervisor)?

## 8. Observability & error semantics — Agent A

- An error is logged once at the boundary that handles it — not re-logged at every frame on the
  way up (double/triple logging of the same failure).
- Trace/span/correlation context is propagated, not dropped; spans opened are always ended.
- `context.Context` is threaded through every I/O call; not silently replaced with
  `context.Background()`/`TODO()` mid-chain (which severs deadlines and trace context).
- Errors wrapped with `%w` (not `%v`) where callers may `errors.Is`/`errors.As`; the wrapping
  chain is intact and sentinel errors are re-exported by the facade for cross-package `errors.Is`.

## Auth / crypto / SQL — Agent A (only if the package handles secrets, tokens, crypto or builds SQL)

- Secrets/tokens/MACs compared with `subtle.ConstantTimeCompare`, not `==`/`bytes.Equal` (timing).
- Randomness for keys/tokens/nonces comes from `crypto/rand`, never `math/rand`.
- JWT/signatures: `alg=none` rejected; algorithm taken from a trusted allow-list (no alg-confusion
  asymmetric↔symmetric); `kid`/algorithm validated before use.
- Secrets and private keys never reach logs, error texts, or traces.
- SQL injection via **dynamic identifiers**: table/column names can't be parameterized (`$1`); if
  they come from the caller or are built dynamically, the identifier must be validated/quoted before
  interpolation.

## 5–6. Memory / high-load — Agent B

- `ctx = derive(ctx)` inside `for {}` → unbounded `valueCtx` chain. Use a loop-local var.
- `time.NewTicker`/`NewTimer` always `Stop`ped (defer). `NewTicker(d)` panics if `d <= 0` —
  is `d` floored, and is the floored value the one actually passed?
- Channels: buffered vs unbounded; blocking sends always have a `<-done`/`<-ctx.Done()` escape.
- Batch buffers bounded and reset/reused; note the dominant steady-state allocation.
- Unclosed `rows.Close()`/`Body.Close()`, missing `defer tx.Rollback()`, leaked transactions,
  connection-pool exhaustion.

## 2–3. Architecture / concurrency — Agent C

- Shared mutable struct fields across goroutines — synchronized? Reason manually if `-race` can't run.
- `atomic.*` used for non-atomic read-modify-write (Load…Store) → lost updates; needs CAS loop.
- Dead packages (`grep -r '<module>/<pkg>'` returns nothing) — flag.
- Repeated contract across N types but no declared interface → only add a **public** interface
  to production code if something actually consumes it (Go favours accept-interfaces-return-structs;
  an unused exported interface is dead surface). To document/lock the shared shape, declare a
  **private** interface in each external `_test` package and assert conformance there:
  `assert.Implements(t, (*iface)(nil), &pkg.Type{})` (pattern `TestXImplementsY`). Do **not** add
  `var _ Iface = (*T)(nil)` compile-asserts in package files, and never force a new prod import
  just to satisfy an assert.
- Duplicated worker-pool / boilerplate that has already drifted → extract to `internal/...`.
- Facade re-exports public sentinel errors so callers can `errors.Is` without importing subpackages.

## 7. Tests — Agent C

- **Coverage gap analysis**: from `go test -coverprofile` find exported functions and error branches
  with no test; list the missing tests as findings (severity by the risk of the uncovered code).
- Tests live in an **external `_test` package** (`package foo_test`) and exercise the public
  contract, not internals; prefer **table-driven** style.
- Assertions are **meaningful** — they check a concrete result/error (`errors.Is`/value), not just
  `require.NoError`.
- **I/O layers** (repositories, clients, filesystem) have integration tests, or an explicitly noted
  reason for their absence.
- **Race tests are mandatory for concurrent types**: a type whose methods are called from multiple
  goroutines needs a test that drives parallel access under `-race`. If `-race` can't run, the test
  must still exist, and `AUDIT.md` records that it wasn't run under the race detector.
- Determinism: no reliance on wall-clock time, network, or map-iteration order without control;
  a flaky test is itself a finding.

## 9. Comments / docs — Agent D

- Comments contradicting code (precision claims, old names after a rename, "to the second" vs ms).
- Missing godoc on exported symbols; duplicated doc text on option pairs (`WithX`/`WithXStrategy`).
- Generated mocks: stale `Source:`/`mockgen` header, archived deps (`github.com/golang/mock` →
  `go.uber.org/mock`).
