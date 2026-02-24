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
