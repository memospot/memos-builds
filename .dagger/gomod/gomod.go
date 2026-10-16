// Package gomod holds the pure go.mod text rules for modernc.org/sqlite ↔
// libc pinning. It imports no Dagger SDK types, so its tests run with a plain
// `go test`.
//
// # Reasoning
//
// Fixes compilation failures and runtime errors on alternate platforms due to
// a mismatch between the version of `modernc.org/sqlite` and `modernc.org/libc`
// that happens when `go get -u` is run without this issue in mind.
//
// # See
//
//   - <https://pkg.go.dev/modernc.org/sqlite#hdr-Fragile_modernc_org_libc_dependency>
//
//   - <https://gitlab.com/cznic/sqlite/-/issues/177>
package gomod

import (
	"fmt"
	"regexp"
)

// sqliteLibcMap pins known-good modernc.org/libc versions for each
// modernc.org/sqlite release. Known versions from
// https://gitlab.com/cznic/sqlite/-/blob/master/go.mod
var sqliteLibcMap = map[string]string{
	"v1.37.0": "v1.62.1", // Memos v0.24.3
	"v1.37.1": "v1.65.8", // v0.24.4-v0.25.0
	"v1.38.2": "v1.66.3", // v0.25.1-v0.26.2
	"v1.46.1": "v1.67.6", // v0.26.3+
	"v1.50.0": "v1.72.0", // v0.29.0-v0.30.0
	"v1.56.0": "v1.74.4", // v0.30.0+
	"v1.57.0": "v1.74.4",
}

// Regex patterns for parsing go.mod
var (
	reSqlite       = regexp.MustCompile(`modernc\.org/sqlite\s+(v[\d.]+)`)
	reLibc         = regexp.MustCompile(`(modernc\.org/libc\s+)(v[\d.]+)(\s+//\s+indirect)`)
	reLibcUpstream = regexp.MustCompile(`modernc\.org/libc\s+(v[\d.]+)`)
)

// SqliteVersion extracts the modernc.org/sqlite version from go.mod contents.
// Returns the version string and true if found, empty string and false otherwise.
func SqliteVersion(contents string) (string, bool) {
	matches := reSqlite.FindStringSubmatch(contents)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}

// PinnedLibc returns the known-good modernc.org/libc version for a given
// modernc.org/sqlite version from the hardcoded map. The second return value
// reports whether the version is pinned; otherwise the caller falls back to an
// upstream fetch.
func PinnedLibc(sqliteVersion string) (string, bool) {
	libcVersion, ok := sqliteLibcMap[sqliteVersion]
	return libcVersion, ok
}

// UpstreamGoModURL returns the URL of the upstream modernc.org/sqlite go.mod
// for a given sqlite version, used as the fallback source for unpinned
// versions.
func UpstreamGoModURL(sqliteVersion string) string {
	return fmt.Sprintf("https://gitlab.com/cznic/sqlite/-/raw/%s/go.mod", sqliteVersion)
}

// LibcFromUpstreamGoMod extracts the modernc.org/libc version from the contents
// of an upstream modernc.org/sqlite go.mod.
// Returns the version string and true if found, empty string and false otherwise.
func LibcFromUpstreamGoMod(contents string) (string, bool) {
	matches := reLibcUpstream.FindStringSubmatch(contents)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}

// PatchLibc replaces the modernc.org/libc version in go.mod contents with the
// expected version, preserving the `// indirect` suffix.
// Returns the modified contents and true if a change was made, original contents and false otherwise.
func PatchLibc(contents, expectedVersion string) (string, bool) {
	matches := reLibc.FindStringSubmatch(contents)
	if len(matches) < 4 {
		return contents, false
	}

	currentVersion := matches[2]
	if currentVersion == expectedVersion {
		return contents, false
	}

	replacement := "${1}" + expectedVersion + "${3}"
	return reLibc.ReplaceAllString(contents, replacement), true
}
