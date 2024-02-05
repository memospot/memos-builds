"""
Memos utility functions.
"""

import re
from contextlib import suppress
from pathlib import Path

# Where to find the version file. Relative to the repository root.
VERSION_FILE = "memos/server/version/version.go"
VERSION_REGEX = re.compile(r'^var\s+Version\s*=\s*"v?([0-9.]+)"$', re.MULTILINE)
DEVVERSION_REGEX = re.compile(r'^var\s+DevVersion\s*=\s*"v?([0-9.]+)"$', re.MULTILINE)
SEMVER_REGEX = re.compile(r"^v?\d+\.\d+\.\d+(-\w+)?$")


def _get_version_from_file(version_file: str | Path, pattern: re.Pattern[str]) -> str:
    with suppress(FileNotFoundError), open(version_file) as file:
        content = file.read()
        match = pattern.search(content)
        version = match.group(1) if match else ""
        if validate_semver(version):
            return version if version[0] == "v" else f"v{version}"
    return ""


def get_version(file: str | Path = "") -> str:
    """
    Get current Memos version from version.go file.

    May return an empty string if the file is not found or the version is not valid.

    Returns
    -------
    str
        The version string.
    """
    return _get_version_from_file(file or VERSION_FILE, VERSION_REGEX)


def get_dev_version(file: str | Path = "") -> str:
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
    return _get_version_from_file(file or VERSION_FILE, DEVVERSION_REGEX)


def validate_semver(version: str) -> bool:
    """
    Validate a semantic version string.
    """
    match = SEMVER_REGEX.match(version)
    return match is not None and match.string == version
