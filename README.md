# Memos Anywhere

Multiplatform builds for [Memos](https://github.com/usememos/memos), a beautiful, lightweight, and privacy-first note-taking service.

Some of these builds are utilized by [Memospot](https://github.com/lincolnthalles/memospot), an independent desktop app for Memos available on macOS, Linux, and Windows.

<div align="center" width="100%" style="display: flex; justify-content: center;">
  <p align="center" width="100%">

[![Downloads](https://img.shields.io/github/downloads/lincolnthalles/memos-builds/total?logo=github)](https://github.com/lincolnthalles/memos-builds/releases) [![GitHub Stars](https://img.shields.io/github/stars/lincolnthalles/memos-builds?logo=github)](https://github.com/lincolnthalles/memos-builds)

  </p>
</div>

<p align="center" width="100%">
  <a href="https://www.usememos.com/">
    <picture>
      <source
        media="(prefers-color-scheme: dark)"
        srcset="assets/powered_by_memos_dark.webp"
      />
      <source
        media="(prefers-color-scheme: light)"
        srcset="assets/powered_by_memos.webp"
      />
      <img height="128"
        alt="powered by memos"
        src="assets/powered_by_memos.webp"
      />
    </picture>
  </a>
</p>

<div align="center" width="100%" style="display: flex; justify-content: center;">
  <p align="center" width="100%">

[![Homepage](https://img.shields.io/badge/Home-blue)](https://www.usememos.com) [![Blog](https://img.shields.io/badge/Blog-gray)](https://www.usememos.com/blog) [![Docs](https://img.shields.io/badge/Docs-blue)](https://www.usememos.com/docs) [![Live Demo](https://img.shields.io/badge/Live-Demo-blue)](https://demo.usememos.com/) [![Memos Discord](https://img.shields.io/badge/Discord-chat-5865f2?logo=discord&logoColor=f5f5f5)](https://discord.gg/tfPJa4UmAv) [![GitHub Stars](https://img.shields.io/github/stars/usememos/memos?logo=github)](https://github.com/usememos/memos)

  </p>
</div>

## Docker

This project provides optimized Memos images for the following platforms:
|      amd64     |     arm32    |     other     |
| -------------- | ------------ | ------------- |
|  linux/amd64   | linux/arm/v5 |   linux/386   |
| linux/amd64/v2 | linux/arm/v6 |  linux/arm64  |
| linux/amd64/v3 | linux/arm/v7 | linux/ppc64le |
|                |              | linux/riscv64 |
|                |              |  linux/s390x  |

To use an image for a specific CPU architecture, add `--platform=<platform>` to the `docker` command line, before the image specifier. Read more at [Platform variants](#platform-variants)

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

Please be aware that Memos does not currently follow a consistent versioning process. This means that sometimes a new release may include changes without updating the patch version. As a result, extra effort may be required to keep builds up to date, and there may be differences when compared to official images.

- Versioned images are checked out to Memos's upstream `release/version` branch.

- Nightly images use whatever is available at Memos's `main` branch at build time.

- Image packages are auto-upgraded at build time.

- Nightly images are built daily at 00:00 UTC.

- Images are published at the same time to [Docker Hub](https://hub.docker.com/r/lincolnthalles/memos) and [GitHub Container Registry](https://github.com/lincolnthalles/memos-builds/pkgs/container/memos-builds).

|  Platform |         Image         |
| --------- | --------------------- |
|  arm/v5   | busybox:stable-uclibc |
|  riscv64  |      alpine:edge      |
| All other |     alpine:latest     |

> Up to v0.19.0, `arm32v5` images were based on debian:stable-slim.

## Platform variants

There are multiple builds for `arm` and `amd64` platforms, with different hardware optimizations. Choose the one that best suits the host CPU.

Run `cat /proc/cpuinfo` and `uname -m` to find out your CPU model and architecture. For an `ARMv8` or `aarch64` CPU, use the ARM64 build.

⚠ Avoid using the `arm/v5` variant unless the host CPU can't handle anything newer. While it works, the lack of VFP hinders the performance of several applications.

### amd64

| Suffix | Target CPUs                                            |
| ------ | ------------------------------------------------------ |
| v1     | Runs on all AMD64/Intel 64 CPUs                        |
| v2     | Intel Nehalem (1st gen from 2009) / AMD Jaguar (2013+) |
| v3     | Intel Haswell (4th gen) / AMD Excavator (2015+)        |

### arm

| Suffix | Target CPUs                                   |
| ------ | --------------------------------------------- |
| v5     | Older ARM without VFP (Vector Floating Point) |
| v6     | VFPv1 only: ARM11 or better cores             |
| v7     | VFPv3: Cortex-A cores                         |

## Notes

Linux binaries are packed with [UPX](https://upx.github.io/). This may trigger false positives on some antivirus software. You can unpack the binaries with `upx -d memos*`, if you will.

It's currently not possible to build Memos for Windows i386 and any sort of MIPS architecture, because [modernc.org/libc](https://pkg.go.dev/modernc.org/sqlite#hdr-Supported_platforms_and_architectures) (used by SQLite driver) is not compatible with these targets.

## Support

Memos official first-class [support](https://github.com/usememos/memos/issues) is for its [Docker container](https://hub.docker.com/r/neosmemo/memos).
These binaries and images are provided as a convenience for some specific use cases. They may work fine, and they may not. Use them at your own discretion.

Please do not open issues on the official Memos repository regarding these builds, unless you can reproduce the issue on the official Docker container.

## Running as a Service

You should manually set up a system service to start Memos on boot.

[Memos Service Guide](docs/service.md)

[Memos Windows Service Guide](docs/windows-service.md)

## Running on Android

To run Memos using [Termux](https://play.google.com/store/apps/details?id=com.termux) on Android:

- Download a Linux build suiting device CPU architecture (most modern devices are `arm64`)

- Extract the downloaded file and copy the `memos` binary to internal storage

Run this on Termux:

```sh
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

If you like this project, don't forget to [⭐star](https://github.com/lincolnthalles/memos-builds) it and consider sponsoring it.
