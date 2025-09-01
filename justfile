# https://just.systems

# Typical internal recipe workflow for a Memos build:
# git-subtree-pull -> build-buf -> build-backend-tidy -> build-frontend -> [goreleaser] -> [docker] -> [publish]
#
# Run `just` in the root of the project to see a list of recipes relevant to manual builds.

set shell := ["bash", "-c"]
CI := env_var_or_default("CI", "false")
NPROC := env_var_or_default("NPROC", num_cpus())
GITHUB_ENV := env_var_or_default("GITHUB_ENV", ".GITHUB_ENV")

PATH := if os() == "windows" {
		env_var_or_default('PROGRAMFILES', 'C:\Program Files') + '\Git\usr\bin;' + env_var_or_default('PATH','')
	} else {
		env_var_or_default('PATH','')
	}
bash := if os() == "windows" { "env -S bash -euo pipefail" } else { "/usr/bin/env -S bash -euo pipefail" }

RESET := '\033[0m'
BOLD := '\033[1m'
DIM := '\033[2m'
UNDERLINE := '\033[4m'
BLACK := '\033[30m'
RED := '\033[31m'
GREEN := '\033[32m'
YELLOW := '\033[33m'
BLUE := '\033[34m'
MAGENTA := '\033[35m'
CYAN := '\033[36m'
WHITE := '\033[37m'

set export

[private]
default:
    #!{{bash}}
    echo -e "${BOLD}This justfile contains recipes for building v0.21.1 and onwards versions of ${UNDERLINE}https://github.com/usememos/memos${RESET}.\n"
    if [[ "{{os()}}" == "windows" ]]; then
        program_files="{{replace(env_var_or_default('PROGRAMFILES', 'C:\Program Files'), '\\', '\\\\')}}"
        echo -e "To use this justfile on Windows, make sure Git is installed under ${BOLD}${UNDERLINE}$program_files\\Git${RESET}."
        echo -e "${BOLD}${UNDERLINE}https://git-scm.com/download/win${RESET}"
        echo ""
    fi
    deps=(
        "buf"
        "git"
        "go"
        "goreleaser"
        "node"
        "pnpm"
    )
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            echo -e "${RED}ERROR:${RESET} Please install ${MAGENTA}${BOLD}${UNDERLINE}$dep${RESET}." & exit 1
        fi
    done
    echo -e "${GREEN}Basic project tooling is installed.${RESET}"
    echo -e "${YELLOW}This quick test does not verify tool versions. If you experience any errors, consider updating the related tool.${RESET}\n"
    just --list

# Tidy go.mod, ensuring compatibility with Go 1.25.
[private]
[group('CI')]
build-backend-tidy:
    #!{{bash}}
    set -euo pipefail
    cd memos
    go mod tidy -go=1.25

# Generate protobuf-related code.
[private]
[group('CI')]
build-buf:
    #!{{bash}}
    set -euo pipefail
    # `pnpm install` already does this at `build-frontend` recipe, but the front-end is built separately on CI.
    cd memos/proto
    buf generate

# Build front-end.
[private]
[group('CI')]
build-frontend:
    #!{{bash}}
    set -euo pipefail
    cd memos/web
    pnpm install
    pnpm release # added on v0.22.1

# Build nightly binaries. Front-end must be built beforehand.
[private]
[group('CI')]
build-nightly-backend-only: build-buf build-backend-tidy
    #!{{bash}}
    set -euo pipefail
    goreleaser release --snapshot --clean --timeout 60m

# Add variables to the GitHub Actions environment.
[private]
[group('CI')]
add-to-env KEY VALUE:
    #!{{bash}}
    set -euo pipefail
    echo -e " ${BOLD}${MAGENTA}$KEY${YELLOW}=${CYAN}$VALUE${RESET}"
    echo "$KEY=$VALUE" >> $GITHUB_ENV

[private]
[group('CI')]
setup-env NIGHTLY='':
    #!{{bash}}
    set -euo pipefail
    version_from_file=v$(grep --color=never -Po 'var Version = "\K[^"]+' < ./memos/internal/version/version.go)
    devversion_from_file=v$(grep --color=never -Po 'var DevVersion = "\K[^"]+' < ./memos/internal/version/version.go)
    version_from_git_tag=$(git describe --tags --abbrev=0)
    version_from_ref="{{replace(env_var_or_default('GITHUB_REF_NAME', 'NOT_SET'), 'release/', '')}}"
    git_previous_tag=""
    if [[ "{{NIGHTLY}}" == "--nightly" ]]; then
        echo -e "\n: Setting up ${BLUE}Nightly${RESET} build environment…"
        version_from_file=$(test -n "$devversion_from_file" && echo $devversion_from_file || echo $version_from_file)
    else
        echo -e "\n: Setting up ${GREEN}Release${RESET} build environment…"
    fi
    if [[ -n "$git_previous_tag" ]]; then
        git_previous_tag=$(git describe --tags --abbrev=0 --exclude=*-pre || echo "")
    fi
    version="v$(date +%Y.%m.%d).0"
    for v in $version_from_file $version_from_git_tag $version_from_ref; do
        if [[ $v =~ ^v[0-9]+\.[0-9]+\.[0-9]+ ]]; then
            version=$v
            break
        fi
    done
    echo -e "${MAGENTA}Current${RESET} Memos version is ${GREEN}$version${RESET}"
    just add-to-env "CACHE_KEY" $(date +%Y%m%d)
    just add-to-env "MEMOS_VERSION" $version_from_file
    just add-to-env "GIT_TAG" $version_from_git_tag
    just add-to-env "REF_VERSION" $version_from_ref
    canonical_version=$(echo $version | sed -E 's/^v//g')
    if [[ "{{NIGHTLY}}" == "--nightly" ]]; then
        canonical_version=$(echo $canonical_version | sed -E 's/-pre$//g')
        major=$(echo $canonical_version | grep -Po '^([0-9]+)')
        minor=$(echo $canonical_version | grep -Po '(?<=\.)([0-9]+)(?=\.)')
        patch=$(echo $canonical_version | grep -Po '(?<=\.)([0-9]+)$')
        patch=$((patch + 1))
        if [[ $patch -gt 99 ]]; then
            major=$((major + 1))
            patch=0
        fi
        nightly_version="v${major}.${minor}.${patch}-pre"
        echo -e "${MAGENTA}Build${RESET} version set to ${GREEN}$nightly_version${RESET}"
        just add-to-env "BUILD_VERSION" $nightly_version
        # If not set, goreleaser will infer tags from the git history.
        just add-to-env "GORELEASER_PREVIOUS_TAG" $version-pre
        just add-to-env "GORELEASER_CURRENT_TAG" $nightly_version
    else
        echo -e "${MAGENTA}Build${RESET} version set to ${GREEN}$version${RESET}"
        just add-to-env "BUILD_VERSION" $version
        just add-to-env "GORELEASER_CURRENT_TAG" $version
        major=$(echo $canonical_version | grep -Po '^([0-9]+)')
        minor=$(echo $canonical_version | grep -Po '(?<=\.)([0-9]+)(?=\.)')
        patch=$(echo $canonical_version | grep -Po '(?<=\.)([0-9]+)$')
        if [[ $patch -gt 0 ]]; then
            patch=$((patch - 1))
            previous_version="v${major}.${minor}.${patch}"
            just add-to-env "GORELEASER_PREVIOUS_TAG" $previous_version
        elif [[ -n "$git_previous_tag" ]]; then
            just add-to-env "GORELEASER_PREVIOUS_TAG" $git_previous_tag
        fi
    fi
    echo -e ": Build environment setup complete.\n"

# Commit any changes
[private]
[group('CI')]
git-commit-any MESSAGE='chore(ci): ensure clean git state': git-setup
    #!{{bash}}
    git add -A && git commit -m "{{MESSAGE}}" || true

# Remove a git tag and push it again.
[private]
[group('CI')]
git-retag TAG:
    #!{{bash}}
    set +e
    TAG="v{{trim_start_matches(TAG, 'v')}}"
    git push origin :refs/tags/$TAG
    git tag -d $TAG
    git tag -a $TAG -m "Tag $TAG"
    git push origin $TAG

[private]
[group('CI')]
git-setup:
    #!{{bash}}
    git_email="$(git config user.email || true)"
    git_user="$(git config user.name || true)"
    if [ -z "$git_email" ] || [ -z "$git_user" ]; then
        if [ "{{CI}}" == "true" ]; then
            git config --global user.email "github-actions[bot]@users.noreply.github.com"
            git config --global user.name "github-actions[bot]"
        else
            echo -e "${RED}ERROR: Git user email or name is not set.${RESET}"
            echo "To fix this issue, run the following commands:"
            echo -e " ${BOLD}${UNDERLINE}git config [--global] user.email <your-email>${RESET}"
            echo -e " ${BOLD}${UNDERLINE}git config [--global] user.name <your-name>${RESET}"
            exit 1
        fi
    fi

# Pull a specific branch or commit from usememos/memos.
[private]
[group('CI')]
git-subtree-pull COMMIT: git-setup
    #!{{bash}}
    set -euo pipefail
    git subtree pull --prefix=memos "https://github.com/usememos/memos.git" "{{COMMIT}}" --squash --message="chore(ci): pull {{COMMIT}} from usememos/memos"

    if [ -d "{{justfile_directory()}}/patches" ]; then
        echo -e "${MAGENTA}Applying patches from ./patches to ./memos…${RESET}"
        for patch in {{justfile_directory()}}/patches/*.patch ; do
            echo -e "  Checking for patch ${CYAN}$patch${RESET}…"
            if [ -f "$patch" ]; then
                echo -e "  Applying ${CYAN}$patch${RESET}…"
                if git -C memos apply "$patch"; then
                    echo -e "${GREEN}Patch $patch applied successfully.${RESET}"
                    git add memos && git commit -m "chore(ci): apply patches" || true
                else
                    echo -e "${YELLOW}Failed to apply patch $patch. Skipping.${RESET}"
                    continue
                fi
            fi
        done
    fi

# Pull a specific tag from usememos/memos.
[private]
[group('CI')]
git-subtree-pull-tag TAG:
    #!{{bash}}
    TAG="v{{trim_start_matches(TAG, 'v')}}"
    just git-subtree-pull "tags/$TAG"

# Rename goreleaser build artifacts to the format expected by the Dockerfile.
[private]
[group('CI')]
rename-to-docker:
    #!{{bash}}
    declare -A go_to_docker
    go_to_docker["amd64_v1"]="amd64"
    go_to_docker["amd64_v([2-4])"]="amd64v\1"
    go_to_docker["arm_([5-7])"]="armv\1"
    # GoReleaser >= v2.4.0
    go_to_docker["386_sse2"]="386"
    go_to_docker["arm64_v8.0"]="arm64"
    go_to_docker["mips64le_hardfloat"]="mips64le"
    go_to_docker["ppc64le_power8"]="ppc64le"
    go_to_docker["riscv64_rva20u64"]="riscv64"

    echo -e "\n${MAGENTA}: Renaming goreleaser builds to the format expected by the Dockerfile…${RESET}"
    for folder in $(find ./build/memos_linux* -type d); do
        folder_name=$(basename $folder)
        for go in "${!go_to_docker[@]}"; do
            if [[ $folder_name =~ $go ]]; then
                new_name=$(echo "$folder_name" | sed -E "s/$go/${go_to_docker[$go]}/g")
                if [[ $new_name == $folder_name ]]; then
                    echo -e "  ${CYAN}Skipping${RESET} ${BLUE}$folder_name${RESET}"
                    continue
                fi
                if [ -z "$new_name" ]; then
                    echo -e "  ${RED}Failed to rename${RESET} ${YELLOW}$folder_name${RESET}"
                    continue
                fi
                new_folder="./build/$new_name"
                echo -en "  Renaming ${CYAN}$folder_name${RESET} to ${BLUE}$new_name${RESET}…"
                if mv "$folder" "$new_folder"; then
                    echo -e " ${GREEN}SUCCESS${RESET}"
                    continue
                fi
                echo -e " ${RED}FAILED${RESET}"
            fi
        done
    done
    echo -e ": Renaming complete.\n"

# Build Memos tag (v*.*.* or nightly). Use `--cross` to cross-compile for all supported platforms.
build TAG CROSS='':
    #!{{bash}}
    set -euo pipefail
    TAG="v{{trim_start_matches(TAG, 'v')}}"
    CROSS_COMPILE=false
    if [[ "{{TAG}}" == "nightly" ]]; then
        just git-subtree-pull main
    elif [[ "{{TAG}}" == "testing" ]]; then
        echo -e "${YELLOW}Testing mode. Skipping subtree update.${RESET}"
    else
        just git-subtree-pull "tags/$TAG"
    fi
    just build-buf build-backend-tidy build-frontend
    if [[ "{{CROSS}}" == "--cross" ]] || [[ "{{CROSS}}" == "-c" ]]; then
        CROSS_COMPILE=true
    fi
    if $CROSS_COMPILE; then
        echo -e "${MAGENTA}Cross-compiling for all supported platforms.${RESET}\n"
        goreleaser release --skip=publish --clean --skip=validate --parallelism={{NPROC}} --timeout=60m
    else
        echo -e "${MAGENTA}Building for the current platform.${RESET}"
        echo -e "Use ${BOLD}${UNDERLINE}--cross${RESET} to cross-compile for all supported platforms.${RESET}\n"
        goreleaser build  --single-target --clean --skip=validate --parallelism={{NPROC}}
    fi

# Build local Docker images. If the current OS is not Linux, `--cross` will be passed implicitly.
build-docker TAG CROSS='':
    #!{{bash}}
    set -euo pipefail
    CROSS="{{CROSS}}"
    if [[ "{{os()}}" != "linux" ]] && [[ -z "$CROSS" ]]; then
        CROSS="--cross"
    fi
    TAG="v{{trim_start_matches(TAG, 'v')}}"
    just build "{{TAG}}" "$CROSS"
    just rename-to-docker
    export DOCKER_BUILDKIT=1
    export DOCKER_CLI_EXPERIMENTAL=enabled
    docker_platforms=()
    for plat in $(find ./build/memos_linux* -type d); do
        go_name=$(basename $plat)
        docker_platform="${go_name#memos_}"
        docker_platform="${docker_platform//_//}"
        docker_platform=$(echo "$docker_platform" | sed -E 's/^(armv[5-7])|(v[1-4])/\L\1\2/')
        docker_platforms+=("$docker_platform")
    done
    echo -e "${GREEN}>> Will build Docker images for the following platforms:${RESET}"
    for platform in "${docker_platforms[@]}"; do
        echo -e "  ${GREEN}${platform}${RESET}"
    done
    for platform in "${docker_platforms[@]}"; do
        container_name="memos-${TAG}-${platform//\//-}"
        docker stop $container_name >/dev/null 2>&1 || true
        docker rm -f $container_name >/dev/null 2>&1 || true
        docker ps -a --filter ancestor=$container_name --format="\{\{.ID\}\}" \
            | xargs -r docker rm -f >/dev/null 2>&1 || true
        docker image rm -f $container_name >/dev/null 2>&1 || true
    done
    docker image prune -f
    docker volume prune -f
    declare -A listen_ports
    for platform in "${docker_platforms[@]}"; do
        container_name="memos-${TAG}-${platform//\//-}"
        echo -e "${MAGENTA}Building Docker image ${container_name} for ${platform}.${RESET}"
        docker buildx build --tag $container_name --file ./docker/Dockerfile . --platform=$platform
        if [ $? -ne 0 ]; then
            echo -e "${RED}>> Docker build failed for ${platform} <<${RESET}"
        fi
        docker run --detach --init --rm --name $container_name \
            --publish "0:5230" --env MEMOS_MODE=demo \
            --env TZ=America/Sao_Paulo \
            --platform $platform \
            $container_name
        if [ $? -ne 0 ]; then
            echo -e "${RED}>> Failed to run container ${platform} <<${RESET}"
        else
            port=$(docker port $container_name 5230 | cut -d':' -f2)
            listen_ports["$container_name"]="$port"
        fi
    done
    docker ps
    if [ ${#listen_ports[@]} -eq 0 ]; then
        echo -e "${RED}No containers are running.${RESET}"
        exit 1
    fi
    echo "Build results:"
    for ct in "${!listen_ports[@]}"; do
        echo -e "  ${GREEN}${ct}${RESET}\tis listening on ${CYAN}http://localhost:${listen_ports[$ct]}${RESET}"
    done

# Clean built Docker images
[confirm('This will stop and remove all Docker containers and images starting with memos-v*. Are you sure?')]
clean-docker:
    #!{{bash}}
    set -euo pipefail
    containers=$(docker ps -a --filter "name=memos-v*" --format={\{.ID\}} | xargs -er)
    if [ ! -z "$containers" ]; then
        for ct in $containers; do
            echo -e "${MAGENTA}Cleaning ${ct}.${RESET}"
            docker stop $ct >/dev/null 2>&1 || true
            docker rm $ct >/dev/null 2>&1 || true
        done
    fi
    images=$(docker images --filter "reference=memos-v*" --format={\{.ID\}} | xargs -er)
    if [ ! -z "$images" ]; then
        for img in $images; do
            echo -e "${MAGENTA}Cleaning ${img}.${RESET}"
            docker image rm $img >/dev/null 2>&1 || true
        done
    fi
    echo -e "${GREEN}Cleaning complete.${RESET}"

# Clean-up build artifacts, dangling Docker images, volumes and go cache.
[confirm('This will clean-up build artifacts, Go cache, dangling Docker images and Docker build cache. Are you sure?')]
clean:
    #!{{bash}}
    set +e
    artifacts=(
        "./.task"
        "./build"
        "./memos/web/node_modules"
        "./memos/web/dist"
        "./memos/server/frontend/dist"
        "./memos/server/frontend/dist.bak"
    )
    for artifact in "${artifacts[@]}"; do
        if [ -d "$artifact" ]; then
            rm -rf "$artifact"
        fi
    done
    docker builder prune -f
    docker buildx prune -f
    docker image prune -f
    echo -e "${MAGENTA}Cleaning Go cache…${RESET} This may take a while."
    go clean -cache -modcache

# Reset main branch to origin/main. Excludes ALL untracked files and changes.
[confirm('This will exclude ANY changes and untracked files on the working tree, resetting the local repo to origin/main. Are you sure?')]
git-reset:
    #!{{bash}}
    git fetch origin
    git checkout main
    git reset --hard origin/main
    git clean -fdx
    git checkout -- .
    git checkout -

# Update the `memos` subtree, tag and push to GitHub, triggering the `release` workflow.
publish TAG:
    #!{{bash}}
    set -euo pipefail
    TAG="v{{trim_start_matches(TAG, 'v')}}"
    just git-subtree-pull "tags/$TAG"
    just git-retag "{{TAG}}"
    git push origin main

# Update README.md captures. Requires `https://github.com/sindresorhus/capture-website-cli` and a running Memos instance.
update-captures PORT='5230' TOKEN='':
    #!{{bash}}
    COOKIE=''
    if ! [ -z "{{TOKEN}}" ]; then
        COOKIE='--cookie="memos.access-token={{TOKEN}}"'
    fi
    capture-website --overwrite --type=webp --output=assets/capture_dark.webp --dark-mode $COOKIE http://localhost:{{PORT}}/ &
    capture-website --overwrite --type=webp --output=assets/capture_light.webp $COOKIE http://localhost:{{PORT}}/
