import sys
import json
import unittest
from model import Run
from pathlib import Path
from dataclasses import asdict
from collections.abc import Sequence
from runner import KeratTestRunner


def run_tests(filenames: Sequence[str], dir: Path):
    sys.path.insert(0, dir.as_posix())

    loader = unittest.TestLoader()
    suite = loader.loadTestsFromNames(filenames)
    runner = KeratTestRunner()

    res = runner.run(suite)
    res = Run("", True, res)
    print(json.dumps(asdict(res)))
