from os import chdir
from pathlib import Path
from shutil import move, rmtree

from utils import git
from utils.colors import CYAN, GREEN, MAGENTA, RED, RESET
from utils.exec import Exec

FRONTEND_BUILD_COMMANDS = (
    ("pnpm", "install", "--frozen-lockfile"),
    ("pnpm", "type-gen"),
    ("pnpm", "build"),
)


def build_frontend(source: str, dist: str, final: str) -> None:
    """
    Build the front-end and move the build files to the final location.

    Arguments
    ---------
    source: str
        Front-end root folder. Usually 'web'.
    dist: str
        Usually 'web/dist'.
    final: str
        Where to move the final build folder.
    """
    if not all((source, dist, final)):
        msg = "Missing required arguments!"
        raise ValueError(msg)

    source_path = Path(source).resolve()
    dist_path = Path(dist).resolve()
    final_path = Path(final).resolve()

    print(f"{MAGENTA}Source: {CYAN}{source_path}{RESET}")
    print(f"{MAGENTA}Dist: {CYAN}{dist_path}{RESET}")
    print(f"{MAGENTA}Final: {CYAN}{final_path}{RESET}")

    repo_root = git.get_repo_root()
    print(f"{MAGENTA}Repository root: {CYAN}{repo_root}", RESET)
    chdir(repo_root)

    # A few sanity checks
    for folder in (source_path, dist_path, final_path):
        if not folder.is_relative_to(repo_root):
            msg = f"Folder `{RED}{folder}{RESET}` is not within the repository root `{repo_root}`."
            raise RuntimeError(msg)
    if not source_path.exists():
        msg = f"Source directory {source_path} does not exist."
        raise FileNotFoundError(msg)

    print(f"{GREEN}>> Begin Front-end build <<", RESET)

    chdir(source_path)
    if not Exec(FRONTEND_BUILD_COMMANDS, stop_on_error=True).success:
        chdir(repo_root)
        msg = f"{RED}Front-end build failed!{RESET}"
        raise RuntimeError(msg)

    print(f"{MAGENTA}Emptying `{final_path}`...{RESET}")
    rmtree(final_path, ignore_errors=True)
    final_path.mkdir(parents=True, exist_ok=True)

    print(f"{MAGENTA}Moving build files...{RESET}")
    move(dist_path, final_path)

    if final_path.joinpath("dist", "index.html").exists():
        print(f"{GREEN}>> Front-end build complete! <<{RESET}")
    else:
        print(f"{RED}Front-end build failed!{RESET}")

    chdir(repo_root)
