#!/usr/bin/env python3
"""Run P04 only inside a unique private stage, with explicitly supplied X11 metadata.
Purpose: real Ubuntu Desktop GUI/UDS/headless lifecycle probe. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import pwd
import re
import shutil
import signal
import stat
import subprocess
import sys
import tempfile
import time
import uuid

parser = argparse.ArgumentParser()
parser.add_argument("--deb", required=True, type=Path)
parser.add_argument("--core", required=True, type=Path)
parser.add_argument("--session-env", required=True, type=Path)
parser.add_argument("--user-unit", action="store_true", help="Use an isolated transient systemd --user unit for Core")
parser.add_argument("--installed-desktop", type=Path, help="Use the verified installed HopNode .desktop through Gio; never installs/removes packages")
args = parser.parse_args()
os.umask(0o077)
sys.dont_write_bytecode = True
uid = os.getuid()
if uid == 0:
    raise SystemExit("Run as the authorized desktop user, without sudo")
descriptor = os.open(args.session_env, os.O_RDONLY | os.O_NOFOLLOW)
with os.fdopen(descriptor, "r", encoding="utf-8") as input_file:
    metadata_stat = os.fstat(input_file.fileno())
    if metadata_stat.st_uid != uid or stat.S_IMODE(metadata_stat.st_mode) & 0o077:
        raise SystemExit("Session metadata must be current-user-owned and private (0600)")
    metadata = json.load(input_file)
if metadata.get("uid") != uid:
    raise SystemExit("Session metadata UID does not match the executing user")
session = metadata.get("environment", {})
required = ("DISPLAY", "XAUTHORITY", "DBUS_SESSION_BUS_ADDRESS", "XDG_RUNTIME_DIR")
if any(not isinstance(session.get(key), str) or not session[key] for key in required):
    raise SystemExit("Explicit session metadata is incomplete; DISPLAY/XAUTHORITY/bus were not guessed")
if not re.fullmatch(r":[0-9]+(?:\.[0-9]+)?", session["DISPLAY"]):
    raise SystemExit("Only the explicitly authorized local X11 DISPLAY is accepted")
if not os.access(session["XAUTHORITY"], os.R_OK) or not session["DBUS_SESSION_BUS_ADDRESS"].startswith("unix:"):
    raise SystemExit("Explicit X11 authority or local session bus is unavailable")
runtime = Path(session["XDG_RUNTIME_DIR"]).stat()
if runtime.st_uid != uid or stat.S_IMODE(runtime.st_mode) & 0o077:
    raise SystemExit("Existing user runtime directory is not private/current-user-owned")
for artifact in (args.deb, args.core):
    if not artifact.is_file():
        raise SystemExit("An explicitly supplied input artifact is missing")

stage = Path(tempfile.mkdtemp(prefix=f"hopnode-p04-{uid}-", dir="/tmp"))
stage.chmod(0o700)
for name in ("runtime", "bin", "config", "data", "cache", "state", "extracted"):
    (stage / name).mkdir(mode=0o700)
report = {"status": "FAIL", "uid": uid, "checks": {}, "cleanup": {}, "artifact_sha256": {key: hashlib.sha256(path.read_bytes()).hexdigest() for key, path in (("gui_deb", args.deb), ("core", args.core))}}
processes = []
streams = []
group_cleanup = {}
unit = "hopnode-p04-" + uuid.uuid4().hex + ".service" if args.user_unit else None
unit_started = False
manifest = {"uid": uid, "runner_pid": os.getpid(), "stage": str(stage), "process_groups": [], "transient_user_unit": unit}
with open(stage / "owned-processes.json", "x", encoding="utf-8") as output:
    json.dump(manifest, output)
environment = {"PATH": "/usr/local/bin:/usr/bin:/bin", "LANG": "C.UTF-8", "HOME": pwd.getpwuid(uid).pw_dir, "USER": pwd.getpwuid(uid).pw_name}
environment.update({key: session[key] for key in required})
environment.update({"XDG_CONFIG_HOME": str(stage / "config"), "XDG_DATA_HOME": str(stage / "data"), "XDG_CACHE_HOME": str(stage / "cache"), "XDG_STATE_HOME": str(stage / "state"), "HOPNODE_POC_ENDPOINT": str(stage / "runtime/core.sock")})
core_environment = dict(environment)
for key in ("DISPLAY", "XAUTHORITY", "DBUS_SESSION_BUS_ADDRESS"):
    core_environment.pop(key, None)
core_environment["HOPNODE_POC_AUDIT_FILE"] = str(stage / "ipc-audit.jsonl")
core_binary = stage / "bin/hopnodectl-poc"


def run(command, *, env=None, timeout=15):
    return subprocess.run(command, env=env or environment, capture_output=True, text=True, timeout=timeout)


def launch(command, env, prefix):
    output = open(stage / (prefix + ".stdout"), "xb")
    error = open(stage / (prefix + ".stderr"), "xb")
    streams.extend((output, error))
    process = subprocess.Popen(command, env=env, stdout=output, stderr=error, start_new_session=True)
    processes.append(process)
    manifest["process_groups"].append({"leader_pid": process.pid, "purpose": prefix})
    with open(stage / "owned-processes.json", "w", encoding="utf-8") as output:
        json.dump(manifest, output)
    return process


def launch_installed_gui(entry_module):
    output = open(stage / "gui.stdout", "xb")
    error = open(stage / "gui.stderr", "xb")
    streams.extend((output, error))
    def remember_child(process):
        processes.append(process)
        manifest["process_groups"].append({"leader_pid": process.pid, "purpose": "installed-gui-desktop-entry"})
        with open(stage / "owned-processes.json", "w", encoding="utf-8") as output:
            json.dump(manifest, output)

    process = entry_module.launch_desktop_entry(args.installed_desktop, environment, output, error, remember_child)
    if os.getpgid(process.pid) != process.pid or os.getsid(process.pid) != process.pid:
        raise RuntimeError("Gio desktop child did not enter its own session/process group")
    deadline = time.monotonic() + 5
    actual_executable = Path(f"/proc/{process.pid}/exe")
    while time.monotonic() < deadline and process.poll() is None:
        if actual_executable.resolve() == Path("/usr/bin/hopnode-desktop"):
            report["installed_desktop_exec_verified"] = True
            return process
        time.sleep(0.05)
    raise RuntimeError("Gio child did not execute the actual installed HopNode binary")


def query():
    response = run([str(core_binary), "status", "--json"], env=core_environment)
    if response.returncode:
        raise RuntimeError("Independent Core CLI query failed; see private local logs")
    return json.loads(response.stdout)


def audit_events(gui_pid):
    try:
        lines = (stage / "ipc-audit.jsonl").read_text(encoding="utf-8").splitlines()
    except FileNotFoundError:
        return []
    parsed = []
    for line in lines:
        try:
            event = json.loads(line)
        except json.JSONDecodeError:
            continue
        if event.get("peer_identity_verified") and event.get("peer_pid") == gui_pid and event.get("peer_uid") == uid and event.get("method") == "GET" and event.get("path") == "/v0/status" and event.get("http_status") == 200:
            parsed.append(event)
    return parsed


def group_exists(group_id):
    try:
        os.killpg(group_id, 0)
        return True
    except ProcessLookupError:
        return False
    except PermissionError:
        return True


def group_members(group_id):
    members = []
    for entry in Path("/proc").glob("[0-9]*/stat"):
        try:
            fields = entry.read_text().rpartition(")")[2].split()
            if int(fields[2]) == group_id:
                members.append({"pid": int(entry.parent.name), "uid": entry.parent.stat().st_uid, "session": int(fields[3]), "state": fields[0]})
        except (FileNotFoundError, ProcessLookupError, PermissionError, ValueError, IndexError):
            continue
    return members


def signal_owned_group(group_id, selected_signal):
    members = group_members(group_id)
    if not members:
        if group_exists(group_id):
            raise RuntimeError("Owned process group still exists but membership cannot be verified")
        return
    if any(member["uid"] != uid or member["session"] != group_id for member in members):
        raise RuntimeError("Refusing to signal a process group with unexpected UID/session")
    try:
        os.killpg(group_id, selected_signal)
    except ProcessLookupError:
        pass


def wait_group_gone(group_id, seconds):
    deadline = time.monotonic() + seconds
    while time.monotonic() < deadline:
        if not group_exists(group_id) and not group_members(group_id):
            return True
        time.sleep(0.1)
    return not group_exists(group_id) and not group_members(group_id)


def stop_owned(process):
    # New-session groups contain only this launch and its WebKit descendants.
    group_id = process.pid
    group_cleanup[str(group_id)] = {"gone": False}
    signal_owned_group(group_id, signal.SIGTERM)
    if process.poll() is None:
        try:
            process.wait(timeout=8)
        except subprocess.TimeoutExpired:
            signal_owned_group(group_id, signal.SIGKILL)
            process.wait(timeout=5)
    gone = wait_group_gone(group_id, 3)
    if not gone:
        signal_owned_group(group_id, signal.SIGKILL)
        gone = wait_group_gone(group_id, 5)
    group_cleanup[str(group_id)] = {"gone": gone, "remaining_states": sorted({member["state"] for member in group_members(group_id)})}
    if not gone:
        raise RuntimeError("Owned process group has remaining members; stage will be retained")


def interrupted(signum, frame):
    raise InterruptedError("Owned desktop probe interrupted; cleanup follows")


signal.signal(signal.SIGTERM, interrupted)
signal.signal(signal.SIGINT, interrupted)


try:
    extracted = run(["dpkg-deb", "-x", str(args.deb.resolve()), str(stage / "extracted")], timeout=30)
    if extracted.returncode:
        raise RuntimeError("dpkg-deb extraction failed; no package installation was attempted")
    candidates = [path for path in (stage / "extracted").rglob("hopnode-desktop") if path.is_file() and os.access(path, os.X_OK)]
    if len(candidates) != 1:
        raise RuntimeError("Expected exactly one extracted HopNode desktop executable")
    if not candidates[0].resolve().is_relative_to((stage / "extracted").resolve()):
        raise RuntimeError("Extracted desktop executable resolves outside the private stage")
    entry_module = None
    if args.installed_desktop:
        spec = importlib.util.spec_from_file_location("hopnode_desktop_entry", Path(__file__).with_name("linux-desktop-entry.py"))
        entry_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(entry_module)
        report["installed_payload_sha256"] = entry_module.verify_installed_entry(args.installed_desktop, stage / "extracted")
        report["gui_launch_method"] = "Gio.DesktopAppInfo installed desktop entry"
    shutil.copyfile(args.core, core_binary)
    core_binary.chmod(0o700)
    if unit:
        command = ["systemd-run", "--user", "--unit=" + unit, "--collect", "--property=RuntimeMaxSec=180", "--property=KillMode=control-group", "--setenv=HOPNODE_POC_ENDPOINT=" + environment["HOPNODE_POC_ENDPOINT"], "--setenv=HOPNODE_POC_AUDIT_FILE=" + core_environment["HOPNODE_POC_AUDIT_FILE"], "/usr/bin/env", "-u", "DISPLAY", "-u", "WAYLAND_DISPLAY", "-u", "XAUTHORITY", str(core_binary), "serve"]
        started = run(command)
        if started.returncode:
            raise RuntimeError("Isolated user transient unit could not start; no system unit was changed")
        unit_started = True
        core_pid = int(run(["systemctl", "--user", "show", unit, "-p", "MainPID", "--value"]).stdout.strip())
        report["core_unit"] = unit
    else:
        core_process = launch([str(core_binary), "serve"], core_environment, "core")
        core_pid = core_process.pid
    deadline = time.monotonic() + 10
    first_status = None
    while time.monotonic() < deadline:
        try:
            first_status = query()
            break
        except (RuntimeError, json.JSONDecodeError):
            time.sleep(0.2)
    if not first_status or first_status["pid"] != core_pid:
        raise RuntimeError("Headless CLI did not identify the independently started Core")
    report["checks"]["headless_cli_before_gui"] = "PASS"
    report["checks"]["uds_mode_0600"] = "PASS" if stat.S_IMODE((stage / "runtime/core.sock").stat().st_mode) == 0o600 else "FAIL"
    gui = launch_installed_gui(entry_module) if entry_module else launch([str(candidates[0])], environment, "gui")
    report["gui_pid"] = gui.pid
    report["core_pid"] = core_pid
    deadline = time.monotonic() + 30
    while time.monotonic() < deadline and gui.poll() is None and not audit_events(gui.pid):
        time.sleep(0.25)
    if not audit_events(gui.pid):
        raise RuntimeError("No kernel-verified status request from the actual GUI PID; process existence is insufficient")
    report["checks"]["gui_to_real_uds"] = "PASS"
    before_refresh = len(audit_events(gui.pid))
    helper = Path(__file__).with_name("linux-atspi-own-window.py")
    try:
        run(["python3", str(helper), "--pid", str(gui.pid), "--core-pid", str(core_pid), "--report", str(stage / "ui-accessibility.json")], timeout=30)
    except subprocess.TimeoutExpired:
        report["ui_helper_timeout"] = True
    if (stage / "ui-accessibility.json").exists():
        ui_report = json.loads((stage / "ui-accessibility.json").read_text(encoding="utf-8"))
    else:
        ui_report = {"status": "NOT_RUN", "refresh_clicked": False, "reason": "Owned application accessibility helper unavailable"}
    report["ui_accessibility"] = ui_report
    report["checks"]["visible_ui_and_refresh"] = "NOT_RUN"
    if ui_report.get("refresh_clicked"):
        deadline = time.monotonic() + 5
        while time.monotonic() < deadline and len(audit_events(gui.pid)) <= before_refresh:
            time.sleep(0.2)
        report["checks"]["visible_ui_and_refresh"] = "PASS" if len(audit_events(gui.pid)) > before_refresh else "FAIL"
    stop_owned(gui)
    after_gui = query()
    if after_gui["pid"] != core_pid or after_gui["instance_id"] != first_status["instance_id"]:
        raise RuntimeError("Independent Core instance changed when GUI exited")
    report["checks"]["same_core_after_gui_exit_without_display"] = "PASS"
    report["status"] = "FAIL" if "FAIL" in report["checks"].values() else ("PARTIAL" if "NOT_RUN" in report["checks"].values() else "PASS")
except Exception as error:
    report["error"] = type(error).__name__ + ": " + str(error)
finally:
    for process in reversed(processes):
        try:
            stop_owned(process)
        except Exception as error:
            report["cleanup"]["process_error"] = type(error).__name__
    if unit_started:
        try:
            stopped = run(["systemctl", "--user", "stop", unit])
            state_result = run(["systemctl", "--user", "show", unit, "-p", "MainPID", "-p", "ActiveState", "-p", "LoadState"])
            unit_state = dict(line.split("=", 1) for line in state_result.stdout.splitlines() if "=" in line)
            report["cleanup"]["user_unit_state"] = {key: unit_state.get(key, "") for key in ("MainPID", "ActiveState", "LoadState")}
            report["cleanup"]["owned_user_unit_stopped"] = (unit_state.get("LoadState") == "not-found" and unit_state.get("MainPID", "0") == "0") or (unit_state.get("MainPID") == "0" and unit_state.get("ActiveState") in ("inactive", "failed"))
        except Exception as error:
            report["cleanup"]["owned_user_unit_stopped"] = False
            report["cleanup"]["user_unit_error"] = type(error).__name__
    report["cleanup"]["process_groups"] = group_cleanup
    groups_gone = all(group_cleanup.get(str(process.pid), {}).get("gone") for process in processes)
    report["cleanup"]["all_owned_process_groups_gone"] = groups_gone
    safe_to_remove = groups_gone and report["cleanup"].get("owned_user_unit_stopped", True)
    for stream in streams:
        stream.close()
    for name in ("runtime", "bin", "config", "data", "cache", "state", "extracted"):
        target = stage / name
        if not safe_to_remove:
            report["cleanup"][name] = "RETAINED_UNVERIFIED_PROCESS"
            continue
        # Delete only these script-created children of the canonical private stage.
        if target.is_symlink() or target.resolve().parent != stage.resolve():
            report["cleanup"][name] = "REFUSED"
            continue
        try:
            shutil.rmtree(target)
            report["cleanup"][name] = "REMOVED"
        except FileNotFoundError:
            report["cleanup"][name] = "ABSENT"
        except OSError as error:
            report["cleanup"][name] = "ERROR: " + type(error).__name__
    if any(value is False or (isinstance(value, str) and (value == "REFUSED" or value.startswith("ERROR"))) for value in report["cleanup"].values()) or "process_error" in report["cleanup"]:
        report["status"] = "FAIL"
    with open(stage / "report.json", "x", encoding="utf-8") as output:
        json.dump(report, output, indent=2, ensure_ascii=False)
    # Deliberately print no session variables, UI text, username or source network address.
    print(json.dumps({"status": report["status"], "checks": report["checks"], "evidence_directory": str(stage), "cleanup": report["cleanup"]}, sort_keys=True))
raise SystemExit(0 if report["status"] in ("PASS", "PARTIAL") else 1)
