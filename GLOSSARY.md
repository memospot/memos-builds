# Build Pipeline

The Dagger build pipeline for Memos: it resolves an upstream version, compiles
binaries, and produces archives, checksums, and OCI images.

## Language

**Build plan**:
The static specification of what a build produces: the platform target matrix
and the rules for version, tag, and artifact identity.
_Avoid_: build spec, build metadata

**Target**:
One platform the build produces a binary and, when Linux, a container for,
such as Linux amd64 v3 or Darwin arm64.
_Avoid_: platform entry, build matrix entry
