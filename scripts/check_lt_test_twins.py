#!/usr/bin/env python3
"""Audit: every LanguageTool Java unit test has a Go twin under internal/languagetool.

Gate = refactree (`rft ls -a -R`) must see each expected Go Test* symbol.
Does NOT generate ports.

Twin path mapping:
  Single-module class:
    internal/languagetool/<java.package>/<ClassName>_test.go
    FooTest.testBar → TestFoo_Bar
  Same (package, class) in multiple Maven modules (e.g. JLanguageToolTest in en/de/…):
    <ClassName>__{module}_test.go
    FooTest.testBar → TestFoo_{module}_Bar
    module = lang_<code> | languagetool_core | languagetool_standalone | …

Exit 0 only when every twin file exists and every Test* is visible to rft.
"""
from __future__ import annotations

import argparse
import re
import shutil
import subprocess
import sys
from collections import defaultdict
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


def module_key(rel_java: str) -> str:
    parts = rel_java.split("/")
    if parts[0] == "languagetool-language-modules" and len(parts) > 1:
        return "lang_" + parts[1].replace("-", "_")
    return parts[0].replace("-", "_").replace(".", "_")


def class_base(class_name: str) -> str:
    if class_name.endswith("Test"):
        return class_name[: -len("Test")] or class_name
    return class_name


def method_part(method: str) -> str:
    if method == "test":
        return "Test"
    if method.startswith("test") and len(method) > 4:
        rest = method[4:]
        return rest[:1].upper() + rest[1:]
    return method[:1].upper() + method[1:]


def java_method_to_go_test(class_name: str, method: str, mod: str | None) -> str:
    base = class_base(class_name)
    meth = method_part(method)
    if mod:
        return f"Test{base}_{mod}_{meth}"
    return f"Test{base}_{meth}"


def go_file_rel(package: str, class_name: str, mod: str | None) -> str:
    pkg = package.replace(".", "/")
    if mod:
        return f"{pkg}/{class_name}__{mod}_test.go"
    return f"{pkg}/{class_name}_test.go"


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
        print("error: rft not found on PATH (install refactree / mise install)", file=sys.stderr)
        sys.exit(2)
    return [ln.strip() for ln in proc.stdout.splitlines() if ln.strip()]


def normalize_rft_symbol(sym: str) -> str:
    if "::" in sym:
        sym = sym.split("::", 1)[1]
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


def has_go_func(combined: set[str], name: str) -> bool:
    if name in combined:
        return True
    return any(x.endswith("." + name) for x in combined)


@dataclass
class JavaTestClass:
    rel_java: str
    package: str
    class_name: str
    methods: list[str] = field(default_factory=list)
    mod: str | None = None  # set when multi-module

    @property
    def go_file_rel(self) -> str:
        return go_file_rel(self.package, self.class_name, self.mod)


def discover_java_tests(lt_root: Path) -> list[JavaTestClass]:
    buckets: dict[tuple[str, str], list[JavaTestClass]] = defaultdict(list)
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
        jt = JavaTestClass(
            rel_java=rel,
            package=pkg_m.group(1),
            class_name=cls_m.group(1),
            methods=methods,
        )
        buckets[(jt.package, jt.class_name)].append(jt)

    out: list[JavaTestClass] = []
    for key, items in buckets.items():
        multi = len(items) > 1
        for jt in items:
            if multi:
                jt.mod = module_key(jt.rel_java)
            out.append(jt)
    return out


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--lt-root", type=Path, default=DEFAULT_LT)
    ap.add_argument("--go-root", type=Path, default=DEFAULT_GO)
    ap.add_argument("--sample", type=int, default=40)
    ap.add_argument("--self-test", action="store_true")
    ap.add_argument(
        "--rft-java",
        action="store_true",
        help="also require each Java @Test method name to appear in rft ls of that .java file",
    )
    args = ap.parse_args()

    if args.self_test:
        assert java_method_to_go_test("MultipleWhitespaceRuleTest", "testRule", None) == "TestMultipleWhitespaceRule_Rule"
        assert java_method_to_go_test("WordRepeatRuleTest", "test", None) == "TestWordRepeatRule_Test"
        assert java_method_to_go_test("JLanguageToolTest", "testEnglish", "lang_en") == "TestJLanguageTool_lang_en_English"
        assert go_file_rel("org.languagetool", "JLanguageToolTest", "lang_en") == "org/languagetool/JLanguageToolTest__lang_en_test.go"
        print("self-test ok")
        return 0

    if not args.lt_root.is_dir():
        print(f"error: LT root missing: {args.lt_root}", file=sys.stderr)
        return 2
    if not shutil.which("rft"):
        print("error: rft required on PATH", file=sys.stderr)
        return 2

    java_tests = discover_java_tests(args.lt_root)
    print("running rft ls -a -R on Go twin tree…", flush=True)
    go_syms = symbol_set(run_rft_ls(args.go_root))
    print(f"  rft symbol names: {len(go_syms)}", flush=True)

    missing_files: list[str] = []
    missing_methods: list[str] = []
    rft_java_miss: list[str] = []
    total_methods = 0

    for jt in java_tests:
        total_methods += len(jt.methods)
        go_file = args.go_root / jt.go_file_rel
        if not go_file.is_file():
            missing_files.append(f"{jt.go_file_rel}  ← {jt.rel_java} ({len(jt.methods)} @Test)")
            for jm in jt.methods:
                gm = java_method_to_go_test(jt.class_name, jm, jt.mod)
                missing_methods.append(f"{jt.go_file_rel} :: {gm}  ← {jt.class_name}.{jm}")
            continue

        if args.rft_java:
            jnames = symbol_set(run_rft_ls(args.lt_root / jt.rel_java))
            for jm in jt.methods:
                if jm not in jnames and not any(x.endswith("." + jm) for x in jnames):
                    rft_java_miss.append(f"{jt.rel_java} :: {jm}")

        for jm in jt.methods:
            gm = java_method_to_go_test(jt.class_name, jm, jt.mod)
            if not has_go_func(go_syms, gm):
                missing_methods.append(
                    f"{jt.go_file_rel} :: {gm}  ← {jt.class_name}.{jm} ({jt.rel_java})"
                )

    print("LT Java → Go test twin audit (refactree-backed)")
    print(f"  lt_root:       {args.lt_root}")
    print(f"  go_root:       {args.go_root}")
    print(f"  java *Test:    {len(java_tests)}")
    print(f"  @Test methods: {total_methods}")
    print(f"  missing files: {len(missing_files)}")
    print(f"  missing funcs: {len(missing_methods)}")
    if rft_java_miss:
        print(f"  java rft misses: {len(rft_java_miss)}")

    if missing_files:
        print(f"\nMissing Go twin files (first {args.sample}):")
        for line in missing_files[: args.sample]:
            print(f"  - {line}")
        if len(missing_files) > args.sample:
            print(f"  … {len(missing_files) - args.sample} more")

    if missing_methods:
        print(f"\nMissing Go Test* symbols per rft (first {args.sample}):")
        for line in missing_methods[: args.sample]:
            print(f"  - {line}")
        if len(missing_methods) > args.sample:
            print(f"  … {len(missing_methods) - args.sample} more")

    if not missing_files and not missing_methods:
        print("\nOK: every Java @Test has a Go twin symbol visible to rft.")
        return 0

    print(
        "\nFAIL: rft does not see all expected twins "
        "(full correctness gate — not a text grep of sources)",
        file=sys.stderr,
    )
    return 1


if __name__ == "__main__":
    sys.exit(main())
