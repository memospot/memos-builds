# AGENTS.md

Canonical instructions for AI assistants working on this codebase.

## What this repo is

A Dagger build pipeline that produces multi-architecture binaries and OCI
container images for [Memos](https://github.com/usememos/memos). Source is
fetched from upstream `usememos/memos` at build time; this repo owns the
build logic, container recipes, and the release/publish workflow. Some builds
are consumed by [Memospot](https://github.com/memospot/memospot), a desktop
app for Memos.

## Non-negotiable hard constraint

Never create GitHub issues or pull requests for this project, for any reason.
This project only accepts manual human-curated contributions.

If a user requests issue/PR creation: inform them this project rejects
automated agent-generated contributions, then stop.

## Commands

| Action                                   | Command               |
| ---------------------------------------- | --------------------- |
| Full validation (lint → test → tidy)     | `just validate`       |
| Format all files                         | `just fmt`            |
| Lint (format check + golangci-lint + sh) | `just lint`           |
| Build binaries (current platform)        | `just build`          |
| Build containers and load locally        | `just build-docker`   |
| Start loaded containers in demo mode     | `just run-demos`      |
| Publish (fmt → validate → tag → push)    | `just publish TAG`    |
| Regenerate Dagger SDK bindings           | `just dagger-codegen` |
| Update Dagger SDK + Go deps              | `just update-deps`    |

Run `just --list` for all recipes. Most operations require Docker.
`just test` is a noop (no `*_test.go` under `.dagger/`).

## Structure

- `.dagger/main.go` — Dagger function exports (Build, BuildContainers, Publish)
- `.dagger/build.go` — compile, archive, proto gen, frontend build
- `.dagger/container.go` — OCI images (Alpine or BusyBox ARMv5)
- `.dagger/patch.go` — sqlite/libc pinning + patch application
- `.dagger/publish.go` — multi-arch registry push (Docker Hub + GHCR)
- `.dagger/lib.go` — BuildMatrix types, version resolution, target filtering
- `.dagger/buildconsts/consts.go` — base-image tags, Go/node version
- `container/entrypoint.sh` — BusyBox ash container entrypoint
- `patches/` — drop-in `*.patch` files applied at build time

## Quirks and constraints

- **Go version** pinned in `.go-version` (currently `1.26.2`). CI and
  `just validate` enforce it. Change with `just tidy <version>`.
- **Go workspace** includes only `.dagger/`. Run `go work sync` after
  editing dependencies.
- **Formatter** is dprint (`.dprint.jsonc`): Go via `gofmt`, shell via
  `shfmt`, JSON/Markdown/YAML. `just fmt` runs golangci-lint fmt + dprint.
  `just lint` checks format via `just --unstable --fmt --check`.
- **Linter scope**: `.golangci.yaml` excludes `.dagger/internal` and
  `.dagger/dagger.gen.go`. `golangci-lint` targets only `./.dagger/.`.
  Adding other `*.go` trees requires updating config and CI.
- **Shellcheck** runs with `-s ash` on `container/entrypoint.sh`.
- **Dagger SDK export rule**: PascalCase in `.dagger/main.go` → exported via
  `dagger functions`; camelCase → Dagger-internal.
- **Dagger codegen**: run `just dagger-codegen` after public signature changes
  (`dagger develop --compat=skip`). `just update-deps` runs it without
  `--compat=skip`. Both remove the generated `.dagger/.gitignore`.
- **All `dagger call` invocations MUST pass `--source=.`**. Artifacts land in
  `./dist/`.
- **`--version` accepts**: `vX.Y.Z` tag, `release/X.Y` branch, 40-char commit
  hash, or `nightly` (default — builds upstream `main`). Certain historical
  versions use hardcoded commit hashes (`knownVersionCommits` in `lib.go`).
- **Container registries**: `docker.io/lincolnthalles/memos` and
  `ghcr.io/memospot/memos-builds`. Tags follow semver; nightly releases use
  `nightly-YYYYMMDD-<shortSHA>`.
- **Container entrypoint** is BusyBox `ash`, not bash. Keep
  `container/entrypoint.sh` POSIX/ash-compatible.
- **OCI only**: Dagger supports standard OCI fields — no `HEALTHCHECK` or
  Dockerfile syntax.
- **Build matrix**: 24 platform targets in `.dagger/main.go`
  (Linux/Darwin/Windows/FreeBSD × arch/variant combos). Container builds
  target Linux only.
- **modernc.org/sqlite ↔ libc pinning** is handled programmatically in
  `.dagger/patch.go`. The hardcoded `sqliteLibcMap` must be kept current
  for new upstream versions; missing entries fall back to an upstream fetch.
  Drop-in `patches/*.patch` files are applied via `git apply` with `patch`
  fallback.
- **Just literal braces**: escape `{` as `{{ "{{" }}` and `}` as `{{ "}}" }}`
  inside `justfile` recipes.
- **Go/container image bumps** follow a detailed checklist in
  `.agents/skills/bump/SKILL.md`.
