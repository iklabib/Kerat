import unittest
import traceback
from typing import List
from pathlib import Path
from model import TestResult


class KeratTestResult(unittest.TestResult):
    def __init__(self):
        super().__init__()
        self.results: List[TestResult] = []

    def startTest(self, test):
        self.current_test = TestResult(True, test._testMethodName, "", "")

    def stopTest(self, test):
        self.results.append(self.current_test)

    def addError(self, test, err):
        exc_type, exc_value, tb = err
        frame = traceback.extract_tb(tb)[-1]
        self.current_test.passed = False
        self.current_test.stack_trace = f'File "{Path(frame.filename).name}", line {frame.lineno}, in {frame.name}\n    {frame.line}\n{exc_type.__name__}: {exc_value}'

    def addFailure(self, test, err):
        exc_type, exc_value, _ = err
        self.current_test.passed = False
        self.current_test.message = "".join(traceback.format_exception_only(exc_type, exc_value))



class KeratTestRunner:
    def __init__(self, failfast=False):
        self.failfast = failfast

    def run(self, test) -> List[TestResult]:
        result = KeratTestResult()
        result.failfast = self.failfast
        test.run(result)
        return result.results
