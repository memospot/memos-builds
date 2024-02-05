"""
Minimal console colors.

- Only return color codes if the terminal appears to support ANSI colors and if
colors aren't explicitly disabled with the `NO_COLOR` environment variable.
"""

from os import environ, pathsep


def can_use_colors() -> bool:
    # https://no-color.org/
    no_color = environ.get("NO_COLOR", "")
    if len(no_color) > 0 and no_color.lower() not in (
        "0",
        "disabled",
        "false",
        "n",
        "no",
        "off",
    ):
        return False

    if environ.get("TERM", "").lower() == "xterm-256color":
        return True

    match environ.get("COLORTERM", "").lower():
        case "truecolor":
            return True
        case "256":
            return True
        case _:
            pass

    # powershell
    if len(environ.get("PSMODULEPATH", "").split(pathsep)) >= 1:
        return True

    # cmd
    return environ.get("PROMPT", "") != "$P$G"


use_colors = can_use_colors()


def get_color(code: int) -> str:
    return f"\033[{code}m" if use_colors else ""


BOLD = get_color(1)
RESET = get_color(0)
UNDERLINE = get_color(4)
BLUE = get_color(94)
CYAN = get_color(96)
END = RESET
GREEN = get_color(92)
MAGENTA = get_color(95)
RED = get_color(91)
YELLOW = get_color(93)
DARK_BLUE = get_color(34)
DARK_CYAN = get_color(36)
DARK_GREEN = get_color(32)
DARK_MAGENTA = get_color(35)
DARK_RED = get_color(31)
DARK_YELLOW = get_color(33)


if __name__ == "__main__":
    for var in dir():
        if not var.startswith("__") and var.isupper():
            value = eval(var)  # noqa: S307, PGH001
            print(f"{value}This is {var}" + RESET)
