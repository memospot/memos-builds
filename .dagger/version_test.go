package main

import (
	"slices"
	"testing"
	"time"
)

func TestDevBuildVersion(t *testing.T) {
	// GIVEN
	cases := []struct {
		name string
		when time.Time
		sha  string
		want string
	}{
		{"stable date and lowercase SHA", time.Date(2026, 10, 15, 7, 25, 0, 0, time.UTC), "1511fd1", "26.10.15-dev.1511fd1"},
		{"uppercase SHA is lowercased", time.Date(2026, 10, 15, 7, 25, 0, 0, time.UTC), "ABC1234", "26.10.15-dev.abc1234"},
		{"single-digit day carries no leading zero", time.Date(2026, 10, 1, 7, 25, 0, 0, time.UTC), "1511fd1", "26.10.1-dev.1511fd1"},
		{"zero-padded month is preserved literally", time.Date(2026, 1, 15, 7, 25, 0, 0, time.UTC), "1511fd1", "26.01.15-dev.1511fd1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, err := devBuildVersion(tc.when, tc.sha)

			// THEN
			if err != nil {
				t.Fatalf("devBuildVersion returned error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("devBuildVersion = %q, want %q", got, tc.want)
			}
			if !devVersionPattern.MatchString(got) {
				t.Fatalf("version %q fails the dev pattern", got)
			}
		})
	}
}

func TestDevBuildVersionRejectsBadSHA(t *testing.T) {
	// GIVEN
	now := time.Date(2026, 10, 15, 7, 25, 0, 0, time.UTC)

	for _, sha := range []string{"", "unknown", "abc", "1511fd125", "ZZZZZZZ"} {
		t.Run("sha="+sha, func(t *testing.T) {
			// WHEN
			got, err := devBuildVersion(now, sha)

			// THEN
			if err == nil {
				t.Fatalf("devBuildVersion(%q) = %q, want error", sha, got)
			}
		})
	}
}

func TestContainerTags(t *testing.T) {
	// GIVEN
	m := &MemosBuilds{}
	cases := []struct {
		name    string
		version string
		want    []string
	}{
		{"CalVer stable", "26.10.0", []string{"latest", "26.10", "26.10.0"}},
		{"v-prefixed CalVer stable", "v26.10.0", []string{"latest", "26.10", "26.10.0"}},
		{"bare month stable", "26.10", []string{"latest", "26.10"}},
		{"legacy stable", "v0.25.3", []string{"latest", "0.25", "0.25.3"}},
		{"release candidate", "26.10.0-rc.1", []string{"26.10.0-rc.1"}},
		{"live upstream candidate form", "26.10-rc.1", []string{"26.10-rc.1"}},
		{"dev build keeps dated tag plus alias", "26.10.15-dev.1511fd1", []string{"26.10.15-dev.1511fd1", "dev"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got := m.containerTags(tc.version)

			// THEN
			if !slices.Equal(got, tc.want) {
				t.Fatalf("containerTags(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}

func TestTagsForRegistry(t *testing.T) {
	// GIVEN
	m := &MemosBuilds{}
	ghcr := "ghcr.io/memospot/memos-builds"
	hub := "docker.io/lincolnthalles/memos"

	cases := []struct {
		name     string
		version  string
		registry string
		want     []string
	}{
		{"dev on GHCR", "26.10.15-dev.1511fd1", ghcr, []string{"26.10.15-dev.1511fd1", "dev"}},
		{"dev on Hub", "26.10.15-dev.1511fd1", hub, []string{"dev"}},
		{"candidate on GHCR", "26.10.0-rc.1", ghcr, []string{"26.10.0-rc.1"}},
		{"live upstream candidate on GHCR", "26.10-rc.1", ghcr, []string{"26.10-rc.1"}},
		{"stable on GHCR", "26.10.0", ghcr, []string{"latest", "26.10", "26.10.0"}},
		{"legacy stable on Hub", "v0.25.3", hub, []string{"latest", "0.25", "0.25.3"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got := m.tagsForRegistry(tc.version, tc.registry)

			// THEN
			if !slices.Equal(got, tc.want) {
				t.Fatalf("tagsForRegistry(%q, %q) = %v, want %v", tc.version, tc.registry, got, tc.want)
			}
		})
	}
}

func TestIsReleaseVersion(t *testing.T) {
	// GIVEN
	cases := []struct {
		version string
		want    bool
	}{
		{"26.10.0", true},
		{"26.10", false},
		{"26.10.0-rc.1", true},
		{"26.10-rc.1", true},
		{"v0.25.3", true},
		{"26.10.15-dev.1511fd1", false},
		{"1511fd125f1a2b3c4d5e6f70819293a4b5c6d7e8", false},
		{"", false},
	}

	for _, tc := range cases {
		t.Run("version="+tc.version, func(t *testing.T) {
			// WHEN
			got := isReleaseVersion(tc.version)

			// THEN
			if got != tc.want {
				t.Fatalf("isReleaseVersion(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}

func TestShortCommitHash(t *testing.T) {
	// GIVEN full and short hashes
	// WHEN shortened
	// THEN exactly 7 hex chars are returned, else empty
	if got := shortCommitHash("1511fd125f1a2b3c4d5e6f70819293a4b5c6d7e8"); got != "1511fd1" {
		t.Fatalf("shortCommitHash = %q, want 1511fd1", got)
	}
	if got := shortCommitHash("ABC1234"); got != "ABC1234" {
		t.Fatalf("shortCommitHash = %q, want ABC1234", got)
	}

	for _, bad := range []string{"", "abc", "zzzzzzz"} {
		if got := shortCommitHash(bad); got != "" {
			t.Fatalf("shortCommitHash(%q) = %q, want empty", bad, got)
		}
	}
}
