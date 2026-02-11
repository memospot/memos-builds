#!/usr/bin/env -S just --justfile
# https://just.systems
# Build recipes for https://github.com/usememos/memos
#
# This justfile supports:
# - Building binary archives via Dagger
# - Building containers and loading them locally for testing
# - Screenshot capture for documentation
# - Version publishing workflows

set script-interpreter := ['bash', '-euo', 'pipefail']
set shell := ["bash", "-c"]

DOCKER_NAMES_FMT := "{{" + ".Names" + "}}"
DEFAULT_BUILD_OS := if os() == 'macos' { 'darwin' } else { os() }
DEFAULT_BUILD_ARCH := if arch() == 'aarch64' { 'arm64' } else if arch() == 'x86_64' { 'amd64' } else { arch() }
DEFAULT_BUILD_TARGET := DEFAULT_BUILD_OS + '/' + DEFAULT_BUILD_ARCH
GIT_WIN := join(env('PROGRAMFILES', ''), 'Git', 'usr', 'bin')
export PATH := if os() == 'windows' { GIT_WIN + ';' + env('PATH') } else { env('PATH') }
export CI := env("CI", "false")
export DAGGER_NO_NAG := "1"
export DO_NOT_TRACK := "1"

[private]
[script]
default:
    echo -e "{{ BOLD }}This justfile contains recipes for building {{ UNDERLINE }}https://github.com/usememos/memos{{ NORMAL }}\n"
    if [[ "{{ os() }}" == "windows" ]]; then
        program_files="{{ replace(env('PROGRAMFILES', 'C:\\Program Files'), '\\', '\\\\') }}"
        echo -e "To use this justfile on Windows, make sure Git is installed under {{ BOLD }}{{ UNDERLINE }}$program_files\\Git{{ NORMAL }}."
        echo -e "{{ BOLD }}{{ UNDERLINE }}https://git-scm.com/download/win{{ NORMAL }}"
        echo ""
    fi
    deps=(
        "dagger"
    )
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            echo -e "{{ RED }}ERROR:{{ NORMAL }} Please install {{ MAGENTA }}{{ BOLD }}{{ UNDERLINE }}$dep{{ NORMAL }}." && exit 1
        fi
    done
    echo -e "{{ GREEN }}Basic project tooling is installed.{{ NORMAL }}"
    echo -e "{{ YELLOW }}If you experience any errors, consider updating the related tool.{{ NORMAL }}\n"
    just --list

# Stop the Dagger engine container
dagger-stop:
    #!/usr/bin/env bash
    containers=$(docker container list --all --filter "name=^dagger-engine-*" --format '{{ DOCKER_NAMES_FMT }}')
    if [ -n "$containers" ]; then
        docker container stop $containers
    else
        echo "No Dagger engine containers found."
    fi

# Stop and remove the Dagger engine container
dagger-rm: dagger-stop
    #!/usr/bin/env bash
    containers=$(docker container list --all --filter 'name=^dagger-engine-*' --format '{{ DOCKER_NAMES_FMT }}')
    if [ -n "$containers" ]; then
        docker container rm $containers
    else
        echo "No Dagger engine containers to remove."
    fi

# Clean Dagger cache and Docker build cache
[confirm('This will aggressively prune ALL Docker resources (not just Dagger). This may delete data from other projects. Are you sure?')]
dagger-clean:
    #!/usr/bin/env bash
    echo "Pruning Dagger cache…"
    dagger core engine local-cache prune || true
    echo "Pruning Docker build cache…"
    docker builder prune -a -f || true
    echo "Pruning Docker images…"
    docker image prune -a -f || true
    echo "Pruning Docker volumes…"
    docker volume prune -a -f || true
    echo "Pruning Docker system…"
    docker system prune -a -f || true
    echo "Dagger and Docker cleanup complete."

# Regenerate Dagger files after SDK changes
dagger-dev:
    dagger develop --compat=skip

[doc('Build Memos binaries for the specified version and platforms.
    - VERSION: v*.*.*, nightly, or commit hash.
    - PLATFORMS: Comma-separated list (e.g., "linux/amd64,darwin/arm64") or "all".')]
build VERSION='nightly' PLATFORMS='':
    #!/usr/bin/env bash
    if [ -n "{{ PLATFORMS }}" ]; then
        PLATFORMS="{{ PLATFORMS }}"
    else
        PLATFORMS="{{ DEFAULT_BUILD_TARGET }}"
    fi
    echo -e "Building {{ BLUE }}{{ VERSION }}{{ NORMAL }} for {{ BLUE }}${PLATFORMS}{{ NORMAL }}…"
    dagger call build --source=. --version="{{ VERSION }}" --platforms="${PLATFORMS}" export --path=./dist
    echo -e "{{ GREEN }}Build complete. Artifacts in ./dist/{{ NORMAL }}"

[doc('Build Memos containers for the specified version and platforms.
    - VERSION: v*.*.*, nightly, or commit hash.
    - PLATFORMS: Comma-separated list (e.g., "linux/amd64,darwin/arm64") or "all".')]
build-docker VERSION='nightly' PLATFORMS='':
    #!/usr/bin/env bash
    set -euo pipefail

    # Version to commit mapping for known versions
    declare -A known_versions=(
        ["0.25.2"]="bfad0708e2c8062664e852f6f18223fd943ad5f5"
        ["0.25.3"]="07a030ddfdbe5ac8a22c235be7b5771cc01f8498"
        ["0.26.0"]="43b5a51ec73214d3c56aa48c82783ccfeec1a127"
        ["0.26.1"]="b623162d37f87f9f174d8f6cd8e54c7034cfc789"
    )

    version="{{ VERSION }}"
    original_version="$version"

    # Resolve version to commit if it's a known version
    if [[ "$version" =~ ^v?([0-9]+\.[0-9]+\.[0-9]+)$ ]]; then
        ver="${BASH_REMATCH[1]}"
        if [[ -n "${known_versions[$ver]:-}" ]]; then
            commit="${known_versions[$ver]}"
            echo -e "{{ YELLOW }}Using known commit hash ${commit} for version ${ver}{{ NORMAL }}"
            version="$commit"
        fi
    fi

    # Determine platforms
    platforms="{{ PLATFORMS }}"
    if [ -z "$platforms" ]; then
        platforms="{{ DEFAULT_BUILD_TARGET }}"
    fi

    echo -e "{{ CYAN }}Building containers for version ${original_version} (${version})…{{ NORMAL }}"

    # Clean old build directory
    rm -rf ./dist/container
    mkdir -p ./dist/container

    # Build containers via Dagger
    echo -e "{{ BLUE }}Calling Dagger to build containers…{{ NORMAL }}"
    if ! dagger call build-containers \
        --source . \
        --version "$version" \
        --platforms "$platforms" \
        export \
        --path ./dist/container; then
        echo -e "{{ RED }}ERROR: Dagger build failed{{ NORMAL }}"
        exit 1
    fi

    # Find all exported tarballs
    tarballs=(./dist/container/*.tar)
    if [ ! -f "${tarballs[0]:-}" ]; then
        echo -e "{{ RED }}ERROR: No container tarballs found in ./dist/container/{{ NORMAL }}"
        exit 1
    fi

    echo -e "{{ GREEN }}Found ${#tarballs[@]} container tarball(s){{ NORMAL }}"

    # Arrays to track containers
    declare -a container_names=()
    declare -a container_ports=()
    declare -a container_platforms=()

    for tarball in "${tarballs[@]}"; do
        filename=$(basename "$tarball")
        echo -e "\n{{ CYAN }}Processing ${filename}…{{ NORMAL }}"

        # Extract platform from filename (memos-linux-amd64.tar -> linux-amd64)
        platform_dashed="${filename#memos-}"
        platform_dashed="${platform_dashed%.tar}"

        # Load the image
        echo -e "{{ BLUE }}Loading image from tarball…{{ NORMAL }}"
        load_output=$(docker load -i "$tarball" 2>&1)

        # Extract image ID from docker load output
        image_id=""
        if [[ "$load_output" =~ Loaded\ image\ ID:\ (sha256:[a-f0-9]+) ]]; then
            image_id="${BASH_REMATCH[1]}"
        elif [[ "$load_output" =~ Loaded\ image:\ ([^[:space:]]+) ]]; then
            image_id="${BASH_REMATCH[1]}"
        fi

        if [ -z "$image_id" ]; then
            echo -e "{{ RED }}ERROR: Failed to extract image ID from docker load output{{ NORMAL }}"
            echo "$load_output"
            continue
        fi

        echo -e "{{ GREEN }}Loaded image: ${image_id}{{ NORMAL }}"

        # Create tag and container name
        container_name="memos-${original_version}-${platform_dashed}"
        image_tag="${container_name}:local"

        # Tag the image
        if ! docker tag "$image_id" "$image_tag" 2>/dev/null; then
            echo -e "{{ YELLOW }}Warning: Failed to tag image (may already be tagged){{ NORMAL }}"
        fi

        # Cleanup old container with same name
        echo -e "{{ BLUE }}Cleaning up old container ${container_name}…{{ NORMAL }}"
        docker stop "$container_name" 2>/dev/null || true
        docker rm -f "$container_name" 2>/dev/null || true

        # Get platform info from image
        echo -e "{{ BLUE }}Inspecting image platform…{{ NORMAL }}"
        inspect_output=$(docker inspect "$image_tag" 2>&1) || inspect_output=$(docker inspect "$image_id" 2>&1)

        img_os=$(echo "$inspect_output" | grep -o '"Os": "[^"]*"' | head -1 | cut -d'"' -f4) || img_os=""
        img_arch=$(echo "$inspect_output" | grep -o '"Architecture": "[^"]*"' | head -1 | cut -d'"' -f4) || img_arch=""
        img_variant=$(echo "$inspect_output" | grep -o '"Variant": "[^"]*"' | head -1 | cut -d'"' -f4) || img_variant=""

        docker_platform="${img_os}/${img_arch}"
        if [ -n "$img_variant" ]; then
            docker_platform="${docker_platform}/${img_variant}"
        fi

        echo -e "{{ CYAN }}Platform: ${docker_platform}{{ NORMAL }}"

        # Run the container
        echo -e "{{ BLUE }}Starting container ${container_name}…{{ NORMAL }}"
        if ! docker run \
            --rm \
            --detach \
            --init \
            --name "$container_name" \
            --publish "0:5230" \
            --env "MEMOS_DEMO=true" \
            --env "TZ=America/Sao_Paulo" \
            --platform "$docker_platform" \
            "$image_tag" 2>/dev/null; then
            echo -e "{{ RED }}ERROR: Failed to start container ${container_name}{{ NORMAL }}"
            continue
        fi

        # Wait for container to start
        sleep 2

        # Check if container is still running
        if ! docker inspect -f '{{{{.State.Running}}}}' "$container_name" 2>/dev/null | grep -q "true"; then
            echo -e "{{ RED }}ERROR: Container ${container_name} exited unexpectedly{{ NORMAL }}"
            echo -e "{{ YELLOW }}Container logs:{{ NORMAL }}"
            docker logs "$container_name" 2>&1 || true
            continue
        fi

        # Get the mapped port
        port_output=$(docker port "$container_name" 5230 2>/dev/null || true)
        if [ -z "$port_output" ]; then
            echo -e "{{ YELLOW }}Warning: No public port found for ${container_name} (init might be slow){{ NORMAL }}"
            continue
        fi

        # Extract port number (format: 0.0.0.0:PORT or :::PORT)
        port=$(echo "$port_output" | grep -oE '[0-9]+$' | head -1)

        if [ -z "$port" ]; then
            echo -e "{{ YELLOW }}Warning: Could not parse port for ${container_name}{{ NORMAL }}"
            continue
        fi

        container_names+=("$container_name")
        container_ports+=("$port")
        container_platforms+=("$docker_platform")

        echo -e "{{ GREEN }}Container ${container_name} running on port ${port}{{ NORMAL }}"
    done

    # Show running containers
    echo -e "\n{{ CYAN }}=== Running Containers ==={{ NORMAL }}"
    docker ps --filter "name=memos-"

    # Final report
    echo -e "\n{{ BOLD }}Build Results:{{ NORMAL }}"
    if [ ${#container_names[@]} -eq 0 ]; then
        echo -e "{{ RED }}No containers are running.{{ NORMAL }}"
        exit 1
    fi

    for i in "${!container_names[@]}"; do
        echo -e "  {{ GREEN }}${container_names[$i]} {{ CYAN }}→ http://localhost:${container_ports[$i]}{{ NORMAL }}"
    done

    echo -e "\n{{ GREEN }}✓ Build and load complete!{{ NORMAL }}"

# Clean built Docker containers and images
[confirm('This will stop and remove all Docker containers and images starting with memos-. Are you sure?')]
clean-docker:
    #!/usr/bin/env bash
    set -euo pipefail
    containers=$(docker ps -a --filter "name=^memos-" --format='{{ "{{.Names}}" }}' || true)
    if [ -n "$containers" ]; then
        echo "$containers" | while read -r ct; do
            echo "Stopping container ${ct}..."
            docker stop "$ct" >/dev/null 2>&1 || true
            docker rm "$ct" >/dev/null 2>&1 || true
        done
    fi
    images=$(docker images --filter "reference=memos-*" --format '{{ "{{.Repository}}:{{.Tag}}" }}' || true)
    if [ -n "$images" ]; then
        echo "$images" | while read -r img; do
            echo "Removing image ${img}..."
            docker rmi "$img" >/dev/null 2>&1 || true
        done
    fi
    docker image prune -f
    echo "Cleaning complete."

# Clean build artifacts, Docker cache, and optionally Go cache
[confirm('This will clean build artifacts, Go cache, dangling Docker images and Docker build cache. Are you sure?')]
[script]
clean GOCACHE='false':
    echo "Cleaning build artifacts…"
    rm -rf dist/

    echo "Cleaning Docker…"
    docker builder prune -f
    docker buildx prune -f 2>/dev/null || true
    docker image prune -f

    if [[ "{{ GOCACHE }}" == "true" ]]; then
        echo "Cleaning Go cache…"
        go clean -cache -modcache
    fi

    echo "Cleaning complete."

# Reset main branch to origin/main (destructive)
[confirm('This will exclude ANY changes and untracked files on the working tree, resetting the local repo to origin/main. Are you sure?')]
[script]
git-reset:
    git fetch origin
    git checkout main
    git reset --hard origin/main
    git clean -fdx
    git checkout -- .
    git checkout -

# Remove a git tag and push it again (internal use)
[private]
[script]
git-retag TAG:
    set +e
    TAG="v{{ trim_start_matches(TAG, 'v') }}"
    git push origin :refs/tags/$TAG
    git tag -d $TAG
    git tag -a $TAG -m "Tag $TAG"
    git push origin $TAG

[doc('Tag and push to GitHub, triggering the release workflow.')]
[script]
publish TAG:
    TAG="v{{ trim_start_matches(TAG, 'v') }}"
    just git-retag "$TAG"
    git push origin main

[doc('Update README.md captures. Requires a running Memos instance and bunx.')]
update-captures PORT='':
    #!/usr/bin/env bash
    port="{{ PORT }}"
    if [ -z "${port}" ]; then
        container="$(docker ps --format '{{ DOCKER_NAMES_FMT }}' | grep -m1 '^memos-' || true)"
        if [ -z "${container}" ]; then
            echo "ERROR: No running Docker containers found matching memos-*."
            exit 1
        fi
        port="$(docker port "$container" 5230/tcp | head -n1 | cut -d: -f2)"
    fi
    ARGS=(
        --launch-options='{"args":["--accept-lang=en-US"]}'
        --module='document.querySelector("article").remove()'
        --width=1280
        --height=800
        --scale-factor=3
        --overwrite
        --type=webp
        http://localhost:${port}/
    )

    bunx capture-website \
        ${ARGS[@]} \
        --local-storage="memos-theme=midnight" \
        --output=assets/capture_dark.webp --dark-mode &

    bunx capture-website \
        ${ARGS[@]} \
        --local-storage="memos-theme=paper" \
        --output=assets/capture_light.webp \
        http://localhost:${port}/
