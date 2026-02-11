/*
 * https://docs.dagger.io/extending/modules/
 *
 * Regenerate the Dagger files with `dagger develop --compat=0.19.10`.
 *
 * NOTES
 * - Dagger requires passing host paths explicitly, so `--source .` is a must.
 * - Dagger only supports standard OCI fields, so Docker-specific features like healthchecks are not supported.
 */
package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"

	"github.com/Masterminds/semver/v3"
)

var TARGETS []BuildMatrix = []BuildMatrix{
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

var commitHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

type MemosBuilds struct{}

// generateNightlyVersion bumps patch and adds -pre suffix with commit metadata.
func (m *MemosBuilds) generateNightlyVersion(
	ctx context.Context,
	git *dagger.GitRepository,
	baseVersion string,
) *semver.Version {
	v, err := semver.NewVersion(baseVersion)
	if err != nil {
		v, _ = semver.NewVersion("0.0.0")
	}

	baseVer, _ := v.SetPrerelease("")
	nextVer := baseVer.IncPatch()

	if commit, err := git.Branch("main").Commit(ctx); err == nil {
		if v, err := nextVer.SetMetadata(commit[:7]); err == nil {
			nextVer = v
		}
	}

	nightlyVer, _ := nextVer.SetPrerelease("pre")
	return &nightlyVer
}

// prepareSource resolves version and applies all patches.
func (m *MemosBuilds) prepareSource(
	ctx context.Context,
	source *dagger.Directory,
	version string,
) (*dagger.Directory, string, error) {
	if source == nil {
		return nil, "", fmt.Errorf("source directory must be passed explicitly by the user")
	}

	gitSrc, resolvedVersion, err := m.resolveVersion(ctx, version)
	if err != nil {
		return nil, "", err
	}

	gitSrc, err = m.patchModerncSqlite(ctx, gitSrc)
	if err != nil {
		return nil, "", fmt.Errorf("failed to patch go.mod: %w", err)
	}

	patchesDir := source.Directory("patches")
	gitSrc, err = m.applyPatches(ctx, gitSrc, patchesDir)
	if err != nil {
		return nil, "", fmt.Errorf("failed to apply patches: %w", err)
	}

	return gitSrc, resolvedVersion, nil
}

// Build compiles Memos binaries, creates release archives, and generates checksums.
func (m *MemosBuilds) Build(
	ctx context.Context,
	source *dagger.Directory,
	version string,
	platforms string,
) (*dagger.Directory, error) {
	if version == "" {
		version = "nightly"
	}

	targets, err := filterTargets(platforms)
	if err != nil {
		return nil, fmt.Errorf("invalid platforms: %w", err)
	}

	gitSrc, version, err := m.prepareSource(ctx, source, version)
	if err != nil {
		return nil, err
	}

	gitSrc = m.generateProto(gitSrc)
	frontendDist := m.buildFrontend(gitSrc)

	binaries, err := m.buildBackend(ctx, gitSrc, frontendDist, version, targets)
	if err != nil {
		return nil, err
	}

	archives := m.createReleaseArchives(binaries, version, targets)
	checksums := m.generateChecksums(archives, version)
	out := archives.WithFile(fmt.Sprintf(buildconsts.CHECKSUM_FILE_FORMAT, version), checksums)

	return out, nil
}

// Publish builds release artifacts and optionally publishes containers.
func (m *MemosBuilds) Publish(
	ctx context.Context,
	source *dagger.Directory,
	version string,
	dockerHubUser string,
	dockerHubPassword *dagger.Secret,
	ghcrUser string,
	ghcrPassword *dagger.Secret,
) (*dagger.Directory, error) {
	if version == "" {
		version = "nightly"
	}

	// Resolve version first to ensure consistent versioning across build and publish.
	_, resolvedVersion, err := m.resolveVersion(ctx, version)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve version: %w", err)
	}

	out, err := m.Build(ctx, source, resolvedVersion, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	if (dockerHubUser != "" && dockerHubPassword != nil) || (ghcrUser != "" && ghcrPassword != nil) {
		_, err := m.publishContainers(ctx, source, resolvedVersion, dockerHubUser, dockerHubPassword, ghcrUser, ghcrPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to publish containers: %w", err)
		}
	}

	return out, nil
}

// BuildContainers builds Docker images for the specified platforms and returns them as tarballs.
func (m *MemosBuilds) BuildContainers(
	ctx context.Context,
	source *dagger.Directory,
	version string,
	platforms string,
) (*dagger.Directory, error) {
	if version == "" {
		version = "nightly"
	}

	targets, err := filterTargets(platforms)
	if err != nil {
		return nil, fmt.Errorf("invalid platforms: %w", err)
	}

	containerTargets := filterLinuxTargets(targets)
	if len(containerTargets) == 0 {
		return nil, fmt.Errorf("no Linux platforms in the selected targets")
	}

	containers, err := m.buildContainers(ctx, source, version, containerTargets)
	if err != nil {
		return nil, err
	}

	out := dag.Directory()
	for i, ctr := range containers {
		t := containerTargets[i]
		platform := strings.ReplaceAll(t.DockerPlatform(), "/", "-")
		imageName := fmt.Sprintf("memos-%s.tar", platform)
		out = out.WithFile(imageName, ctr.AsTarball())
	}

	return out, nil
}
