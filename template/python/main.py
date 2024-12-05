import json
import tempfile
from pathlib import Path
from kerat import run_tests
from util import write, exit
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
    payload = input()
    source_code = load_source_code(payload)

    with tempfile.TemporaryDirectory() as tmpdir:
        dir = Path(tmpdir)

        sources = source_code.src + source_code.src_test
        for src in sources:
            write(dir / src.filename, src.src)

        filenames = [Path(x.filename).stem for x in source_code.src_test]
        run_tests(filenames, dir.absolute())
