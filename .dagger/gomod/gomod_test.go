package gomod

import (
	"strings"
	"testing"
)

// maxLibcVersion is the newest libc version in sqliteLibcMap, used to model an
// unpinned sqlite version in tests.
const maxLibcVersion = "v1.74.4"

// memosGoMod is a trimmed go.mod from usememos/memos (main, 2026-10-16):
// modernc.org/sqlite is a direct dependency, modernc.org/libc its indirect pin.
const memosGoMod = `module github.com/usememos/memos

go 1.26.0

require (
	github.com/google/uuid v1.6.0
	modernc.org/sqlite v1.56.0
)

require (
	modernc.org/libc v1.75.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.0 // indirect
)
`

// upstreamSqliteGoMod is a trimmed go.mod of modernc.org/sqlite v1.56.0,
// declaring the libc version it needs.
const upstreamSqliteGoMod = `module modernc.org/sqlite

go 1.25.0

require (
	modernc.org/libc v1.74.4
	modernc.org/mathutil v1.7.1
)
`

func TestSqliteVersion(t *testing.T) {
	// GIVEN go.mod contents with and without a sqlite requirement
	cases := []struct {
		name     string
		contents string
		want     string
		found    bool
	}{
		{"pinned memos go.mod", memosGoMod, "v1.56.0", true},
		{"unparseable contents", "module example.com/m\n", "", false},
		{"empty contents", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, found := SqliteVersion(tc.contents)

			// THEN
			if got != tc.want || found != tc.found {
				t.Fatalf("SqliteVersion() = (%q, %v), want (%q, %v)", got, found, tc.want, tc.found)
			}
		})
	}
}

func TestPinnedLibc(t *testing.T) {
	// GIVEN sqlite versions with and without a map entry
	cases := []struct {
		name          string
		sqliteVersion string
		want          string
		found         bool
	}{
		{"newest pinned version", "v1.56.0", maxLibcVersion, true},
		{"oldest pinned version", "v1.37.0", "v1.62.1", true},
		{"unpinned version", "v1.59.0", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, found := PinnedLibc(tc.sqliteVersion)

			// THEN
			if got != tc.want || found != tc.found {
				t.Fatalf("PinnedLibc(%q) = (%q, %v), want (%q, %v)", tc.sqliteVersion, got, found, tc.want, tc.found)
			}
		})
	}
}

func TestUpstreamGoModURL(t *testing.T) {
	// GIVEN an unpinned sqlite version
	// WHEN
	got := UpstreamGoModURL("v1.59.0")

	// THEN
	want := "https://gitlab.com/cznic/sqlite/-/raw/v1.59.0/go.mod"
	if got != want {
		t.Fatalf("UpstreamGoModURL() = %q, want %q", got, want)
	}
}

func TestLibcFromUpstreamGoMod(t *testing.T) {
	// GIVEN upstream sqlite go.mod contents with and without a direct libc line
	cases := []struct {
		name     string
		contents string
		want     string
		found    bool
	}{
		{"upstream sqlite go.mod", upstreamSqliteGoMod, "v1.74.4", true},
		{"unparseable contents", "module modernc.org/sqlite\n", "", false},
		{"empty contents", "", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got, found := LibcFromUpstreamGoMod(tc.contents)

			// THEN
			if got != tc.want || found != tc.found {
				t.Fatalf("LibcFromUpstreamGoMod() = (%q, %v), want (%q, %v)", got, found, tc.want, tc.found)
			}
		})
	}
}

func TestPatchLibc(t *testing.T) {
	// GIVEN memos go.mod contents pinning libc v1.75.3
	t.Run("downgrades to the pinned libc version", func(t *testing.T) {
		// WHEN
		got, modified := PatchLibc(memosGoMod, maxLibcVersion)

		// THEN
		if !modified {
			t.Fatal("PatchLibc reported no change, want a rewrite")
		}
		// AND only the libc version changed, keeping its indirect marker
		want := strings.Replace(memosGoMod, "modernc.org/libc v1.75.3", "modernc.org/libc "+maxLibcVersion, 1)
		if got != want {
			t.Fatalf("PatchLibc() =\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("no change when the version already matches", func(t *testing.T) {
		// GIVEN contents already at the expected version
		contents := strings.Replace(memosGoMod, "modernc.org/libc v1.75.3", "modernc.org/libc "+maxLibcVersion, 1)

		// WHEN
		got, modified := PatchLibc(contents, maxLibcVersion)

		// THEN
		if modified {
			t.Fatal("PatchLibc reported a change, want none")
		}
		if got != contents {
			t.Fatalf("PatchLibc returned different contents without a change:\n%s", got)
		}
	})

	t.Run("no change without an indirect libc line", func(t *testing.T) {
		// GIVEN contents where libc is not marked indirect
		// AND reLibc only matches indirect requirements: if upstream ever
		// promotes libc to a direct dependency, pinning silently stops.
		contents := strings.Replace(memosGoMod, "modernc.org/libc v1.75.3 // indirect", "modernc.org/libc v1.75.3", 1)

		// WHEN
		got, modified := PatchLibc(contents, maxLibcVersion)

		// THEN
		if modified {
			t.Fatal("PatchLibc reported a change, want none")
		}
		if got != contents {
			t.Fatalf("PatchLibc returned different contents without a change:\n%s", got)
		}
	})

	t.Run("no change without any libc line", func(t *testing.T) {
		// GIVEN contents without a libc requirement
		contents := "module example.com/m\n\ngo 1.26.0\n"

		// WHEN
		got, modified := PatchLibc(contents, maxLibcVersion)

		// THEN
		if modified {
			t.Fatal("PatchLibc reported a change, want none")
		}
		if got != contents {
			t.Fatalf("PatchLibc returned different contents without a change:\n%s", got)
		}
	})
}
