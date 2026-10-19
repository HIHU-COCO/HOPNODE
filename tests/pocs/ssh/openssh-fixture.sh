#!/usr/bin/env bash
# P02 isolated Ubuntu 22.04 OpenSSH fixture. Author: HIHU-COCO, 2026-10-19.
# Usage: sudo bash openssh-fixture.sh prepare /absolute/new-exchange-dir [port]
#        sudo bash openssh-fixture.sh run /absolute/exchange-dir
set -euo pipefail
umask 077

fail() { printf 'P02 fixture: %s\n' "$*" >&2; exit 1; }

[[ ${EUID} == 0 ]] || fail 'run as root on Ubuntu 22.04 amd64'
[[ $# -ge 2 && $# -le 3 ]] || fail 'expected prepare EXCHANGE [PORT] or run EXCHANGE'
task_action=$1
task_input=$2
[[ "$task_input" == /* && "$task_input" != / ]] || fail 'EXCHANGE must be an absolute directory path'
case "$task_action" in prepare|run) ;; *) fail 'unknown action' ;; esac
[[ "$task_action" == prepare || $# == 2 ]] || fail 'run reads its prepared port; do not pass a port'

for task_command in realpath stat ss chroot mount findmnt; do
  command -v "$task_command" >/dev/null || fail "missing prerequisite: $task_command"
done

check_port() {
  [[ "$task_port" =~ ^[0-9]{1,5}$ ]] || fail 'port must be numeric'
  task_port=$((10#$task_port))
  (( task_port >= 1024 && task_port <= 65535 )) || fail 'port must be between 1024 and 65535'
  task_listeners=$(ss -H -ltn "sport = :$task_port") || fail 'unable to inspect listening ports'
  [[ -z "$task_listeners" ]] || fail "port $task_port already has a listener; choose a different port"
}

validate_jail() {
  [[ "$task_jail" =~ ^/var/tmp/hopnode-p02-jail-[A-Za-z0-9]{8}$ ]] || fail 'unexpected jail path'
  [[ -d "$task_jail" && ! -L "$task_jail" ]] || fail 'jail must be an existing native directory'
  [[ "$(realpath -e -- "$task_jail")" == "$task_jail" ]] || fail 'jail path must not traverse symlinks'
  [[ "$(stat -c %u -- "$task_jail")" == 0 ]] || fail 'jail is not owned by root'
  [[ -f "$task_jail/.hopnode-p02-exchange" && ! -L "$task_jail/.hopnode-p02-exchange" ]] || fail 'jail marker missing'
  [[ "$(cat -- "$task_jail/.hopnode-p02-exchange")" == "$task_exchange" ]] || fail 'jail marker does not match EXCHANGE'
}

if [[ "$task_action" == run ]]; then
  task_exchange=$(realpath -e -- "$task_input")
  [[ -d "$task_exchange" ]] || fail 'EXCHANGE does not exist'
  [[ -f "$task_exchange/jail-path" && ! -L "$task_exchange/jail-path" ]] || fail 'jail-path missing or symlinked'
  [[ -f "$task_exchange/port" && ! -L "$task_exchange/port" ]] || fail 'port metadata missing or symlinked'
  task_jail=$(cat -- "$task_exchange/jail-path")
  task_port=$(cat -- "$task_exchange/port")
  validate_jail
  check_port
  [[ "$(findmnt -rn --mountpoint "$task_jail/dev/pts" -o FSTYPE)" == devpts ]] || fail 'dedicated devpts mount missing'
  chroot "$task_jail" /usr/sbin/sshd -t -f /etc/hopnode/sshd_config
  printf 'Foreground sshd on 127.0.0.1:%s; stop this foreground process with Ctrl+C.\n' "$task_port"
  exec chroot "$task_jail" /usr/sbin/sshd -D -e -f /etc/hopnode/sshd_config
fi

# Read only the system version metadata, never a user-provided shell file.
[[ -r /etc/os-release ]] || fail '/etc/os-release unavailable'
task_os=$(awk -F= '$1=="ID" {gsub(/"/,"",$2); print $2}' /etc/os-release)
task_release=$(awk -F= '$1=="VERSION_ID" {gsub(/"/,"",$2); print $2}' /etc/os-release)
[[ "$task_os" == ubuntu && "$task_release" == 22.04 ]] || fail 'this pinned fixture requires Ubuntu 22.04'
for task_command in apt-get dpkg dpkg-deb ssh-keygen ldd awk cp mktemp mknod chown sha256sum find; do
  command -v "$task_command" >/dev/null || fail "missing prerequisite: $task_command"
done
[[ "$(dpkg --print-architecture)" == amd64 ]] || fail 'this fixture requires amd64'
task_port=${3:-22476}
check_port

[[ ! -e "$task_input" && ! -L "$task_input" ]] || fail 'prepare requires a new EXCHANGE directory; existing data is preserved'
task_parent=$(realpath -e -- "$(dirname -- "$task_input")")
task_name=$(basename -- "$task_input")
[[ "$task_name" != . && "$task_name" != .. ]] || fail 'invalid EXCHANGE name'
task_exchange="$task_parent/$task_name"
mkdir -m 700 -- "$task_exchange"
task_jail=$(mktemp -d /var/tmp/hopnode-p02-jail-XXXXXXXX)
chmod 755 -- "$task_jail"
printf '%s\n' "$task_exchange" > "$task_jail/.hopnode-p02-exchange"
printf '%s\n' "$task_jail" > "$task_exchange/jail-path"
printf '%s\n' "$task_port" > "$task_exchange/port"
trap 'printf "Preparation failed; retain EXCHANGE and inspect its jail-path. A devpts mount may require the manual unmount steps in README.md.\n" >&2' ERR

task_download="$task_exchange/downloads"
task_package="$task_exchange/package"
mkdir -p -- "$task_download" "$task_package"
(
  cd -- "$task_download"
  # Downloads only: no apt install, service registration, or host account changes.
  apt-get download 'openssh-server=1:8.9p1-3ubuntu0.18' 'libwrap0=7.6.q-31build2'
)
task_server_count=0
task_wrap_count=0
for task_deb in "$task_download"/*.deb; do
  [[ -f "$task_deb" ]] || fail 'package download produced no .deb files'
  [[ "$(dpkg-deb -f "$task_deb" Architecture)" == amd64 ]] || fail 'unexpected package architecture'
  task_package_name=$(dpkg-deb -f "$task_deb" Package)
  task_package_version=$(dpkg-deb -f "$task_deb" Version)
  case "$task_package_name:$task_package_version" in
    openssh-server:1:8.9p1-3ubuntu0.18) task_server_count=$((task_server_count + 1)) ;;
    libwrap0:7.6.q-31build2) task_wrap_count=$((task_wrap_count + 1)) ;;
    *) fail 'unexpected package name or version' ;;
  esac
  dpkg-deb -x "$task_deb" "$task_package"
  sha256sum -- "$task_deb" >> "$task_exchange/package-sha256.txt"
done
[[ "$task_server_count" == 1 && "$task_wrap_count" == 1 ]] || fail 'expected exactly one fixed-version package of each type'

mkdir -p "$task_jail"/{etc/hopnode,run/sshd,home/fixture,workspace,dev/pts,lib/x86_64-linux-gnu,lib64,usr/sbin,bin,usr/bin}
for task_binary in "$task_package/usr/sbin/sshd" /bin/bash /usr/bin/whoami /usr/bin/id /usr/bin/sleep; do
  [[ -x "$task_binary" ]] || fail "binary unavailable: $task_binary"
  if [[ "$task_binary" == "$task_package/usr/sbin/sshd" ]]; then
    cp -p -- "$task_binary" "$task_jail/usr/sbin/sshd"
  else
    cp -p --parents -L -- "$task_binary" "$task_jail"
  fi
  task_libraries=$(LD_LIBRARY_PATH="$task_package/usr/lib/x86_64-linux-gnu:$task_package/lib/x86_64-linux-gnu" ldd "$task_binary")
  [[ "$task_libraries" != *'not found'* ]] || fail "unresolved shared library for $task_binary"
  while IFS= read -r task_library; do
    if [[ "$task_library" == "$task_package"/* ]]; then
      cp -p -L -- "$task_library" "$task_jail/lib/x86_64-linux-gnu/$(basename -- "$task_library")"
    else
      cp -p --parents -L -- "$task_library" "$task_jail"
    fi
  done < <(printf '%s\n' "$task_libraries" | awk '{ for(i=1;i<=NF;i++) if($i ~ /^\//) print $i }')
done
cp -p -L -- /lib64/ld-linux-x86-64.so.2 "$task_jail/lib64/ld-linux-x86-64.so.2"
find "$task_jail" -type d -exec chmod 755 -- {} +

printf '%s\n' 'root:x:0:0:root:/root:/bin/bash' 'fixture:x:1000:1000:POC:/home/fixture:/bin/bash' 'sshd:x:1001:1001:privsep:/run/sshd:/bin/bash' > "$task_jail/etc/passwd"
printf '%s\n' 'root:!:20000:0:99999:7:::' 'fixture:NP:20000:0:99999:7:::' 'sshd:!:20000:0:99999:7:::' > "$task_jail/etc/shadow"
chmod 600 -- "$task_jail/etc/shadow"
printf '%s\n' 'root:x:0:' 'fixture:x:1000:' 'sshd:x:1001:' > "$task_jail/etc/group"
printf '%s\n' 'passwd: files' 'group: files' 'shadow: files' 'hosts: files dns' > "$task_jail/etc/nsswitch.conf"
chmod 644 -- "$task_jail/etc/passwd" "$task_jail/etc/group" "$task_jail/etc/nsswitch.conf"
chmod 755 -- "$task_jail/etc" "$task_jail/etc/hopnode" "$task_jail/run" "$task_jail/run/sshd" "$task_jail/home" "$task_jail/home/fixture" "$task_jail/workspace" "$task_jail/dev" "$task_jail/dev/pts"
ssh-keygen -q -t ed25519 -N '' -f "$task_jail/etc/hopnode/host_ed25519"
ssh-keygen -q -t ed25519 -N '' -f "$task_exchange/client_ed25519"
cp -- "$task_exchange/client_ed25519.pub" "$task_jail/etc/hopnode/authorized_keys"
chmod 644 -- "$task_jail/etc/hopnode/authorized_keys"
chmod 600 -- "$task_exchange/client_ed25519" "$task_jail/etc/hopnode/host_ed25519"
ssh-keygen -lf "$task_jail/etc/hopnode/host_ed25519.pub" | awk '{print $2}' > "$task_exchange/fingerprint"
: > "$task_exchange/known_hosts"
chown 1000:1000 -- "$task_jail/home/fixture" "$task_jail/workspace"
mknod -m 666 "$task_jail/dev/null" c 1 3
mknod -m 666 "$task_jail/dev/urandom" c 1 9
mknod -m 666 "$task_jail/dev/tty" c 5 0
ln -s pts/ptmx "$task_jail/dev/ptmx"

cat > "$task_jail/etc/hopnode/sshd_config" <<EOF
Port $task_port
ListenAddress 127.0.0.1
HostKey /etc/hopnode/host_ed25519
PidFile /run/sshd/pid
AuthorizedKeysFile /etc/hopnode/authorized_keys
PasswordAuthentication no
KbdInteractiveAuthentication no
PubkeyAuthentication yes
PermitRootLogin no
AllowUsers fixture
StrictModes yes
UsePAM no
UseDNS no
AllowTcpForwarding no
AllowAgentForwarding no
X11Forwarding no
PermitTunnel no
Subsystem sftp internal-sftp
LogLevel ERROR
EOF
chmod 600 -- "$task_jail/etc/hopnode/sshd_config"
chroot "$task_jail" /usr/sbin/sshd -t -f /etc/hopnode/sshd_config
mount -t devpts -o newinstance,ptmxmode=0666,mode=620,gid=5 devpts "$task_jail/dev/pts"
trap - ERR
printf 'Prepared isolated fixture on 127.0.0.1:%s.\nEXCHANGE: %s\nJAIL: %s\nRun this script with: run EXCHANGE\n' "$task_port" "$task_exchange" "$task_jail"
