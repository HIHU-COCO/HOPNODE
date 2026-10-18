"""文件用途：P01 只读、脱敏的网络前置检查。创建者：HIHU-COCO；日期：2026-10-19。"""
import argparse
import json
import shutil
import subprocess
import sys


def summarize_status(raw):
    if not isinstance(raw, dict):
        raise ValueError("status must be an object")
    peers = raw.get("Peer") or {}
    self_node = raw.get("Self") or {}
    if not isinstance(peers, dict) or not isinstance(self_node, dict):
        raise ValueError("invalid status fields")
    backend = raw.get("BackendState", "Unknown")
    if backend not in {"NoState", "NeedsLogin", "NeedsMachineAuth", "Stopped", "Starting", "Running", "InUseOtherUser"}:
        backend = "Unknown"
    return {
        "backend_state": backend,
        "self_online": self_node.get("Online") is True,
        "peer_count": len(peers),
        "connection_type": "unknown",
        "cross_network_verified": False,
        "derp_verified": False,
    }


def main(argv=None):
    parser = argparse.ArgumentParser(description="P01 read-only preflight. Does not change the existing network.")
    parser.add_argument("--tailscale", help="explicit Tailscale executable path")
    args = parser.parse_args(argv)
    executable = args.tailscale or shutil.which("tailscale")
    if not executable:
        print(json.dumps({"status": "BLOCKED", "code": "TAILSCALE_NOT_FOUND"}))
        return 2
    try:
        result = subprocess.run([executable, "status", "--json"], capture_output=True, text=True, timeout=15, check=False)
        if result.returncode:
            print(json.dumps({"status": "BLOCKED", "code": "TAILSCALE_STATUS_UNAVAILABLE"}))
            return 2
        summary = summarize_status(json.loads(result.stdout))
    except (OSError, subprocess.TimeoutExpired, ValueError):
        print(json.dumps({"status": "BLOCKED", "code": "TAILSCALE_STATUS_INVALID_OR_UNAVAILABLE"}))
        return 2
    print(json.dumps({"status": "OBSERVED", "data": summary}))
    return 0


if __name__ == "__main__":
    sys.exit(main())
