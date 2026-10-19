#!/usr/bin/env python3
"""Launch only the selected desktop entry with a new, owned Linux session.
Purpose: P04 installed-menu evidence, never package installation. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import hashlib
import os
from pathlib import Path
import stat
import subprocess
import time


class DesktopEntryChild:
    """Reap the direct child returned by Gio's DO_NOT_REAP_CHILD callback."""

    def __init__(self, pid):
        self.pid = pid
        self.returncode = None

    def poll(self):
        if self.returncode is None:
            waited, status = os.waitpid(self.pid, os.WNOHANG)
            if waited:
                self.returncode = os.waitstatus_to_exitcode(status)
        return self.returncode

    def wait(self, timeout):
        deadline = time.monotonic() + timeout
        while self.poll() is None:
            if time.monotonic() >= deadline:
                raise subprocess.TimeoutExpired("owned Gio desktop entry", timeout)
            time.sleep(0.05)
        return self.returncode


def verify_installed_entry(entry_path, extracted_root):
    """Accept only the known, root-owned payload matching this exact deb."""
    expected_entry = Path("/usr/share/applications/hopnode-desktop-poc.desktop")
    executable = Path("/usr/bin/hopnode-desktop")
    if Path(entry_path) != expected_entry:
        raise RuntimeError("Installed mode accepts only the known HopNode desktop entry")
    hashes = {}
    for key, installed in (("desktop_entry", expected_entry), ("executable", executable)):
        if installed.resolve() != installed or not installed.is_file():
            raise RuntimeError("Installed HopNode payload must be a regular file without symlink parents")
        for ancestor in (installed, *installed.parents):
            info = ancestor.stat()
            if info.st_uid != 0 or stat.S_IMODE(info.st_mode) & 0o022:
                raise RuntimeError("Installed HopNode payload or parent is not protected and root-owned")
        packaged = Path(extracted_root) / installed.relative_to("/")
        digest = hashlib.sha256(installed.read_bytes()).hexdigest()
        if not packaged.is_file() or digest != hashlib.sha256(packaged.read_bytes()).hexdigest():
            raise RuntimeError("Installed HopNode payload differs from the supplied deb")
        hashes[key] = digest
    if not os.access(executable, os.X_OK):
        raise RuntimeError("Installed HopNode binary is not executable")
    from gi.repository import Gio
    entry = Gio.DesktopAppInfo.new_from_filename(str(expected_entry))
    if not entry or entry.get_string("Type") != "Application" or entry.get_string("Exec") != "hopnode-desktop":
        raise RuntimeError("Installed desktop entry has an unexpected application command")
    if any(entry.get_boolean(key) for key in ("Terminal", "DBusActivatable", "Hidden", "NoDisplay")):
        raise RuntimeError("Installed entry must launch directly without terminal or DBus activation")
    return hashes


def launch_desktop_entry(entry_path, environment, output, error, on_child=None):
    """Use the actual desktop entry; callbacks provide the child PID and setsid."""
    from gi.repository import Gio, GLib
    entry = Gio.DesktopAppInfo.new_from_filename(str(entry_path))
    if entry is None:
        raise RuntimeError("Gio could not read the selected desktop entry")
    context = Gio.AppLaunchContext()
    for inherited in context.get_environment():
        context.unsetenv(inherited.split("=", 1)[0])
    for key, value in environment.items():
        context.setenv(key, value)
    context.setenv("PATH", "/usr/bin:/bin")
    children = []

    def new_session(*unused):
        os.setsid()

    def child_pid(app, pid, *unused):
        child = DesktopEntryChild(int(pid))
        children.append(child)
        if on_child is not None:
            on_child(child)

    launched = entry.launch_uris_as_manager_with_fds(
        [], context, GLib.SpawnFlags.DO_NOT_REAP_CHILD | GLib.SpawnFlags.SEARCH_PATH_FROM_ENVP,
        new_session, None, child_pid, None, -1, output.fileno(), error.fileno(),
    )
    if not launched or len(children) != 1:
        raise RuntimeError("Gio did not return exactly one owned desktop-entry child")
    return children[0]
