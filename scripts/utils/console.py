"""
Console utilities.
"""


class Color:
    """
    Minimal console colors.
    """

    blue = "\033[94m"
    bold = "\033[1m"
    end = "\033[0m"
    green = "\033[92m"
    magenta = "\033[95m"
    red = "\033[91m"
    reset = "\033[0m"
    underline = "\033[4m"
    yellow = "\033[93m"


if __name__ == "__main__":
    # test colors
    for color in dir(Color):
        if not color.startswith("__"):
            print(getattr(Color, color) + f"This is {color}" + Color.reset)
