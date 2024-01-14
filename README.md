# memos-builds

This project hosts builds for [Memos](https://github.com/usememos/memos), a beautiful, privacy-first, lightweight note-taking service.

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

This side-project provides optimized images for `amd64v2`, `armv6`, `arm64` and `riscv64`. If you need support for legacy `amd64v1` CPUs, use the official `neosmemo/memos` image.


##### Latest

```sh
docker run -d --name memos -p 5230:5230 -v ~/.memos/:/var/opt/memos lincolnthalles/memos:latest
# or
docker run -d --name memos -p 5230:5230 -v ~/.memos/:/var/opt/memos ghcr.io/lincolnthalles/memos-builds:latest
```

##### Nightly

```sh
docker run -d --name memos-nightly -p 5231:5230 -v ~/.memos-nightly/:/var/opt/memos lincolnthalles/memos:nightly
# or
docker run -d --name memos-nightly -p 5231:5230 -v ~/.memos-nightly/:/var/opt/memos ghcr.io/lincolnthalles/memos-builds:nightly
```

[Docker Hub](https://hub.docker.com/r/lincolnthalles/memos)


## Platform variants

`arm` and `amd64` platforms have multiple builds, with different hardware optimizations. Choose the one that best suits the host CPU.

> Run `cat /proc/cpuinfo` and `uname -m` to find out your CPU model and architecture. For an `ARMv8` CPU, use the ARM64 build.

### amd64

| Suffix | Target CPUs                                       |
| ------ | ------------------------------------------------- |
|   v1   | Runs on all AMD64/Intel 64 CPUs                   |
|   v2   | Intel Nehalem (1st geN) / AMD Jaguar and newer    |
|   v3   | Intel Haswell (4th gen) / AMD Excavator and newer |

### arm

| Suffix | Target CPUs                       |
| ------ | ----------------------------------|
|   v5   | Older ARM without VFP             |
|   v6   | VFPv1 only: ARM11 or better cores |
|   v7   | VFPv3: Cortex-A cores             |

## ⚠ Notes

Linux binaries are packed with [UPX](https://upx.github.io/). This may trigger false-positives on some antivirus software. You can unpack the binaries with `upx -d memos*`, if you will.

It's currently not possible to build Memos for Windows i386 and any sort of MIPS architecture, because [modernc.org/libc](https://pkg.go.dev/modernc.org/sqlite#hdr-Supported_platforms_and_architectures) (used by SQLite driver) is not compatible with these targets.

## Support

Memos official first-class [support](https://github.com/usememos/memos/issues) is for its Docker container.
These binaries are provided as a convenience for some specific use cases. They may work fine, and they may not. Use them at your own discretion.

Please do not open issues on the official Memos repository regarding these builds, unless you can reproduce the issue on the official Docker container.

## Running as a Service

You should manually setup a system service to start Memos on boot.

[Memos Service Guide](docs/service.md)

[Memos Windows Service Guide](docs/windows-service.md)

## Running on Android
To run Memos using [Termux](https://play.google.com/store/apps/details?id=com.termux) on Android:

- Download a Linux build suiting device CPU architecture (most modern devices are `arm64`)

- Extract downloaded file and copy the `memos` binary to internal storage

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

⚠ As stated at [Termux Wiki](https://wiki.termux.com/wiki/Internal_and_external_storage), all data under Termux home directory will be deleted if you uninstall the app.

## Star History

<picture>
  <source
    media="(prefers-color-scheme: dark)"
    srcset="
      https://api.star-history.com/svg?repos=lincolnthalles/memos-builds&type=Date&theme=dark
    "
  />
  <source
    media="(prefers-color-scheme: light)"
    srcset="
      https://api.star-history.com/svg?repos=lincolnthalles/memos-builds&type=Date
    "
  />
  <img
    alt="Star History Chart"
    src="https://api.star-history.com/svg?repos=lincolnthalles/memos-builds&type=Date"
  />
</picture>

## Support

If you like this project, don't forget to [⭐star](https://github.com/lincolnthalles/memos-builds) it and consider supporting my work:

<p align="center" width="100%">

  <a href="https://www.buymeacoffee.com/lincolnthalles">
    <img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" />
  </a>
</p>
