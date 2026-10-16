package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/buildplan"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"time"
)

// extractVersionFromSource reads the injected version from upstream source code.
// Returns "" when the source predates version injection (legacy 0.x with
// `var Version = "…"`) or carries no injected default.
func (m *MemosBuilds) extractVersionFromSource(ctx context.Context, src *dagger.Directory) (string, bool) {
	contents, err := src.File(buildconsts.VERSION_FILE).Contents(ctx)
	if err != nil {
		return "", false
	}

	return buildplan.ParseDeclaredVersion(contents)
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
	plan, err := buildplan.PlanVersion(version, time.Now().UTC())
	if err != nil {
		return nil, "", "", "", err
	}

	git := dag.Git("https://github.com/usememos/memos.git")
	treeOpts := dagger.GitRefTreeOpts{Depth: 1}

	switch plan.RefKind {
	case buildplan.RefCommit:
		gitSrc = git.Commit(plan.Ref).Tree(treeOpts)
		commit = plan.Ref
	case buildplan.RefTag:
		ref := git.Tag(plan.Ref)
		commit, err = ref.Commit(ctx)
		if err != nil {
			return nil, "", "", "", fmt.Errorf("failed to resolve tag %q: %w", plan.Ref, err)
		}
		gitSrc = ref.Tree(treeOpts)
	case buildplan.RefBranch:
		ref := git.Branch(plan.Ref)
		commit, err = ref.Commit(ctx)
		if err != nil {
			return nil, "", "", "", fmt.Errorf("failed to resolve %s branch: %w", plan.Ref, err)
		}
		gitSrc = ref.Tree(treeOpts)
	}

	srcVersion, srcFound := "", false
	if plan.SourceOverride != buildplan.OverrideNone {
		srcVersion, srcFound = m.extractVersionFromSource(ctx, gitSrc)
	}

	buildVersion, releaseVersion, err = plan.Finalize(srcVersion, srcFound, commit)
	if err != nil {
		return nil, "", "", "", err
	}

	return gitSrc, buildVersion, releaseVersion, commit, nil
}
