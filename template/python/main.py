import json
from kerat import run_tests
from pathlib import Path
from util import exit
from model import SourceCode, SourceFile


def load_source_code(json_data: str) -> SourceCode:
    try:
        data = json.loads(json_data)

        src_test = [SourceFile(**item) for item in data.get("src_test", [])]
        src = [SourceFile(**item) for item in data.get("src", [])]

        return SourceCode(src_test=src_test, src=src)
    except Exception:
        exit("failed to read source codes")


if __name__ == "__main__":
    workdir = Path("/") / "workspace"
    filenames = [Path(x.name).stem for x in workdir.glob("*.py")]
    run_tests(filenames, workdir.absolute())
