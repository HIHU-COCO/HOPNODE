#!/usr/bin/env python3
"""Read/click only the owned P04 application's accessibility tree.
Purpose: P04 optional genuine UI evidence. Creator: HIHU-COCO. Created: 2026-10-19.
"""
import argparse
import json
import os
import re
import time

parser = argparse.ArgumentParser()
parser.add_argument("--pid", required=True, type=int)
parser.add_argument("--core-pid", required=True, type=int)
parser.add_argument("--report", required=True)
args = parser.parse_args()
result = {"status": "NOT_RUN", "reason": "AT-SPI unavailable", "refresh_clicked": False}


def save():
    descriptor = os.open(args.report, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    with os.fdopen(descriptor, "w", encoding="utf-8") as output:
        json.dump(result, output, ensure_ascii=False)


try:
    import gi
    gi.require_version("Atspi", "2.0")
    from gi.repository import Atspi

    desktop = Atspi.get_desktop(0)
    application = None
    deadline = time.monotonic() + 12
    while application is None and time.monotonic() < deadline:
        for index in range(desktop.get_child_count()):
            candidate = desktop.get_child_at_index(index)
            if candidate and candidate.get_process_id() == args.pid:
                application = candidate
                break
        if application is None:
            time.sleep(0.25)
    if application is None:
        raise RuntimeError("Owned GUI PID has no accessible application; other applications were not inspected")

    def inspect_owned():
        pending = [(application, 0)]
        words, button = [], None
        visited = 0
        while pending and visited < 800:
            node, depth = pending.pop()
            visited += 1
            try:
                name = node.get_name() or ""
                words.append(name)
                if name == "刷新状态":
                    button = node
                try:
                    text = node.get_text_iface()
                    if text:
                        words.append(text.get_text(0, min(text.get_character_count(), 4096)))
                except Exception:
                    pass
                if depth < 20:
                    for index in range(min(node.get_child_count(), 200)):
                        child = node.get_child_at_index(index)
                        if child:
                            pending.append((child, depth + 1))
            except Exception:
                continue
        return "\n".join(words), button

    text, refresh = "", None
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline:
        text, refresh = inspect_owned()
        if refresh and "已连接" in text and re.search(r"\b" + str(args.core_pid) + r"\b", text):
            break
        time.sleep(0.25)
    result["owned_gui_pid"] = args.pid
    result["connected_text_visible"] = "已连接" in text
    result["core_pid_text_visible"] = bool(re.search(r"\b" + str(args.core_pid) + r"\b", text))
    if not refresh or not result["connected_text_visible"] or not result["core_pid_text_visible"]:
        raise RuntimeError("Owned UI text or refresh action was not exposed; IPC audit remains independent evidence")
    action = refresh.get_action_iface()
    if not action or action.get_n_actions() < 1 or not action.do_action(0):
        raise RuntimeError("Owned refresh button did not expose a usable action")
    result.update(status="PASS", reason="Owned GUI connected text/PID read and real refresh action invoked", refresh_clicked=True)
except Exception as error:
    # Do not include application text, user identity or session environment in reports.
    result["reason"] = type(error).__name__ + ": " + str(error)
finally:
    save()
