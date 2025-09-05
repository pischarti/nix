# nix

## Nix Installation

- https://nix.dev/install-nix.html

## Enter flake shell

```sh
nix develop
```

## Delete .db files recursively

Safe CLI to find and delete `.db` files recursively.

- Run dry-run (default):
```sh
python ./py/sql/drop_db.py ./py
```

- Actually delete (confirm):
```sh
python ./sql/drop_db.py ./py --yes
```

- Options:
  - `--ext .db .sqlite`: match additional extensions
  - `--quiet`: only show summary

- Behavior:
  - Recurses from the folder
  - Lists matched files
  - Deletes only with `--yes`
  - Prints a summary and any failures

## Unit tests for drop_db.py

- Coverage:
  - Find `.db` files (single/multiple extensions, case-insensitive)
  - Deletion behavior
  - CLI dry-run vs `--yes`
  - `--ext` filtering

- Run tests:
```sh
python3 -m unittest -v py/sql/tests/test_drop_db.py
```

## Run with uv

Dependencies are listed in example `requirements.txt`.

1. Install uv (macOS if don't have nix):
```sh
brew install uv
```

2. Create a virtual environment:
```sh
uv venv
```

3. Activate it (zsh/bash):
```sh
source .venv/bin/activate
```

4. Install dependencies:
```sh
uv pip install -r requirements.txt
```

5. Run the script:
```sh
uv run py/panda/join_agg_vis.py
```
