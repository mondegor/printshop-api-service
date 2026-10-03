---
name: audit-go-package
description: Use when asked to "audit" a Go package or module, to re-audit one against a prior AUDIT.md, or to do a deep correctness/robustness pass (concurrency, leaks, panics, security) that ordinary diff review misses. Not for normal PR diffs — use /code-review for those.
---

# Audit a Go package (multi-dimensional)

A repeatable methodology for a *full* audit of a Go package/module across several dimensions.
Distinct from `/code-review` (diff-focused): this audits a whole package and produces a
persistent findings document.

The dense per-dimension traps live in [checklist.md](checklist.md) — that file is the
**per-agent brief** pasted to the dispatched subagents. This file is the **orchestrator's**
methodology: scope, process, severity, consolidation, output.

## Dimensions

One line per dimension; the concrete checks live in [checklist.md](checklist.md):

1. **Security** — input validation, overflow/narrowing casts, injection, panics from untrusted
   input, secret/token/PII leakage. Auth/crypto/SQL packages also get the **Auth / crypto / SQL** block.
2. **Architecture** — interfaces/contracts, duplication, dead code, facade/re-export hygiene, naming.
3. **Concurrency correctness** — shared state synchronized, atomics correct, no data races.
4. **Fault-tolerance** — per-item panic recovery, idempotent `Shutdown`/`Close`, ctx-cancellation, nothing swallowed.
5. **Resource & memory leaks** — goroutine exits, `valueCtx` growth, unstopped timers, unbounded buffers, unclosed `rows`/`Body`/`tx`.
6. **High-load / performance** — backpressure, channel buffering, hot-path allocations, lock contention.
7. **Tests & coverage** — error-path coverage, table-driven external `_test` packages, I/O integration tests, meaningful assertions.
8. **Observability & error semantics** — single-point logging, trace propagation, `%w` chains, `context.Context` threaded through I/O.
9. **Code comments / docs** — stale/contradictory comments, missing godoc, godot, language consistency, generated-mock headers.

### Agent assignment

| Subagent | Dimensions covered |
|----------|--------------------|
| A | 1 Security · 4 Fault-tolerance · 8 Observability/error-semantics (incl. secret/token leakage) |
| B | 5 Resource & memory leaks · 6 High-load/performance |
| C | 2 Architecture · 3 Concurrency correctness · 7 Tests & coverage |
| D | 9 Comments & documentation |

> Agent A carries the heaviest load (1 · 4 · 8); D the lightest (9). Keep the split —
> docs need a different reading mindset — but size A's brief accordingly.
>
> Agent A also owns the **Auth / crypto / SQL** block when the package handles
> secrets/tokens/crypto or builds SQL — paste it into A's brief alongside dimension 1.

## Process (orchestrator)

1. **Read the prior `<pkg>/AUDIT.md` first** if it exists. Treat its "✅ fixed" marks as
   *claims to verify*, not facts. On a re-audit, also focus new effort on what changed since the
   last round — diff against the date/commit recorded in the prior `AUDIT.md` header
   (`git log --since=<date>` or `git diff <commit>..HEAD -- <pkg>`).
2. List the package files and understand structure (`find <pkg> -name '*.go'`; read entry points,
   constructors, the main loops/goroutines). **Fix the scope here**: decide which adjacent layers
   are in or out (composition-root/wiring, `_sample` migrations, generated mocks) and record it for
   the `AUDIT.md` header. For a whole **module**, audit package-by-package with one `AUDIT.md` per
   package rather than one giant report.
3. **Dispatch parallel subagents — one agent per dimension group** per the Agent-assignment table
   above (in a single message; **REQUIRED SUB-SKILL:** superpowers:dispatching-parallel-agents).
   Give each agent: the package path, module path, its dimension block(s) from
   [checklist.md](checklist.md), the prior `AUDIT.md` path, and these instructions — read actual
   source (don't trust AUDIT.md), separate **"Verification of prior findings"** from
   **"New findings"**, cite `file:line`, assign severity, do not modify files. For a large package,
   tell agents to read file-by-file (not the whole tree at once) to avoid context exhaustion.
4. **Personally re-verify every headline finding against the source before finalizing.** Read the
   exact lines. This catches both false "fixed" claims and agent false positives (see Lessons).
5. **Consolidate into `<pkg>/AUDIT.md`** per the **Output** format below.

> **Environment caveat — `-race`/CGO:** the race detector needs gcc/CGO. If it can't run in this
> environment, agents reason about data races manually, race tests must still *exist*, and
> `AUDIT.md` records that they weren't run under the detector. This applies wherever `-race` is
> mentioned below.

## Severity scale

- 🔴 critical — crashes the process / loses data / security hole; fix before ship.
- 🟠 major — correctness/robustness risk under realistic conditions (races, double-shutdown,
   architectural debt that will rot).
- 🟡 minor — comments, docs, consistency, low-risk edge cases.
- ℹ️ info/nit — purely stylistic or non-actionable observation.

For security findings, also note **likelihood × impact** (e.g. "low likelihood / high impact") —
a remotely reachable, easy-to-trigger hole ranks above a theoretical one with the same impact.

## Verification commands

```bash
go build ./...
go vet ./<pkg>/...
go test -coverprofile=cover.out ./<pkg>/...   # profile feeds the Tests & coverage gap analysis
go tool cover -func=cover.out                 # per-function coverage; find untested branches
golangci-lint run ./<pkg>/...     # strict project config; expect 0 issues
govulncheck ./<pkg>/...           # known CVEs in dependencies (go install golang.org/x/vuln/cmd/govulncheck@latest; if installed)
gosec ./<pkg>/...                 # security rules (G-codes), if installed
go mod tidy -diff                 # read-only: report unused / outdated deps without rewriting go.mod
CGO_ENABLED=1 go test -race ./<pkg>/...   # needs gcc; see Environment caveat above
grep -rn '<module>/<deadpkg>'     # confirm dead-code claims
```

## Output

Write/overwrite `<pkg>/AUDIT.md` with:

- **Header** — scope (state which adjacent layers are in/out, e.g. composition-root/wiring, sample
  migrations, generated mocks), plus branch, **commit SHA** (so the next round can diff against it),
  date, round.
- **How to verify** — the commands above.
- **Verification table for prior findings** — claimed vs actual-by-code with `file:line`.
- **New-findings list** — each finding carries a **stable ID** (`AUDIT-001…`, reused across rounds),
  severity, `file:line`, and a **Suggested fix** line; for **security findings** also record
  **likelihood × impact** (per the Severity scale). **Deduplicate** findings that overlapping
  agents raised for the same `file:line`, keeping the more actionable framing.
- **Rejected false positives** — with reasoning, so the next round doesn't re-raise them.
- **Priority summary**, and a clear statement of whether code was changed or the round is report-only.

## Lessons

Why the disciplines above exist — recurring traps from past rounds:

- **A "✅ fixed" mark is not proof** — fixes get reverted, partially applied, or claimed without landing.
- **Agent false positives are common** — they usually evaporate when you open the cited `file:line`
  (the code already guards the case, or the agent misread the control flow).
- **Overlapping agents double-raise the same `file:line`** under different dimensions.
- **Severity inflation** — a theoretical, hard-to-reach issue is not 🔴.
