# Dagger Build Pipeline

Reproducible, containerized build pipeline for [memospot/memos-builds](https://github.com/memospot/memos-builds).

## Prerequisites

- [Dagger CLI](https://docs.dagger.io/install) (v0.19+)
- Docker or a compatible container runtime

## Quick Start

```bash
# Clone the project
git clone https://github.com/memospot/memos-builds.git
cd memos-builds

# Build release artifacts for all targets (dev)
dagger call build --source=. --version=dev export --path=./dist

# Build for specific platforms only
dagger call build --source=. --version=26.10.0 --platforms=linux/amd64 export --path=./dist

# Build containers (Linux only) and export as tarballs
dagger call build-containers --source=. export --path=./containers

# Full publish (archives + container push)
dagger call publish --source=. --version=26.10.0 \
  --docker-hub-user=USER --docker-hub-password=env:DOCKER_TOKEN \
  --ghcr-user=USER --ghcr-password=env:GITHUB_TOKEN \
  export --path=./dist
```

## Function Map

```text
dagger call build
  ├── resolveVersion         # Determine git ref, resolve dev version
  │   ├── patchModerncSqlite # Fix libc/sqlite version mismatch
  │   └── applyPatches       # Apply local .patch files
  ├── generateProto          # buf generate (protobuf)
  ├── buildFrontend          # pnpm install + build (Node)
  ├── buildBackend           # Cross-compile Go binaries per target
  ├── createReleaseArchives  # tar.gz / zip per binary
  └── generateChecksums      # SHA256SUMS file

dagger call build-containers
  ├── resolveVersion
  ├── generateProto
  ├── buildFrontend
  ├── buildBackend           # Linux targets only
  └── buildContainer         # Per-platform container (Alpine or BusyBox)
      ├── buildAlpineContainer
      └── buildBusyBoxARMv5Container

dagger call publish
  ├── build                  # Full artifact pipeline (all targets)
  └── publishContainers      # Multi-arch push to Docker Hub + GHCR
      ├── resolveVersion
      ├── buildBackend       # Linux targets only
      └── buildContainer     # Shared with build-containers
```

## Parameters

| Function           | Parameter               | Default  | Description                                                |
| ------------------ | ----------------------- | -------- | ---------------------------------------------------------- |
| `build`            | `--source`              | `.`      | Host source directory                                      |
|                    | `--version`             | `dev`    | Git ref: tag (`26.10.0`, `v0.25.3`), commit hash, or `dev` |
|                    | `--platforms`           | all      | `all`, or comma-separated: `linux/amd64,darwin/arm64`      |
| `build-containers` | `--source`              | `.`      | Host source directory                                      |
|                    | `--version`             | `dev`    | Same as `build`                                            |
|                    | `--platforms`           | all      | Same as `build`; non-Linux entries are silently ignored    |
| `publish`          | `--source`              | `.`      | Host source directory                                      |
|                    | `--version`             | required | Git tag for the release                                    |
|                    | `--docker-hub-user`     | —        | Docker Hub username                                        |
|                    | `--docker-hub-password` | —        | Docker Hub token (use `env:VAR`)                           |
|                    | `--ghcr-user`           | —        | GHCR username                                              |
|                    | `--ghcr-password`       | —        | GHCR token (use `env:VAR`)                                 |

## Build Targets

Targets are defined in `buildplan.Targets` (`.dagger/buildplan/targets.go`). Each entry is a `{OS, Arch, ArchLevel}` tuple.

Uncomment entries to enable additional platforms. The full matrix includes Linux, Darwin, Windows, and FreeBSD across amd64, arm64, arm, 386, ppc64le, riscv64, and s390x.

## Output Structure

`dagger call build` produces:

```bash
memos-26.10.0-linux-x86_64.tar.gz
memos-26.10.0-darwin-arm64.tar.gz
memos-26.10.0-windows-x86_64.zip
memos-26.10.0_SHA256SUMS.txt
```

`dagger call build-containers` produces:

```bash
memos-linux-amd64.tar        # OCI tarball
memos-linux-arm64.tar
memos-linux-arm-v7.tar
```

## Maintenance Guide

### Updating for upstream changes

1. **Dagger module Go**: must match the Dagger engine's `go.mod` (`cat dagger.json | jq -r .engineVersion`, then fetch that tag's `go.mod`). Apply with `just tidy X.Y.Z` — updates `.dagger/go.mod`, `go.work`, `.go-version` only.
2. **Memos build Go**: tracks upstream Memos. Update `MEMOS_GO_VERSION` and `GOLANG_BUILD_IMAGE` in `buildconsts/consts.go`. Do NOT run `just tidy` with this version.
3. **Node version**: Update `NODE_BUILD_IMAGE` in `buildconsts/consts.go`.
4. **Protobuf tooling**: Update `BUF_IMAGE` in `buildconsts/consts.go`.
5. **Base container image**: Update `PRIMARY_IMAGE` in `buildconsts/consts.go`.
6. **Version path**: If the upstream project moves the version variable, update `VERSION_FILE` and `VERSION_IMPORT_PATH` in `buildconsts/consts.go`.

### Adding/removing platforms

Edit the `Targets` slice in `.dagger/buildplan/targets.go`. Each entry maps to:

- A cross-compiled binary (via `buildBackend`)
- A release archive (via `createReleaseArchives`)
- A container image if Linux (via `buildContainer`)

### SQLite/libc patching

The `sqliteLibcMap` in `patch.go` pins known-good `modernc.org/libc` versions for each `modernc.org/sqlite` release. When the upstream project bumps SQLite:

1. Check if the new version is already in the map.
2. If not, the build will attempt to fetch the correct libc version from upstream automatically.
3. After confirming, add the mapping to `sqliteLibcMap` for reproducibility.

### Applying custom patches

Place `.patch` files in the `patches/` directory. They are applied (via `git apply` with `patch` fallback) to the upstream source after checkout.

### Regenerating Dagger bindings

After changing any public function signature:

```bash
dagger develop --compat=0.19.10
```

## File Layout

```text
.dagger/
├── main.go          # Entrypoints: Build, BuildContainers, Publish
├── build.go         # generateProto, buildFrontend, buildBackend
├── container.go     # buildContainer, buildAlpineContainer, buildBusyBoxARMv5Container
├── publish.go       # Archives, checksums, container publishing
├── patch.go         # SQLite/libc patching, custom patch application
├── resolve.go       # Version resolution and git source preparation
├── buildplan/       # Pure build plan: version/tag rules, platform matrix, artifact names
│   ├── version.go   # Version, tag, and release-identity rules
│   ├── plan.go      # Version plan: ref selection and version policy
│   └── targets.go   # Platform matrix and artifact naming
└── buildconsts/
    └── consts.go    # All configurable build constants
```
