package buildplan

import (
	"regexp"
	"testing"
)

// upstreamBuildVersion is the validator from usememos/memos
// internal/version/version.go (fetched 2026-10-16). version.Validate() runs at
// startup and exits on mismatch, so every embedded CalVer or dev version must
// match it exactly.
var upstreamBuildVersion = regexp.MustCompile(`^[1-9][0-9]\.(0[1-9]|1[0-2])\.(0|[1-9][0-9]*)(-rc\.[1-9][0-9]*|-dev\.[0-9a-f]{7})?$`)

func TestAppVersion(t *testing.T) {
	// GIVEN build versions spanning every shape resolveVersion can produce,
	// including single-digit months, whose zero padding a semver re-render
	// would strip
	cases := []struct {
		name    string
		version string
		want    string
	}{
		{"calendar release is literal", "26.10.0", "26.10.0"},
		{"zero-padded September month survives", "26.09.0", "26.09.0"},
		{"bare month gains a patch", "26.10", "26.10.0"},
		{"candidate without patch gains one", "26.10-rc.1", "26.10.0-rc.1"},
		{"candidate with patch is literal", "26.10.0-rc.1", "26.10.0-rc.1"},
		{"dev version keeps its v-prefixed build form stripped", "v26.10.15-dev.1511fd1", "26.10.15-dev.1511fd1"},
		{"zero-padded January month survives", "26.01.15-dev.1511fd1", "26.01.15-dev.1511fd1"},
		{"legacy version drops the v prefix", "v0.25.3", "0.25.3"},
		{"build metadata is dropped", "26.10.0+1511fd1", "26.10.0"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, ok := AppVersion(tc.version)

			// THEN
			if !ok {
				t.Fatalf("AppVersion(%q) reported failure", tc.version)
			}
			if got != tc.want {
				t.Fatalf("AppVersion(%q) = %q, want %q", tc.version, got, tc.want)
			}
			// AND calendar and dev versions satisfy upstream's startup validator
			if tc.version != "v0.25.3" && !upstreamBuildVersion.MatchString(got) {
				t.Fatalf("AppVersion(%q) = %q, which upstream rejects", tc.version, got)
			}
		})
	}
}

func TestAppVersionRejectsUnknown(t *testing.T) {
	// GIVEN inputs that are not semantic versions
	for _, input := range []string{"", "main", "dev", "1511fd125f1a2b3c4d5e6f70819293a4b5c6d7e8"} {
		t.Run("version="+input, func(t *testing.T) {
			// WHEN
			got, ok := AppVersion(input)

			// THEN
			if ok {
				t.Fatalf("AppVersion(%q) = %q, want failure", input, got)
			}
		})
	}
}

func TestParseDeclaredVersion(t *testing.T) {
	// GIVEN source file contents in the shapes upstream ships
	cases := []struct {
		name     string
		contents string
		want     string
		found    bool
	}{
		{
			"legacy declaration is found",
			"package version\n\nvar Version = \"0.25.3\"\n",
			"0.25.3", true,
		},
		{
			"calendar declaration is found",
			"var Version = \"26.10.0\"\n",
			"26.10.0", true,
		},
		{
			"declaration without a value has no default",
			"package version\n\nvar Version string\n",
			"", false,
		},
		{
			"unparseable value is rejected",
			"var Version = \"not-a-version\"\n",
			"", false,
		},
		{
			"empty contents are rejected",
			"",
			"", false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, found := ParseDeclaredVersion(tc.contents)

			// THEN
			if got != tc.want || found != tc.found {
				t.Fatalf("ParseDeclaredVersion() = (%q, %v), want (%q, %v)", got, found, tc.want, tc.found)
			}
		})
	}
}
