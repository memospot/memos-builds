<!-- markdownlint-disable blanks-around-headings blanks-around-lists no-duplicate-heading -->

# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!--
Types of changes
----------------
Added: for new features.
Changed: for changes in existing functionality.
Deprecated: for soon-to-be removed features.
Removed: for now removed features.
Fixed: for any bug fixes.
Security: in case of vulnerabilities.
-->

<!-- next-header -->

## [Unreleased] - ReleaseDate

### Added

- Daily development builds now follow upstream calendar versioning (`YY.MM.D-dev.<sha7>`), so they start without extra patching.
- Release candidate builds (for example, `26.10.0-rc.1`) get their own prerelease with matching container tags.

### Changed

- **Breaking:** daily development releases and container tags moved from `nightly-YYYYMMDD-<sha>` to upstream-style `YY.MM.D-dev.<sha7>` versions, with a floating `dev` container tag instead of `nightly`.
- **Breaking:** future stable and candidate tags drop the `v` prefix to match upstream (for example, `26.10.0` instead of `v26.10.0`); legacy `v0.x` tags are unchanged.
- **Breaking:** calendar build inputs require the patch component (`26.10.0`, `26.10.0-rc.1`), matching upstream's release tooling; unpatched forms (`26.10`, and the historical `26.10-rc.1` tag) are rejected instead of failing mid-build. The `26.10` container alias is still published from `26.10.N` releases.
- Old daily development releases are now pruned by release date instead of by tag name.

### Removed

- Dropped the `release/` branch build path; upstream no longer uses it.

### Fixed

- Fixed daily builds failing to start with `invalid build version` after upstream moved to calendar versioning.
- Fixed release names disagreeing with the published container tags when a dated development build was requested.
- Fixed the test job failing on every commit: the tests no longer need a live Dagger session, so a plain `go test` runs them anywhere.
- Fixed January–September builds failing to start with `invalid build metadata`: the embedded version keeps its zero-padded month (`26.09.0` no longer becomes `26.9.0`).
- Fixed calendar container tags losing the zero-padded month: a `26.09.0` release publishes `26.09` and `26.09.0`, never `26.9` and `26.9.0`.
- Fixed `TagsForRegistry` returning `dev` for `v`-prefixed calendar versions such as `v26.10.0`; stability is now decided on the bare version.
- Fixed `FilterTargets` aliasing the global target list for empty and `all` selections; it now returns a copy.
- Warn during the build when `modernc.org/libc` becomes a direct dependency, since the libc pin only rewrites the `// indirect` form.

## [0.26.0] - 2026-02-02

### Added

- (container) Can read an env file from `$MEMOS_DATA/memos.env` to pass environment variables to the container. This file have precedence over environment variables passed to the container.

- (container) `MEMOS_DSN` can be passed as a secret mount. It will be loaded automatically from the default Docker secret mount `/run/secrets/MEMOS_DSN`.

### Removed

- (container) Removed Health checks. Containers are now built strictly adhering to Open Container Initiative (OCI) image specification.

### Changed

- (container) Run as `nonroot` user by default. Group and user can be overriden by passing `PGID` and `PUID` environment variables to the container.

- (upstream) `MEMOS_MODE` is now retired. Database is always in `prod` mode unless `MEMOS_DEMO=true` is set.

- Builds are now using [Dagger](https://dagger.io).

<!-- next-url -->

[Unreleased]: https://github.com/memospot/memos-builds/compare/v0.26.0...HEAD
[0.26.0]: https://github.com/memospot/memos-builds/releases/tag/v0.26.0
