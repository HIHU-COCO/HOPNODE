#!/usr/bin/env python3
"""Sanitized lifecycle assertions. Creator: HIHU-COCO. Created: 2026-10-19."""

import hashlib
import json
import os
from pathlib import Path
import stat
import subprocess
import sys


def require(condition, message):
    if not condition:
        raise ValueError(message)


def process_token(pid):
    try:
        fields = Path(f"/proc/{pid}/stat").read_text().rsplit(")", 1)[1].split()
        return fields[19]  # Linux /proc field 22: starttime, after pid and comm.
    except FileNotFoundError:
        return None


def query(core, endpoint, command):
    environment = os.environ.copy()
    environment.pop("DISPLAY", None)
    environment["HOPNODE_POC_ENDPOINT"] = endpoint
    # Never echo or persist raw CLI output: the fixture includes an actual identity.
    return subprocess.run(
        ["/usr/bin/env", "-u", "DISPLAY", core, command, "--json"], env=environment, capture_output=True,
        text=True, timeout=5, check=False,
    )


def read_result(path):
    return json.loads(Path(path).read_text())


def save_result(path, result):
    # Only a fresh task-owned artifact file may be created.
    with open(path, "x", encoding="utf-8") as output:
        json.dump(result, output, indent=2, sort_keys=True)
        output.write("\n")


def check_running(core, endpoint, expected_pid, output):
    parent = Path(endpoint).parent.stat()
    socket = Path(endpoint).stat()
    require(stat.S_IMODE(parent.st_mode) == 0o700, "runtime mode must equal 0700")
    require(stat.S_IMODE(socket.st_mode) == 0o600 and stat.S_ISSOCK(socket.st_mode), "socket must equal 0600")
    require(parent.st_uid == os.geteuid() and socket.st_uid == os.geteuid(), "IPC owner differs from test identity")
    snapshots = {}
    for command in ("status", "doctor"):
        completed = query(core, endpoint, command)
        require(completed.returncode == 0, "headless CLI request failed")
        result = json.loads(completed.stdout)
        require(result.get("api_version") == "poc.v0" and result.get("os") == "linux", "wrong fixture protocol or OS")
        require(result.get("pid") == int(expected_pid), "CLI PID differs from systemd MainPID")
        require(bool(result.get("identity")), "fixture did not report an identity")
        require(bool(result.get("instance_id")), "fixture did not report an instance")
        snapshots[command] = {key: result[key] for key in ("api_version", "version", "pid", "instance_id", "os")}
        snapshots[command]["identity"] = "REDACTED"
    require(snapshots["status"] == snapshots["doctor"], "status and doctor disagree")
    token = process_token(int(expected_pid))
    require(token is not None, "MainPID is absent")
    save_result(output, {
        **snapshots, "process_start_ticks": token, "runtime_mode": "0700",
        "socket_mode": "0600", "display_unset_for_cli": True,
    })
    print("HEADLESS_STATUS_DOCTOR_AND_PRIVATE_IPC=PASS")


def check_stopped(core, endpoint, previous, output):
    old = read_result(previous)
    old_pid = old["status"]["pid"]
    require(process_token(old_pid) != old["process_start_ticks"], "owned service process still exists")
    for command in ("status", "doctor"):
        require(query(core, endpoint, command).returncode != 0, "stopped IPC unexpectedly accepted request")
    require(not Path(endpoint).exists(), "owned runtime socket remained after stop")
    save_result(output, {"old_pid": old_pid, "owned_process_absent": True, "status_rejected": True, "doctor_rejected": True})
    print("STOP_PROCESS_ABSENT_AND_REQUESTS_REJECTED=PASS")


def check_different(previous, current):
    old, new = read_result(previous), read_result(current)
    require(old["status"]["pid"] != new["status"]["pid"], "restart did not produce a new PID")
    require(old["status"]["instance_id"] != new["status"]["instance_id"], "restart did not produce a new instance")
    print("RESTART_NEW_PID_AND_INSTANCE=PASS")


def main():
    command, *args = sys.argv[1:]
    if command == "running" and len(args) == 4:
        check_running(*args)
    elif command == "stopped" and len(args) == 4:
        check_stopped(*args)
    elif command == "different" and len(args) == 2:
        check_different(*args)
    elif command == "metadata" and len(args) == 4:
        core, unit, endpoint, output = args
        save_result(output, {"unit": unit, "endpoint": endpoint, "core_sha256": hashlib.sha256(Path(core).read_bytes()).hexdigest()})
    else:
        raise ValueError("invalid verifier arguments")


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, KeyError, IndexError, subprocess.SubprocessError):
        # No traceback, raw response, environment, identity, or subprocess stderr.
        print("SERVICE_LIFECYCLE_ASSERTION=FAIL", file=sys.stderr)
        sys.exit(1)
