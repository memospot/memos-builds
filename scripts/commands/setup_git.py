from utils import git
from utils.colors import BLUE, GREEN, RESET


def setup_git() -> None:
    """
    Set up git for the build environment.
    """
    print(f"{GREEN}>> Setting up git <<", RESET)
    git_setup = git.setup()
    for line in (
        f"Git version: {GREEN}{git_setup.version}",
        f"Git user email: {GREEN}{git_setup.user_email}",
        f"Git user name: {GREEN}{git_setup.user_name}",
    ):
        print(f"{BLUE}{line}{RESET}")
    print(f"{GREEN}>> Git setup complete <<\n", RESET)
