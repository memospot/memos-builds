// # Patch utilities.
//
// Applies patches to the source code before building.
package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"regexp"
)

// sqlite-libc map: known versions from
// https://gitlab.com/cznic/sqlite/-/blob/master/go.mod
var sqliteLibcMap = map[string]string{
	"v1.37.0": "v1.62.1", // Memos v0.24.3
	"v1.37.1": "v1.65.8", // v0.24.4-v0.25.0
	"v1.38.2": "v1.66.3", // v0.25.1-v0.26.1
	"v1.39.0": "v1.66.3",
	"v1.39.1": "v1.66.10",
	"v1.40.0": "v1.66.10",
	"v1.40.1": "v1.66.10",
	"v1.42.2": "v1.66.10",
	"v1.43.0": "v1.66.10",
	"v1.44.0": "v1.67.4",
	"v1.44.1": "v1.67.6",
	"v1.44.2": "v1.67.6",
	"v1.44.3": "v1.67.6",
}

// Regex patterns for parsing go.mod
var (
	reSqlite       = regexp.MustCompile(`modernc\.org/sqlite\s+(v[\d.]+)`)
	reLibc         = regexp.MustCompile(`(modernc\.org/libc\s+)(v[\d.]+)(\s+//\s+indirect)`)
	reLibcUpstream = regexp.MustCompile(`modernc\.org/libc\s+(v[\d.]+)`)
)

// getSqliteVersion extracts the modernc.org/sqlite version from go.mod contents.
// Returns the version string and true if found, empty string and false otherwise.
func getSqliteVersion(goModContents string) (string, bool) {
	matches := reSqlite.FindStringSubmatch(goModContents)
	if len(matches) < 2 {
		return "", false
	}
	return matches[1], true
}

// getLibcVersionForSqlite returns the expected modernc.org/libc version for a given SQLite version.
//
// First checks the hardcoded map, then falls back to fetching from upstream.
// Returns an error if the version cannot be determined.
func getLibcVersionForSqlite(ctx context.Context, sqliteVersion string) (string, error) {
	// Try hardcoded map first.
	if libcVersion, ok := sqliteLibcMap[sqliteVersion]; ok {
		return libcVersion, nil
	}

	// Fallback: fetch from upstream.
	upstreamURL := fmt.Sprintf("https://gitlab.com/cznic/sqlite/-/raw/%s/go.mod", sqliteVersion)
	upstreamGoMod := dag.HTTP(upstreamURL)
	upstreamContents, err := upstreamGoMod.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("sqlite %s not in hardcoded map and upstream fetch failed (%w); please update sqliteLibcMap in patch.go", sqliteVersion, err)
	}

	matches := reLibcUpstream.FindStringSubmatch(upstreamContents)
	if len(matches) < 2 {
		return "", fmt.Errorf("sqlite %s: could not extract libc version from upstream go.mod. If this persists, something big has changed and this project will need intervention.", sqliteVersion)
	}

	return matches[1], nil
}

// patchLibcVersion replaces the modernc.org/libc version in go.mod contents.
// Returns the modified contents and true if a change was made, original contents and false otherwise.
func patchLibcVersion(goModContents, expectedVersion string) (string, bool) {
	matches := reLibc.FindStringSubmatch(goModContents)
	if len(matches) < 4 {
		return goModContents, false
	}

	currentVersion := matches[2]
	if currentVersion == expectedVersion {
		return goModContents, false
	}

	replacement := "${1}" + expectedVersion + "${3}"
	return reLibc.ReplaceAllString(goModContents, replacement), true
}

// Patch modernc.org/sqlite.
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
func (m *MemosBuilds) patchModerncSqlite(ctx context.Context, sourceCode *dagger.Directory) (*dagger.Directory, error) {
	goModContents, err := sourceCode.File("go.mod").Contents(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w. If this persists, check if the upstream project structure has changed.", err)
	}

	sqliteVersion, found := getSqliteVersion(goModContents)
	if !found {
		return sourceCode, nil
	}

	expectedLibcVersion, err := getLibcVersionForSqlite(ctx, sqliteVersion)
	if err != nil {
		return nil, err
	}

	newContents, modified := patchLibcVersion(goModContents, expectedLibcVersion)
	if !modified {
		return sourceCode, nil
	}

	result := sourceCode.WithNewFile("go.mod", newContents)
	return result, nil
}

// Apply diff patches to the source code.
func (m *MemosBuilds) applyPatches(ctx context.Context, source *dagger.Directory, patches *dagger.Directory) (*dagger.Directory, error) {
	if patches == nil {
		return source, nil
	}

	patches = patches.Filter(dagger.DirectoryFilterOpts{Include: []string{"*.patch"}})
	entries, err := patches.Entries(ctx)
	if err != nil || len(entries) == 0 {
		return source, nil
	}

	return dag.Container().
		From(buildconsts.PRIMARY_IMAGE).
		WithExec([]string{"apk", "add", "git", "patch"}).
		WithDirectory("/src", source).
		WithDirectory("/patches", patches).
		WithWorkdir("/src").
		WithExec([]string{"sh", "-c", `
			[ -d "/patches" ] || exit 0
			for patchfile in /patches/*.patch; do
				if [ -f "$patchfile" ]; then
					printf "-> Applying %s… " "$patchfile"
					if git -C /src apply "$patchfile" > /dev/null 2>&1; then
						printf "SUCCESS (via git apply)\n"
						continue
					fi
					if patch -d /src --fuzz=5 > /dev/null 2>&1 < "$patchfile"; then
						printf "SUCCESS (via patch)\n"
						continue
					fi
					printf "Failed! Continuing…\n"
				fi
			done
		`}).
		Directory("/src"), nil
}
