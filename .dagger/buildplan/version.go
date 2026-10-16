// Package buildplan holds the build's static specification: the platform
// matrix and the version, tag, and artifact-naming rules. It imports no
// Dagger SDK types, so its tests run with a plain `go test`.
package buildplan

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

var (
	// DevVersionPattern mirrors the upstream Go validator for development
	// builds: YY.MM.D-dev.<sha7> (day without leading zero, lowercase hex).
	DevVersionPattern = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])\.([1-9][0-9]*)-dev\.([0-9a-f]{7})$`)

	// LegacyVersionPattern matches legacy `v0.x` git tags.
	LegacyVersionPattern = regexp.MustCompile(`^v0\.[0-9]+\.[0-9]+$`)

	// CalVerStablePattern matches calendar releases with an optional patch
	// component: YY.MM[.N]. Upstream never tags bare YY.MM, but the old
	// validator accepts it, so it stays accepted here.
	CalVerStablePattern = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])(\.(0|[1-9][0-9]*))?$`)

	// CalVerRCPattern matches release candidates YY.MM[.N]-rc.K, including
	// the live 26.10-rc.1 form. Upstream tags candidates bare.
	CalVerRCPattern = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])(\.(0|[1-9][0-9]*))?-rc\.[1-9][0-9]*$`)

	shortSHA7Pattern = regexp.MustCompile(`^[0-9a-fA-F]{7}$`)
)

// ShortCommitHash returns the 7-char short commit hash, or an empty string
// when the input cannot carry one.
func ShortCommitHash(commit string) string {
	if len(commit) < 7 {
		return ""
	}

	short := commit[:7]
	if !shortSHA7Pattern.MatchString(short) {
		return ""
	}
	return short
}

// DevBuildVersion returns the upstream-style development version string for
// the given UTC date and short SHA: YY.MM.D-dev.<sha7> (e.g.
// 26.10.15-dev.1511fd1). The returned string is validated against the upstream
// Go validator, never against Masterminds semver: Masterminds coerces numeric
// segments to uint64 and re-renders them, so 26.01.15-dev.… becomes the
// invalid 26.1.15-dev.…. The SHA must already be exactly 7 hex chars;
// anything else returns an error instead of an invalid version.
func DevBuildVersion(now time.Time, shortSHA string) (string, error) {
	sha := strings.ToLower(shortSHA)
	if !shortSHA7Pattern.MatchString(sha) {
		return "", fmt.Errorf("invalid short SHA %q: expected 7 hex chars", shortSHA)
	}
	version := fmt.Sprintf("%02d.%02d.%d-dev.%s", now.Year()%100, int(now.Month()), now.Day(), sha)
	if !DevVersionPattern.MatchString(version) {
		return "", fmt.Errorf("invalid dev version %q: fails upstream validation", version)
	}
	return version, nil
}

// IsStableRelease reports whether a version denotes a stable GitHub release:
// legacy `v0.x` or calendar `YY.MM.N`. Release candidates and dev builds are
// prereleases, never stable. Bare `YY.MM` is accepted as an input alias but
// is not itself a release.
func IsStableRelease(releaseVersion string) bool {
	if LegacyVersionPattern.MatchString(releaseVersion) {
		return true
	}
	m := CalVerStablePattern.FindStringSubmatch(releaseVersion)
	return len(m) > 3 && m[3] != ""
}

// IsReleaseCandidate reports whether a version is an upstream release
// candidate: YY.MM[.N]-rc.K.
func IsReleaseCandidate(version string) bool {
	return CalVerRCPattern.MatchString(strings.TrimPrefix(version, "v"))
}

// ContainerTags returns tags for release images.
//
// CalVer releases (26.10.0): ["latest", "26.10", "26.10.0"]
// Legacy releases (v0.25.3): ["latest", "0.25", "0.25.3"]
// Release candidates (26.10.0-rc.1, 26.10-rc.1): the bare version itself.
// Dev builds (26.10.15-dev.1511fd1): ["26.10.15-dev.1511fd1", "dev"]
func ContainerTags(version string) []string {
	bare := strings.TrimPrefix(version, "v")
	switch {
	case LegacyVersionPattern.MatchString(version):
		v, err := semver.NewVersion(strings.TrimPrefix(version, "v"))
		if err != nil {
			return []string{"latest"}
		}
		return []string{
			"latest",
			fmt.Sprintf("%d.%d", v.Major(), v.Minor()),
			fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch()),
		}
	case IsStableRelease(bare) || CalVerStablePattern.MatchString(bare):
		parts := strings.Split(bare, ".")
		if len(parts) == 2 {
			return []string{"latest", bare}
		}
		v, err := semver.NewVersion(bare)
		if err != nil {
			return []string{"latest"}
		}
		return []string{
			"latest",
			fmt.Sprintf("%d.%d", v.Major(), v.Minor()),
			fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch()),
		}
	case IsReleaseCandidate(bare):
		return []string{bare}
	case DevVersionPattern.MatchString(bare):
		return []string{bare, "dev"}
	default:
		return []string{"dev"}
	}
}

// TagsForRegistry returns the tags to push to a specific registry.
func TagsForRegistry(version string, registry string) []string {
	bare := strings.TrimPrefix(version, "v")

	// Dev builds: dated dev tag on GHCR plus a floating alias on both registries.
	if DevVersionPattern.MatchString(bare) {
		if strings.HasPrefix(registry, "ghcr.io") {
			return []string{bare, "dev"}
		}
		return []string{"dev"}
	}

	if IsStableRelease(version) || LegacyVersionPattern.MatchString(version) {
		return ContainerTags(version)
	}

	if IsReleaseCandidate(version) {
		return []string{bare}
	}

	return []string{"dev"}
}
