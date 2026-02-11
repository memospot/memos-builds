# Building with Dagger

For automated, reproducible builds using Dagger, see the [Dagger Pipeline Documentation](../.dagger/README.md).

## Quick Reference

```bash
# Build binaries for current platform
dagger call build --source=. --version=nightly export --path=./dist

# Build for specific platforms
dagger call build --source=. --version=v0.26.1 --platforms=linux/amd64 export --path=./dist

# Build containers
dagger call build-containers --source=. export --path=./containers
```

## Prerequisites

- [Docker](https://docs.docker.com/get-started/) or [Podman](https://podman.io/)
- [Dagger CLI](https://docs.dagger.io/install) (v0.19+)

## Cross-Platform Emulation

To run images built for different architectures:

```bash
docker run --privileged --rm tonistiigi/binfmt --install all
```

## Simpler Alternative

For most use cases, use the [justfile](../justfile) commands instead:

```bash
just build          # Build binaries
just build-docker   # Build and test containers
```
