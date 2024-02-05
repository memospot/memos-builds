import re
from glob import glob
from os import chdir, rename
from pathlib import Path

from utils import git
from utils.colors import BLUE, BOLD, DARK_YELLOW, GREEN, RED, RESET

GORELEASER_GLOB = "./build/backend/memos*"
GORELEASER_PREFIX = "memos_linux_"

# Mapping from goreleaser/GOARCH to Docker architecture tags.
GO_TO_DOCKER = {
    r"amd64_v1": r"amd64",
    r"amd64_v([2-4])": r"amd64v\1",
    r"arm_([5-7])": r"armv\1",
}


# Validating patterns for Docker platform names.
DOCKER_PATTERNS = (
    r"386",
    r"amd64",
    r"amd64v[2-4]",
    r"armv[5-7]",
    r"arm64",
    r"ppc64le",
    r"riscv64",
    r"s390x",
)


def rename_to_docker() -> None:
    """
    Rename goreleaser builds to the format expected by the Dockerfile.
    """
    repo_root = git.get_repo_root()
    chdir(repo_root)
    print(f"Repository root is {BOLD}{repo_root}", RESET)

    print(
        f"{GREEN}>> Renaming goreleaser builds to the format expected by the Dockerfile <<",
        RESET,
    )

    renames = 0
    for item in glob(GORELEASER_GLOB):
        folder = Path(item)
        if not folder.is_dir():
            continue

        folder_name = folder.name

        patterns = "|".join(rf"^{GORELEASER_PREFIX}{p}$" for p in DOCKER_PATTERNS)
        if re.match(patterns, folder_name) is not None:
            continue

        new_name = folder_name
        for go, docker in GO_TO_DOCKER.items():
            new_name = re.sub(
                rf"^{GORELEASER_PREFIX}{go}$", f"{GORELEASER_PREFIX}{docker}", new_name
            )

        if new_name == folder_name:
            continue

        new_folder = folder.parent.joinpath(new_name)
        rename(folder, new_folder)
        if new_folder.exists():
            print(
                f"=> {GREEN}Renamed",
                f"{BLUE}{folder_name}{RESET}",
                f"to {GREEN}{new_name}{RESET}",
            )
            renames += 1
            continue

        print(
            f"=> {RED}Failed to rename",
            f"{DARK_YELLOW}{folder_name}{RESET}",
            f"to {RED}{new_name}{RESET}",
        )

    if renames == 0:
        print(f"{DARK_YELLOW}>> No folders were renamed. <<{RESET}")
        return

    print(f"{GREEN}>> Renamed {renames} folders. <<\n", RESET)
