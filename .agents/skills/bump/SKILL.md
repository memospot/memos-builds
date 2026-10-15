---
name: bump
disable-model-invocation: true
description: Version bump checklists for Go and container images.
---

# bump

Two independent Go versions. Never sync them to a single value.

- **Dagger module Go** (`.dagger/go.mod`, `go.work`, `.go-version`): must match the
  Dagger engine's `go.mod`. Mismatch causes bootstrap errors like
  `existing go.mod has unsupported version 1.27 (highest supported version is 1.26.8)`.
- **Memos build Go** (`MEMOS_GO_VERSION` + `GOLANG_BUILD_IMAGE` in
  `.dagger/buildconsts/consts.go`): tracks upstream Memos. Applied inside the
  build container via `go mod tidy -go=<MEMOS_GO_VERSION>`.

## Dagger module Go bump

1. Get engine version:

   ```bash
   cat dagger.json | jq -r '.engineVersion'
   ```

2. Get the engine's Go version (source of truth — that's the target):

   ```bash
   curl -s "https://raw.githubusercontent.com/dagger/dagger/refs/tags/{engineVersion}/go.mod" | head -n 4 | grep -E '^(go|toolchain)'
   ```

   Match the `go X.Y.Z` line. If a `toolchain` line is present, note it but still
   pass the `go` line to `just tidy`.

3. Apply to the module only:

   ```bash
   just tidy X.Y.Z
   ```

   This updates `.dagger/go.mod`, `go.work`, and `.go-version`. It never touches
   the Memos build Go.

4. Commit:

   ```bash
   git commit -s -m "chore: bump dagger module to go <major.minor>"
   ```

   Example: `chore: bump dagger module to go 1.26`

## Memos build Go bump

1. Get upstream Go version:

   ```bash
   curl -s "https://raw.githubusercontent.com/usememos/memos/refs/heads/main/go.mod" | head -n 4 | grep go
   ```

   Match the `go X.Y.Z` line — that's the target.

2. Get latest Go alpine image:

   ```bash
   curl -s "https://hub.docker.com/v2/repositories/library/golang/tags?page_size=100" | jq -r '.results[] | select(.name | test("^[0-9]+\\.[0-9]+\\.[0-9]+-alpine$")) | .name' | sort -V | tail -n 1
   ```

   The image may not yet match the upstream Go version — use upstream as source of truth.

3. Update `.dagger/buildconsts/consts.go`:
   - `MEMOS_GO_VERSION = "X.Y.Z"`
   - `GOLANG_BUILD_IMAGE = "golang:X.Y.Z-alpine"`

   Do NOT run `just tidy` with this version — the Dagger module must stay on the
   engine's Go.

4. Commit:

   ```bash
   git commit -s -m "chore: bump memos to go <major.minor>"
   ```

   Example: `chore: bump memos to go 1.27`

## Container images bump

1. Get latest versions:

   ```bash
   # Alpine
   curl -s "https://hub.docker.com/v2/repositories/library/alpine/tags?page_size=100" | jq -r '.results[] | select(.name | test("^[0-9]+\\.[0-9]+(\\.[0-9]+)*$")) | .name' | sort -V | tail -n 1

   # BusyBox
   curl -s "https://hub.docker.com/v2/repositories/arm32v5/busybox/tags?page_size=100" | jq -r '.results[] | select(.name | test("^[0-9]+\\.[0-9]+(\\.[0-9]+)*-glibc$")) | .name' | sort -V | tail -n 1

   # buf
   curl -s "https://hub.docker.com/v2/repositories/bufbuild/buf/tags?page_size=100" | jq -r '.results[] | select(.name | test("^[0-9]+\\.[0-9]+(\\.[0-9]+)*$")) | .name' | sort -V | tail -n 1

   # Node (major only)
   curl -s "https://raw.githubusercontent.com/usememos/memos/refs/heads/main/web/package.json" | jq -r '.engines.node | capture("=\\s*(?<v>[0-9]+)") | .v'
   ```

2. Update `.dagger/buildconsts/consts.go`:
   - `PRIMARY_IMAGE = "alpine:X.Y.Z"`
   - `ALTERNATE_IMAGE = "arm32v5/busybox:X.Y.Z-glibc"`
   - `BUF_IMAGE = "bufbuild/buf:X.Y.Z"`
   - `NODE_IMAGE = "node:X-alpine"`
