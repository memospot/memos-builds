"""
Run CI actions.

This script is used to run CI actions. It is called by GitHub Actions workflows.

The following actions are supported:
- `pull-subtree`: Pull the subtree from the memos repository.
- `rename-to-docker`: Rename the goreleaser builds to the format expected by the Dockerfile.
- `setup-env`: Set up the build environment.
- `setup-git`: Set up git.

Usage
-----
`python scripts/ci.py <command>`
"""

from argparse import ArgumentParser
from sys import exit, version as python_version

from commands.build_frontend import build_frontend
from commands.rename_to_docker import rename_to_docker
from commands.setup_env import setup_env
from commands.setup_git import setup_git
from utils import git

if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument(
        "command",
        help="Action to perform.",
        choices=(
            "build-frontend",
            "pull-subtree",
            "rename-to-docker",
            "retag",
            "setup-env",
            "setup-git",
        ),
    )
    parser.add_argument("positional", nargs="?", help="Optional value to pass to action.")

    # build-frontend
    parser.add_argument("--source", default="memos/web", help="Front-end root folder.")
    parser.add_argument("--dist", default="memos/web/dist", help="Usually 'web/dist'.")
    parser.add_argument(
        "--final", default="memos/server/router/frontend", help="Where to move the final build."
    )

    # setup-env
    parser.add_argument("--nightly", action="store_true", help="Set up for nightly build.")

    # pull-subtree
    parser.add_argument("--branch", help="Branch to pull the subtree from.")

    # retag
    parser.add_argument("--tag", help="Tag to retag.")
    parser.add_argument("--push", action="store_true", help="Push the tag after retagging.")

    args = parser.parse_args()

    print(f"Python version is {python_version}")

    match args.command:
        case "build-frontend":
            if not all((args.source, args.dist, args.final)):
                parser.print_help()
                exit("Missing required arguments!")

            build_frontend(
                source=args.source,
                dist=args.dist,
                final=args.final,
            )

        case "rename-to-docker":
            rename_to_docker()

        case "retag":
            git.retag(tag=args.tag or args.positional, push=args.push)

        case "pull-subtree":  # ci.py pull-subtree main
            git.setup()  # Ensure git is set up to prevent CI issues.
            git.commit_any_changes()

            prefix = "memos"
            branch = args.branch or args.positional
            if not branch:
                exit("Missing required argument: --branch")

            git.subtree_pull(
                prefix=prefix,
                repo="https://github.com/usememos/memos.git",
                branch=branch,
            )
            git.clean(prefix)

        case "setup-env":
            setup_env(nightly=args.nightly)

        case "setup-git":
            setup_git()

        case _:
            pass

    exit(0)
