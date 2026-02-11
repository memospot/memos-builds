# Building Memos from Source (manual)

This guide explains how to build Memos binaries from source for your platform or cross-compile for other platforms.

## Requirements

Before building Memos, ensure you have the following tools installed:

- [**Git**](https://git-scm.com/) - Version control system for cloning the repository
- [**Go 1.25+**](https://go.dev/dl/) - Go programming language compiler
- [**Node.js 22 LTS**](https://nodejs.org/en/download/) - JavaScript runtime for building the frontend
- [**pnpm**](https://pnpm.io/installation) - Fast, disk space efficient package manager

> [!TIP]
> You can install these tools via [homebrew](https://brew.sh/) or [winget](https://learn.microsoft.com/en-us/windows/package-manager/winget/)
>
> ```bash
> brew install git go@1.25 node@22 pnpm@9
> ```
>
> ```powershell
> winget install Git.Git GoLang.Go OpenJS.NodeJS.22 pnpm.pnpm
> ```

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/usememos/memos.git
cd memos
```

### 2. Checkout a Specific Version (Optional)

To build a specific release version:

```bash
# Checkout a specific tag
git checkout v0.25.2

# Or checkout a specific commit
git checkout <commit-hash>
```

### 3. Enable pnpm

If you haven't already enabled pnpm with Node.js:

```bash
corepack enable pnpm
```

### 4. Build the Frontend

Before compiling the Go code, you need to build the frontend assets:

```bash
# Install frontend dependencies
pnpm install

# Build the frontend for production
pnpm run release
```

> [!IMPORTANT]
> If you skip this step, you'll get a `No embeddable frontend found.` when accessing the web interface.

### 5. Build the Backend Binary

Now build the Go binary for your current platform:

```bash
CGO_ENABLED=0 go build -o memos ./bin/memos/main.go
```

> [!TIP]
> On PowerShell, you can set environment variables prefixing them with `$Env:`:
>
> ```powershell
> $Env:CGO_ENABLED = "0"; go build -o memos.exe ./bin/memos/main.go
> ```

The binary will be created as `memos` (or `memos.exe` on Windows) in the current directory.

## Cross-Platform Building

You can build binaries for different platforms by setting the `GOOS` and `GOARCH` environment variables.

> [!NOTE]
> Make sure you've completed the frontend build steps (`pnpm install` and `pnpm run release`) before cross-compiling, as the frontend assets need to be embedded in all binaries.

### Common Platform Examples

**Linux AMD64:**

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o memos-linux-amd64 ./bin/memos/main.go
```

**Windows AMD64:**

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o memos-windows-amd64.exe ./bin/memos/main.go
```

**macOS ARM64 (Apple Silicon):**

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o memos-darwin-arm64 ./bin/memos/main.go
```

**macOS AMD64 (Intel):**

```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o memos-darwin-amd64 ./bin/memos/main.go
```

**Linux ARM with specific variants:**

```bash
# ARMv5
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -o memos-linux-armv5 ./bin/memos/main.go

# ARMv6
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -o memos-linux-armv6 ./bin/memos/main.go

# ARMv7
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o memos-linux-armv7 ./bin/memos/main.go
```

## Environment Variables Reference

| Variable  | Description                   | Example Values                          |
|-----------|-------------------------------|-----------------------------------------|
| `GOOS`    | Target operating system       | `linux`, `windows`, `darwin`, `freebsd` |
| `GOARCH`  | Target architecture           | `amd64`, `arm64`, `arm`, `386`          |
| `GOAMD64` | AMD64 microarchitecture level | `v1` (default), `v2`, `v3`, `v4`        |
| `GOARM`   | ARM architecture version      | `5`, `6`, `7`                           |

## Build with Optimizations

For production builds with size and performance optimizations:

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o memos ./bin/memos/main.go
```

The `-ldflags="-s -w"` flags strip debug information and symbol tables, reducing binary size.

The `-trimpath` flag removes the build path from the binary, which is useful for security and portability.

## Verifying Your Build

After building, verify if the binary works:

```bash
./memos --demo=true --data=. --port=5230
```

Then open `http://localhost:5230` in your browser.

## Troubleshooting

### Issue: "command not found" errors

- Ensure Go, Node.js, and pnpm are properly installed and in your PATH
- Verify versions: `go version`, `node --version`, and `pnpm --version`
- If pnpm is not found, run `corepack enable pnpm`

### Issue: Build fails with dependency errors

- Run `go mod download` to ensure all Go dependencies are downloaded
- Run `pnpm install` to ensure all frontend dependencies are installed
- Check your Go version meets the minimum requirement (1.25+)
- Check your Node.js version is 22 LTS or compatible

## Next Steps

- [Configuration Guide](./configuration.md) - Learn how to configure Memos
- [Running as a Service (Linux)](./service-linux.md) - Set up Memos as a system service
- [Running as a Service (Windows)](./service-windows.md) - Set up Memos as a system service
