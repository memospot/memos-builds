# memos-builds

This project hosts builds for [Memos](https://github.com/usememos/memos), a beautiful, privacy-first, lightweight note-taking service.

Some of these builds are consumed by [Memospot](https://github.com/lincolnthalles/memospot), a self-contained, single user desktop client for Memos.


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
        srcset="assets/powered-by-memos_dark.webp"
      />
      <source
        media="(prefers-color-scheme: light)"
        srcset="assets/powered-by-memos.webp"
      />
    </picture>
  </a>
</p>

<div align="center" width="100%" style="display: flex; justify-content: center;">
  <p align="center" width="100%">

[![Homepage](https://img.shields.io/badge/Home-blue)](https://www.usememos.com) [![Blog](https://img.shields.io/badge/Blog-gray)](https://www.usememos.com/blog) [![Docs](https://img.shields.io/badge/Docs-blue)](https://www.usememos.com/docs) [![Live Demo](https://img.shields.io/badge/Live-Demo-blue)](https://demo.usememos.com/) [![Memos Discord](https://img.shields.io/badge/Discord-chat-5865f2?logo=discord&logoColor=f5f5f5)](https://discord.gg/tfPJa4UmAv) [![GitHub Stars](https://img.shields.io/github/stars/usememos/memos?logo=github)](https://github.com/usememos/memos)

  </p>
</div>

## Notes

All `amd64` binaries are built with SSE4.2 support, requiring at least a CPU launched on 2009 for Intel (first gen. Core), and 2013 for AMD.

Most binaries are packed with [UPX](https://upx.github.io/). This may trigger false-positives on some antivirus software. You can unpack the binaries with `upx -d memos*`, if you will.

⚠ It's currently not possible to build Memos for Windows i386 and any sort of MIPS architecture, because [modernc.org/libc](https://pkg.go.dev/modernc.org/sqlite#hdr-Supported_platforms_and_architectures) (used by SQLite driver) is not compatible with these targets.

## Support

⚠ Memos official first-class [support](https://github.com/usememos/memos/issues) is for the official Docker container.
These binaries are provided as a convenience for some specific use cases. They may work fine, and they may not. Use them at your own discretion.

Please do not open issues on the official Memos repository regarding these builds, unless you can reproduce the issue on the official Docker container.

## Running as a Service

You should manually setup a system service to start Memos on boot.
[Memos Windows Service Guide](https://github.com/usememos/memos/blob/main/docs/windows-service.md)

Sample service environment setup. Adjust to openrc or systemd as needed.
```sh
MEMOS_MODE="prod" # dev, prod, demo
MEMOS_PORT="5230"
MEMOS_ADDR="" # set this to 127.0.0.1 to restrict access
MEMOS_DATA="/opt/memos" # data directory
MEMOS_DRIVER="sqlite" # sqlite, mysql
MEMOS_DSN="" # database connection string
./memos

# Alternatively:
./memos --mode=prod --port=5230 --addr=127.0.0.1 --data=/opt/memos
```


## Star History

<picture>
  <source
    media="(prefers-color-scheme: dark)"
    srcset="
      https://api.star-history.com/svg?repos=usememos/memos,lincolnthalles/memos-builds,lincolnthalles/memospot&type=Date&theme=dark
    "
  />
  <source
    media="(prefers-color-scheme: light)"
    srcset="
      https://api.star-history.com/svg?repos=usememos/memos,lincolnthalles/memos-builds,lincolnthalles/memospot&type=Date
    "
  />
</picture>

## Support

If you like this project, don't forget to [⭐star](https://github.com/lincolnthalles/memos-builds) it and consider supporting my work:

<p align="center" width="100%">

  <a href="https://www.buymeacoffee.com/lincolnthalles">
    <img src="https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png" alt="Buy Me A Coffee" />
  </a>
</p>
