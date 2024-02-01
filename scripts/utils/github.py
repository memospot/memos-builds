import os


def add_to_env(key: str, value: str) -> None:
    """
    Add a key-value pair to the GitHub environment (`$GITHUB_ENV`).

    Raises
    ------
    OSError
        If the file cannot be written to.
    """
    github_env = os.getenv("GITHUB_ENV", ".GITHUB_ENV")
    with open(github_env, "a") as file:
        file.write(f"{key.strip()}={value.strip()}\n")
