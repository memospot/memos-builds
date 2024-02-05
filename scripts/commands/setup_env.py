from datetime import datetime, timezone
from os import chdir, environ

from utils import git, github, memos
from utils.colors import BLUE, BOLD, CYAN, DARK_YELLOW, GREEN, MAGENTA, RESET


def setup_env(*, nightly: bool = False) -> None:
    """
    Set up the build environment.

    - Determine the build version.
    - Set environment variables for the build.
    """
    print(f"{GREEN}>> Setting up build environment <<", RESET)
    now = datetime.now(tz=timezone.utc)

    repo_root = git.get_repo_root()
    chdir(repo_root)
    print(f"Repository root is {BOLD}{repo_root}", RESET)

    file_version = memos.get_version()
    git_tag = git.get_current_tag()
    git_previous_tag = git.get_previous_tag()
    ref_version = environ.get("GITHUB_REF_NAME", "NOT_SET").replace("release/", "")

    # Falls back from file_version -> git_tag -> ref_version -> date-based
    version = "v" + now.strftime("%Y.%m.%d") + ".0"
    for v in (file_version, git_tag, ref_version):
        if v and memos.validate_semver(v):
            version = v
            break

    print(f"{MAGENTA}Discovered versions:", RESET)
    print(f"  version.go: {CYAN}{file_version}", RESET)
    print(f"  Git tag: {CYAN}{git_tag}", RESET)
    print(f"  GitHub Ref: {CYAN}{ref_version}", RESET)
    print(f"Selected {GREEN}{version}{RESET} as {MAGENTA}BUILD_VERSION", RESET)

    date_string = now.strftime("%Y%m%d")
    github.add_to_env("CACHE_KEY", date_string)
    github.add_to_env("MEMOS_VERSION", file_version or version)
    github.add_to_env("GIT_TAG", git_tag)
    github.add_to_env("REF_VERSION", ref_version)

    if nightly:
        # Increment the last number in the version string and append "-dev".
        nightly_version = (
            ".".join(version.split(".")[:-1] + [str(int(version.split(".")[-1]) + 1)]) + "-dev"
        )
        print(
            f"{BLUE}Nightly{RESET} is set to True:",
            f"{MAGENTA}BUILD_VERSION{RESET} is now {DARK_YELLOW}{nightly_version}",
            RESET,
        )
        github.add_to_env("BUILD_VERSION", nightly_version)

        # If not set, goreleaser will infer tags from the git history.
        # This is not desired when running a workflow manually, as "memos-builds"
        # tags may not match "memos" repo tags.
        github.add_to_env("GORELEASER_PREVIOUS_TAG", version + "-dev")
        github.add_to_env("GORELEASER_CURRENT_TAG", nightly_version)
    else:
        github.add_to_env("BUILD_VERSION", version)
        github.add_to_env("GORELEASER_PREVIOUS_TAG", git_previous_tag)
        github.add_to_env("GORELEASER_CURRENT_TAG", version)

    print(f"{GREEN}>> Build environment setup complete <<\n", RESET)
