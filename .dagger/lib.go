package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"
)

var versionVarPattern = regexp.MustCompile(`var Version = "([^"]+)"`)

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

// extractVersionFromSource reads version from upstream source code.
func (m *MemosBuilds) extractVersionFromSource(ctx context.Context, src *dagger.Directory) string {
	contents, err := src.File(buildconsts.VERSION_FILE).Contents(ctx)
	if err != nil {
		return "0.0.0"
	}

	match := versionVarPattern.FindStringSubmatch(contents)
	if len(match) < 2 {
		return "0.0.0"
	}

	versionFromFile := match[1]
	if _, err := semver.NewVersion(versionFromFile); err != nil {
		return "0.0.0"
	}

	return versionFromFile
}

// resolveVersion determines the git source and version string from user input.
//
// Handles semver tags, release branches, commit hashes, or defaults to nightly.
func (m *MemosBuilds) resolveVersion(
	ctx context.Context,
	version string,
) (gitSrc *dagger.Directory, resolvedVersion string, err error) {
	git := dag.Git("https://github.com/usememos/memos.git")
	treeOpts := dagger.GitRefTreeOpts{Depth: 1}

	if v, err := semver.NewVersion(version); err == nil {
		gitSrc = git.Tag("v" + v.String()).Tree(treeOpts)
		return gitSrc, "v" + v.String(), nil
	}

	if after, ok := strings.CutPrefix(version, "release/"); ok {
		ver := strings.TrimPrefix(after, "v")
		gitSrc = git.Ref("heads/release/" + ver).Tree(treeOpts)
		return gitSrc, "v" + ver, nil
	}

	if commitHashPattern.MatchString(version) {
		gitSrc = git.Commit(version).Tree(treeOpts)
		srcVersion := m.extractVersionFromSource(ctx, gitSrc)
		return gitSrc, srcVersion, nil
	}

	// Use nightly as default version.
	gitSrc = git.Branch("main").Tree(treeOpts)

	baseVersion := m.extractVersionFromSource(ctx, gitSrc)
	nightlyVer := m.generateNightlyVersion(ctx, git, baseVersion)

	return gitSrc, "v" + nightlyVer.String(), nil
}
