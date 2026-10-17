# AGENTS.md

Canonical instructions for AI assistants working on this codebase.

## Overview

Dagger build pipeline that produces multi-architecture binaries and OCI container images for [Memos](https://github.com/usememos/memos).

Source is fetched from upstream `usememos/memos` at build time; this repo owns the build logic, container recipes, and the release/publish workflow.

Some builds are consumed by [Memospot](https://github.com/memospot/memospot), a desktop
app for Memos.

## Rules

- Never push directly to `main` through the agent. Use `just publish TAG` for releases.
- Deliver code changes only after `just gate` succeeds.

## Code Style

- Abbreviations only for short-lived variables in <=5-line scopes (single-letter preferred). Exported names stay descriptive.
- Qualifiers carry meaning: big-endian for non-booleans (`latencyMsMax`), little-endian for booleans (`isEnabled`). Prefer `api *Fetcher` over `fetcher *Fetcher`.
- Readability: state invariants positively. Functional for small iterator logic, imperative on hot paths. Early returns over reassignment, `else` blocks, or nested `if`s. Pass critical options explicitly at call sites.
- Do not create single-use helpers under ~10 LoC simply to shorten a function.
- Treat `index`, `count`, and `size` as distinct types.

## Architecture & Modules

- One clear responsibility per module. Entry points handle wiring and startup flow only.
- Do not create or expand catch-all `utils`/`helpers` modules; place logic in its domain module.
- Separate pure domain logic from I/O and platform integration. Make dependencies explicit through signatures.
- Split on diverging responsibilities. Extract a helper only if reused or independently meaningful through a public contract.

## API & Error Handling

- Add context when propagating errors. No silent failures.
- Catch at architectural boundaries, not randomly. Fail fast.
- Warnings for recoverable errors; descriptive errors for fatal failures.

## Concurrency & State

- Prefer immutable data; scope mutability narrowly.

## Comments

- Leave surrounding comments unchanged unless directly related to the task.
- Omit new comments by default.
- Never add redundant docstrings (i.e `Build the frontend` for a `buildFrontend()` function).
- Explain intent, not steps. Inline comments only for non-obvious logic. Never address the user in code comments.
- When editing code, remove or rewrite related comments that violate the above rules or are no longer accurate.

## Tests

- Only high-value regression tests covering success, failure, and realistic edge cases. No tests for statically defined values, no negative tests for removed logic.
- Test public contracts and behavior only, with deep equality over whole objects. Do not duplicate implementation logic in the test.
- Deterministic: no wall-clock, network, or environment coupling unless the behavior under test requires it.
- Isolation: mock only at the system boundary (subprocess, filesystem, network), never domain logic.
- Inline BDD-style comments `// GIVEN` / `// WHEN` / `// THEN` (`// AND` as needed) + high-signal description, splitting each test into logical blocks whenever it doesn't get redundant and improves readability.
- **Test split: `go test` vs Dagger test module**:
  - `go test -v ./.dagger/. ./.dagger/buildplan/. ./.dagger/gomod/.` covers ONLY pure Go helpers
    that take/return stdlib values: no `dag.*`, no `*dagger.Directory|Container|File|Secret`,
    no `.Sync/.Contents/.Entries/.Publish`, no `ctx`-to-engine calls. All package
    patterns are kept: `.dagger/buildplan` and `.dagger/gomod` hold the pure logic, and `.dagger`
    stays in the list so an accidental test there fails loudly instead of never running.
  - Anything engine-bound is tested ONLY via a dedicated Dagger `tests` module with an
    `All(ctx) error` aggregator, run as `dagger call -m tests all --source=.`. Do not add engine-dependent tests to `*_test.go` — package `main` init panics without `DAGGER_SESSION_PORT`. See <https://docs.dagger.io/0.21/reference/best-practices/modules>.
  - Design rule: extract decisions into pure helpers, keep SDK wrappers thin.
    E.g. platform/name/tag/version logic lives in `.dagger/buildplan`, go.mod text rules in
    `.dagger/gomod`, and both get table-driven `*_test.go`; `build*.go`, `container.go`,
    `publishContainers`, `resolveVersion`, `prepareSource`, `applyPatches` stay thin orchestrators.
  - No `tests/` module exists yet. Do not create one for pure logic. If one is added
    later, also update `go.work`, `.golangci.yaml` lint scope, `justfile` (`test-dagger`),
    and CI so `just test` stays engine-free.

## Commands

| Action                                | Command               |
| ------------------------------------- | --------------------- |
| Concise agent-oriented validation     | `just gate`           |
| Full validation (lint → test → tidy)  | `just validate`       |
| Format all files                      | `just fmt`            |
| Lint (golangci-lint + shellcheck)     | `just lint`           |
| Build binaries (current platform)     | `just build`          |
| Build containers and load locally     | `just build-docker`   |
| Start loaded containers in demo mode  | `just run-demos`      |
| Publish (fmt → validate → tag → push) | `just publish TAG`    |
| Regenerate Dagger SDK bindings        | `just dagger-codegen` |
| Update Dagger SDK + Go deps           | `just update-deps`    |
| Install Git hooks (run once)          | `just hooks`          |
| Run prek hooks on all files           | `just hooks-run`      |
| Clean Dagger cache and containers     | `just dagger-clean`   |

Run `just --list` for all recipes. Most operations require Docker.

## Structure

- `.dagger/main.go` — Dagger function exports (Build, BuildContainers, Publish)
- `.dagger/build.go` — compile, archive, proto gen, frontend build
- `.dagger/container.go` — OCI images (Alpine or BusyBox ARMv5)
- `.dagger/patch.go` — sqlite/libc pinning wrapper + patch application
- `.dagger/gomod/gomod.go` — go.mod parsing and libc pinning rules
- `.dagger/publish.go` — multi-arch registry push (Docker Hub + GHCR)
- `.dagger/resolve.go` — version resolution and git source preparation
- `.dagger/buildplan/version.go` — version, tag, and release-identity rules
- `.dagger/buildplan/plan.go` — version plan: ref selection and version policy
- `.dagger/buildplan/targets.go` — platform matrix and artifact naming
- `.dagger/buildconsts/consts.go` — base-image tags, Go/node version
- `container/entrypoint.sh` — BusyBox ash container entrypoint
- `patches/` — drop-in `*.patch` files applied at build time

## Quirks and constraints

- **Dagger module Go** pinned in `.go-version` (currently `1.26.1`). Must match
  the Dagger engine's `go.mod` (`cat dagger.json | jq -r .engineVersion`, then
  fetch that tag's `go.mod`). CI and `just validate` enforce it. Change with
  `just tidy <version>`. Never set this to the Memos Go version.
- **Memos build Go** pinned as `MEMOS_GO_VERSION` + `GOLANG_BUILD_IMAGE` in
  `.dagger/buildconsts/consts.go` (currently `1.27.2`). Tracks upstream Memos,
  independently of the module Go. Applied inside the build container
  (`build.go` runs `go mod tidy -go=<MEMOS_GO_VERSION>`).
- **Go workspace** includes only `.dagger/`. Run `go work sync` after
  editing dependencies.
- **Formatter**: dprint (`.dprint.jsonc`) with gofmt, shfmt, JSON/Markdown/YAML.
  `just fmt` runs golangci-lint fmt + dprint. No recipe or CI checks formatting.
- **Linter scope**: `.golangci.yaml` excludes `.dagger/internal` and
  `.dagger/dagger.gen.go`. `golangci-lint` targets `./.dagger/.`,
  `./.dagger/buildplan/.`, and `./.dagger/gomod/.`. Adding other `*.go` trees
  requires updating config and CI.
- **Shellcheck** runs with `-s ash` on `container/entrypoint.sh`.
- **prek hooks** configured in `prek.toml`: trailing-whitespace, end-of-file-fixer,
  YAML/JSON/TOML checks, large-file guard, zizmor (GitHub Actions security),
  and shellcheck.
- **Dagger SDK export rule**: PascalCase in `.dagger/main.go` → exported via
  `dagger functions`; camelCase → Dagger-internal.
- **Dagger engine** version pinned in `dagger.json` (currently `v0.21.10`).
- **Dagger codegen**: run `just dagger-codegen` after public signature changes
  (`dagger develop --compat=skip`). `just update-deps` runs codegen without
  `--compat=skip`. Both remove the generated `.dagger/.gitignore`.
- **All `dagger call` invocations MUST pass `--source=.`**. Artifacts land in
  `./dist/`.
- **`--version` accepts**: legacy `v0.x` tags, bare or `v`-prefixed calendar
  tags (`26.10.0`, `26.10.0-rc.1`), bare dev versions (`26.10.15-dev.1511fd1`),
  40-char commit hashes, or `dev` (default — builds upstream `main`). Calendar
  inputs require the patch component, matching upstream's release tooling
  (`YY.MM.N[-rc.K]`): unpatched forms (`26.10`, and the historical
  `26.10-rc.1` tag) are rejected.
  Certain historical versions use hardcoded commit hashes
  (`knownVersionCommits` in `.dagger/buildplan/plan.go`). Dev builds embed an
  upstream-style `YY.MM.D-dev.<sha7>` version so they pass upstream validation.
  The embedded app version is rendered by `buildplan.AppVersion`, which
  preserves literal segments: upstream's startup validator requires the
  zero-padded month (`26.09.0`), which a semver re-render would strip to the
  rejected `26.9.0`.
- **Container registries**: `docker.io/lincolnthalles/memos` and
  `ghcr.io/memospot/memos-builds`. CalVer releases tag `latest`, `YY.MM` and
  `YY.MM.N` (bare, no `v` prefix; the `YY.MM` month alias is derived from the
  release); legacy `v0.x` releases tag `latest`,
  `0.YY` and `0.YY.N` (bare patch — the `v` lives only on the git tag);
  dev releases use their `YY.MM.D-dev.<sha7>` tag on
  GHCR plus a floating `dev` alias on both registries.
- **Container entrypoint** is BusyBox `ash`, not bash. Keep
  `container/entrypoint.sh` POSIX/ash-compatible.
- **OCI only**: Dagger supports standard OCI fields — no `HEALTHCHECK` or
  Dockerfile syntax.
- **Build matrix**: 24 platform targets in `.dagger/buildplan/targets.go`
  (`buildplan.Targets`; Linux/Darwin/Windows/FreeBSD × arch/variant combos).
  Container builds target Linux only.
- **modernc.org/sqlite ↔ libc pinning** is handled programmatically in
  `.dagger/patch.go`; the map and go.mod text rules live in `.dagger/gomod`.
  The hardcoded `sqliteLibcMap` must be kept current
  for new upstream versions; missing entries fall back to an upstream fetch.
  Drop-in `patches/*.patch` files are applied via `git apply` with `patch`
  fallback.
- **Just literal braces**: escape `{` as `{{ "{{" }}` and `}` as `{{ "}}" }}`
  inside `justfile` recipes.
- **Go/container image bumps** follow a detailed checklist in
  `.agents/skills/bump/SKILL.md`.

## Contributing

Agent-driven GitHub contributions on the user's behalf. No `CONTRIBUTING.md` or `.github/` templates.

- Style (contributions only): plain high-signal English, no emojis. Complete sentences; headings, lists, or tables when clearer. Keep responses concise; completeness wins over brevity.
- Never create an issue or PR unless the user directly requests that specific issue/PR in this session. Inferred intent is not permission. Covers `gh issue/pr create`, API, and UI paths.
- Issues: problem simply, then suggested solution briefly. Omit implementation detail unless material. Upload evidence to GitHub.
- PRs: subject `type(scope): imperative summary`, ≤72 chars, short lowercase scope. Body states the problem in a sentence or two, then the fix. State why over how.
- PR scope: one problem per PR. Adjacent cleanup needs its own PR unless required for the fix. Bump only deps the fix requires.
- PR evidence: upload to GitHub, never commit screenshots. UI changes need before/after images. Report material outcomes, risks, caveats, and blockers; omit inventories the diff already shows.
