package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"

	"github.com/Masterminds/semver/v3"
)

// PublishedImage describes a container image published to a registry.
type PublishedImage struct {
	Registry string `json:"registry"`         // e.g. "ghcr.io/memospot/memos-builds"
	Tag      string `json:"tag"`              // e.g. "nightly"
	Digest   string `json:"digest,omitempty"` // e.g. "sha256:abc123..."
	Ref      string `json:"ref"`              // full ref with digest
}

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

// containerTags returns tags for release images.
// Releases (v0.25.3): ["latest", "0.25", "0.25.3"]
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

func (m *MemosBuilds) ghcrNightlyTags(version string) []string {
	v, err := semver.NewVersion(version)
	if err != nil {
		return []string{"nightly"}
	}

	tag, _ := nightlyReleaseTag("nightly", v.Metadata())
	if tag == "nightly" {
		return []string{"nightly"}
	}

	return []string{tag, "nightly"}
}

func (m *MemosBuilds) tagsForRegistry(version string, registry string) []string {
	if nightlyTag, ok := nightlyReleaseTag(version, ""); ok {
		if strings.HasPrefix(registry, "ghcr.io") {
			if nightlyTag == "nightly" {
				return []string{"nightly"}
			}

			return []string{nightlyTag, "nightly"}
		}

		return []string{"nightly"}
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		return []string{"latest"}
	}

	if v.Prerelease() == "" {
		return m.containerTags(version)
	}

	if strings.HasPrefix(registry, "ghcr.io") {
		return m.ghcrNightlyTags(version)
	}

	return []string{"nightly"}
}

// parsePublishedRef extracts the digest from a fully-qualified image reference.
// Input: "ghcr.io/memospot/memos-builds:nightly@sha256:abc123..."
// Output: "sha256:abc123..."
func parsePublishedRef(ref string) string {
	if idx := strings.LastIndex(ref, "@"); idx != -1 {
		return ref[idx+1:]
	}
	return ""
}

// publishContainers builds and publishes multi-arch Docker images to registries.
// Called internally by Publish — not exposed as a standalone Dagger function
// to avoid redundant Build calls.
func (m *MemosBuilds) publishContainers(
	ctx context.Context,
	source *dagger.Directory,
	gitSrc *dagger.Directory,
	buildVersion string,
	releaseVersion string,
	commit string,
	dockerHubUser string,
	dockerHubPassword *dagger.Secret,
	ghcrUser string,
	ghcrPassword *dagger.Secret,
) ([]PublishedImage, error) {
	// Build binaries for Linux targets only (containers are Linux-only).
	linuxTargets := filterLinuxTargets(TARGETS)
	if len(linuxTargets) == 0 {
		return nil, nil
	}

	platformVariants, err := m.buildContainers(ctx, source, gitSrc, buildVersion, commit, linuxTargets)
	if err != nil {
		return nil, fmt.Errorf("failed to build containers: %w", err)
	}

	var allPublished []PublishedImage

	publishTargets := []struct {
		registry string
		user     string
		password *dagger.Secret
		media    dagger.ImageMediaTypes
	}{
		{
			"docker.io/lincolnthalles/memos",
			dockerHubUser,
			dockerHubPassword,
			dagger.ImageMediaTypesOcimediaTypes,
		},
		{
			"ghcr.io/memospot/memos-builds",
			ghcrUser,
			ghcrPassword,
			dagger.ImageMediaTypesOcimediaTypes,
		},
	}
	for _, target := range publishTargets {
		if target.user != "" && target.password != nil {
			address := strings.Split(target.registry, "/")[0]
			tags := m.tagsForRegistry(releaseVersion, target.registry)
			publisher := platformVariants[0].
				WithRegistryAuth(address, target.user, target.password)
			for _, tag := range tags {
				addr := fmt.Sprintf("%s:%s", target.registry, tag)
				ref, err := publisher.Publish(ctx, addr, dagger.ContainerPublishOpts{
					PlatformVariants: platformVariants[1:],
					MediaTypes:       target.media,
				})
				if err != nil {
					return nil, fmt.Errorf("failed to publish to %s: %w", target.registry, err)
				}
				allPublished = append(allPublished, PublishedImage{
					Registry: target.registry,
					Tag:      tag,
					Digest:   parsePublishedRef(ref),
					Ref:      ref,
				})
			}
		}
	}

	return allPublished, nil
}

// PublishedImagesJSON returns a JSON-serialised []PublishedImage suitable for
// writing to a file in the dist directory.
func PublishedImagesJSON(images []PublishedImage) (string, error) {
	b, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
