#!/usr/bin/env python3
"""Account for an owned transient unit before a potentially ambiguous launch.
Purpose: P04 failure-safe user-unit cleanup. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import re


class OwnedUserUnit:
    def __init__(self, unit, run):
        if not re.fullmatch(r"hopnode-p04-[0-9a-f]{32}\.service", unit):
            raise ValueError("Only a unique P04-owned transient user unit is accepted")
        self.unit = unit
        self.run = run
        self.start_attempted = False

    def start(self, command):
        # systemd may create the unit even if the parent command times out/raises.
        self.start_attempted = True
        result = self.run(command)
        if result.returncode:
            raise RuntimeError("Isolated user transient unit could not start; cleanup still follows")
        return result

    def stop_and_verify(self):
        result = {"owned_user_unit_stopped": False}
        try:
            self.run(["systemctl", "--user", "stop", self.unit])
        except Exception as error:
            result["user_unit_stop_error"] = type(error).__name__
        # Even a failed/timed-out stop must be followed by an independent read.
        try:
            state_result = self.run(["systemctl", "--user", "show", self.unit, "-p", "MainPID", "-p", "ActiveState", "-p", "LoadState"])
            state = dict(line.split("=", 1) for line in state_result.stdout.splitlines() if "=" in line)
            result["user_unit_state"] = {key: state.get(key, "") for key in ("MainPID", "ActiveState", "LoadState")}
            result["owned_user_unit_stopped"] = (state.get("LoadState") == "not-found" and state.get("MainPID", "0") == "0") or (state.get("MainPID") == "0" and state.get("ActiveState") in ("inactive", "failed"))
        except Exception as error:
            result["user_unit_query_error"] = type(error).__name__
        return result
