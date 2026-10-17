package buildplan

import (
	"fmt"
	"slices"
	"strings"
)

// BuildMatrix describes one build target: OS, architecture, and architecture
// level (e.g. amd64 microarchitecture or ARM variant).
type BuildMatrix struct {
	OS        string
	Arch      string
	ArchLevel string
}

// Targets lists every supported build target.
var Targets []BuildMatrix = []BuildMatrix{
	// Linux
	{"linux", "amd64", "v1"},
	{"linux", "amd64", "v2"},
	{"linux", "amd64", "v3"},
	{"linux", "arm", "v5"},
	{"linux", "arm", "v6"},
	{"linux", "arm", "v7"},
	{"linux", "arm64", ""},
	{"linux", "386", "sse2"},
	{"linux", "ppc64le", "power8"},
	{"linux", "riscv64", "rva20u64"},
	{"linux", "s390x", ""},

	// Darwin
	{"darwin", "amd64", "v1"},
	{"darwin", "amd64", "v2"},
	{"darwin", "amd64", "v3"},
	{"darwin", "arm64", ""},

	// Windows
	{"windows", "amd64", "v1"},
	{"windows", "amd64", "v2"},
	{"windows", "amd64", "v3"},
	{"windows", "arm64", ""},
	{"windows", "386", "sse2"},

	// FreeBSD
	{"freebsd", "amd64", "v1"},
	{"freebsd", "amd64", "v2"},
	{"freebsd", "amd64", "v3"},
	{"freebsd", "arm64", ""},
}

// DockerNoVariants lists Docker platforms that do not support variants.
var DockerNoVariants []string = []string{"386", "ppc64le", "riscv64"}

// DockerPlatform returns a platform string in a format accepted by Docker.
//
// E.g. "linux/amd64", "linux/arm/v7".
func (m *BuildMatrix) DockerPlatform() string {
	if m.ArchLevel == "" || slices.Contains(DockerNoVariants, m.Arch) {
		return fmt.Sprintf("%s/%s", m.OS, m.Arch)
	}
	return fmt.Sprintf("%s/%s/%s", m.OS, m.Arch, m.ArchLevel)
}

// GoEnv returns the GO* environment variable mapping for this target
// (e.g. {"GOARM": "7"}). Returns nil when the target needs no arch-level
// variables. The caller applies the map verbatim, learning nothing about
// individual arches.
func (m *BuildMatrix) GoEnv() map[string]string {
	if m.ArchLevel == "" {
		return nil
	}
	switch m.Arch {
	case "arm":
		return map[string]string{"GOARM": strings.TrimPrefix(m.ArchLevel, "v")}
	case "amd64":
		return map[string]string{"GOAMD64": m.ArchLevel}
	case "386":
		return map[string]string{"GO386": m.ArchLevel}
	case "ppc64le":
		return map[string]string{"GOPPC64": m.ArchLevel}
	case "riscv64":
		return map[string]string{"GORISCV64": m.ArchLevel}
	default:
		return nil
	}
}

// IsARMv5 reports whether this target needs the BusyBox container:
// Linux arm at level v5, the only ARMv5 row in Targets.
func (m BuildMatrix) IsARMv5() bool {
	return m.OS == "linux" && m.Arch == "arm" && m.ArchLevel == "v5"
}

// IsAmd64Variant reports whether this target is an amd64 microarchitecture
// variant above the base level, whose platform the engine would otherwise
// collapse to plain linux/amd64.
func (m BuildMatrix) IsAmd64Variant() bool {
	return m.Arch == "amd64" && m.ArchLevel != "" && m.ArchLevel != "v1"
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
		return "armv" + m.GoEnv()["GOARM"] + "l"
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

// FilterTargets returns a subset of Targets matching the given platforms string.
// The returned slice is a copy: mutating it never affects Targets.
//
// Accepted formats:
//   - "" or "all": returns all Targets
//   - Comma-separated list: "linux/amd64,darwin/arm64,linux/arm/v7"
//
// Platform strings use the Docker format: os/arch or os/arch/variant.
func FilterTargets(platforms string) ([]BuildMatrix, error) {
	platforms = strings.TrimSpace(platforms)
	if platforms == "" || platforms == "all" {
		return slices.Clone(Targets), nil
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
		for _, t := range Targets {
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
			for _, t := range Targets {
				if t.OS == parts[0] && t.Arch == parts[1] {
					filtered = append(filtered, t)
					found = true
					break
				}
			}
		}

		if !found {
			var available []string
			for _, t := range Targets {
				available = append(available, t.DockerPlatform())
			}
			return nil, fmt.Errorf("platform %q not found in Targets (available: %s)", p, strings.Join(available, ", "))
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no valid platforms specified")
	}

	return filtered, nil
}

// FilterLinuxTargets returns only Linux targets from the given slice.
func FilterLinuxTargets(targets []BuildMatrix) []BuildMatrix {
	var linux []BuildMatrix
	for _, t := range targets {
		if t.OS == "linux" {
			linux = append(linux, t)
		}
	}
	return linux
}
