---
name: check-go-style
description: Use when asked to check Go code against the project Style Guide — "проверь стиль", "check style compliance", "соответствует ли style guide", "нарушения стиля" — over the whole repo, a package/path, or the uncommitted diff. Report-only. For deep correctness/robustness audits use audit-go-package; for reviewing a PR diff use /code-review.
---

# Style compliance check

Audit Go code (and `contracts/`) for compliance with this project's Style Guide.
**Report-only** — do not modify any files unless the user explicitly asks to fix the findings
afterwards.

## Core principle

The `go-style-guide` skill is the **source of truth** — do not invent rules, and do not report
anything you cannot point at a line of that guide for. `golangci-lint` is the objective baseline;
this skill exists for the layer above it: the conventions the strict lint config does **not**
enforce. Everything in the Sweep and Checklist below is deliberately linter-invisible.

## Scope

Arguments control what gets checked:

- **empty** → uncommitted changes (`git diff` staged + unstaged); if there are none, the whole repo
- **a path or package** (`mrauth/`, `mrqueue/repository/`) → just that
- `--diff` → only `git diff` against the working tree
- `--all` → the entire repository

Ignore `examples/`, `docs/`, `vendor/`, `.github/` and generated `mock/` directories for
source-style rules (relaxed linting / generated code), but still report obvious breakages there.

## Process

1. **Load the rules** — invoke the `go-style-guide` skill.
2. **Determine the file set** from the scope above (`git status` / `git diff` / `Glob`).
3. **Run the linters** as the objective baseline:
   - `make lint` (or `golangci-lint run` scoped to the paths if `make` is unavailable)
   - `make test` (or `go test ./...`) when behaviour is in question
   Map each finding to the guideline it violates.
4. **Run the detection sweep** below — seconds, narrows the manual pass to real candidates.
5. **Work the manual checklist** on the file set + sweep candidates.

## Detection sweep

Every pattern below was validated against this repository. Each yields *candidates*, not verdicts —
open the hit and judge it. Restrict paths to the scope from step 2.

```bash
# 1. type declaration outside a grouped `type ( … )` block (tests may inline `type testCase`)
grep -rn --include="*.go" --exclude-dir=mock --exclude="*_test.go" "^type [A-Za-z]" . | grep -v " ($"

# 2. functional options declared outside their own *options.go file
grep -rln --include="*.go" "^func With[A-Z]" . | grep -v "options.go"

# 3. anti-join rule: NOT EXISTS / NOT IN instead of LEFT JOIN … IS NULL
grep -rn --include="*.go" "NOT EXISTS\|NOT IN" .

# 4. LIMIT passed as a bind parameter instead of inlined via mrstorage.NonZeroLimit
grep -rn --include="*.go" 'LIMIT \$' .

# 5. nested maps → flat map keyed by a small private struct
grep -rn --include="*.go" 'map\[[^]]*\]map\[' .

# 6. map-typed fields in input DTOs / wire config (the `grep -v` drops local vars, which are fine)
find . -type d \( -name dto -o -path "*/wire/*/config" \) -exec grep -rn --include="*.go" 'map\[' {} + \
    | grep -v ':=\|make(\|_test.go'

# 7. ExecAffected with a discarded count — only for genuinely idempotent/bulk ops, and say so
grep -rn --include="*.go" '_, err :\?= .*ExecAffected' .

# 8. sentinels compared with == instead of errors.Is
grep -rn --include="*.go" '== errors\.Err' .

# 9. happy path buried inside `err == nil` (allowed only as last resort, with a comment)
grep -rn --include="*.go" 'err == nil {' .

# 10. mockgen directives in production source instead of the consuming _test.go
grep -rn --include="*.go" "go:generate mockgen" . | grep -v "_test.go"

# 11. test-file forms this style forbids (testpackage's skip-regexp lets them through; we don't)
find . -name "*_internal_test.go" -o -name "*_export_test.go" -o -name "*_mock_test.go"

# 12. non-canonical spellings of a domain abbreviation (worked example: 2FA)
grep -rn --include="*.go" --include="*.yaml" "secondFactor\|SecondFactor\|TwoFA\|TwoFa\|two_factor" .

# 13. contracts: request model/parameter with maxLength but no paired minLength
for f in $(grep -rl "maxLength" contracts/ | grep -iE "request|parameters"); do
    grep -q "minLength" "$f" || echo "$f"
done

# 14. contracts: `example` listing alternatives instead of one valid value
grep -rn "example:.*|" contracts/

# 15. contracts: filename prefix must match the directory (_shared/ → Api., components/ → <Domain>.)
find contracts -path "*_shared/components/*" -name "*.yaml" ! -name "Api.*"

# 16. missing trailing newline
for f in $(git ls-files '*.go' '*.yaml'); do [ -n "$(tail -c1 "$f")" ] && echo "$f"; done
```

## Manual checklist

Grouped by Style Guide section. These need reading, not grepping.

### Declarations

- Zero-value config-like constructor params normalized to a `const ( defaultXxx … )` default
  **inside the constructor**; validation that *rejects* values stays separate.
- An `// OPTIONAL` interface/callback dependency defaults to a no-op **after** the options loop
  (covers both "not set" and `WithXxx(nil)`), and the no-op is an unexported type, not a package
  `var`.
- Narrow YAML-bound types (`uint8`, `int8`, `uint16`) stay in `wire/<comp>/config`; object
  constructors take the widened standard type (`int` for sizes/limits, `uint64` for identifiers),
  and the `wire` factory converts.
- Repository/storage methods return `[]entity.X` / `[]uint32` / an entity — **never** a `map`;
  any lookup index is built by the consuming layer.
- Interface methods with native return types (`[]uint32`, `string`, `bool`) have **named** results;
  when any result is named, the error is named `err` too.

### Comments

- Exported symbols: `// Name - описание.` — name, space-dash-space, text, terminating period.
  `godot` catches the period, nothing catches the missing `Name - ` prefix.
- Internal comments (inside bodies) starting lowercase must **not** end with a period —
  `godot` doesn't reach them.
- Non-trivial constructors document params as a bulleted `// Параметры:` list.
- Comment language (EN/RU) matches the surrounding file.
- In every file you touch, **all** stale or copy-pasted doc comments are fixed, not just the ones
  on changed lines.

### Naming

- Layer split: `usecase`/`service` name values `item`/`items` with business-intent method names;
  `repository` names them `row`/`rows` with technical CRUD verbs (`Insert`, `Update`, `Delete`,
  `Fetch`) — a count next to `FetchOpenSessionIDs` is `FetchOpenSessionCount`, not `Count…`.
- One canonical spelling per domain abbreviation, in every context (exported / sentinel /
  unexported / package dir / SQL / JSON / URL). A wire error **code** is the sentinel's own name
  minus `Err` (`ErrAuth2FAIsDisabled` → `"Auth2FAIsDisabled"`), so one grep finds var, code and
  contract mention.

### Errors

- The `mrpostgres` call matches the sentinel the caller guards on — `ExecRow` for exactly-one-row
  when absence is meaningful, plain `Exec` for many rows, `ExecAffected` **only** when the caller
  genuinely uses the number. A count that only ever gets compared to zero means `Exec` was wanted.
- The sentinel is documented in the method's doc comment when callers depend on it.
- Wrapped **once**, at the boundary that returns to the caller — not inside a `txManager.Do`
  closure *and* again on its result.
- A file whose every `errors.*` reference is `New`/`Is`/`As`/`Unwrap`/`Join` imports **stdlib**
  `errors`, not the core package.
- "The identifier the client presented doesn't work" is normalized in the **usecase**, never in an
  HTTP adapter; `errors.ErrRecordNotFound` has no place in a credential flow; the storage
  not-found sentinel is translated at exactly **one** call site (the lookup/consume of the
  client-presented identifier) — everywhere else a missing row is an internal error.
- Any branch keyed on a storage sentinel has a **repository integration test** pinning it; a gomock
  stub returning a sentinel the real repo never emits passes over dead code.

### Control flow

- The happy path stays unindented: branch on `if err != nil`, fall through to success. Each
  `err == nil` from sweep #9 either reads as a genuine last resort **with a comment explaining
  why**, or is a finding.

### Tests

- External `package foo_test` only, driven through the public API — no `_internal_test.go` /
  `_export_test.go` escape hatch (sweep #11), no reaching for unexported symbols.
- `require.*` for checks that must abort the test, `assert.*` for checks it can survive.
- Repeated mock/fixture setup across a package uses a `testify/suite` with `SetupTest`, not
  duplicated per-test wiring; mocks are `go.uber.org/mock/gomock` only.
- Mocks generated into a nested `mock/` package next to the consumer.
- `t.Parallel()` at the top of every test and subtest — **except** `PostgresTester` integration
  suites, where it must be absent *and* a comment must say why, so nobody re-adds it.
- Table-driven tests use a local `type testCase struct` with named cases; benchmarks live in
  `*_bench_test.go`.

### Contracts (OpenAPI)

- Filename prefix matches the directory: `_shared/components/` → `Api.`,
  `components/` → `<Domain>.` (sub-domains extend it).
- A `$ref` to a shared field pulls in its `description` too — reuse only when that description is
  the one you want verbatim.
- An optional request field is a **pointer** in Go with `omitnil`, and still carries its paired
  `min…` in the contract (sweep #13).
- `required` matches what the handler actually emits — a field with Go `omitempty` must not be
  `required`; zero-value `time.Time` is the usual trap.

## Output

Report concisely, grouped:

- **Lint failures** — the linter's own messages verbatim, with `file:line` and the rule name.
- **Style-guide violations** — `file:line` + the specific guideline + a one-line fix.
- **OK** — the categories that passed, so the reader knows what was actually covered.

Then a verdict: **compliant** / **needs fixes (N issues)**.

Sweep hits are candidates. Do not report one without opening the line and confirming it breaks a
guideline — a generated file, a test's inline `type testCase`, or a documented last-resort
`err == nil` are all legitimate.
