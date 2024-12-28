from typing import List
from dataclasses import dataclass


@dataclass
class SourceFile:
    filename: str
    src: str


@dataclass
class SourceCode:
    src_test: List[SourceFile]
    src: List[SourceFile]


@dataclass
class TestResult:
    passed: bool
    name: str
    message: str
    stack_trace: str


@dataclass
class Run:
    message: str
    success: bool
    output: List[TestResult]
