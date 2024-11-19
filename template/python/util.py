import os
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
    with path.open('w') as w:
        w.write(content)

def get_timeout() -> int:
    v = os.environ.get('TIMEOUT')
    if v is None or v == "":
        exit("env timeout not defined")
    
    try:
        return int(v)
    except ValueError:
        exit("failed to parse env timeout")