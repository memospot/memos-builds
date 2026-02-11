// # Memos project build constants.
//
// Quick tooling updates can be done by changing the constants below.
//
// Deeper changes must be made by updating the related build steps.
package buildconsts

// Container base image to use for the project.
//
// Note: An Alpine image is expected at several build steps.
const PRIMARY_IMAGE string = "alpine:3.23"

// Container base image to use for the ARMv5 build.
//
// Note: The uclibc variant is smaller, but has issues with timezones.
const ALTERNATE_IMAGE string = "arm32v5/busybox:1.37-glibc"

// Container image to use for the Go build.
const GOLANG_BUILD_IMAGE string = "golang:1.25.7-alpine3.23"

// Container image to use for frontend builds.
const NODE_BUILD_IMAGE string = "node:22-alpine"

// Container image to use for proto builds.
const BUF_IMAGE string = "bufbuild/buf:1.65.0"

// Passed to `go mod tidy`.
const GO_VERSION string = "1.25.7"

// Where the semantic version is defined in the source code.
const VERSION_FILE string = "internal/version/version.go"

// The import path of the semantic version variable. Will be used at build time to override nightly versions to the format `v0.26.2-pre+b623162`.
const VERSION_IMPORT_PATH string = "github.com/usememos/memos/internal/version.Version"

// Passed to `go build` as the entrypoint of the application.
const APP_ENTRYPOINT string = "./cmd/memos/main.go"

// String format for the checksum file.
const CHECKSUM_FILE_FORMAT string = "memos-%s_SHA256SUMS.txt"
