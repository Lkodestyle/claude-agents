#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
lint-gate: PostToolUse hook on Edit/Write. Format the file in place and feed
remaining lint errors back to Claude, BEFORE CI gets a chance to reject them.

Opt-in by repo convention, not by flag: a tool only runs when the repo already
uses it (a ruff/prettier/eslint config exists, or the tool is the standard one
for the language, like gofmt). A repo with no tooling is left alone, so the
hook is safe to install globally.

Exit codes (PostToolUse contract):
  0 -> nothing to say (file formatted silently, or no tool applies)
  2 -> stderr is shown to Claude (lint errors remain after formatting)
Anything that goes wrong in the gate itself exits 0: the gate must never block
an edit because a linter is missing or crashed.
"""

import json
import os
import shutil
import subprocess
import sys
from pathlib import Path

TIMEOUT = 20


def run(argv, cwd=None):
    return subprocess.run(argv, capture_output=True, text=True, timeout=TIMEOUT, cwd=cwd)


def repo_root(path: Path) -> Path:
    for parent in (path.parent, *path.parent.parents):
        if (parent / ".git").exists():
            return parent
    return path.parent


def has_any(root: Path, *names: str) -> bool:
    return any((root / n).exists() for n in names)


def pkg_has(root: Path, key: str) -> bool:
    try:
        pkg = json.loads((root / "package.json").read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return False
    return key in pkg.get("devDependencies", {}) or key in pkg.get("dependencies", {})


def pyproject_mentions(root: Path, word: str) -> bool:
    try:
        return word in (root / "pyproject.toml").read_text(encoding="utf-8")
    except OSError:
        return False


def gate(path: Path) -> tuple[bool, str]:
    """Returns (ran_something, remaining_errors)."""
    root = repo_root(path)
    suffix = path.suffix.lower()
    errors = []
    ran = False

    if suffix == ".py":
        uses_ruff = has_any(root, "ruff.toml", ".ruff.toml") or pyproject_mentions(root, "ruff")
        if uses_ruff and shutil.which("ruff"):
            ran = True
            run(["ruff", "format", "-q", str(path)], cwd=root)
            check = run(["ruff", "check", "--fix", "-q", str(path)], cwd=root)
            if check.returncode != 0:
                errors.append(check.stdout.strip() or check.stderr.strip())
        # Syntax is free and universal: a broken .py should never survive the turn.
        syntax = run([sys.executable, "-m", "py_compile", str(path)])
        if syntax.returncode != 0:
            ran = True
            errors.append(syntax.stderr.strip())

    elif suffix in (".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs", ".json", ".css", ".scss", ".vue", ".svelte"):
        uses_prettier = has_any(root, ".prettierrc", ".prettierrc.json", ".prettierrc.js", "prettier.config.js") or pkg_has(root, "prettier")
        if uses_prettier and shutil.which("npx"):
            ran = True
            run(["npx", "--no-install", "prettier", "--write", "--log-level", "error", str(path)], cwd=root)
        uses_eslint = has_any(root, ".eslintrc.json", ".eslintrc.js", "eslint.config.js", "eslint.config.mjs") or pkg_has(root, "eslint")
        if uses_eslint and shutil.which("npx") and suffix not in (".json", ".css", ".scss"):
            ran = True
            check = run(["npx", "--no-install", "eslint", "--fix", str(path)], cwd=root)
            if check.returncode != 0:
                errors.append(check.stdout.strip() or check.stderr.strip())

    elif suffix == ".go" and shutil.which("gofmt"):
        ran = True
        run(["gofmt", "-w", str(path)])
        if shutil.which("go") and has_any(root, "go.mod"):
            vet = run(["go", "vet", "./" + str(path.parent.relative_to(root)) + "/..."], cwd=root)
            if vet.returncode != 0:
                errors.append(vet.stderr.strip())

    elif suffix in (".tf", ".tfvars") and shutil.which("terraform"):
        ran = True
        run(["terraform", "fmt", str(path)], cwd=root)

    elif suffix in (".sh", ".bash"):
        bash = shutil.which("bash")
        if bash:
            ran = True
            syntax = run([bash, "-n", str(path)])
            if syntax.returncode != 0:
                errors.append(syntax.stderr.strip())
        if shutil.which("shellcheck"):
            check = run(["shellcheck", "-S", "error", str(path)])
            if check.returncode != 0:
                errors.append(check.stdout.strip())

    return ran, "\n".join(e for e in errors if e)


def main() -> int:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0
    file_path = (payload.get("tool_input") or {}).get("file_path", "")
    if not file_path:
        return 0
    path = Path(file_path)
    if not path.is_file():
        return 0
    try:
        _, errors = gate(path)
    except Exception:
        return 0
    if errors:
        print(f"[lint-gate] {path.name} still has issues after auto-format:\n{errors[:3000]}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())
