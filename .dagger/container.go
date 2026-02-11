// Container build definitions.
package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"time"
)

// addContainerAnnotations adds OCI labels to a container.
func (m *MemosBuilds) addContainerAnnotations(c *dagger.Container) *dagger.Container {
	labels := map[string]string{
		"title":       "Memos",
		"description": "A privacy-first, lightweight note-taking service.",
		"licenses":    "MIT",
		"url":         "https://usememos.com",
		"vendor":      "Memospot",
		"source":      "https://github.com/memospot/memos-builds",
		"created":     time.Now().Format(time.RFC3339),
	}
	for k, v := range labels {
		c = c.WithLabel("org.opencontainers.image."+k, v)
	}
	return c
}

// addContainerEnv adds environment variables to a container.
func (m *MemosBuilds) addContainerEnv(c *dagger.Container) *dagger.Container {
	vars := map[string]string{
		"TZ":         "UTC",
		"MEMOS_DATA": "/var/opt/memos",
		"MEMOS_PORT": "5230",
	}
	for k, v := range vars {
		c = c.WithEnvVariable(k, v)
	}
	return c
}

// buildContainer creates a container for a specific platform using the built binary.
func (m *MemosBuilds) buildContainer(
	// The specific binary file for this platform
	binary *dagger.File,
	// Platform string (e.g. linux/amd64, linux/arm/v5)
	platform string,
	source *dagger.Directory,
) *dagger.Container {
	// ARMv5 requires special handling:
	// 	- It's only supported on BusyBox.
	// 	- BusyBox lacks package manager, tz-data and needs a pre-built su-exec.
	if platform == "linux/arm/v5" {
		return m.buildBusyBoxARMv5Container(binary, platform, source)
	}

	// All other platforms use a standard Alpine-based container.
	return m.buildAlpineContainer(binary, platform, source)
}

// buildAlpineContainer creates an Alpine-based container.
func (m *MemosBuilds) buildAlpineContainer(
	binary *dagger.File,
	platform string,
	source *dagger.Directory,
) *dagger.Container {

	entrypoint := source.Directory("container").File("entrypoint.sh")
	newFilePerms := dagger.ContainerWithFileOpts{Permissions: 0755}

	return dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(platform)}).
		From(buildconsts.PRIMARY_IMAGE).
		With(m.addContainerAnnotations).
		WithExec([]string{"apk", "add", "--no-cache", "tzdata", "ca-certificates", "su-exec"}).
		WithDirectory("/var/opt/memos", dag.Directory()).
		WithWorkdir("/usr/local/bin").
		WithFile("/usr/local/bin/memos", binary, newFilePerms).
		WithNewFile("/usr/share/memos/buildinfo", fmt.Sprintf("TARGETPLATFORM=%s\n", platform)).
		WithExec([]string{"sh", "-c", `
			echo "SHA256SUM=$(sha256sum /usr/local/bin/memos | cut -d' ' -f1)" >> /usr/share/memos/buildinfo
		`}).
		WithFile("/init", entrypoint, newFilePerms).
		With(m.addContainerEnv).
		WithExposedPort(5230).
		WithUser("root").
		WithEntrypoint([]string{"/init"}).
		WithDefaultArgs([]string{"/usr/local/bin/memos"})
}

// buildBusyBoxARMv5Container creates a BusyBox-based container.
//
// # Notes
//   - Updated tz-data and ca-certificates are copied from Google's distroless image.
//   - su-exec is injected from a precompiled ARMv5 binary.
//   - Isolated to make it easier to remove if ARMv5 gets too difficult to support.
func (m *MemosBuilds) buildBusyBoxARMv5Container(
	binary *dagger.File,
	platform string,
	source *dagger.Directory,
) *dagger.Container {

	entrypointFile := source.Directory("container").File("entrypoint.sh")
	suExecFile := source.Directory("container/armv5/bin").File("su-exec")
	newFilePerms := dagger.ContainerWithFileOpts{Permissions: 0755}

	// Get updated tzdata and ca-certificates from Google's distroless image.
	distrolessCt := dag.Container().From("gcr.io/distroless/static:latest")

	return dag.Container(dagger.ContainerOpts{Platform: dagger.Platform(platform)}).
		From(buildconsts.ALTERNATE_IMAGE).
		With(m.addContainerAnnotations).
		WithFile("/init", entrypointFile, newFilePerms).
		WithFile("/usr/local/bin/su-exec", suExecFile, newFilePerms).
		WithDirectory("/usr/share/zoneinfo", distrolessCt.Directory("/usr/share/zoneinfo")).
		WithDirectory("/etc/ssl/certs", distrolessCt.Directory("/etc/ssl/certs")).
		WithDirectory("/var/opt/memos", dag.Directory()).
		WithWorkdir("/usr/local/bin").
		WithFile("/usr/local/bin/memos", binary, newFilePerms).
		WithNewFile("/usr/share/memos/buildinfo", fmt.Sprintf("TARGETPLATFORM=%s\n", platform)).
		// Add `BusyBox v* (build date) multi-call binary.` to `/etc/os-release`
		WithExec([]string{"sh", "-c", `
				if ! [ -f /etc/os-release ]; then
						busybox=$(find --help 2>&1 | head -1)
						if [ "$(echo "$busybox" | cut -d' ' -f1)" = "BusyBox" ]; then
								echo "PRETTY_NAME=\"$busybox\"" > /etc/os-release
						fi
				fi
		`}).
		WithExec([]string{"sh", "-c", `
			echo "SHA256SUM=$(sha256sum /usr/local/bin/memos | cut -d' ' -f1)" >> /usr/share/memos/buildinfo
		`}).
		With(m.addContainerEnv).
		WithExposedPort(5230).
		WithUser("root").
		WithEntrypoint([]string{"/init"}).
		WithDefaultArgs([]string{"/usr/local/bin/memos"})
}

// buildContainers creates multiple containers for the specified targets.
// It shares the build logic (proto, frontend, backend) to avoid redundancy.
func (m *MemosBuilds) buildContainers(
	ctx context.Context,
	source *dagger.Directory,
	version string,
	targets []BuildMatrix,
) ([]*dagger.Container, error) {
	// 1. Resolve version and source
	gitSrc, version, err := m.prepareSource(ctx, source, version)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare source: %w", err)
	}

	// 2. Generate proto and build frontend (shared across all targets)
	gitSrc = m.generateProto(gitSrc)
	frontendDist := m.buildFrontend(gitSrc)

	// 3. Build backend binaries for all requested targets
	binaries, err := m.buildBackend(ctx, gitSrc, frontendDist, version, targets)
	if err != nil {
		return nil, fmt.Errorf("failed to build binaries: %w", err)
	}

	// 4. Create container instances for each target
	var containers []*dagger.Container
	for _, t := range targets {
		binary := binaries.File(t.BinaryName())
		ctr := m.buildContainer(binary, t.DockerPlatform(), source)
		containers = append(containers, ctr)
	}

	return containers, nil
}
