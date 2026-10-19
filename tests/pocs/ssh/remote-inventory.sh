#!/usr/bin/env bash
# Read-only P02 inventory. Author: HIHU-COCO, 2026-10-19.
# Run as the already-authorized SSH user. No sudo, writes, service changes,
# hostnames, addresses, user names, environment dumps, or configuration secrets.
set -u
set -o pipefail
export LC_ALL=C
export PATH=/usr/sbin:/usr/bin:/sbin:/bin

emit() { printf '%s=%s\n' "$1" "$2"; }
token() {
  if [[ "$2" =~ ^[A-Za-z0-9._+:~/-]{1,160}$ ]]; then emit "$1" "$2"; else emit "$1" unknown; fi
}
available() { if command -v "$2" >/dev/null 2>&1; then emit "$1" yes; else emit "$1" no; fi; }
flag_set() { if [[ -n "$2" ]]; then emit "$1" set; else emit "$1" unset; fi; }
os_field() {
  task_value=$(awk -F= -v field="$2" '$1==field {gsub(/"/,"",$2); print $2; exit}' /etc/os-release 2>/dev/null || true)
  token "$1" "$task_value"
}

emit format hopnode-p02-inventory-v1
os_field os.id ID
os_field os.version_id VERSION_ID
os_field os.codename VERSION_CODENAME
token os.arch "$(uname -m 2>/dev/null || true)"
token os.kernel "$(uname -r 2>/dev/null || true)"
token os.long_bit "$(getconf LONG_BIT 2>/dev/null || true)"
token identity.uid "$(id -u 2>/dev/null || true)"
token identity.gid "$(id -g 2>/dev/null || true)"
if [[ ${EUID} == 0 ]]; then emit identity.is_root yes; else emit identity.is_root no; fi
task_groups=" $(id -Gn 2>/dev/null || true) "
case "$task_groups" in *' sudo '*|*' wheel '*) emit identity.admin_group_member yes ;; *) emit identity.admin_group_member no ;; esac
emit privilege.sudo_execution NOT_RUN
emit privilege.service_mutation NOT_RUN
for task_location in tmp var_tmp home; do
  case "$task_location" in tmp) task_path=/tmp ;; var_tmp) task_path=/var/tmp ;; home) task_path=${HOME-} ;; esac
  if [[ -n "$task_path" && -d "$task_path" && -w "$task_path" && -x "$task_path" ]]; then
    emit "permission.$task_location.access_writable" yes
  else
    emit "permission.$task_location.access_writable" no
  fi
done
emit permission.write_probe NOT_RUN

emit binary.search_scope system_path_only
for task_binary in bash ssh sshd dpkg-query systemctl loginctl pgrep mktemp stat rmdir tty stty sleep nohup setsid sha256sum gcc make pkg-config node pnpm rustc cargo go codex tailscale hopnodectl hopnode-core; do
  available "binary.$task_binary.present" "$task_binary"
done

case "${XDG_SESSION_TYPE-}" in x11|wayland|tty) emit gui.ssh_session_type "$XDG_SESSION_TYPE" ;; '') emit gui.ssh_session_type unset ;; *) emit gui.ssh_session_type other ;; esac
flag_set gui.display_variable "${DISPLAY-}"
flag_set gui.wayland_variable "${WAYLAND_DISPLAY-}"
flag_set gui.runtime_directory_variable "${XDG_RUNTIME_DIR-}"
task_desktop=${XDG_CURRENT_DESKTOP-}
case "${task_desktop,,}" in *gnome*) emit gui.ssh_desktop_hint gnome ;; *kde*) emit gui.ssh_desktop_hint kde ;; *xfce*) emit gui.ssh_desktop_hint xfce ;; '') emit gui.ssh_desktop_hint unset ;; *) emit gui.ssh_desktop_hint other ;; esac
shopt -s nullglob
task_sessions=(/usr/share/xsessions/*.desktop)
emit gui.x11_session_entry_count "${#task_sessions[@]}"
task_sessions=(/usr/share/wayland-sessions/*.desktop)
emit gui.wayland_session_entry_count "${#task_sessions[@]}"
for task_process in gnome-shell plasmashell Xorg Xwayland; do
  if command -v pgrep >/dev/null 2>&1; then
    if pgrep -x "$task_process" >/dev/null 2>&1; then emit "gui.process.$task_process.present" yes; else emit "gui.process.$task_process.present" no; fi
  else
    emit "gui.process.$task_process.present" unknown
  fi
done
emit gui.interactive_access NOT_RUN

task_pid1=$(cat /proc/1/comm 2>/dev/null || true)
if [[ "$task_pid1" == systemd ]]; then emit systemd.is_pid1 yes; else emit systemd.is_pid1 no; fi
if command -v systemctl >/dev/null 2>&1; then
  task_system_state=$(systemctl is-system-running 2>/dev/null || true)
  case "$task_system_state" in initializing|starting|running|degraded|maintenance|stopping|offline|unknown) emit systemd.system_state "$task_system_state" ;; *) emit systemd.system_state unknown ;; esac
  for task_unit in ssh.service tailscaled.service display-manager.service hopnode-core.service hopnode-agent.service; do
    task_unit_info=$(systemctl show --no-pager -p LoadState -p ActiveState -p SubState "$task_unit" 2>/dev/null || true)
    for task_property in LoadState ActiveState SubState; do
      task_value=$(printf '%s\n' "$task_unit_info" | awk -F= -v property="$task_property" '$1==property {print $2; exit}')
      token "service.$task_unit.$task_property" "$task_value"
    done
  done
else
  emit systemd.system_state unknown
fi

for task_package in openssh-server openssh-client systemd dbus sudo libc6 libgtk-3-0 libwebkit2gtk-4.1-0 libayatana-appindicator3-1 libssl3 gnome-shell gdm3 tailscale build-essential pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev libayatana-appindicator3-dev librsvg2-dev patchelf; do
  if ! command -v dpkg-query >/dev/null 2>&1; then
    emit "package.$task_package.installed" unknown
    continue
  fi
  if task_package_info=$(dpkg-query -W -f='${Status}\t${Version}\t${Architecture}' "$task_package" 2>/dev/null); then
    IFS=$'\t' read -r task_status task_version task_arch <<< "$task_package_info"
    if [[ "$task_status" == 'install ok installed' ]]; then
      emit "package.$task_package.installed" yes
      token "package.$task_package.version" "$task_version"
      token "package.$task_package.arch" "$task_arch"
    else
      emit "package.$task_package.installed" no
    fi
  else
    emit "package.$task_package.installed" no
  fi
done
emit package.installation NOT_RUN
emit inventory.completed yes
