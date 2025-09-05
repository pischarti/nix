import io
import os
import sys
import types
import shutil
import tempfile
import unittest
from contextlib import redirect_stdout
from pathlib import Path
import importlib.util


def _load_drop_db_module() -> types.ModuleType:
    here = Path(__file__).resolve()
    module_path = here.parents[1] / "drop_db.py"
    spec = importlib.util.spec_from_file_location("drop_db", str(module_path))
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)  # type: ignore[attr-defined]
    return module


class DropDbTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmpdir = tempfile.mkdtemp(prefix="dropdb-tests-")
        self.root = Path(self.tmpdir)
        # Create a nested structure with various files
        (self.root / "sub").mkdir(parents=True, exist_ok=True)
        (self.root / "a.db").write_text("dummy")
        (self.root / "b.sqlite").write_text("dummy")
        (self.root / "c.txt").write_text("dummy")
        (self.root / "sub" / "n.DB").write_text("dummy")

        # Load module under test
        self.drop_db = _load_drop_db_module()

    def tearDown(self) -> None:
        shutil.rmtree(self.tmpdir, ignore_errors=True)

    def test_find_db_files_single_extension(self) -> None:
        files = self.drop_db.find_db_files(self.root, [".db"])  # only .db
        found = {p.name for p in files}
        self.assertIn("a.db", found)
        self.assertIn("n.DB", found)  # case-insensitive match
        self.assertNotIn("b.sqlite", found)
        self.assertNotIn("c.txt", found)

    def test_find_db_files_multiple_extensions(self) -> None:
        files = self.drop_db.find_db_files(self.root, [".db", ".sqlite"])  # both
        found = {p.name for p in files}
        self.assertEqual(found, {"a.db", "b.sqlite", "n.DB"})

    def test_delete_files(self) -> None:
        targets = [self.root / "a.db", self.root / "sub" / "n.DB"]
        deleted, failures = self.drop_db.delete_files(targets)
        self.assertEqual(len(failures), 0)
        for p in targets:
            self.assertFalse(p.exists())
        self.assertEqual({p.name for p in deleted}, {"a.db", "n.DB"})

    def test_main_dry_run_does_not_delete(self) -> None:
        # Ensure files exist
        self.assertTrue((self.root / "a.db").exists())
        self.assertTrue((self.root / "b.sqlite").exists())

        stdout = io.StringIO()
        with redirect_stdout(stdout):
            code = self.drop_db.main([str(self.root)])

        self.assertEqual(code, 0)
        # Files should still exist after dry-run
        self.assertTrue((self.root / "a.db").exists())
        self.assertTrue((self.root / "b.sqlite").exists())
        self.assertIn("Dry-run", stdout.getvalue())

    def test_main_yes_deletes(self) -> None:
        # Ensure files exist
        (self.root / "extra.db").write_text("dummy")
        files_before = list(self.root.rglob("*.db"))
        self.assertGreaterEqual(len(files_before), 2)

        stdout = io.StringIO()
        with redirect_stdout(stdout):
            code = self.drop_db.main([str(self.root), "--yes"])  # delete

        self.assertEqual(code, 0)
        # All .db files should be gone; .sqlite remains
        self.assertEqual(list(self.root.rglob("*.db")), [])
        self.assertTrue((self.root / "b.sqlite").exists())
        self.assertIn("Summary:", stdout.getvalue())

    def test_main_ext_only_sqlite(self) -> None:
        # Delete only .sqlite
        stdout = io.StringIO()
        with redirect_stdout(stdout):
            code = self.drop_db.main([str(self.root), "--yes", "--ext", ".sqlite"])  # only sqlite

        self.assertEqual(code, 0)
        self.assertFalse((self.root / "b.sqlite").exists())
        # .db files remain
        self.assertTrue((self.root / "a.db").exists())
        self.assertTrue((self.root / "sub" / "n.DB").exists())


if __name__ == "__main__":
    unittest.main()


