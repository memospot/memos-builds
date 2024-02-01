import argparse
import os
from datetime import datetime, timezone
from sys import exit

from utils import git, github, memos
from utils.console import Color


def setup_git() -> None:
    git_setup = git.setup()
    for line in (
        f"{Color.blue}Git version: {Color.green}{git_setup.version}",
        f"{Color.blue}Git user email: {Color.green}{git_setup.user_email}",
        f"{Color.blue}Git user name: {Color.green}{git_setup.user_name}",
    ):
        print(line, Color.reset)


def setup_env(*, nightly: bool = False) -> None:
    repo_root = git.get_repo_root()
    os.chdir(repo_root)
    print(f"Repository root is {Color.bold}{repo_root}", Color.reset)

    file_version = memos.get_version()
    print(
        f"{Color.blue}Version from `version.go`: {Color.green}{file_version}",
        Color.reset,
    )

    git_tag = git.get_current_tag()
    print(f"{Color.blue}Git tag: {Color.green}{git_tag}", Color.reset)

    ref_version = os.getenv("GITHUB_REF_NAME", "").replace("release/", "")
    print(f"{Color.blue}GitHub Ref version: {Color.green}{ref_version}", Color.reset)

    if file_version:
        version = file_version
    elif git_tag:
        version = git_tag
    elif ref_version:
        version = ref_version
    else:
        version = "v" + datetime.now(timezone.utc).strftime("%Y.%m.%d") + ".0"

    print(f"{Color.blue}Version: {Color.green}{version}", Color.reset)

    date_string = datetime.now(tz=timezone.utc).strftime("%Y%m%d")
    github.add_to_env("DATE_STRING", date_string)
    github.add_to_env("GIT_TAG", git_tag)
    github.add_to_env("VERSION", version)
    github.add_to_env("MEMOS_VERSION", file_version)
    github.add_to_env("REF_VERSION", ref_version)

    if nightly:
        # Increment the last number in the version string and append "-dev".
        nightly_version = (
            ".".join(version.split(".")[:-1] + [str(int(version.split(".")[-1]) + 1)]) + "-dev"
        )
        print(f"{Color.blue}Nightly version: {Color.green}{nightly_version}", Color.reset)
        github.add_to_env("NIGHTLY_VERSION", nightly_version)
        github.add_to_env("GORELEASER_CURRENT_TAG", nightly_version)
    else:
        github.add_to_env("GORELEASER_CURRENT_TAG", version)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--setup-git", action="store_true")
    parser.add_argument("--setup-env", action="store_true")
    parser.add_argument("--nightly", action="store_true")
    args = parser.parse_args()

    if args.setup_git:
        setup_git()
    elif args.setup_env:
        setup_env(nightly=args.nightly)
    else:
        setup_git()
        setup_env(nightly=args.nightly)

    exit(0)
