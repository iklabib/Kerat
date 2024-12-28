import sys
import json
from model import Run
from pathlib import Path
from dataclasses import asdict


def exit(msg: str):
    res = json.dumps(asdict(Run(msg, False, [])))
    print(res)
    sys.exit(0)


def write(path: Path, content: str):
    with path.open("w") as w:
        w.write(content)