package main

import (
	"context"
	"fmt"
	"strings"

	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"

	"github.com/Masterminds/semver/v3"
)

// createArchive creates a tar.gz or zip archive from a binary file.
func (m *MemosBuilds) createArchive(
	binary *dagger.File,
	archiveName string,
) *dagger.File {
	isZip := strings.HasSuffix(archiveName, ".zip")

	binaryInArchive := "memos"
	if isZip {
		binaryInArchive = "memos.exe"
	}

	ctr := dag.Container().
		From(buildconsts.PRIMARY_IMAGE).
		WithWorkdir("/work").
		WithFile("/work/"+binaryInArchive, binary).
		WithExec([]string{"chmod", "+x", binaryInArchive})

	if isZip {
		ctr = ctr.
			WithExec([]string{"apk", "add", "--no-cache", "zip"}).
			WithExec([]string{"zip", "-9", archiveName, binaryInArchive})
	} else {
		ctr = ctr.WithExec([]string{"tar", "-czvf", archiveName, binaryInArchive})
	}

	return ctr.File("/work/" + archiveName)
}

// createReleaseArchives creates release archives for the given targets.
// Returns a directory containing all archives.
func (m *MemosBuilds) createReleaseArchives(
	binaries *dagger.Directory,
	version string,
	targets []BuildMatrix,
) *dagger.Directory {
	out := dag.Directory()

	for _, t := range targets {
		binaryName := t.BinaryName()
		archiveName := t.ArchiveName(version)

		binary := binaries.File(binaryName)
		archive := m.createArchive(binary, archiveName)
		out = out.WithFile(archiveName, archive)
	}

	return out
}

// generateChecksums creates a SHA256SUMS file for all archives in the directory.
func (m *MemosBuilds) generateChecksums(
	archives *dagger.Directory,
	version string,
) *dagger.File {
	checksumFile := fmt.Sprintf(buildconsts.CHECKSUM_FILE_FORMAT, version)

	ctr := dag.Container().
		From(buildconsts.PRIMARY_IMAGE).
		WithWorkdir("/work").
		WithDirectory("/work", archives).
		// Generate checksums for all archive files
		// Output format: <hash>  <filename> (two spaces)
		WithExec([]string{"sh", "-c", "sha256sum *.tar.gz *.zip 2>/dev/null | sort > " + checksumFile})

	return ctr.File("/work/" + checksumFile)
}

// containerTags returns the tags to use for container images based on version.
// For releases (v0.25.3): ["latest", "0.25", "0.25.3"]
// For nightlies (v0.25.4-pre): ["nightly"]
func (m *MemosBuilds) containerTags(version string) []string {
	v, err := semver.NewVersion(version)
	if err != nil {
		// Fallback for unparseable versions
		return []string{"latest"}
	}

	if v.Prerelease() != "" {
		return []string{"nightly"}
	}

	// Release tags: latest, major.minor, major.minor.patch
	return []string{
		"latest",
		fmt.Sprintf("%d.%d", v.Major(), v.Minor()),
		fmt.Sprintf("%d.%d.%d", v.Major(), v.Minor(), v.Patch()),
	}
}

// publishContainers builds and publishes multi-arch Docker images to registries.
// Called internally by Publish — not exposed as a standalone Dagger function
// to avoid redundant Build calls.
func (m *MemosBuilds) publishContainers(
	ctx context.Context,
	source *dagger.Directory,
	gitSrc *dagger.Directory,
	version string,
	dockerHubUser string,
	dockerHubPassword *dagger.Secret,
	ghcrUser string,
	ghcrPassword *dagger.Secret,
) (string, error) {
	// Build binaries for Linux targets only (containers are Linux-only).
	linuxTargets := filterLinuxTargets(TARGETS)
	if len(linuxTargets) == 0 {
		return "No Linux targets configured, skipping container publish", nil
	}

	platformVariants, err := m.buildContainers(ctx, source, gitSrc, version, linuxTargets)
	if err != nil {
		return "", fmt.Errorf("failed to build containers: %w", err)
	}

	tags := m.containerTags(version)
	var allPublished []string

	publishTargets := []struct {
		registry string
		user     string
		password *dagger.Secret
	}{
		{"docker.io/lincolnthalles/memos", dockerHubUser, dockerHubPassword},
		{"ghcr.io/memospot/memos-builds", ghcrUser, ghcrPassword},
	}
	for _, target := range publishTargets {
		if target.user != "" && target.password != nil {
			address := strings.Split(target.registry, "/")[0]
			for _, tag := range tags {
				addr := fmt.Sprintf("%s:%s", target.registry, tag)
				ref, err := dag.Container().
					WithRegistryAuth(address, target.user, target.password).
					Publish(ctx, addr, dagger.ContainerPublishOpts{
						PlatformVariants: platformVariants,
					})
				if err != nil {
					return "", fmt.Errorf("failed to publish to %s: %w", target.registry, err)
				}
				allPublished = append(allPublished, ref)
			}
		}
	}

	if len(allPublished) == 0 {
		return "No registry credentials provided, skipping container publish", nil
	}

	return fmt.Sprintf("Published %d images: %s", len(allPublished), strings.Join(allPublished, ", ")), nil
}
