# Memos Anywhere

Multiplatform builds for [Memos](https://github.com/usememos/memos), a beautiful, lightweight, and privacy-first note-taking service.

Some of these builds are utilized by [Memospot](https://github.com/memospot/memospot), a self-contained desktop app for Memos, available on macOS, Linux and Windows.

[![Downloads](https://img.shields.io/github/downloads/memospot/memos-builds/total?logo=github)](https://github.com/memospot/memos-builds/releases) [![GitHub Stars](https://img.shields.io/github/stars/memospot/memos-builds?logo=github)](https://github.com/memospot/memos-builds)

<div align="center">

<a href="https://www.usememos.com/">
    <picture>
      <source
        media="(prefers-color-scheme: dark)"
        srcset="https://raw.githubusercontent.com/memospot/memos-builds/main/assets/capture_dark.webp"
      />
      <source
        media="(prefers-color-scheme: light)"
        srcset="https://raw.githubusercontent.com/memospot/memos-builds/main/assets/capture_light.webp"
      />
      <img
        alt="demo"
        src="https://www.usememos.com/demo.png"
      />
    </picture>
  </a>

[![Homepage](https://img.shields.io/badge/Home-blue)](https://www.usememos.com) [![Blog](https://img.shields.io/badge/Blog-gray)](https://www.usememos.com/blog) [![Docs](https://img.shields.io/badge/Docs-blue)](https://www.usememos.com/docs) [![Live Demo](https://img.shields.io/badge/Live-Demo-blue)](https://demo.usememos.com/) [![Memos Discord](https://img.shields.io/badge/Discord-chat-5865f2?logo=discord&logoColor=f5f5f5)](https://discord.gg/tfPJa4UmAv) [![GitHub Stars](https://img.shields.io/github/stars/usememos/memos?logo=github)](https://github.com/usememos/memos)

</div>

## Docker

This project provides optimized Memos images for the following platforms:

| amd64          | arm32        | other         |
| -------------- | ------------ | ------------- |
| linux/amd64    | linux/arm/v5 | linux/386     |
| linux/amd64/v2 | linux/arm/v6 | linux/ppc64le |
| linux/amd64/v3 | linux/arm/v7 | linux/riscv64 |
|                | linux/arm64  | linux/s390x   |

To use an image for a specific CPU architecture, add `--platform=<platform>` to the `docker` command line, before the image specifier. Read more at [Platform variants](#platform-variants).

> [!TIP]
> The optimizations include build flags, smaller images, and health check.

### Quick start

#### Docker run (latest)

```sh
docker run --detach --name memos --publish 5230:5230 \
  --volume ~/.memos/:/var/opt/memos lincolnthalles/memos:latest
```

#### Docker run (nightly)

```sh
docker run --detach --name memos-nightly --publish 5231:5230 \
  --volume ~/.memos-nightly/:/var/opt/memos lincolnthalles/memos:nightly
```

#### Docker run (throwaway nightly in demo mode)

```sh
docker run --detach --rm --name memos-throwaway --publish 5232:5230 \
  --env MEMOS_MODE=demo lincolnthalles/memos:nightly
```

#### Keeping containers up-to-date

Use [Watchtower](https://containrrr.dev/watchtower/).

```sh
docker run --detach --name watchtower \
  --volume /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower
```

### About images

- Versioned images are checked out to the matching Memos upstream tag.

- Nightly images use whatever is available at Memos `main` branch at build time.

- Image packages are auto-upgraded at build time.

- Nightly images are built daily at 00:25 UTC.

- Images are published at the same time to [Docker Hub](https://hub.docker.com/r/lincolnthalles/memos) and [GitHub Container Registry](https://github.com/memospot/memos-builds/pkgs/container/memos-builds).

| Platform  | Image                 |
| --------- | --------------------- |
| arm/v5    | busybox:1.36.1-uclibc |
| All other | alpine:3.21           |

## Platform variants

Multiple builds for `arm` and `amd64` platforms exists, with different hardware optimizations. Choose the build that best suits the host CPU.

Run `cat /proc/cpuinfo` and `uname -m` to find out your CPU model and architecture. For an `ARMv8` or `aarch64` CPU, use the ARM64 build.

⚠ Avoid using the `arm/v5` variant unless the host CPU can't handle anything newer. While it works, the lack of VFP hinders the performance of applications that were not specifically written to this architecture.

| Variant  | Target CPUs                                            |
| -------- | ------------------------------------------------------ |
| amd64    | Runs on all AMD64/Intel 64 CPUs. Also known as x86_64  |
| amd64/v2 | Intel Nehalem (1st gen from 2009) / AMD Jaguar (2013+) |
| amd64/v3 | Intel Haswell (4th gen) / AMD Excavator (2015+)        |
| arm/v5   | Older ARM without VFP (Vector Floating Point)          |
| arm/v6   | VFPv1 only: ARM11 or better cores                      |
| arm/v7   | VFPv3: Cortex-A cores                                  |
| arm64    | Recent ARM64/AArch64 CPUs                              |

## Notes

Linux binaries are packed with [UPX](https://upx.github.io/). This may trigger false positives on some antivirus software. You can unpack the binaries with `upx -d memos*`, if you will.

## Support

Memos official first-class [support](https://github.com/usememos/memos/issues) is for its [Docker container](https://hub.docker.com/r/neosmemo/memos).
These binaries and images are provided as a convenience for some specific use cases. They may work fine, and they may not. Use them at your discretion.

Please do not open issues on the official Memos repository regarding these builds, unless you can reproduce the issue on the official Docker container.

## Running as a Service

To start Memos at system boot, you must manually set up a system service.

[Memos Service Guide](docs/service.md)

[Memos Windows Service Guide](docs/windows-service.md)

## Running on Android

To run Memos using [Termux](https://play.google.com/store/apps/details?id=com.termux) on Android:

- Download a Linux build suiting device CPU architecture (most modern devices are `arm64`).

- Extract the downloaded file and copy the `memos` binary to internal storage.

Run this on Termux:

```bash
# This will prompt you for storage access permission
termux-setup-storage

# Copy the binary to Termux home directory
cp ~/storage/shared/memos .

# Make it executable
chmod +x ./memos

# Run Memos
MEMOS_MODE=prod MEMOS_DATA=. MEMOS_PORT=5230 ./memos

# Memos will be available at http://localhost:5230
# and on local network at http://<device-ip>:5230
```

⚠ As stated at [Termux Wiki](https://wiki.termux.com/wiki/Internal_and_external_storage), all data under Termux's home directory will be deleted if you uninstall the app.

## Supporting

If you appreciate this project, be sure to [⭐star](https://github.com/memospot/memos-builds) it on GitHub.
