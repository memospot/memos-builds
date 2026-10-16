package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

var (
	versionVarPattern = regexp.MustCompile(`var Version = "([^"]+)"`)
	shortSHA7Pattern  = regexp.MustCompile(`^[0-9a-fA-F]{7}$`)

	// devVersionPattern mirrors the upstream Go validator for development
	// builds: YY.MM.D-dev.<sha7> (day without leading zero, lowercase hex).
	devVersionPattern = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])\.([1-9][0-9]*)-dev\.([0-9a-f]{7})$`)

	// legacyVersionPattern matches legacy `v0.x` git tags.
	legacyVersionPattern = regexp.MustCompile(`^v0\.[0-9]+\.[0-9]+$`)

	// calVerStablePattern matches calendar releases with an optional patch
	// component: YY.MM[.N]. Upstream never tags bare YY.MM, but the old
	// validator accepts it, so it stays accepted here.
	calVerStablePattern = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])(\.(0|[1-9][0-9]*))?$`)

	// calVerRCPattern matches release candidates YY.MM[.N]-rc.K, including
	// the live 26.10-rc.1 form. Upstream tags candidates bare.
	calVerRCPattern = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])(\.(0|[1-9][0-9]*))?-rc\.[1-9][0-9]*$`)

	// knownVersionCommits maps versions to specific commit hashes.
	// Used to address upstream inconsistencies in tagged releases.
	knownVersionCommits = map[string]string{
		"0.25.2": "bfad0708e2c8062664e852f6f18223fd943ad5f5",
		"0.25.3": "07a030ddfdbe5ac8a22c235be7b5771cc01f8498",
		"0.26.0": "43b5a51ec73214d3c56aa48c82783ccfeec1a127",
		"0.26.1": "b623162d37f87f9f174d8f6cd8e54c7034cfc789",
	}
)

type BuildMatrix struct {
	OS        string
	Arch      string
	ArchLevel string
}

// Docker platforms that do not support variants.
var DockerNoVariants []string = []string{"386", "ppc64le", "riscv64"}

// Platform returns a platform string in a format accepted by Docker.
//
// E.g. "linux/amd64", "linux/arm/v7".
func (m *BuildMatrix) DockerPlatform() string {
	if m.ArchLevel == "" || slices.Contains(DockerNoVariants, m.Arch) {
		return fmt.Sprintf("%s/%s", m.OS, m.Arch)
	}
	return fmt.Sprintf("%s/%s/%s", m.OS, m.Arch, m.ArchLevel)
}

// GoArm returns the GOARM environment variable value (e.g., "5", "6", "7")
// Returns empty string for non-ARM architectures.
func (m *BuildMatrix) GoArm() string {
	if m.Arch != "arm" {
		return ""
	}
	return strings.TrimPrefix(m.ArchLevel, "v")
}

// GoAmd64 returns the GOAMD64 environment variable value (e.g., "v1", "v2", "v3")
// Returns empty string for non-AMD64 architectures.
func (m *BuildMatrix) GoAmd64() string {
	if m.Arch != "amd64" {
		return ""
	}
	return m.ArchLevel
}

// Go386 returns the GO386 environment variable value (e.g., "sse2")
// Returns empty string for non-386 architectures
func (m *BuildMatrix) Go386() string {
	if m.Arch != "386" {
		return ""
	}
	return m.ArchLevel
}

// GoPpc64 returns the GOPPC64 environment variable value (e.g., "power8")
// Returns empty string for non-ppc64le architectures.
func (m *BuildMatrix) GoPpc64() string {
	if m.Arch != "ppc64le" {
		return ""
	}
	return m.ArchLevel
}

// GoRiscv64 returns the GORISCV64 environment variable value (e.g., "rva20u64")
// Returns empty string for non-riscv64 architectures
func (m *BuildMatrix) GoRiscv64() string {
	if m.Arch != "riscv64" {
		return ""
	}
	return m.ArchLevel
}

// UnameArch returns uname-compatible architecture name.
// (e.g., "x86_64", "i386", "arm64", "armv7l")
func (m *BuildMatrix) UnameArch() string {
	switch m.Arch {
	case "amd64":
		return "x86_64"
	case "386":
		return "i386"
	case "arm":
		return "armv" + m.GoArm() + "l"
	case "arm64":
		return "arm64"
	default:
		// ppc64le, riscv64, s390x stay as-is
		return m.Arch
	}
}

// BinaryName returns the binary filename for this target
// (e.g., "memos-linux-amd64v2", "memos-windows-arm64.exe")
func (m *BuildMatrix) BinaryName() string {
	name := fmt.Sprintf("memos-%s-%s", m.OS, m.Arch)
	if m.ArchLevel != "" {
		name += m.ArchLevel
	}
	if m.OS == "windows" {
		name += ".exe"
	}
	return name
}

// ArchiveName returns the archive filename for this target matching goreleaser format
// (e.g., "memos-v0.25.3-linux-x86_64.tar.gz", "memos-v0.25.3-windows-x86_64.zip")
func (m *BuildMatrix) ArchiveName(version string) string {
	arch := m.UnameArch()

	// Add amd64 level suffix if not v1
	if m.Arch == "amd64" && m.ArchLevel != "" && m.ArchLevel != "v1" {
		arch += "_" + m.ArchLevel
	}

	ext := "tar.gz"
	if m.OS == "windows" {
		ext = "zip"
	}

	return fmt.Sprintf("memos-%s-%s-%s.%s", version, m.OS, arch, ext)
}

// filterTargets returns a subset of TARGETS matching the given platforms string.
//
// Accepted formats:
//   - "" or "all": returns all TARGETS
//   - Comma-separated list: "linux/amd64,darwin/arm64,linux/arm/v7"
//
// Platform strings use the Docker format: os/arch or os/arch/variant.
func filterTargets(platforms string) ([]BuildMatrix, error) {
	platforms = strings.TrimSpace(platforms)
	if platforms == "" || platforms == "all" {
		return TARGETS, nil
	}

	requested := strings.Split(platforms, ",")
	var filtered []BuildMatrix

	for _, p := range requested {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		// Try exact match first (e.g., "linux/amd64/v3").
		found := false
		for _, t := range TARGETS {
			if t.DockerPlatform() == p {
				filtered = append(filtered, t)
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Fallback: match os/arch without variant, picking the first (base) entry.
		// This allows "linux/amd64" to match "linux/amd64/v1".
		parts := strings.Split(p, "/")
		if len(parts) == 2 {
			for _, t := range TARGETS {
				if t.OS == parts[0] && t.Arch == parts[1] {
					filtered = append(filtered, t)
					found = true
					break
				}
			}
		}

		if !found {
			var available []string
			for _, t := range TARGETS {
				available = append(available, t.DockerPlatform())
			}
			return nil, fmt.Errorf("platform %q not found in TARGETS (available: %s)", p, strings.Join(available, ", "))
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no valid platforms specified")
	}

	return filtered, nil
}

// filterLinuxTargets returns only Linux targets from the given slice.
func filterLinuxTargets(targets []BuildMatrix) []BuildMatrix {
	var linux []BuildMatrix
	for _, t := range targets {
		if t.OS == "linux" {
			linux = append(linux, t)
		}
	}
	return linux
}

// extractVersionFromSource reads the injected version from upstream source code.
// Returns "" when the source predates version injection (legacy 0.x with
// `var Version = "…"`) or carries no injected default.
func (m *MemosBuilds) extractVersionFromSource(ctx context.Context, src *dagger.Directory) (string, bool) {
	contents, err := src.File(buildconsts.VERSION_FILE).Contents(ctx)
	if err != nil {
		return "", false
	}

	match := versionVarPattern.FindStringSubmatch(contents)
	if len(match) < 2 {
		return "", false
	}

	versionFromFile := match[1]
	if _, err := semver.NewVersion(versionFromFile); err != nil {
		return "", false
	}

	return versionFromFile, true
}

// devBuildVersion returns the upstream-style development version string for
// the given UTC date and short SHA: YY.MM.D-dev.<sha7> (e.g.
// 26.10.15-dev.1511fd1). The returned string is validated against the upstream
// Go validator, never against Masterminds semver: Masterminds coerces numeric
// segments to uint64 and re-renders them, so 26.01.15-dev.… becomes the
// invalid 26.1.15-dev.…. The SHA must already be exactly 7 hex chars;
// anything else returns an error instead of an invalid version.
func devBuildVersion(now time.Time, shortSHA string) (string, error) {
	sha := strings.ToLower(shortSHA)
	if !shortSHA7Pattern.MatchString(sha) {
		return "", fmt.Errorf("invalid short SHA %q: expected 7 hex chars", shortSHA)
	}
	version := fmt.Sprintf("%02d.%02d.%d-dev.%s", now.Year()%100, int(now.Month()), now.Day(), sha)
	if !devVersionPattern.MatchString(version) {
		return "", fmt.Errorf("invalid dev version %q: fails upstream validation", version)
	}
	return version, nil
}

// isStableRelease reports whether a version denotes a stable GitHub release:
// legacy `v0.x` or calendar `YY.MM.N`. Release candidates and dev builds are
// prereleases, never stable. Bare `YY.MM` is accepted as an input alias but
// is not itself a release.
func isStableRelease(releaseVersion string) bool {
	if legacyVersionPattern.MatchString(releaseVersion) {
		return true
	}
	m := calVerStablePattern.FindStringSubmatch(releaseVersion)
	return len(m) > 3 && m[3] != ""
}

// isReleaseCandidate reports whether a version is an upstream release
// candidate: YY.MM[.N]-rc.K.
func isReleaseCandidate(version string) bool {
	return calVerRCPattern.MatchString(strings.TrimPrefix(version, "v"))
}

// isReleaseVersion reports whether a version gets a GitHub release at all:
// stable releases and release candidates. Dev builds and anything
// unparseable are excluded; commit hashes are handled by the caller.
func isReleaseVersion(releaseVersion string) bool {
	return isStableRelease(releaseVersion) || isReleaseCandidate(releaseVersion)
}

// resolveVersion determines the git source and build/release versions from user input.
//
// Accepted inputs: legacy `v0.x` tags, bare or `v`-prefixed calendar tags
// (`26.10.0`, `v26.10.0`, `26.10-rc.1`, `26.10.0-rc.1`), bare or `v`-prefixed
// dev versions (`26.10.15-dev.1511fd1`), 40-char commit hashes, or empty
// (defaults to a dev build from upstream `main` with an upstream-style
// `YY.MM.D-dev.<sha7>` build version, using the current UTC date).
func (m *MemosBuilds) resolveVersion(
	ctx context.Context,
	version string,
) (gitSrc *dagger.Directory, buildVersion string, releaseVersion string, commit string, err error) {
	git := dag.Git("https://github.com/usememos/memos.git")
	treeOpts := dagger.GitRefTreeOpts{Depth: 1}

	// Normalise version input to a git tag, preserving the literal segments.
	// Upstream calendar tags are bare (e.g. 26.10.0, 26.10-rc.1); legacy 0.x
	// tags keep the prefix (e.g. v0.25.3). A leading "v" on calendar versions
	// is accepted for convenience. The literal form is preserved throughout:
	// Masterminds semver coerces numeric segments to integers and re-renders
	// them, turning 26.09.0 into the nonexistent tag 26.9.0. Anything else is
	// rejected: silently building main for a typo'd version ships the
	// wrong code under the wrong tag.
	normaliseTag := func(input string) (tag string, release string, err error) {
		bare := strings.TrimPrefix(input, "v")
		switch {
		case legacyVersionPattern.MatchString(input):
			return input, input, nil
		case legacyVersionPattern.MatchString("v" + bare):
			return "v" + bare, "v" + bare, nil
		case devVersionPattern.MatchString(bare):
			return bare, bare, nil
		case calVerRCPattern.MatchString(bare):
			return bare, bare, nil
		case calVerStablePattern.MatchString(bare):
			return bare, bare, nil
		default:
			return "", "", fmt.Errorf("invalid version %q: expected a release tag, dev version, or commit hash", input)
		}
	}

	tag, release, err := normaliseTag(version)
	if err == nil {
		// Check if this version has a known commit hash to use instead of the tag.
		// Keys are bare (e.g. "0.25.3") while release keeps the legacy "v"
		// prefix, so strip it for the lookup.
		if commitHash, ok := knownVersionCommits[strings.TrimPrefix(release, "v")]; ok {
			gitSrc = git.Commit(commitHash).Tree(treeOpts)
			if srcVersion, found := m.extractVersionFromSource(ctx, gitSrc); found {
				return gitSrc, srcVersion, release, commitHash, nil
			}
			return gitSrc, release, release, commitHash, nil
		}
		if devVersionPattern.MatchString(release) {
			// Dev versions are minted per commit, not tagged upstream:
			// build from main. The embedded version matches the input tag
			// exactly, so rebuilds are deterministic.
			ref := git.Branch("main")
			mainCommit, commitErr := ref.Commit(ctx)
			if commitErr != nil {
				return nil, "", "", "", fmt.Errorf("failed to resolve main branch: %w", commitErr)
			}
			gitSrc = ref.Tree(treeOpts)
			return gitSrc, "v" + release, release, mainCommit, nil
		}
		ref := git.Tag(tag)
		tagCommit, commitErr := ref.Commit(ctx)
		if commitErr != nil {
			return nil, "", "", "", fmt.Errorf("failed to resolve tag %q: %w", tag, commitErr)
		}
		gitSrc = ref.Tree(treeOpts)
		// Build from the release tag itself: it satisfies upstream validation.
		return gitSrc, release, release, tagCommit, nil
	}

	if version != "" && version != "dev" && !commitHashPattern.MatchString(version) {
		return nil, "", "", "", err
	}

	if commitHashPattern.MatchString(version) {
		gitSrc = git.Commit(version).Tree(treeOpts)
		if srcVersion, found := m.extractVersionFromSource(ctx, gitSrc); found {
			return gitSrc, srcVersion, srcVersion, version, nil
		}
		// Modern upstream sources carry no injected default; synthesise a
		// dev version so the binary passes upstream validation.
		devVer, synErr := devBuildVersion(time.Now().UTC(), shortCommitHash(version))
		if synErr != nil {
			return nil, "", "", "", fmt.Errorf("cannot synthesise dev version for %q: %w", version, synErr)
		}
		return gitSrc, "v" + devVer, devVer, version, nil
	}

	// Default: dev build from upstream main, minted from the current UTC date
	// and the resolved commit's short SHA.
	ref := git.Branch("main")
	commit, commitErr := ref.Commit(ctx)
	if commitErr != nil {
		return nil, "", "", "", fmt.Errorf("failed to resolve main branch: %w", commitErr)
	}
	gitSrc = ref.Tree(treeOpts)

	devVer, synErr := devBuildVersion(time.Now().UTC(), shortCommitHash(commit))
	if synErr != nil {
		return nil, "", "", "", fmt.Errorf("cannot synthesise dev version for main@%q: %w", commit, synErr)
	}
	return gitSrc, "v" + devVer, devVer, commit, nil
}
