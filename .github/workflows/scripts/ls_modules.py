#!/usr/bin/python3

"""Script to list Go modules."""

import argparse
import json
import os


def args() -> argparse.Namespace:
    """ Reads the stdin arguments.

        Returns: arguments namespace.
    """
    args = argparse.ArgumentParser(
        description="Script to list Go modules and prints them as JSON-encoded array string to stdout",
    )
    args.add_argument("-p", "--path", help="Base directory to list.", required=True, type=str)
    o = args.parse_args()
    return o


def get_flag(path: str, prefix: str = "") -> str:
    """ Generates the flag to submit to codecov.com. See details: https://docs.codecov.com/docs/flags.

    Args:
        path: Path to the module.
        prefix: Common prefix to remove.

    Return:
        Flag value.
    """
    return path\
        .removeprefix(prefix)\
        .removeprefix(".")\
        .removeprefix("/") \
               .replace("/", "-")\
        .replace(".", "-")[-45:]


def main(path: str) -> None:
    """ Script entrypoint.

    It prints JSON-encoded array of plugins.

    Args:
        path: Base dir to list.
    """
    modules: list[dict[str, str]] = []

    for p, _, files in os.walk(path):
        if "go.mod" in files:

            flag = get_flag(p, path)
            if flag == "":
                flag = "root"

            modules.append({"path": p, "flag": flag})

    print(json.dumps(modules))


if __name__ == "__main__":
    main(args().path)
