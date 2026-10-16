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
	"dagger/memos-builds/buildplan"
	"dagger/memos-builds/internal/dagger"
)

var commitHashPattern = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

type MemosBuilds struct{}

// prepareSource resolves version and applies all patches.
func (m *MemosBuilds) prepareSource(
	ctx context.Context,
	source *dagger.Directory,
	version string,
) (*dagger.Directory, string, string, string, error) {
	if source == nil {
		return nil, "", "", "", fmt.Errorf("source directory must be passed explicitly by the user")
	}

	gitSrc, buildVersion, releaseVersion, commit, err := m.resolveVersion(ctx, version)
	if err != nil {
		return nil, "", "", "", err
	}

	gitSrc, err = m.patchModerncSqlite(ctx, gitSrc)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("failed to patch go.mod: %w", err)
	}

	patchesDir := source.Directory("patches")
	gitSrc, err = m.applyPatches(ctx, gitSrc, patchesDir)
	if err != nil {
		return nil, "", "", "", fmt.Errorf("failed to apply patches: %w", err)
	}

	return gitSrc, buildVersion, releaseVersion, commit, nil
}

// Build compiles Memos binaries, creates release archives, and generates checksums.
func (m *MemosBuilds) Build(
	ctx context.Context,
	source *dagger.Directory,
	version string,
	platforms string,
) (*dagger.Directory, error) {
	out, _, _, _, _, err := m.buildInternal(ctx, source, version, platforms)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// buildInternal is the core build logic.
// It returns artifacts, source, build version, release version, source commit, and an error.
func (m *MemosBuilds) buildInternal(
	ctx context.Context,
	source *dagger.Directory,
	version string,
	platforms string,
) (*dagger.Directory, *dagger.Directory, string, string, string, error) {
	if version == "" {
		version = "dev"
	}

	targets, err := buildplan.FilterTargets(platforms)
	if err != nil {
		return nil, nil, "", "", "", fmt.Errorf("invalid platforms: %w", err)
	}

	gitSrc, buildVersion, releaseVersion, commit, err := m.prepareSource(ctx, source, version)
	if err != nil {
		return nil, nil, "", "", "", err
	}

	gitSrc = m.generateProto(gitSrc)
	frontendDist := m.buildFrontend(gitSrc)

	binaries, err := m.buildBackend(ctx, gitSrc, frontendDist, buildVersion, commit, targets)
	if err != nil {
		return nil, nil, "", "", "", err
	}

	archives := m.createReleaseArchives(binaries, buildVersion, targets)
	checksums := m.generateChecksums(archives, buildVersion)
	out := archives.WithFile(fmt.Sprintf(buildconsts.CHECKSUM_FILE_FORMAT, buildVersion), checksums)

	return out, gitSrc, buildVersion, releaseVersion, commit, nil
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
	out, gitSrc, buildVersion, releaseVersion, commit, err := m.buildInternal(ctx, source, version, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build: %w", err)
	}

	if (dockerHubUser != "" && dockerHubPassword != nil) || (ghcrUser != "" && ghcrPassword != nil) {
		images, err := m.publishContainers(ctx, source, gitSrc, buildVersion, releaseVersion, commit, dockerHubUser, dockerHubPassword, ghcrUser, ghcrPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to publish containers: %w", err)
		}

		if len(images) > 0 {
			jsonData, err := PublishedImagesJSON(images)
			if err != nil {
				return nil, fmt.Errorf("failed to serialise published image metadata: %w", err)
			}
			out = out.WithNewFile("container-digests.json", jsonData)
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
		version = "dev"
	}

	targets, err := buildplan.FilterTargets(platforms)
	if err != nil {
		return nil, fmt.Errorf("invalid platforms: %w", err)
	}

	containerTargets := buildplan.FilterLinuxTargets(targets)
	if len(containerTargets) == 0 {
		return nil, fmt.Errorf("no Linux platforms in the selected targets")
	}

	gitSrc, buildVersion, _, commit, err := m.prepareSource(ctx, source, version)
	if err != nil {
		return nil, err
	}

	containers, err := m.buildContainers(ctx, source, gitSrc, buildVersion, commit, containerTargets)
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
