package main

import (
	"context"
	"dagger/memos-builds/buildconsts"
	"dagger/memos-builds/internal/dagger"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
)

// Generate Proto code
func (m *MemosBuilds) generateProto(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From(buildconsts.BUF_IMAGE).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithWorkdir("/src").
		WithDirectory("/src", source).
		WithWorkdir("/src/proto").
		WithExec([]string{"buf", "generate"}).
		WithWorkdir("/src").
		Directory("/src")
}

// Build the frontend
func (m *MemosBuilds) buildFrontend(source *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From(buildconsts.NODE_BUILD_IMAGE).
		WithExec([]string{"corepack", "enable"}).
		WithMountedCache("/root/.local/share/pnpm/store", dag.CacheVolume("pnpm-store")).
		WithDirectory("/app/web", source.Directory("web")).
		WithWorkdir("/app/web").
		WithMountedCache("/app/web/node_modules", dag.CacheVolume("node-modules")).
		WithExec([]string{"pnpm", "install"}).
		WithWorkdir("/app/web").
		WithExec([]string{"pnpm", "run", "release"}).
		Directory("/app/server/router/frontend/dist")
}

// Build the backend binaries for the given targets.
// Builds are dispatched in batches to control resource usage.
// Batch size defaults to NumCPU-1, or NumCPU when CI=true.
func (m *MemosBuilds) buildBackend(
	ctx context.Context,
	source *dagger.Directory,
	frontendDist *dagger.Directory,
	version string,
	targets []BuildMatrix,
) (*dagger.Directory, error) {
	maxConcurrent := max(runtime.NumCPU()-1, 1)
	if os.Getenv("CI") == "true" {
		maxConcurrent = runtime.NumCPU()
	}

	ldflags := []string{
		"-s",
		"-w",
		"-extldflags '-static'",
	}

	// Memos migrations will fail if we override this field with gibberish.
	v, err := semver.NewVersion(version)
	if err == nil {
		// https://pkg.go.dev/cmd/link
		ldflags = append(ldflags, fmt.Sprintf("-X %s=%s", buildconsts.VERSION_IMPORT_PATH, v.String()))
	}

	buildOne := func(c *dagger.Container, t BuildMatrix) *dagger.File {
		name := t.BinaryName()

		// Set architecture-specific environment variables.
		ctr := c.
			WithEnvVariable("GOMAXPROCS", fmt.Sprint(maxConcurrent)).
			WithEnvVariable("CGO_ENABLED", "0").
			WithEnvVariable("GOOS", t.OS).
			WithEnvVariable("GOARCH", t.Arch)

		// Set architecture level variables based on the architecture.
		if goarm := t.GoArm(); goarm != "" {
			ctr = ctr.WithEnvVariable("GOARM", goarm)
		}
		if goamd64 := t.GoAmd64(); goamd64 != "" {
			ctr = ctr.WithEnvVariable("GOAMD64", goamd64)
		}
		if go386 := t.Go386(); go386 != "" {
			ctr = ctr.WithEnvVariable("GO386", go386)
		}
		if goppc64 := t.GoPpc64(); goppc64 != "" {
			ctr = ctr.WithEnvVariable("GOPPC64", goppc64)
		}
		if goriscv64 := t.GoRiscv64(); goriscv64 != "" {
			ctr = ctr.WithEnvVariable("GORISCV64", goriscv64)
		}

		return ctr.
			WithExec([]string{"go", "build",
				"-trimpath",
				"-ldflags", strings.Join(ldflags, " "),
				"-tags", "netgo,osusergo",
				"-o", "/out/" + name,
				buildconsts.APP_ENTRYPOINT,
			}).
			File("/out/" + name)
	}

	base := dag.Container().
		From(buildconsts.GOLANG_BUILD_IMAGE).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir("/src").
		WithDirectory("/src", source).
		WithDirectory("/src/server/router/frontend/dist", frontendDist).
		// Tidy is required as go.mod may have been patched at earlier steps.
		WithExec([]string{"go", "mod", "tidy", "-go=" + buildconsts.GO_VERSION}).
		WithDirectory("/out", dag.Directory())

	out := dag.Directory()
	for _, t := range targets {
		f := buildOne(base, t)
		out = out.WithFile(t.BinaryName(), f)
		// Sync forces this build to complete before starting the next.
		out, err = out.Sync(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to build %s: %w", t.BinaryName(), err)
		}
	}

	return out, nil
}
