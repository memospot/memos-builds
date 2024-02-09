"""
Ported from Go's semver package.

All tests are passing, though it's still possible to clean-up and improve the code
using Python's own features.

Package semver implements comparison of semantic version strings.
In this package, semantic version strings must begin with a leading "v", as in "v1.0.0".

The general form of a semantic version string accepted by this package is

        vMAJOR[.MINOR[.PATCH[-PRERELEASE][+BUILD]]]

    where square brackets indicate optional parts of the syntax;
    MAJOR, MINOR, and PATCH are decimal integers without extra leading zeros;
    PRERELEASE and BUILD are each a series of non-empty dot-separated identifiers
    using only alphanumeric characters and hyphens; and
    all-numeric PRERELEASE identifiers must not have leading zeros.

This package follows Semantic Versioning 2.0.0 (see semver.org)
with two exceptions. First, it requires the "v" prefix. Second, it recognizes
vMAJOR and vMAJOR.MINOR (with no prerelease or build suffixes)
as shorthands for vMAJOR.0.0 and vMAJOR.MINOR.0.

Notes
-----
- Removed deprecated `max` function.
"""

from dataclasses import dataclass


@dataclass
class Version:
    """
    Version represents the parsed form of a semantic version string.
    """

    major: str
    minor: str
    patch: str
    short: str
    prerelease: str
    build: str
    is_valid: bool = False

    def __str__(self) -> str:
        """
        Return the semantic version string in its parsed form.
        """
        return canonical(
            f"v{self.major}.{self.minor}.{self.patch}{self.short}{self.prerelease}{self.build}"
        )

    def __repr__(self) -> str:
        """
        Return a object representation of the parsed semantic version.
        """
        rep = ""
        for field in self.__dataclass_fields__:
            rep += f"{field}={getattr(self, field)}, "
        return f"Parsed({rep[:-2]})"


def is_valid(v: str) -> bool:
    """
    Report whether v is a valid semantic version string.
    """
    return parse(v).is_valid


def canonical(v: str) -> str:
    """
    Return the canonical formatting of the semantic version v.

    It fills in any missing .MINOR or .PATCH and discards build metadata.
    Two semantic versions compare equal only if their canonical formattings
    are identical strings.
    The canonical invalid semantic version is the empty string.
    """
    p = parse(v)
    if not p.is_valid:
        return ""
    if p.build:
        return v[: len(v) - len(p.build)]
    if p.short:
        return v + p.short
    return v


def major(v: str) -> str:
    """
    Return the major version prefix of the semantic version v.

    For example, major("v2.1.0") == "v2".
    If v is an invalid semantic version string, Major returns the empty string.

    Arguments
    ---------
    v: str
        Semantic version string.
    """
    pv = parse(v)
    if not pv.is_valid:
        return ""
    return v[: 1 + len(pv.major)]


def major_minor(v: str) -> str:
    """
    Return the major.minor version prefix of the semantic version v.

    For example, major_minor("v2.1.0") == "v2.1".
    If v is an invalid semantic version string, major_minor returns the empty string.
    """
    pv = parse(v)
    if not pv.is_valid:
        return ""
    i = 1 + len(pv.major)
    if (
        (j := i + 1 + len(pv.minor))
        and j <= len(v)
        and v[i] == "."
        and v[i + 1 : j] == pv.minor
    ):
        return v[:j]
    return v[:i] + "." + pv.minor


def prerelease(v: str) -> str:
    """
    Return the prerelease suffix of the semantic version v.

    For example, Prerelease("v2.1.0-pre+meta") == "-pre".
    If v is an invalid semantic version string, Prerelease returns the empty string.
    """
    pv = parse(v)
    return pv.prerelease if pv.is_valid else ""


def build(v: str) -> str:
    """
    Return the build suffix of the semantic version v.

    For example, build("v2.1.0+meta") == "+meta".
    If v is an invalid semantic version string, build returns the empty string.
    """
    pv = parse(v)
    return pv.build if pv.is_valid else ""


def compare(v: str, w: str) -> int:
    """
    Return an integer comparing two versions according to semantic version precedence.

    The result will be 0 if v == w, -1 if v < w, or +1 if v > w.

    An invalid semantic version string is considered less than a valid one.
    All invalid semantic version strings compare equal to each other.
    """
    pv = parse(v)
    pw = parse(w)
    if not pv.is_valid and not pw.is_valid:
        return 0
    if not pv.is_valid:
        return -1
    if not pw.is_valid:
        return +1

    if (c := compare_int(pv.major, pw.major)) and c != 0:
        return c
    if (c := compare_int(pv.minor, pw.minor)) and c != 0:
        return c
    if (c := compare_int(pv.patch, pw.patch)) and c != 0:
        return c
    return compare_prerelease(pv.prerelease, pw.prerelease)


class ByVersion(list[str]):
    """
    Implement [sort.Interface] for sorting semantic version strings.
    """

    def __init__(self, versions: list[str]) -> None:
        self.versions = versions

    def __len__(self) -> int:
        """
        Return the number of versions.
        """
        return len(self.versions)

    def swap(self, i: int, j: int) -> None:
        """
        Swap the versions at indices i and j.
        """
        self.versions[i], self.versions[j] = self.versions[j], self.versions[i]

    def less(self, i: int, j: int) -> bool:
        """
        Return whether the version at index i is less than the version at index j.
        """
        cmp = compare(self.versions[i], self.versions[j])
        if cmp != 0:
            return cmp < 0
        return self.versions[i] < self.versions[j]


def sort(lst: list[str]) -> None:
    """
    Sort a list of semantic version strings in place.
    """
    ByVersion(lst).sort()


def parse(v: str) -> Version:
    p = Version("", "", "", "", "", "")
    if not v or v[0] != "v":
        return p
    p.major, v, ok = parse_int(v[1:])
    if not ok:
        return p
    if not v:
        p.minor = "0"
        p.patch = "0"
        p.short = ".0.0"
        p.is_valid = True
        return p
    if v[0] != ".":
        return p
    p.minor, v, ok = parse_int(v[1:])
    if not ok:
        return p
    if not v:
        p.patch = "0"
        p.short = ".0"
        p.is_valid = True
        return p
    if v[0] != ".":
        return p
    p.patch, v, ok = parse_int(v[1:])
    if not ok:
        return p
    if len(v) > 0 and v[0] == "-":
        p.prerelease, v, ok = parse_prerelease(v)
        if not ok:
            return p
    if len(v) > 0 and v[0] == "+":
        p.build, v, ok = parse_build(v)
        if not ok:
            return p
    if v:
        return p
    p.is_valid = True
    return p


def parse_int(v: str) -> tuple[str, str, bool]:
    if not v or not v[0].isdigit():
        return "", "", False
    i = 1
    while i < len(v) and v[i].isdigit():
        i += 1
    if v[0] == "0" and i != 1:
        return "", "", False
    return v[:i], v[i:], True


def parse_prerelease(v: str) -> tuple[str, str, bool]:
    """
    Parse a pre-release version.

    A pre-release version MAY be denoted by appending a hyphen and a series of dot separated
    identifiers immediately following the patch version. Identifiers MUST comprise only ASCII
    alphanumerics and hyphen [0-9A-Za-z-]. Identifiers MUST NOT be empty. Numeric identifiers
    MUST NOT include leading zeroes.
    """
    if not v or v[0] != "-":
        return "", "", False
    i = 1
    start = 1
    while i < len(v) and v[i] != "+":
        if not is_ident_char(v[i]) and v[i] != ".":
            return "", "", False
        if v[i] == ".":
            if start == i or is_bad_num(v[start:i]):
                return "", "", False
            start = i + 1
        i += 1
    if start == i or is_bad_num(v[start:i]):
        return "", "", False
    return v[:i], v[i:], True


def parse_build(v: str) -> tuple[str, str, bool]:
    if not v or v[0] != "+":
        return "", "", False
    i = 1
    start = 1
    while i < len(v):
        if not is_ident_char(v[i]) and v[i] != ".":
            return "", "", False
        if v[i] == ".":
            if start == i:
                return "", "", False
            start = i + 1
        i += 1
    if start == i:
        return "", "", False
    return v[:i], v[i:], True


def is_ident_char(c: str) -> bool:
    return c.isalnum() or c == "-"


def is_bad_num(v: str) -> bool:
    i = 0
    while i < len(v) and "0" <= v[i] <= "9":
        i += 1
    return i == len(v) and i > 1 and v[0] == "0"


def is_num(v: str) -> bool:
    i = 0
    while i < len(v) and "0" <= v[i] <= "9":
        i += 1
    return i == len(v)


def compare_int(x: str, y: str) -> int:
    if x == y:
        return 0
    if len(x) < len(y):
        return -1
    if len(x) > len(y):
        return 1
    return -1 if x < y else 1


def compare_prerelease(x: str, y: str) -> int:
    """
    Compare pre-release versions.

        When major, minor, and patch are equal, a pre-release version has lower precedence than
    a normal version. Example: 1.0.0-alpha < 1.0.0.

        Precedence for two pre-release versions with the same major, minor, and patch version
    MUST be determined by comparing each dot separated identifier from left to right until
    a difference is found as follows: identifiers consisting of only digits are compared
    numerically and identifiers with letters or hyphens are compared lexically in ASCII sort
    order.
        Numeric identifiers always have lower precedence than non-numeric identifiers.
        A larger set of pre-release fields has a higher precedence than a smaller set, if all
    of the preceding identifiers are equal.
        Example: 1.0.0-alpha < 1.0.0-alpha.1 < 1.0.0-alpha.beta < 1.0.0-beta < 1.0.0-beta.2 <
        1.0.0-beta.11 < 1.0.0-rc.1 < 1.0.0.
    """
    if x == y:
        return 0
    if x == "":
        return 1
    if y == "":
        return -1
    while x != "" and y != "":
        x = x[1:]  # skip - or .
        y = y[1:]  # skip - or .
        dx, x = next_ident(x)
        dy, y = next_ident(y)
        if dx != dy:
            ix = is_num(dx)
            iy = is_num(dy)
            if ix != iy:
                return -1 if ix else 1
            if ix:
                if len(dx) < len(dy):
                    return -1
                if len(dx) > len(dy):
                    return 1
            return -1 if dx < dy else 1
    return -1 if x == "" else 1


def next_ident(x: str) -> tuple[str, str]:
    """
    Return the next identifier in x and the remainder of x after the identifier.
    """
    ident, _, remainder = x.partition(".")
    return ident, remainder
