// # Memos project build constants.
//
// Quick tooling updates can be done by changing the constants below.
//
// Deeper changes must be made by updating the related build steps.
package buildconsts

// Container base image to use for the project.
//
// Note: An Alpine image is expected at several build steps.
const PRIMARY_IMAGE string = "alpine:3.24.2"

// Container base image to use for the ARMv5 build.
//
// Note: The uclibc variant is smaller, but has issues with timezones.
const ALTERNATE_IMAGE string = "arm32v5/busybox:1.38.0-glibc"

// Container image to use for the Go build (Memos build Go toolchain).
const GOLANG_BUILD_IMAGE string = "golang:1.27.2-alpine"

// Container image to use for frontend builds.
const NODE_BUILD_IMAGE string = "node:24-alpine"

// Container image to use for proto builds.
const BUF_IMAGE string = "bufbuild/buf:1.73.0"

// Memos build Go version.
//
// Passed to `go mod tidy` inside the Memos build container (`build.go`).
// Tracks upstream Memos, independently of the Dagger module Go version
// (`.dagger/go.mod`, `go.work`, `.go-version`), which must match the
// Dagger engine's `go.mod` instead.
const MEMOS_GO_VERSION string = "1.27.2"

// Where the semantic version is defined in the source code.
const VERSION_FILE string = "internal/version/version.go"

// The import path of the semantic version variable.
const VERSION_IMPORT_PATH string = "github.com/usememos/memos/internal/version.Version"

// The import path of the source commit variable.
const COMMIT_IMPORT_PATH string = "github.com/usememos/memos/internal/version.Commit"

// Passed to `go build` as the entrypoint of the application.
const APP_ENTRYPOINT string = "./cmd/memos"

// String format for the checksum file.
const CHECKSUM_FILE_FORMAT string = "memos-%s_SHA256SUMS.txt"
