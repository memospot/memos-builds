"""
Memos utility functions.
"""


import re
from pathlib import Path

# Where to find the version file. Relative to the repository root.
VERSION_FILE = "memos-upstream/server/version/version.go"
VERSION_REGEX = r'^var\s+Version\s+=\s+"v?([0-9.]+)"$'
DEVVERSION_REGEX = r'^var\s+DevVersion\s+=\s+"v?([0-9.]+)"$'


def _get_version_from_file(version_file: str | Path, pattern: str) -> str:
    try:
        with open(version_file) as file:
            content = file.read()
            match = re.search(pattern, content, re.MULTILINE)
            version = match.group(1) if match else ""
            if validate_semver(version):
                return version if version[0] == "v" else f"v{version}"
    except FileNotFoundError:
        return ""  # this allows falling back to git tag if upstream changes

    return ""
    # msg = f"Could not find version in {version_file} using pattern {pattern}"
    # raise RuntimeError(msg)


def get_version() -> str:
    """
    Get current Memos version from version.go file.

    May return an empty string if the file is not found or the version is not valid.

    Returns
    -------
    str
        The version string.
    """
    return _get_version_from_file(VERSION_FILE, VERSION_REGEX)


def get_dev_version() -> str:
    """
    Get current Memos development version from version.go file.

    May return an empty string if the file is not found or the version is not valid.

    Parameters
    ----------
    version: str
        The version string, or an empty string if the file is not found.

    Returns
    -------
    str
        The development version string.
    """
    return _get_version_from_file(VERSION_FILE, DEVVERSION_REGEX)


def validate_semver(version: str) -> bool:
    """
    Validate a semantic version string.
    """
    match = re.match(r"^v?\d+\.\d+\.\d+(-\w+)?$", version)
    return match is not None and match.string == version
