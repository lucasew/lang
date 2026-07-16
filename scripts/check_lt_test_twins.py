#!/usr/bin/env python3
"""Audit: every LanguageTool Java unit test has a Go twin under internal/languagetool.

Uses refactree (`rft ls -a`) for Go symbol discovery. Discovers Java @Test methods
from sources (annotation is the ground-truth list of tests). Does NOT generate ports.

Mapping (predictable transform):
  Java:
    .../src/test/java/org/languagetool/rules/FooTest.java
    package org.languagetool.rules;  class FooTest;  @Test void testRule()
  Go twin:
    internal/languagetool/org/languagetool/rules/FooTest_test.go
    func TestRule(t *testing.T)

  FooTest.testRule → TestFoo_Rule | FooTest.test → TestFoo_Test
  FooTest.avoidSomeWords → TestFoo_AvoidSomeWords
  (class prefix avoids same-package collisions e.g. testToString)

Exit 0 only when every twin file and Test* symbol exists (visible to rft).
"""
from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_LT = ROOT / "inspiration" / "languagetool"
DEFAULT_GO = ROOT / "internal" / "languagetool"

RE_TEST_METHOD = re.compile(
    r"@Test\b(?:\([^)]*\))?\s*"
    r"(?:public\s+|protected\s+|private\s+)?"
    r"(?:static\s+)?"
    r"void\s+(\w+)\s*\(",
    re.MULTILINE,
)
RE_PACKAGE = re.compile(r"(?m)^package\s+([\w.]+)\s*;")
RE_CLASS = re.compile(r"(?m)(?:public\s+)?(?:abstract\s+)?class\s+(\w+Test)\b")


def java_method_to_go_test(class_name: str, method: str) -> str:
    """Java @Test method → unique Go Test* name (class-prefixed to avoid package collisions).

    AnalyzedTokenTest.testToString → TestAnalyzedToken_ToString
    WordRepeatRuleTest.test       → TestWordRepeatRule_Test
    FooTest.avoidSomeWords        → TestFoo_AvoidSomeWords
    """
    base = class_name
    if base.endswith("Test"):
        base = base[: -len("Test")]
    if not base:
        base = class_name
    if method == "test":
        meth = "Test"
    elif method.startswith("test") and len(method) > 4:
        rest = method[4:]
        meth = rest[:1].upper() + rest[1:]
    else:
        meth = method[:1].upper() + method[1:]
    return "Test" + base + "_" + meth


def run_rft_ls(path: Path) -> list[str]:
    if not path.exists():
        return []
    cmd = ["rft", "ls", "-a"]
    if path.is_dir():
        cmd.append("-R")
    cmd.append(f"path:{path}::")
    try:
        proc = subprocess.run(
            cmd, cwd=str(ROOT), capture_output=True, text=True, check=False
        )
    except FileNotFoundError:
        print("error: rft not found on PATH (install refactree)", file=sys.stderr)
        sys.exit(2)
    return [ln.strip() for ln in proc.stdout.splitlines() if ln.strip()]


def normalize_rft_symbol(sym: str) -> str:
    if "::" in sym:
        sym = sym.split("::", 1)[1]
    # strip glued source offsets from rft -l style (e.g. testRule11911199)
    sym = re.sub(r"(?<=[A-Za-z_])\d{6,}$", "", sym)
    return sym


def symbol_set(raw: list[str]) -> set[str]:
    names: set[str] = set()
    for s in raw:
        s = normalize_rft_symbol(s)
        names.add(s)
        if "." in s:
            names.add(s.rsplit(".", 1)[-1])
    return names


def go_package_index(go_root: Path) -> dict[str, set[str]]:
    if not go_root.is_dir():
        return {}
    out: dict[str, set[str]] = {}
    for dirpath, dirnames, filenames in os.walk(go_root):
        dirnames[:] = [d for d in dirnames if not d.startswith(".")]
        if not any(f.endswith(".go") for f in filenames):
            continue
        p = Path(dirpath)
        rel = p.relative_to(go_root).as_posix()
        out[rel] = symbol_set(run_rft_ls(p))
    return out


@dataclass
class JavaTestClass:
    rel_java: str
    package: str
    class_name: str
    methods: list[str] = field(default_factory=list)

    @property
    def go_pkg_rel(self) -> str:
        return self.package.replace(".", "/")

    @property
    def go_file_rel(self) -> str:
        return f"{self.go_pkg_rel}/{self.class_name}_test.go"


def discover_java_tests(lt_root: Path) -> list[JavaTestClass]:
    results: list[JavaTestClass] = []
    for path in sorted(lt_root.rglob("*Test.java")):
        if "/src/test/java/" not in path.as_posix():
            continue
        rel = path.relative_to(lt_root).as_posix()
        text = path.read_text(encoding="utf-8", errors="replace")
        pkg_m = RE_PACKAGE.search(text)
        cls_m = RE_CLASS.search(text)
        if not pkg_m or not cls_m:
            print(f"warn: skip (no package/class): {rel}", file=sys.stderr)
            continue
        methods: list[str] = []
        seen: set[str] = set()
        for m in RE_TEST_METHOD.findall(text):
            if m not in seen:
                seen.add(m)
                methods.append(m)
        results.append(
            JavaTestClass(
                rel_java=rel,
                package=pkg_m.group(1),
                class_name=cls_m.group(1),
                methods=methods,
            )
        )
    return results


def has_go_func(combined: set[str], name: str) -> bool:
    if name in combined:
        return True
    return any(x.endswith("." + name) for x in combined)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--lt-root", type=Path, default=DEFAULT_LT)
    ap.add_argument("--go-root", type=Path, default=DEFAULT_GO)
    ap.add_argument("--sample", type=int, default=40)
    ap.add_argument("--self-test", action="store_true", help="run transform unit checks and exit")
    args = ap.parse_args()

    if args.self_test:
        assert java_method_to_go_test("MultipleWhitespaceRuleTest", "testRule") == "TestMultipleWhitespaceRule_Rule"
        assert java_method_to_go_test("WordRepeatRuleTest", "test") == "TestWordRepeatRule_Test"
        assert java_method_to_go_test("GlobalSpellingTest", "avoidSomeWords") == "TestGlobalSpelling_AvoidSomeWords"
        assert java_method_to_go_test("AnalyzedTokenTest", "testToString") == "TestAnalyzedToken_ToString"
        assert java_method_to_go_test("AnalyzedTokenReadingsTest", "testToString") == "TestAnalyzedTokenReadings_ToString"
        print("self-test ok")
        return 0

    if not args.lt_root.is_dir():
        print(f"error: LT root missing: {args.lt_root}", file=sys.stderr)
        return 2

    java_tests = discover_java_tests(args.lt_root)
    go_idx = go_package_index(args.go_root)

    missing_files: list[str] = []
    missing_methods: list[str] = []
    total_methods = 0

    for jt in java_tests:
        total_methods += len(jt.methods)
        go_file = args.go_root / jt.go_file_rel
        pkg_syms = go_idx.get(jt.go_pkg_rel, set())

        if not go_file.is_file():
            missing_files.append(
                f"{jt.go_file_rel}  ← {jt.rel_java} ({len(jt.methods)} @Test)"
            )
            for jm in jt.methods:
                gm = java_method_to_go_test(jt.class_name, jm)
                missing_methods.append(
                    f"{jt.go_file_rel} :: {gm}  ← {jt.class_name}.{jm}"
                )
            continue

        combined = set(pkg_syms) | symbol_set(run_rft_ls(go_file))
        for jm in jt.methods:
            gm = java_method_to_go_test(jt.class_name, jm)
            if not has_go_func(combined, gm):
                missing_methods.append(
                    f"{jt.go_file_rel} :: {gm}  ← {jt.class_name}.{jm}"
                )

    n_classes = len(java_tests)
    print("LT Java → Go test twin audit (refactree-backed)")
    print(f"  lt_root:       {args.lt_root}")
    print(f"  go_root:       {args.go_root}")
    print(f"  java *Test:    {n_classes}")
    print(f"  @Test methods: {total_methods}")
    print(f"  missing files: {len(missing_files)}")
    print(f"  missing funcs: {len(missing_methods)}")

    if missing_files:
        print(f"\nMissing Go twin files (first {args.sample}):")
        for line in missing_files[: args.sample]:
            print(f"  - {line}")
        if len(missing_files) > args.sample:
            print(f"  … {len(missing_files) - args.sample} more")

    if missing_methods:
        print(f"\nMissing Go Test* symbols (first {args.sample}):")
        for line in missing_methods[: args.sample]:
            print(f"  - {line}")
        if len(missing_methods) > args.sample:
            print(f"  … {len(missing_methods) - args.sample} more")

    if not missing_files and not missing_methods:
        print("\nOK: every Java @Test has a Go twin under internal/languagetool.")
        return 0

    print(
        "\nFAIL: hand-port tests into "
        "internal/languagetool/<java.package>/<ClassName>_test.go",
        file=sys.stderr,
    )
    return 1


if __name__ == "__main__":
    sys.exit(main())
