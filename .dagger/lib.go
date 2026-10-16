package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/buildplan"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

var (
	versionVarPattern = regexp.MustCompile(`var Version = "([^"]+)"`)

	// knownVersionCommits maps versions to specific commit hashes.
	// Used to address upstream inconsistencies in tagged releases.
	knownVersionCommits = map[string]string{
		"0.25.2": "bfad0708e2c8062664e852f6f18223fd943ad5f5",
		"0.25.3": "07a030ddfdbe5ac8a22c235be7b5771cc01f8498",
		"0.26.0": "43b5a51ec73214d3c56aa48c82783ccfeec1a127",
		"0.26.1": "b623162d37f87f9f174d8f6cd8e54c7034cfc789",
	}
)

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
		case buildplan.LegacyVersionPattern.MatchString(input):
			return input, input, nil
		case buildplan.LegacyVersionPattern.MatchString("v" + bare):
			return "v" + bare, "v" + bare, nil
		case buildplan.DevVersionPattern.MatchString(bare):
			return bare, bare, nil
		case buildplan.CalVerRCPattern.MatchString(bare):
			return bare, bare, nil
		case buildplan.CalVerStablePattern.MatchString(bare):
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
		if buildplan.DevVersionPattern.MatchString(release) {
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
		devVer, synErr := buildplan.DevBuildVersion(time.Now().UTC(), buildplan.ShortCommitHash(version))
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

	devVer, synErr := buildplan.DevBuildVersion(time.Now().UTC(), buildplan.ShortCommitHash(commit))
	if synErr != nil {
		return nil, "", "", "", fmt.Errorf("cannot synthesise dev version for main@%q: %w", commit, synErr)
	}
	return gitSrc, "v" + devVer, devVer, commit, nil
}
