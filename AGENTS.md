# AGENTS.md

Canonical instructions for AI assistants working on this codebase.

## What this repo is

A Dagger build pipeline that produces multi-architecture binaries and OCI
container images for [Memos](https://github.com/usememos/memos). Source is
fetched from upstream `usememos/memos` at build time; this repo owns the
build logic, container recipes, and the release/publish workflow. Some builds
are consumed by [Memospot](https://github.com/memospot/memospot), a desktop
app for Memos.

## Commands

| Action                                  | Command               |
| --------------------------------------- | --------------------- |
| Full validation (lint+test+tidy)        | `just validate`       |
| Lint                                    | `just lint`           |
| Format                                  | `just fmt`            |
| Build binaries for current platform     | `just build`          |
| Build containers (build + load locally) | `just build-docker`   |
| Start loaded containers in demo mode    | `just run-demos`      |
| Publish a release                       | `just publish TAG`    |
| Clean build artifacts + Docker cache    | `just clean`          |
| Clean Docker containers/images          | `just clean-docker`   |
| Clean Dagger engine state               | `just dagger-clean`   |
| Regenerate Dagger SDK bindings          | `just dagger-codegen` |
| Update Dagger SDK + Go deps             | `just update-deps`    |

Run `just --list` for the full recipe set. Most operations require Docker.
`just test` currently exercises nothing (no `*_test.go` under `.dagger/`).

## Structure

- `.dagger/` — Go Dagger module: `main.go` (public exports), `build.go`
  (compile/archive, proto gen, frontend build), `container.go` (OCI images),
  `patch.go` (sqlite fix + patch apply), `publish.go` (multi-arch release
  push), `buildconsts/consts.go` (base-image tags, Go/node versions),
  `lib.go` (BuildMatrix types, version resolution, target filtering).
- `container/` — Container assets: `entrypoint.sh` (ash) and per-arch dirs.
- `patches/` — Drop-in `*.patch` files applied at build time.
- `.github/workflows/` — CI (`ci.yml`), release (`dagger.yml`), audits.
- `justfile` — All convenience commands.

## Quirks and constraints

- **Go version pinned** in `.go-version` (currently `1.26.2`). CI and
  `just validate` enforce it. Change with `just tidy <version>`.
- **Go workspace** includes only `.dagger/`. Run `go work sync` after
  editing dependencies.
- **Formatter** is dprint (`.dprint.jsonc`): Go via `gofmt`, shell via
  `shfmt`, plus JSON/Markdown/YAML. `just fmt` runs dprint +
  `golangci-lint fmt`. `just lint` checks format first via
  `just --unstable --fmt --check`.
- **Linter scope** is narrow: `.golangci.yaml` excludes `.dagger/internal`
  and `.dagger/dagger.gen.go`. `golangci-lint` targets only `./.dagger/.`.
  Adding other `*.go` trees requires updating `.golangci.yaml` and CI.
- **Dagger SDK export rule**: PascalCase in `.dagger/main.go` → exported to
  `dagger functions`; camelCase → Dagger-internal only.
- **Dagger codegen**: run `just dagger-codegen` after public signature changes
  (`dagger develop --compat=skip`). This also removes the generated
  `.dagger/.gitignore` to prevent accidental commits of generated code.
- **All `dagger call` invocations MUST pass `--source=.`**. Dagger requires
  host paths explicitly. Artifacts land in `./dist/`.
- **`--version` accepts**: `vX.Y.Z` tag, `release/X.Y` branch, 40-char
  commit hash, or `nightly` (default — builds upstream `main`).
- **Container registries**: `docker.io/lincolnthalles/memos` and
  `ghcr.io/memospot/memos-builds`. Tags follow semver; nightly releases use
  `nightly-YYYYMMDD-<shortSHA>`.
- **Container entrypoint** is BusyBox `ash`, not bash. Keep
  `container/entrypoint.sh` POSIX/ash-compatible. `just lint` runs
  `shellcheck -s ash container/entrypoint.sh`.
- **OCI only**: Dagger supports standard OCI fields — no Docker-specific
  features like `HEALTHCHECK` or Dockerfile syntax.
- **Build matrix**: 24 platform targets in `.dagger/main.go` (Linux/Darwin/
  Windows/FreeBSD × amd64/arm/arm64/386/ppc64le/riscv64/s390x). Container
  builds target only Linux.
- **modernc.org/sqlite ↔ libc pinning** is handled programmatically in
  `.dagger/patch.go`. The hardcoded `sqliteLibcMap` must be kept current
  for new upstream versions; missing entries fall back to an upstream fetch.
  Drop-in `patches/*.patch` files are applied separately via `git apply`
  with `patch` fallback.
- **Just literal braces**: escape `{` as `{{ "{{" }}` and `}` as `{{ "}}" }}`
  inside `justfile` recipes.
- **No Go tests** are checked in (no `*_test.go` under `.dagger/`).
  `just test` currently exercises nothing. New tests belong in `.dagger/`.
- **Go/container image bumps** follow a detailed checklist in
  `.agents/skills/bump/SKILL.md`.
