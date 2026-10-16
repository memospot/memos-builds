package buildplan

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var commitHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

// knownVersionCommits maps versions to specific commit hashes.
// Used to address upstream inconsistencies in tagged releases.
var knownVersionCommits = map[string]string{
	"0.25.2": "bfad0708e2c8062664e852f6f18223fd943ad5f5",
	"0.25.3": "07a030ddfdbe5ac8a22c235be7b5771cc01f8498",
	"0.26.0": "43b5a51ec73214d3c56aa48c82783ccfeec1a127",
	"0.26.1": "b623162d37f87f9f174d8f6cd8e54c7034cfc789",
}

// RefKind tells the executor how to resolve the git ref named by a plan.
type RefKind int

const (
	// RefCommit resolves a full commit hash.
	RefCommit RefKind = iota
	// RefTag resolves a git tag.
	RefTag
	// RefBranch resolves a branch tip.
	RefBranch
)

// SourceOverride says what a version declared in the source code replaces.
type SourceOverride int

const (
	// OverrideNone ignores any declared source version.
	OverrideNone SourceOverride = iota
	// OverrideBuild replaces only the build version.
	OverrideBuild
	// OverrideBoth replaces the build and release versions.
	OverrideBoth
)

// VersionPlan is the pure decision of which git ref a build comes from and
// which versions it carries. The executor resolves Ref and then calls Finalize.
type VersionPlan struct {
	Now            time.Time
	Ref            string
	LiteralBuild   string
	LiteralRelease string
	RefKind        RefKind
	SourceOverride SourceOverride
	// MintDev synthesises an upstream-style dev version from the resolved
	// commit when no declared source version overrides it.
	MintDev bool
}

// PlanVersion decides how to resolve a version input without touching the
// engine: which git ref to resolve, which versions to carry, and whether to
// mint a dev version.
//
// Accepted inputs: legacy `v0.x` tags, bare or `v`-prefixed calendar tags
// (`26.10.0`, `26.10-rc.1`), bare or `v`-prefixed dev versions
// (`26.10.15-dev.1511fd1`), 40-char commit hashes, or empty (defaults to a
// dev build from upstream `main`).
func PlanVersion(input string, now time.Time) (VersionPlan, error) {
	plan := VersionPlan{Now: now}

	// Normalise version input to a git tag, preserving the literal segments.
	// Upstream calendar tags are bare (e.g. 26.10.0, 26.10-rc.1); legacy 0.x
	// tags keep the prefix (e.g. v0.25.3). A leading "v" on calendar versions
	// is accepted for convenience. The literal form is preserved throughout:
	// Masterminds semver coerces numeric segments to integers and re-renders
	// them, turning 26.09.0 into the nonexistent tag 26.9.0. Anything else is
	// rejected: silently building main for a typo'd version ships the wrong
	// code under the wrong tag.
	bare := strings.TrimPrefix(input, "v")
	var tag, release string
	switch {
	case LegacyVersionPattern.MatchString(input):
		tag, release = input, input
	case LegacyVersionPattern.MatchString("v" + bare):
		tag, release = "v"+bare, "v"+bare
	case DevVersionPattern.MatchString(bare):
		// Dev versions are minted per commit, not tagged upstream: build from
		// main. The embedded version matches the input tag exactly, so
		// rebuilds are deterministic.
		plan.RefKind, plan.Ref = RefBranch, "main"
		plan.LiteralBuild, plan.LiteralRelease = "v"+bare, bare
		return plan, nil
	case CalVerRCPattern.MatchString(bare), CalVerStablePattern.MatchString(bare):
		tag, release = bare, bare
	}

	if tag != "" {
		// Keys are bare (e.g. "0.25.3") while release keeps the legacy "v"
		// prefix, so strip it for the lookup.
		if commitHash, ok := knownVersionCommits[strings.TrimPrefix(release, "v")]; ok {
			plan.RefKind, plan.Ref = RefCommit, commitHash
			plan.LiteralBuild, plan.LiteralRelease = release, release
			plan.SourceOverride = OverrideBuild
			return plan, nil
		}
		plan.RefKind, plan.Ref = RefTag, tag
		plan.LiteralBuild, plan.LiteralRelease = release, release
		return plan, nil
	}

	if input == "" || input == "dev" {
		// Default: dev build from upstream main, minted from the current UTC
		// date and the resolved commit's short SHA.
		plan.RefKind, plan.Ref = RefBranch, "main"
		plan.MintDev = true
		return plan, nil
	}
	if commitHashPattern.MatchString(input) {
		// Modern upstream sources carry no injected default; Finalize mints a
		// dev version so the binary passes upstream validation.
		plan.RefKind, plan.Ref = RefCommit, input
		plan.SourceOverride = OverrideBoth
		plan.MintDev = true
		return plan, nil
	}

	return VersionPlan{}, fmt.Errorf(`invalid version %q: expected a release tag (v0.25.3, 26.10.0, 26.10-rc.1), a dev version (26.10.15-dev.1511fd1), a 40-character commit hash, or "dev" to build upstream main`, input)
}

// Finalize applies the source-declared version and dev minting to the plan
// once the executor has resolved Ref. srcVersion/srcFound come from
// ParseDeclaredVersion; resolvedCommit is the commit the ref resolved to.
func (p VersionPlan) Finalize(srcVersion string, srcFound bool, resolvedCommit string) (buildVersion string, releaseVersion string, err error) {
	if p.SourceOverride != OverrideNone && srcFound {
		if p.SourceOverride == OverrideBoth {
			return srcVersion, srcVersion, nil
		}
		return srcVersion, p.LiteralRelease, nil
	}

	if p.MintDev {
		dev, mintErr := DevBuildVersion(p.Now, ShortCommitHash(resolvedCommit))
		if mintErr != nil {
			if p.RefKind == RefBranch {
				return "", "", fmt.Errorf("cannot synthesise dev version for %s@%q: %w", p.Ref, resolvedCommit, mintErr)
			}
			return "", "", fmt.Errorf("cannot synthesise dev version for %q: %w", p.Ref, mintErr)
		}
		return "v" + dev, dev, nil
	}

	return p.LiteralBuild, p.LiteralRelease, nil
}
