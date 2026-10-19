#!/usr/bin/env python3
"""Exercise actual Gio child ownership with a private sleep desktop fixture.
Purpose: P04 launcher lifecycle regression, not GUI acceptance. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import importlib.util
import os
from pathlib import Path
import signal
import sys
import tempfile
import time
import unittest

sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("hopnode_desktop_entry", Path(__file__).with_name("linux-desktop-entry.py"))
launcher = importlib.util.module_from_spec(spec)
spec.loader.exec_module(launcher)


class DesktopEntryOwnership(unittest.TestCase):
    def test_actual_gio_callback_session_environment_and_reap(self):
        with tempfile.TemporaryDirectory(prefix="hopnode-gio-owned-") as directory:
            stage = Path(directory)
            desktop = stage / "owned-sleep.desktop"
            desktop.write_text("[Desktop Entry]\nType=Application\nName=Owned P04 launcher fixture\nExec=/usr/bin/sleep 60\nTerminal=false\n", encoding="utf-8")
            child = None
            callbacks = []
            try:
                with open(stage / "stdout", "xb") as output, open(stage / "stderr", "xb") as error:
                    child = launcher.launch_desktop_entry(desktop, {"LANG": "C.UTF-8", "HOPNODE_POC_ENDPOINT": str(stage / "owned.sock")}, output, error, callbacks.append)
                self.assertEqual(callbacks, [child])
                self.assertEqual(os.getpgid(child.pid), child.pid)
                self.assertEqual(os.getsid(child.pid), child.pid)
                self.assertEqual(Path(f"/proc/{child.pid}").stat().st_uid, os.getuid())
                expected_endpoint = ("HOPNODE_POC_ENDPOINT=" + str(stage / "owned.sock")).encode()
                deadline = time.monotonic() + 5
                actual_environment = []
                while expected_endpoint not in actual_environment and time.monotonic() < deadline and child.poll() is None:
                    actual_environment = Path(f"/proc/{child.pid}/environ").read_bytes().split(b"\0")
                    time.sleep(0.02)
                self.assertIn(expected_endpoint, actual_environment)
                self.assertNotIn(b"DISPLAY", [item.split(b"=", 1)[0] for item in actual_environment])
                self.assertNotIn(b"SSH_CONNECTION", [item.split(b"=", 1)[0] for item in actual_environment])
                self.assertIsNone(child.poll())
            finally:
                if child is not None:
                    if child.poll() is None:
                        if os.getpgid(child.pid) != child.pid or os.getsid(child.pid) != child.pid:
                            raise RuntimeError("Fixture ownership changed; refusing group signal")
                        os.killpg(child.pid, signal.SIGTERM)
                        child.wait(timeout=5)
                    with self.assertRaises(ProcessLookupError):
                        os.killpg(child.pid, 0)
            self.assertEqual((stage / "stderr").stat().st_size, 0)


if __name__ == "__main__":
    unittest.main(verbosity=2)
