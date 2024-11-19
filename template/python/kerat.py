import sys
import json
import unittest
from model import Run
from pathlib import Path
from util import get_timeout
from dataclasses import asdict
from collections.abc import Sequence
from runner import KeratTestRunner

def run_tests(filenames: Sequence[str], dir: Path):
    sys.path.insert(0, dir.as_posix())

    loader = unittest.TestLoader()
    suite = loader.loadTestsFromNames(filenames)
    timeout = get_timeout()
    runner = KeratTestRunner(timeout)

    res = runner.run(suite)
    res = Run("", True, res)
    print(json.dumps(asdict(res)))