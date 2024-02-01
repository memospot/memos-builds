import glob
import os
import re
from pathlib import Path

from utils import git
from utils.console import Color

repo_root = git.get_repo_root()
os.chdir(repo_root)
print(f"Repository root is {Color.bold}{repo_root}", Color.reset)

print("Renaming builds to Docker format...")

# Prefix for each build folder.
PREFIX = "memos_linux_"

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

# Mapping from goreleaser/GOARCH to Docker architecture tags.
GO_TO_DOCKER = {
    r"amd64_v1": r"amd64",
    r"amd64_v([2-4])": r"amd64v\1",
    r"arm_([5-7])": r"armv\1",
}

for item in glob.glob("./build/backend/memos*"):
    if not os.path.isdir(item):
        continue

    folder = Path(item)
    folder_name = folder.name

    patterns = "|".join(rf"^{PREFIX}{p}$" for p in DOCKER_PATTERNS)
    if re.match(patterns, folder_name) is not None:
        print(
            f"{Color.green}`{folder_name}` is already in Docker format.",
            Color.reset,
        )
        continue

    new_name = folder_name
    for go, docker in GO_TO_DOCKER.items():
        new_name = re.sub(rf"^{PREFIX}{go}$", f"{PREFIX}{docker}", new_name)

    if new_name == folder_name:
        continue

    print(
        f"=> Renaming {Color.blue}{folder_name}{Color.reset} to {Color.magenta}{new_name}",
        Color.reset,
    )
    os.rename(folder, Path(folder.parent, new_name))
