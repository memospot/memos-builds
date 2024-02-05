from os import environ


def add_to_env(key: str, value: str) -> None:
    """
    Add a key-value pair to the GitHub environment (`$GITHUB_ENV`).

    Falls back to `.GITHUB_ENV` as filename if the environment
    variable is not set.

    Raises
    ------
    OSError
        If the file cannot be written to.
    """
    github_env = environ.get("GITHUB_ENV", ".GITHUB_ENV")
    with open(github_env, "a") as file:
        file.write(f"{key.strip()}={value.strip()}\n")
