// # Patch utilities.
//
// Applies patches to the source code before building.
package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/gomod"
	"dagger/memos-builds/internal/dagger"
	"fmt"
)

// getLibcVersionForSqlite returns the expected modernc.org/libc version for a given SQLite version.
//
// First checks the hardcoded pins, then falls back to fetching from upstream.
// Returns an error if the version cannot be determined.
func getLibcVersionForSqlite(ctx context.Context, sqliteVersion string) (string, error) {
	if libcVersion, ok := gomod.PinnedLibc(sqliteVersion); ok {
		return libcVersion, nil
	}

	// Fallback: fetch from upstream.
	upstreamGoMod := dag.HTTP(gomod.UpstreamGoModURL(sqliteVersion))
	upstreamContents, err := upstreamGoMod.Contents(ctx)
	if err != nil {
		return "", fmt.Errorf("sqlite %s not in hardcoded map and upstream fetch failed (%w); please update sqliteLibcMap in .dagger/gomod", sqliteVersion, err)
	}

	libcVersion, found := gomod.LibcFromUpstreamGoMod(upstreamContents)
	if !found {
		return "", fmt.Errorf("sqlite %s: could not extract libc version from upstream go.mod. If this persists, something big has changed, and developer intervention is required", sqliteVersion)
	}

	return libcVersion, nil
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
		return nil, fmt.Errorf("failed to read go.mod: %w. If this persists, check if the upstream project structure has changed", err)
	}

	sqliteVersion, found := gomod.SqliteVersion(goModContents)
	if !found {
		return sourceCode, nil
	}

	expectedLibcVersion, err := getLibcVersionForSqlite(ctx, sqliteVersion)
	if err != nil {
		return nil, err
	}

	newContents, modified := gomod.PatchLibc(goModContents, expectedLibcVersion)
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
