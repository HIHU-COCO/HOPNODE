#!/usr/bin/env python3
"""Inject systemd parent-command failures without creating real units.
Purpose: P04 ambiguous-start cleanup regression. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import importlib.util
from pathlib import Path
import subprocess
import sys
import unittest

sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location("hopnode_user_unit", Path(__file__).with_name("linux-user-unit.py"))
units = importlib.util.module_from_spec(spec)
spec.loader.exec_module(units)
UNIT = "hopnode-p04-" + "a" * 32 + ".service"


class AmbiguousUnitStart(unittest.TestCase):
    def exercise(self, *, stop_fails=False, query_fails=False):
        state = {"active": False, "calls": []}
        controller = None

        def run(command):
            state["calls"].append(command)
            if command[0] == "systemd-run":
                self.assertTrue(controller.start_attempted)
                state["active"] = True
                raise subprocess.TimeoutExpired(command, 15)
            self.assertEqual(command[0:2], ["systemctl", "--user"])
            if command[2] == "stop":
                self.assertEqual(command[3], UNIT)
                if stop_fails:
                    raise OSError("injected stop failure")
                state["active"] = False
                return subprocess.CompletedProcess(command, 0, "", "")
            self.assertEqual(command[2:4], ["show", UNIT])
            if query_fails:
                raise subprocess.TimeoutExpired(command, 15)
            output = "MainPID=123\nActiveState=active\nLoadState=loaded\n" if state["active"] else "MainPID=0\nActiveState=inactive\nLoadState=not-found\n"
            return subprocess.CompletedProcess(command, 0, output, "")

        controller = units.OwnedUserUnit(UNIT, run)
        with self.assertRaises(subprocess.TimeoutExpired):
            try:
                controller.start(["systemd-run", "--user", "--unit=" + UNIT])
            finally:
                cleanup = controller.stop_and_verify()
        self.assertEqual([call[2] for call in state["calls"][1:]], ["stop", "show"])
        return cleanup, state

    def test_created_unit_is_stopped_even_when_launch_parent_times_out(self):
        cleanup, state = self.exercise()
        self.assertFalse(state["active"])
        self.assertTrue(cleanup["owned_user_unit_stopped"])
        self.assertEqual(cleanup["user_unit_state"]["MainPID"], "0")

    def test_active_unit_after_failed_stop_cannot_authorize_stage_cleanup(self):
        cleanup, state = self.exercise(stop_fails=True)
        self.assertTrue(state["active"])
        self.assertFalse(cleanup["owned_user_unit_stopped"])
        self.assertEqual(cleanup["user_unit_state"]["ActiveState"], "active")

    def test_unverifiable_state_after_stop_cannot_authorize_stage_cleanup(self):
        cleanup, state = self.exercise(query_fails=True)
        self.assertFalse(state["active"])
        self.assertFalse(cleanup["owned_user_unit_stopped"])
        self.assertIn("user_unit_query_error", cleanup)


if __name__ == "__main__":
    unittest.main(verbosity=2)
