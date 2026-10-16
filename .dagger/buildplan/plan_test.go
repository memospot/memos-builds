package buildplan

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

const fortyCharHash = "1511fd125f1a2b3c4d5e6f70819293a4b5c6d7e8"

func TestPlanVersion(t *testing.T) {
	// GIVEN one input per accepted class, at a fixed instant
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name  string
		input string
		want  VersionPlan
	}{
		{
			"legacy tag resolves the tag",
			"v0.31.0",
			VersionPlan{RefKind: RefTag, Ref: "v0.31.0", LiteralBuild: "v0.31.0", LiteralRelease: "v0.31.0", Now: now},
		},
		{
			"bare legacy tag gains the v prefix",
			"0.31.0",
			VersionPlan{RefKind: RefTag, Ref: "v0.31.0", LiteralBuild: "v0.31.0", LiteralRelease: "v0.31.0", Now: now},
		},
		{
			"pinned legacy version resolves its hardcoded commit",
			"v0.25.3",
			VersionPlan{RefKind: RefCommit, Ref: "07a030ddfdbe5ac8a22c235be7b5771cc01f8498", LiteralBuild: "v0.25.3", LiteralRelease: "v0.25.3", SourceOverride: OverrideBuild, Now: now},
		},
		{
			"calendar tag resolves the bare tag",
			"26.10.0",
			VersionPlan{RefKind: RefTag, Ref: "26.10.0", LiteralBuild: "26.10.0", LiteralRelease: "26.10.0", Now: now},
		},
		{
			"v-prefixed calendar tag normalises to bare",
			"v26.10.0",
			VersionPlan{RefKind: RefTag, Ref: "26.10.0", LiteralBuild: "26.10.0", LiteralRelease: "26.10.0", Now: now},
		},
		{
			"release candidate resolves the bare tag",
			"26.10-rc.1",
			VersionPlan{RefKind: RefTag, Ref: "26.10-rc.1", LiteralBuild: "26.10-rc.1", LiteralRelease: "26.10-rc.1", Now: now},
		},
		{
			"dev version builds main and keeps the literal version",
			"26.10.15-dev.1511fd1",
			VersionPlan{RefKind: RefBranch, Ref: "main", LiteralBuild: "v26.10.15-dev.1511fd1", LiteralRelease: "26.10.15-dev.1511fd1", Now: now},
		},
		{
			"commit hash defers to the declared source version",
			fortyCharHash,
			VersionPlan{RefKind: RefCommit, Ref: fortyCharHash, SourceOverride: OverrideBoth, MintDev: true, Now: now},
		},
		{
			"empty input builds a dev version from main",
			"",
			VersionPlan{RefKind: RefBranch, Ref: "main", MintDev: true, Now: now},
		},
		{
			"dev input builds a dev version from main",
			"dev",
			VersionPlan{RefKind: RefBranch, Ref: "main", MintDev: true, Now: now},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, err := PlanVersion(tc.input, now)

			// THEN
			if err != nil {
				t.Fatalf("PlanVersion(%q) returned error: %v", tc.input, err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("PlanVersion(%q) = %+v, want %+v", tc.input, got, tc.want)
			}
		})
	}
}

func TestPlanVersionRejectsUnknownInput(t *testing.T) {
	// GIVEN inputs that are neither tags, dev versions, hashes, nor "dev"
	for _, input := range []string{"26.9.0", "v1.2", "main", "short"} {
		t.Run("input="+input, func(t *testing.T) {
			// WHEN
			_, err := PlanVersion(input, time.Now())

			// THEN
			if err == nil {
				t.Fatalf("PlanVersion(%q) succeeded, want error", input)
			}
			// AND the message names every accepted form
			want := fmt.Sprintf(`invalid version %q: expected a release tag (v0.25.3, 26.10.0, 26.10-rc.1), a dev version (26.10.15-dev.1511fd1), a 40-character commit hash, or "dev" to build upstream main`, input)
			if err.Error() != want {
				t.Fatalf("error = %q, want %q", err, want)
			}
		})
	}
}

func TestFinalize(t *testing.T) {
	// GIVEN plans paired with resolved source state, at a fixed January instant
	// (single-digit month, where literal preservation matters most)
	now := time.Date(2027, 1, 15, 7, 25, 0, 0, time.UTC)

	cases := []struct {
		name        string
		srcVersion  string
		wantBuild   string
		wantRelease string
		commit      string
		plan        VersionPlan
		srcFound    bool
	}{
		{
			name:        "tag plan keeps its literal versions",
			plan:        VersionPlan{RefKind: RefTag, Ref: "26.10.0", LiteralBuild: "26.10.0", LiteralRelease: "26.10.0", Now: now},
			commit:      "d566184",
			wantBuild:   "26.10.0",
			wantRelease: "26.10.0",
		},
		{
			name:        "dev build plan keeps the v-prefixed build version",
			plan:        VersionPlan{RefKind: RefBranch, Ref: "main", LiteralBuild: "v26.10.15-dev.1511fd1", LiteralRelease: "26.10.15-dev.1511fd1", Now: now},
			commit:      "e051129",
			wantBuild:   "v26.10.15-dev.1511fd1",
			wantRelease: "26.10.15-dev.1511fd1",
		},
		{
			name:        "pinned plan takes the declared build version",
			plan:        VersionPlan{RefKind: RefCommit, Ref: "07a030ddfdbe5ac8a22c235be7b5771cc01f8498", LiteralRelease: "v0.25.3", SourceOverride: OverrideBuild, Now: now},
			srcVersion:  "0.25.3",
			srcFound:    true,
			commit:      "07a030ddfdbe5ac8a22c235be7b5771cc01f8498",
			wantBuild:   "0.25.3",
			wantRelease: "v0.25.3",
		},
		{
			name:        "pinned plan falls back to the release version without a declaration",
			plan:        VersionPlan{RefKind: RefCommit, Ref: "07a030ddfdbe5ac8a22c235be7b5771cc01f8498", LiteralBuild: "v0.25.3", LiteralRelease: "v0.25.3", SourceOverride: OverrideBuild, Now: now},
			commit:      "07a030ddfdbe5ac8a22c235be7b5771cc01f8498",
			wantBuild:   "v0.25.3",
			wantRelease: "v0.25.3",
		},
		{
			name:        "commit plan lets the declaration override both versions",
			plan:        VersionPlan{RefKind: RefCommit, Ref: fortyCharHash, SourceOverride: OverrideBoth, MintDev: true, Now: now},
			srcVersion:  "26.10.0",
			srcFound:    true,
			commit:      fortyCharHash,
			wantBuild:   "26.10.0",
			wantRelease: "26.10.0",
		},
		{
			name:        "commit plan mints a dev version without a declaration",
			plan:        VersionPlan{RefKind: RefCommit, Ref: fortyCharHash, SourceOverride: OverrideBoth, MintDev: true, Now: now},
			commit:      fortyCharHash,
			wantBuild:   "v27.01.15-dev.1511fd1",
			wantRelease: "27.01.15-dev.1511fd1",
		},
		{
			name:        "default plan mints from the resolved commit",
			plan:        VersionPlan{RefKind: RefBranch, Ref: "main", MintDev: true, Now: now},
			commit:      "e05112937c05a2b31c4d5e6f70819293a4b5c6d7",
			wantBuild:   "v27.01.15-dev.e051129",
			wantRelease: "27.01.15-dev.e051129",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			build, release, err := tc.plan.Finalize(tc.srcVersion, tc.srcFound, tc.commit)

			// THEN
			if err != nil {
				t.Fatalf("Finalize returned error: %v", err)
			}
			if build != tc.wantBuild || release != tc.wantRelease {
				t.Fatalf("Finalize = (%q, %q), want (%q, %q)", build, release, tc.wantBuild, tc.wantRelease)
			}
		})
	}
}

func TestFinalizeMintErrors(t *testing.T) {
	// GIVEN plans that mint from an unusable commit
	now := time.Date(2027, 1, 15, 7, 25, 0, 0, time.UTC)

	t.Run("default plan names the branch", func(t *testing.T) {
		plan := VersionPlan{RefKind: RefBranch, Ref: "main", MintDev: true, Now: now}

		// WHEN
		_, _, err := plan.Finalize("", false, "short")

		// THEN
		if err == nil || !strings.HasPrefix(err.Error(), "cannot synthesise dev version for main@") {
			t.Fatalf("error = %v, want a branch-scoped synthesis error", err)
		}
	})

	t.Run("commit plan names the commit", func(t *testing.T) {
		plan := VersionPlan{RefKind: RefCommit, Ref: fortyCharHash, SourceOverride: OverrideBoth, MintDev: true, Now: now}

		// WHEN
		_, _, err := plan.Finalize("", false, "short")

		// THEN
		if err == nil || !strings.HasPrefix(err.Error(), fmt.Sprintf("cannot synthesise dev version for %q", fortyCharHash)) {
			t.Fatalf("error = %v, want a commit-scoped synthesis error", err)
		}
	})
}
