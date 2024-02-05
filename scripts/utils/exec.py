"""
Command executor.

Run commands in the system shell, keeping track of their process IDs,
so they get killed on error or user interrupt.

Handles user interrupts on a best-effort basis.
"""

from collections.abc import Sequence
from contextlib import suppress
from dataclasses import dataclass
from os import kill
from pathlib import Path
from signal import SIGINT, SIGTERM
from subprocess import Popen, list2cmdline
from sys import platform, stderr, stdin, stdout

from .colors import BLUE, BOLD, DARK_RED, MAGENTA, RESET


@dataclass
class Cmd:
    """
    A command to be run in the system shell.

    Attributes
    ----------
    cmd: Sequence[Path | str]
        The command to be run.

    cmdline: str
        `cmd`, after being parsed by `subprocess.list2cmdline`. Don't set this manually.

    cwd: Path | str | None
        The working directory for the command.

    env: dict[str, str] | None
        Environment variables for the command.

    ignore_error: bool
        Continue execution of subsequent commands even if the current command fails.

    stream: bool
        Stream command output to the console.
    """

    cmd: Sequence[Path | str]
    cmdline: str = ""
    cwd: Path | str | None = None
    env: dict[str, str] | None = None
    ignore_error: bool = False
    stream: bool = True


class Exec:
    """
    Run commands in the system shell.

    Keeps track of process IDs, so they get killed on error or user interrupt.

    Global settings:
    - `stop_on_error == True`: stops further execution if a command fails.

    - `silent == True`: nothing is printed to the console.

    - `no_kill == True`: processes are not killed on error or user interrupt.

    Command settings:
    - `Cmd.cmd`: the command to be run.

    - `Cmd.cwd`: the working directory for the command.

    - `Cmd.env`: environment variables for the command.

    - `Cmd.ignore_error == True`: execution continues even if the command fails.

    - `Cmd.stream == False`: command output is not streamed to the console.

    Example
    -------
    ```
    Exec(
        (
            Cmd(cmd=("ls", "-l")),
            Cmd(cmd=("echo", "Hello, world!")),
        )
    )

    Exec(
        (
            ("ls", "-l"),
            ("echo", "Hello, world!"),
        )
    )
    ```
    """

    def __init__(
        self,
        commands: Sequence[Cmd] | Sequence[Sequence[str]],
        *,
        stop_on_error: bool = True,
        silent: bool = False,
        no_kill: bool = False,
    ) -> None:
        self.pids: list[int] = []
        self.stop_on_error = stop_on_error
        self.silent = silent
        self.no_kill = no_kill
        self.success = True

        for command in commands:
            cmd = command if isinstance(command, Cmd) else Cmd(cmd=command)
            cmd.cmdline = list2cmdline(cmd.cmd)
            if not self.silent:
                print(f"Running command {BLUE}{cmd.cmdline}{RESET}")
            if not self.run(cmd):
                self.success = False
                if self.stop_on_error:
                    self._cleanup()
                    return
                if not cmd.ignore_error:
                    self._cleanup()
                    return
            self.success = True and self.success

    def _cleanup(self) -> None:
        """
        Kill all started processes.

        Called when a command fails and `stop_on_error == True`
        or when a command's `ignore_error` attribute is False.
        """
        if self.no_kill:
            return

        for pid in self.pids:
            if not self.silent:
                print(f"Killing process {BOLD}{pid}{RESET}.")
            with suppress(ProcessLookupError, OSError):
                for sig in (SIGINT, SIGTERM):
                    kill(pid, sig)
                if platform == "win32":
                    from signal import CTRL_BREAK_EVENT, CTRL_C_EVENT

                    for sig in (CTRL_C_EVENT, CTRL_BREAK_EVENT):
                        kill(pid, sig)

    def run(self, command: Cmd) -> bool:
        """
        Run the command.
        """
        stream = command.stream and not self.silent
        try:
            proc = Popen(
                args=command.cmdline,
                cwd=command.cwd,
                shell=True,
                stdout=stdout if stream else None,
                stderr=stderr if stream else None,
                stdin=stdin if stream else None,
            )
            self.pids.append(proc.pid)
            proc.communicate()
            if proc.returncode != 0:
                if not self.silent:
                    print(
                        f"{DARK_RED}ERROR:{MAGENTA}",
                        command.cmdline,
                        f"{RESET}exited with code {MAGENTA}{proc.returncode}{RESET}",
                    )
                return False
        except (KeyboardInterrupt, ProcessLookupError):
            return False

        return True
