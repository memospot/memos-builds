package buildplan

import (
	"slices"
	"strings"
	"testing"
)

func TestDockerPlatform(t *testing.T) {
	// GIVEN targets with and without variant support
	cases := []struct {
		name   string
		target BuildMatrix
		want   string
	}{
		{"amd64 level is part of the platform", BuildMatrix{"linux", "amd64", "v3"}, "linux/amd64/v3"},
		{"arm level is part of the platform", BuildMatrix{"linux", "arm", "v7"}, "linux/arm/v7"},
		{"no-variant arch drops the level", BuildMatrix{"linux", "386", "sse2"}, "linux/386"},
		{"ppc64le drops the level", BuildMatrix{"linux", "ppc64le", "power8"}, "linux/ppc64le"},
		{"riscv64 drops the level", BuildMatrix{"linux", "riscv64", "rva20u64"}, "linux/riscv64"},
		{"empty level yields os/arch", BuildMatrix{"linux", "arm64", ""}, "linux/arm64"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got := tc.target.DockerPlatform()

			// THEN
			if got != tc.want {
				t.Fatalf("DockerPlatform() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBinaryName(t *testing.T) {
	// GIVEN targets with and without arch level and OS extension
	cases := []struct {
		name   string
		target BuildMatrix
		want   string
	}{
		{"amd64 level is appended verbatim", BuildMatrix{"linux", "amd64", "v2"}, "memos-linux-amd64v2"},
		{"empty level is omitted", BuildMatrix{"linux", "arm64", ""}, "memos-linux-arm64"},
		{"windows binary carries .exe", BuildMatrix{"windows", "arm64", ""}, "memos-windows-arm64.exe"},
		{"windows keeps level before extension", BuildMatrix{"windows", "amd64", "v3"}, "memos-windows-amd64v3.exe"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got := tc.target.BinaryName()

			// THEN
			if got != tc.want {
				t.Fatalf("BinaryName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestArchiveName(t *testing.T) {
	// GIVEN targets spanning the uname and level-suffix rules
	cases := []struct {
		name    string
		target  BuildMatrix
		version string
		want    string
	}{
		{"amd64 v1 carries no level suffix", BuildMatrix{"linux", "amd64", "v1"}, "26.10.0", "memos-26.10.0-linux-x86_64.tar.gz"},
		{"amd64 v2 carries level suffix", BuildMatrix{"linux", "amd64", "v2"}, "26.10.0", "memos-26.10.0-linux-x86_64_v2.tar.gz"},
		{"arm level maps to uname arch", BuildMatrix{"linux", "arm", "v7"}, "26.10.0", "memos-26.10.0-linux-armv7l.tar.gz"},
		{"386 maps to i386", BuildMatrix{"linux", "386", "sse2"}, "26.10.0", "memos-26.10.0-linux-i386.tar.gz"},
		{"unknown arch stays as-is", BuildMatrix{"linux", "s390x", ""}, "26.10.0", "memos-26.10.0-linux-s390x.tar.gz"},
		{"windows uses zip", BuildMatrix{"windows", "amd64", "v1"}, "26.10.0", "memos-26.10.0-windows-x86_64.zip"},
		{"version string is embedded literally", BuildMatrix{"darwin", "arm64", ""}, "26.01.15-dev.1511fd1", "memos-26.01.15-dev.1511fd1-darwin-arm64.tar.gz"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got := tc.target.ArchiveName(tc.version)

			// THEN
			if got != tc.want {
				t.Fatalf("ArchiveName(%q) = %q, want %q", tc.version, got, tc.want)
			}
		})
	}
}

func TestGoEnv(t *testing.T) {
	// GIVEN targets spanning every arch with a level mapping
	cases := []struct {
		name   string
		want   map[string]string
		target BuildMatrix
	}{
		{"arm level maps to GOARM", map[string]string{"GOARM": "7"}, BuildMatrix{"linux", "arm", "v7"}},
		{"amd64 level maps to GOAMD64", map[string]string{"GOAMD64": "v3"}, BuildMatrix{"linux", "amd64", "v3"}},
		{"386 level maps to GO386", map[string]string{"GO386": "sse2"}, BuildMatrix{"linux", "386", "sse2"}},
		{"ppc64le level maps to GOPPC64", map[string]string{"GOPPC64": "power8"}, BuildMatrix{"linux", "ppc64le", "power8"}},
		{"riscv64 level maps to GORISCV64", map[string]string{"GORISCV64": "rva20u64"}, BuildMatrix{"linux", "riscv64", "rva20u64"}},
		{"empty level yields nil", nil, BuildMatrix{"linux", "arm64", ""}},
		{"unknown arch yields nil", nil, BuildMatrix{"linux", "s390x", ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// WHEN
			got := tc.target.GoEnv()

			// THEN
			if len(got) != len(tc.want) {
				t.Fatalf("GoEnv() = %v, want %v", got, tc.want)
			}
			for k, v := range tc.want {
				if got[k] != v {
					t.Fatalf("GoEnv() = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestFilterTargets(t *testing.T) {
	// GIVEN platform selections against the declared matrix
	t.Run("empty selects every target without aliasing the global", func(t *testing.T) {
		// WHEN
		got, err := FilterTargets("")

		// THEN
		if err != nil {
			t.Fatalf("FilterTargets returned error: %v", err)
		}
		if !slices.Equal(got, Targets) {
			t.Fatalf("FilterTargets(\"\") selected %d targets, want all %d", len(got), len(Targets))
		}
		// AND mutating the result leaves the global untouched
		got[0].OS = "plan9"
		if Targets[0].OS == "plan9" {
			t.Fatal("FilterTargets(\"\") aliases Targets: mutating the result mutated the global")
		}
	})

	t.Run("all selects every target", func(t *testing.T) {
		// WHEN
		got, err := FilterTargets("all")

		// THEN
		if err != nil {
			t.Fatalf("FilterTargets returned error: %v", err)
		}
		if !slices.Equal(got, Targets) {
			t.Fatalf("FilterTargets(\"all\") selected %d targets, want all %d", len(got), len(Targets))
		}
	})

	t.Run("bare os/arch picks the base variant", func(t *testing.T) {
		// WHEN
		got, err := FilterTargets("linux/amd64")

		// THEN
		if err != nil {
			t.Fatalf("FilterTargets returned error: %v", err)
		}
		want := []BuildMatrix{{"linux", "amd64", "v1"}}
		if !slices.Equal(got, want) {
			t.Fatalf("FilterTargets(\"linux/amd64\") = %v, want %v", got, want)
		}
	})

	t.Run("exact variant match wins over base", func(t *testing.T) {
		// WHEN
		got, err := FilterTargets("linux/amd64/v3")

		// THEN
		if err != nil {
			t.Fatalf("FilterTargets returned error: %v", err)
		}
		want := []BuildMatrix{{"linux", "amd64", "v3"}}
		if !slices.Equal(got, want) {
			t.Fatalf("FilterTargets(\"linux/amd64/v3\") = %v, want %v", got, want)
		}
	})

	t.Run("multiple platforms, whitespace and empty entries tolerated", func(t *testing.T) {
		// WHEN
		got, err := FilterTargets(" linux/arm/v7 , darwin/arm64 ,, ")

		// THEN
		if err != nil {
			t.Fatalf("FilterTargets returned error: %v", err)
		}
		want := []BuildMatrix{{"linux", "arm", "v7"}, {"darwin", "arm64", ""}}
		if !slices.Equal(got, want) {
			t.Fatalf("FilterTargets() = %v, want %v", got, want)
		}
	})

	t.Run("unknown platform fails listing available targets", func(t *testing.T) {
		// WHEN
		_, err := FilterTargets("plan9/amd64")

		// THEN
		if err == nil {
			t.Fatal("FilterTargets(plan9/amd64) succeeded, want error")
		}
		if !strings.Contains(err.Error(), "plan9/amd64") || !strings.Contains(err.Error(), "linux/amd64/v3") {
			t.Fatalf("error %q does not name the requested platform and the available ones", err)
		}
	})
}

func TestFilterLinuxTargets(t *testing.T) {
	// GIVEN a mixed selection
	input := []BuildMatrix{
		{"linux", "amd64", "v1"},
		{"windows", "amd64", "v1"},
		{"linux", "arm64", ""},
		{"darwin", "arm64", ""},
	}

	// WHEN
	got := FilterLinuxTargets(input)

	// THEN
	want := []BuildMatrix{{"linux", "amd64", "v1"}, {"linux", "arm64", ""}}
	if !slices.Equal(got, want) {
		t.Fatalf("FilterLinuxTargets() = %v, want %v", got, want)
	}
}
