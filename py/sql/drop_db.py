import argparse
import os
import sys
from pathlib import Path
from typing import Iterable, List, Sequence, Tuple


def find_db_files(root_directory: Path, suffixes: Sequence[str]) -> List[Path]:
    """Return a list of database files under root_directory matching suffixes.

    Search is recursive and does not follow directory symlinks.
    Matching is case-insensitive on file extensions.
    """
    normalized_suffixes = tuple(s.lower() for s in suffixes)

    matched_files: List[Path] = []
    for current_root, dir_names, file_names in os.walk(root_directory, followlinks=False):
        # Convert to Path once per directory for efficiency
        current_root_path = Path(current_root)
        for file_name in file_names:
            if file_name.lower().endswith(normalized_suffixes):
                matched_files.append(current_root_path / file_name)
    return matched_files


def delete_files(files: Iterable[Path]) -> Tuple[List[Path], List[Tuple[Path, Exception]]]:
    """Attempt to delete each file. Return (deleted, failures)."""
    deleted: List[Path] = []
    failures: List[Tuple[Path, Exception]] = []

    for file_path in files:
        try:
            file_path.unlink(missing_ok=True)
            # If file still exists after unlink attempt, treat as failure
            if file_path.exists():
                raise RuntimeError("File still exists after delete attempt")
            deleted.append(file_path)
        except Exception as exc:  # noqa: BLE001 - we want to report all exceptions
            failures.append((file_path, exc))

    return deleted, failures


def parse_args(argv: Sequence[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Find and delete all .db files under a folder (dry-run by default)",
    )
    parser.add_argument(
        "path",
        nargs="?",
        default=str(Path.cwd()),
        help="Folder to scan recursively (default: current working directory)",
    )
    parser.add_argument(
        "--yes",
        action="store_true",
        help="Actually delete the files (without this flag, performs a dry-run)",
    )
    parser.add_argument(
        "--ext",
        dest="extensions",
        default=[".db"],
        metavar="EXT",
        nargs="+",
        help="File extensions to match (default: .db). Example: --ext .db .sqlite",
    )
    parser.add_argument(
        "--quiet",
        action="store_true",
        help="Only print the final summary",
    )
    return parser.parse_args(argv)


def main(argv: Sequence[str]) -> int:
    args = parse_args(argv)

    root = Path(args.path).expanduser().resolve()
    if not root.exists() or not root.is_dir():
        print(f"Error: '{root}' is not a directory.", file=sys.stderr)
        return 2

    # Normalize extensions to ensure they all start with a dot
    extensions = tuple(
        ext if ext.startswith(".") else f".{ext}" for ext in args.extensions
    )

    files = find_db_files(root, extensions)

    if not args.quiet:
        if not files:
            print(f"No files with extensions {extensions} found under: {root}")
        else:
            action = "Would delete" if not args.yes else "Deleting"
            print(f"{action} {len(files)} file(s) under: {root}")
            for file_path in files:
                print(f" - {file_path}")

    if not files:
        return 0

    if not args.yes:
        print(
            "Dry-run complete. Pass --yes to actually delete these files.",
        )
        return 0

    deleted, failures = delete_files(files)

    # Summary
    print(
        f"Summary: deleted={len(deleted)} failed={len(failures)} total={len(files)}",
    )
    if failures and not args.quiet:
        print("Failures:")
        for file_path, exc in failures:
            print(f" - {file_path}: {exc}")

    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))


