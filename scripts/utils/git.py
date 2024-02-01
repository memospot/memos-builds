# Handles the git environment setup and provides some helper functions to get
# information about the git repository.

from contextlib import suppress
from subprocess import CalledProcessError, check_output, run
from typing import NamedTuple


class GitSetup(NamedTuple):
    version: str
    user_email: str
    user_name: str


def setup() -> GitSetup:
    """
    Setups git environment.

    Sets the git user email and name to github-actions[bot] if not set.

    Returns
    -------
    GitSetup: NamedTuple
        version, user_email, user_name

    Raises
    ------
    RuntimeError
        If the git version cannot be determined.
    CalledProcessError
        If subprocess call fails.
    """
    git_version = get_version()
    if not git_version:
        msg = "Unable to determine git version"
        raise RuntimeError(msg)

    git_user_email, git_user_name = "", ""
    with suppress(CalledProcessError):
        git_user_email = (
            check_output(
                (
                    "git",
                    "config",
                    "--global",
                    "user.email",
                )
            )
            .decode()
            .strip()
        )
        git_user_name = (
            check_output(
                (
                    "git",
                    "config",
                    "--global",
                    "user.name",
                )
            )
            .decode()
            .strip()
        )

    if not git_user_email:
        git_user_email = "github-actions[bot]@users.noreply.github.com"
        run(
            (
                "git",
                "config",
                "--global",
                "user.email",
                git_user_email,
            ),
            check=True,
        )

    if not git_user_name:
        git_user_name = "github-actions[bot]"
        run(
            (
                "git",
                "config",
                "--global",
                "user.name",
                git_user_name,
            ),
            check=True,
        )

    return GitSetup(git_version, git_user_email, git_user_name)


def get_current_tag() -> str:
    """
    Get the current tag.

    Returns
    -------
    str
        The current tag.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    return (
        check_output(
            (
                "git",
                "describe",
                "--tags",
                "--abbrev=0",
            )
        )
        .decode()
        .strip()
    )


def get_version() -> str:
    """
    Get the git version.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    output = (
        check_output(
            (
                "git",
                "--version",
            )
        )
        .decode()
        .strip()
    )
    return output.replace("git version ", "")


def get_repo_name() -> str:
    """
    Get the name of the git repository.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    return (
        check_output(
            (
                "git",
                "config",
                "--get",
                "remote.origin.url",
            )
        )
        .decode()
        .strip()
        .split("/")[-1]
        .replace(".git", "")
    )


def get_repo_root() -> str:
    """
    Get the root of the git repository.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    return (
        check_output(
            (
                "git",
                "rev-parse",
                "--show-toplevel",
            )
        )
        .decode()
        .strip()
    )
