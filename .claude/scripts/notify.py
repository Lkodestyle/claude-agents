#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Desktop notification for Claude Code's Notification hook.

Claude Code fires this hook when it needs you (a permission prompt) or when it
has been idle waiting for input. We raise a native notification so you notice
while doing something else, on Linux (notify-send), macOS (osascript) and
Windows (PowerShell toast via BurntToast if present, else a message box).

Never fails the hook: any error exits 0 silently.
"""

import json
import os
import platform
import shutil
import subprocess
import sys


def main() -> None:
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        payload = {}
    message = payload.get("message") or payload.get("title") or "Claude Code te necesita"
    title = "Claude Code"
    kind = payload.get("notification_type") or payload.get("type") or ""
    if kind:
        title = f"Claude Code · {kind}"
    cwd = payload.get("cwd") or os.getcwd()
    body = f"{message}\n{os.path.basename(cwd)}"

    system = platform.system()
    if system == "Linux" and shutil.which("notify-send"):
        subprocess.run(["notify-send", "-a", "Claude Code", title, body], timeout=3, check=False)
    elif system == "Darwin":
        script = f'display notification "{body}" with title "{title}"'
        subprocess.run(["osascript", "-e", script], timeout=3, check=False)
    elif system == "Windows":
        ps = (
            "if (Get-Module -ListAvailable -Name BurntToast) { New-BurntToastNotification -Text "
            f"'{title}', '{body}' }} else {{ Add-Type -AssemblyName PresentationFramework; "
            f"[System.Windows.MessageBox]::Show('{body}', '{title}') | Out-Null }}"
        )
        subprocess.run(["powershell", "-NoProfile", "-Command", ps], timeout=5, check=False)


if __name__ == "__main__":
    try:
        main()
    except Exception:
        pass
