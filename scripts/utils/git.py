# Handles the git environment setup and provides some helper functions to get
# information about the git repository.

from contextlib import suppress
from pathlib import Path
from subprocess import CalledProcessError, check_output, run
from typing import NamedTuple


class GitSetup(NamedTuple):
    version: str
    user_email: str
    user_name: str


def setup() -> GitSetup:
    """
    Setups git environment.

    Set the git user email and name to github-actions[bot] if not set.

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
                ("git", "config", "--global", "user.email"),
            )
            .decode()
            .strip()
        )
        git_user_name = (
            check_output(
                ("git", "config", "--global", "user.name"),
            )
            .decode()
            .strip()
        )

    if not git_user_email:
        git_user_email = "github-actions[bot]@users.noreply.github.com"
        run(
            ("git", "config", "--global", "user.email", git_user_email),
            check=True,
        )

    if not git_user_name:
        git_user_name = "github-actions[bot]"
        run(
            ("git", "config", "--global", "user.name", git_user_name),
            check=True,
        )

    return GitSetup(git_version, git_user_email, git_user_name)


def commit_any_changes(message: str = "chore:ci: ensure clean git state") -> None:
    """
    Commit any changes to the git repository.

    Used to ensure a clean git state.
    """
    run(
        ("git", "add", "-A"),
        check=False,
    )
    run(
        ("git", "commit", "-m", message),
        check=False,
    )


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
            ("git", "describe", "--tags", "--abbrev=0"),
        )
        .decode()
        .strip()
    )


def get_previous_tag(*, dev: bool = False) -> str:
    """
    Get the previous tag. Excludes tags ending with -dev.

    Returns
    -------
    str
        The previous tag.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    command = ["git", "describe", "--tags", "--abbrev=0"]
    command.append("--match=*-dev" if dev else "--exclude=*-dev")

    return check_output(command).decode().strip()


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
            ("git", "--version"),
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
        check_output(("git", "config", "--get", "remote.origin.url"))
        .decode()
        .strip()
        .split("/")[-1]
        .replace(".git", "")
    )


def get_repo_root() -> str:
    """
    Get the root of the git repository.

    Determined by `git rev-parse --show-toplevel`.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    FileNotFoundError
        If the repository root does not exist.
    NotADirectoryError
        If the repository root is not a directory.
    """
    root = (
        check_output(
            ("git", "rev-parse", "--show-toplevel"),
        )
        .decode()
        .strip()
    )

    resolved_root = Path(root).resolve()
    if not resolved_root.exists():
        msg = f"Repository root {resolved_root} does not exist."
        raise FileNotFoundError(msg)
    if not resolved_root.is_dir():
        msg = f"Repository root {resolved_root} is not a directory."
        raise NotADirectoryError(msg)

    return str(resolved_root)


def retag(tag: str, *, push: bool = False) -> None:
    """
    Remove a git tag and push it again.

    Used in case the CI pipeline fails and some correction is needed.

    Goreleaser demands a tag commit to be verifiable,
    so the tag must be made right before the build is triggered.

    Parameters
    ----------
    tag : str
        The tag to retag.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    clean_tag = "v" + tag.lstrip("v")
    if push:
        run(
            ("git", "push", "origin", f":refs/tags/{clean_tag}"),
            check=False,
        )
    run(
        ("git", "tag", "-d", clean_tag),
        check=False,
    )
    run(
        ("git", "tag", "-a", clean_tag, "-m", f"Tag {clean_tag}"),
        check=False,
    )
    if push:
        run(
            ("git", "push", "origin", clean_tag),
            check=False,
        )


def subtree_pull(prefix: str, repo: str, branch: str) -> None:
    """
    Pull the subtree from the specified repository and branch.

    Parameters
    ----------
    prefix : str
        The local prefix/directory to pull to.
    repo : str
        The repository to pull from.
    branch : str
        The branch to pull from.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    err = None
    if not prefix:
        err = "prefix cannot be empty"
    elif not repo:
        err = "repo cannot be empty"
    elif not branch:
        err = "branch cannot be empty"
    if err is not None:
        raise ValueError(err)

    run(
        (
            "git",
            "subtree",
            "pull",
            f"--prefix={prefix}",
            repo,
            branch,
            "--squash",
            rf'--message="chore:ci: pull {prefix} from {repo}"',
        ),
        check=True,
    )


def clean(path: str) -> None:
    """
    Clean the git repository.

    Recursively remove files that are not under version control. Bypass gitignore rules.

    Warning: This will remove all untracked files and directories and will commit any changes.

    Parameters
    ----------
    prefix : str
        The prefix to clean.

    Raises
    ------
    CalledProcessError
        If subprocess call fails.
    """
    run(
        ("git", "clean", "-ffx", path),
        check=False,
    )
    commit_any_changes(f"chore:ci: clean {path}")
