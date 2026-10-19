#!/usr/bin/env python3
"""Inspect existing user-session prerequisites without printing private environment.
Purpose: P04 read-only session discovery. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import argparse
import importlib.util
import json
import os
import shutil
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument("--write-environment", help="Create a new private metadata file; never overwrite")
args = parser.parse_args()
if os.getuid() == 0:
    raise SystemExit("Run as the authorized desktop user, without sudo")
manager = subprocess.run(["systemctl", "--user", "show-environment"], capture_output=True, text=True)
if manager.returncode:
    raise SystemExit("User systemd environment unavailable; no environment was guessed")
allowed = {"DISPLAY", "XAUTHORITY", "DBUS_SESSION_BUS_ADDRESS", "XDG_RUNTIME_DIR", "XDG_SESSION_TYPE"}
values = {}
for line in manager.stdout.splitlines():
    key, separator, value = line.partition("=")
    if separator and key in allowed:
        values[key] = value
# The SSH process's actual user runtime/bus values may fill missing manager values.
# DISPLAY/XAUTHORITY are deliberately never inferred or synthesized.
for key in ("XDG_RUNTIME_DIR", "DBUS_SESSION_BUS_ADDRESS"):
    if key not in values and os.environ.get(key):
        values[key] = os.environ[key]
summary = {
    "uid": os.getuid(),
    "tools": {name: bool(shutil.which(name)) for name in ("dpkg-deb", "python3", "systemctl", "systemd-run")},
    "gi_available": importlib.util.find_spec("gi") is not None,
    "session_fields_present": {key: bool(values.get(key)) for key in sorted(allowed)},
    "xauthority_readable": bool(values.get("XAUTHORITY") and os.access(values["XAUTHORITY"], os.R_OK)),
}
if args.write_environment:
    metadata = {"uid": os.getuid(), "environment": values}
    descriptor = os.open(args.write_environment, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(descriptor, "w", encoding="utf-8") as output:
        json.dump(metadata, output)
    summary["private_metadata_created"] = True
print(json.dumps(summary, sort_keys=True))
