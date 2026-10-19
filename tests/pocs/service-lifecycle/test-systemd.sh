#!/usr/bin/env bash
# Purpose: isolated transient systemd lifecycle proof. Creator: HIHU-COCO. Created: 2026-10-19.
set -euo pipefail
umask 077

if [[ $# -ne 1 || $EUID -ne 0 ]]; then
    echo 'BLOCKED: pass one absolute desktop-service POC binary path in an authorized root session' >&2
    exit 2
fi
core=$(realpath -- "$1")
[[ $core = /* && -f $core && -x $core ]] || { echo 'INVALID_POC_BINARY' >&2; exit 2; }
systemctl is-system-running --quiet || { echo 'BLOCKED: systemd is not running' >&2; exit 2; }
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
task_id=$(cat /proc/sys/kernel/random/uuid)
runtime_name="hopnode-p04-lifecycle-${task_id}"
unit="${runtime_name}.service"
endpoint="/run/${runtime_name}/core.sock"
mkdir -p -- "$script_dir/artifacts"
evidence=$(mktemp -d "$script_dir/artifacts/systemd.XXXXXXXX")
unit_owned=0
latest_snapshot=''

cleanup() {
    local result=$? remaining_pid
    trap - EXIT
    if [[ $unit_owned = 1 ]]; then
        # Only this unpredictable task-owned name is ever stopped/reset.
        systemctl stop "$unit" >/dev/null 2>&1 || true
        remaining_pid=$(systemctl show "$unit" --property=MainPID --value 2>/dev/null || true)
        if [[ -n $remaining_pid && $remaining_pid != 0 ]]; then
            echo 'CLEANUP_MAINPID_REMAINS=FAIL' >&2
            result=1
        fi
        if [[ -n $latest_snapshot ]]; then
            if ! python3 "$script_dir/verify-systemd.py" stopped "$core" "$endpoint" "$latest_snapshot" "$evidence/cleanup.json"; then
                result=1
            fi
        fi
        systemctl reset-failed "$unit" >/dev/null 2>&1 || true
    fi
    echo "SANITIZED_ARTIFACTS=$evidence"
    if [[ $result = 0 ]]; then
        echo 'SYSTEMD_LIFECYCLE_POC=PASS'
    else
        echo 'SYSTEMD_LIFECYCLE_POC=FAIL' >&2
    fi
    exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
python3 "$script_dir/verify-systemd.py" metadata "$core" "$unit" "$endpoint" "$evidence/metadata.json"

start_unit() {
    local load_state
    load_state=$(systemctl show "$unit" --property=LoadState --value 2>/dev/null || true)
    if [[ $load_state = loaded ]]; then
        systemctl start "$unit"
    else
        # RuntimeDirectory belongs solely to this unit; systemd removes it on stop.
        systemd-run --quiet --unit="$unit" --service-type=exec \
            --property="RuntimeDirectory=$runtime_name" \
            --property=RuntimeDirectoryMode=0700 \
            --property=RuntimeDirectoryPreserve=no \
            --property=KillMode=control-group --property=KillSignal=SIGINT \
            --property=TimeoutStopSec=10s --property=SendSIGKILL=yes \
            --property=NoNewPrivileges=yes --property=UMask=0077 \
            --property=StandardOutput=null --property=StandardError=null \
            --setenv="HOPNODE_POC_ENDPOINT=$endpoint" \
            "$core" serve
    fi
}

wait_running() {
    local pid attempt
    for attempt in {1..100}; do
        pid=$(systemctl show "$unit" --property=MainPID --value 2>/dev/null || true)
        if [[ $pid =~ ^[1-9][0-9]*$ && -S $endpoint ]]; then
            printf '%s' "$pid"
            return
        fi
        sleep 0.05
    done
    echo 'SERVICE_READINESS_TIMEOUT' >&2
    return 1
}

# Refuse accidental takeover of even a similarly named pre-existing unit.
[[ $(systemctl show "$unit" --property=LoadState --value 2>/dev/null || true) = not-found ]] || { echo 'UNIT_COLLISION' >&2; exit 1; }
unit_owned=1
start_unit
pid_first=$(wait_running)
python3 "$script_dir/verify-systemd.py" running "$core" "$endpoint" "$pid_first" "$evidence/first.json"
latest_snapshot="$evidence/first.json"
echo 'TRANSIENT_DIRECT_EXEC_START=PASS'

systemctl stop "$unit"
python3 "$script_dir/verify-systemd.py" stopped "$core" "$endpoint" "$latest_snapshot" "$evidence/stopped.json"

# A stopped transient unit may have been garbage-collected. Recreate its same
# task-owned definition when necessary; also exercise native restart below.
start_unit
pid_second=$(wait_running)
python3 "$script_dir/verify-systemd.py" running "$core" "$endpoint" "$pid_second" "$evidence/second.json"
latest_snapshot="$evidence/second.json"
python3 "$script_dir/verify-systemd.py" different "$evidence/first.json" "$latest_snapshot"

systemctl restart "$unit"
pid_third=$(wait_running)
python3 "$script_dir/verify-systemd.py" running "$core" "$endpoint" "$pid_third" "$evidence/restarted.json"
latest_snapshot="$evidence/restarted.json"
python3 "$script_dir/verify-systemd.py" different "$evidence/second.json" "$latest_snapshot"
echo 'SYSTEMCTL_RESTART=PASS'
