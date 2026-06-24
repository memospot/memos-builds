---
name: bump
disable-model-invocation: true
description: Version bump checklists for Go and container images.
---

# bump

## Go bump

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
   - `GO_VERSION = "X.Y.Z"`
   - `GOLANG_BUILD_IMAGE = "golang:X.Y.Z-alpine"`

4. Sync Go workspace:

   ```bash
   just tidy X.Y.Z
   ```

5. Commit:

   ```bash
   git commit -s -m "chore: bump to go <major.minor>"
   ```

   Example: `chore: bump to go 1.26`

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
