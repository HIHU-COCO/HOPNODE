#!/usr/bin/env python3
"""Bounded P01 probe with localhost coordination; never use a system daemon.

Run as an ordinary Linux user. All generated credentials/state/raw output stay in
a new private native /var/tmp directory. stdout contains sanitized results only.
This exercises two userspace nodes on one host, not two independent machines.
"""

import argparse
import hashlib
import http.server
import json
import os
from pathlib import Path
import re
import signal
import socket
import ssl
import subprocess
import sys
import tarfile
import tempfile
import threading
import time
import urllib.error
import urllib.request


ASSETS = {
    "headscale": (
        "https://github.com/juanfont/headscale/releases/download/v0.29.5/"
        "headscale_0.29.5_linux_amd64",
        "e979d27fe17d924d33bb6ab7504f6338d0635989a0cfdcfbc0ceb637dbcbe980",
    ),
    "tailscale.tgz": (
        "https://pkgs.tailscale.com/stable/tailscale_1.98.4_amd64.tgz",
        "e6c08a8ee7e63e69aaf1b62ecd12672b3883fbcd2a176bf6cfa42a15fdce0b6b",
    ),
}
MARKER = b"HOPNODE_P01_LOCALHOST_SYNTHETIC_PROOF\n"
TCP_PORTS = (18443, 18444, 15443, 19090, 11055, 11056, 42881, 42882)
UDP_PORTS = (13478, 42641, 42642)


def used_key_auth_classification(data):
    """Classify only fixed diagnostics; never return raw credential output."""
    markers = (b"pre auth key has already been used",
               b"preauth key has already been used",
               b"preauthkey has already been used",
               b"auth key already used", b"authkey already used")
    lines = data.lower().splitlines()
    return {"already_used_backend_auth_error": any(
                b"backend error" in line and any(marker in line for marker in markers)
                for line in lines),
            "x509_error_seen": any(b"x509:" in line for line in lines)}


class MarkerHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        content = MARKER if self.path == "/proof" else b"p01-fixture\n"
        self.send_response(200)
        self.send_header("Content-Length", str(len(content)))
        self.end_headers()
        self.wfile.write(content)

    def log_message(self, *_args):
        pass


class BoundedHTTPServer(http.server.ThreadingHTTPServer):
    daemon_threads = True
    block_on_close = False

    def get_request(self):
        connection, address = self.socket.accept()
        connection.settimeout(2)
        context = getattr(self, "fixture_tls_context", None)
        try:
            if context is not None:
                connection = context.wrap_socket(connection, server_side=True)
            return connection, address
        except Exception:
            connection.close()
            raise

    def handle_error(self, _request, _client_address):
        # Handler tracebacks can contain client addresses; never write them to
        # the sanitized stdout/stderr stream. Keep an in-memory type count only.
        self.fixture_handler_error_types.append(sys.exc_info()[0].__name__)


class Probe:
    def __init__(self):
        if sys.platform != "linux" or os.getuid() == 0:
            raise RuntimeError("ordinary Linux user required")
        os.umask(0o077)
        self.root = Path(tempfile.mkdtemp(prefix="hopnode-p01-localhost-", dir="/var/tmp"))
        self.root.chmod(0o700)
        self.bin = self.root / "bin"
        self.raw = self.root / "raw"
        self.tls = self.root / "tls"
        for directory in (self.bin, self.raw, self.tls, self.root / "empty-ca-dir"):
            directory.mkdir(mode=0o700)
        self.env = {"PATH": "/usr/bin:/bin", "LANG": "C.UTF-8", "HOME": os.environ["HOME"]}
        self.children = []
        self.servers = []
        self.command_no = 0
        self.result = {"scope": "single-host localhost userspace mechanism only",
                       "ordinary_user": True, "checks": {}, "cleanup": {}}

    def check(self, name, value):
        self.result["checks"][name] = bool(value)
        if not value:
            raise RuntimeError("check failed: " + name)

    def run(self, label, argv, timeout=15, env=None):
        self.command_no += 1
        stem = self.raw / (f"{self.command_no:03d}-" + label)
        start = time.monotonic()
        with stem.with_suffix(".out").open("wb") as out, stem.with_suffix(".err").open("wb") as err:
            completed = subprocess.run(argv, cwd=self.root, env=env or self.env,
                                       stdout=out, stderr=err, timeout=timeout, check=False)
        return completed.returncode, stem.with_suffix(".out").read_bytes(), \
            stem.with_suffix(".err").read_bytes(), round((time.monotonic() - start) * 1000)

    @staticmethod
    def identity(pid):
        proc = Path("/proc") / str(pid)
        # comm may contain spaces; split only after its final closing parenthesis.
        fields = (proc / "stat").read_text().rsplit(")", 1)[1].split()
        return (fields[19], os.readlink(proc / "exe"), int(fields[1]))

    def start(self, label, argv, env=None):
        if Path(argv[0]).parent != self.bin or Path(argv[0]).name not in {"headscale", "tailscaled"}:
            raise RuntimeError("long-lived child must be one of the verified Go daemons")
        log_path = self.raw / (label + ".log")
        # This critical section is only for these Go daemons, whose runtime/
        # signal.Notify restores TERM/INT handling. KILL remains a bounded fallback.
        blocked = signal.pthread_sigmask(signal.SIG_BLOCK, {signal.SIGALRM, signal.SIGTERM, signal.SIGINT})
        try:
            with log_path.open("wb") as log:
                process = subprocess.Popen(argv, cwd=self.root, env=env or self.env,
                                           stdout=log, stderr=log)
            # Register before pending cancellation is delivered after unmasking.
            entry = [process, None, label, log_path]
            self.children.append(entry)
        finally:
            signal.pthread_sigmask(signal.SIG_SETMASK, blocked)
        try:
            identity = self.identity(process.pid)
            if identity[2] != os.getpid():
                raise RuntimeError("unexpected child parent")
            entry[1] = identity
        except Exception:
            self.stop(entry)
            raise
        return entry

    def stop(self, entry, budget=4):
        process, identity, _label, _log = entry
        if process.poll() is not None:
            return
        current = self.identity(process.pid)
        # A not-yet-waited direct Popen child cannot have its PID recycled. In the
        # provisional startup failure case additionally require current PPID.
        if (identity is None and current[2] != os.getpid()) or (identity is not None and current != identity):
            raise RuntimeError("child identity changed; refusing signal")
        identity = current
        entry[1] = identity
        process.terminate()
        try:
            process.wait(timeout=max(0.05, min(2, budget / 2)))
        except subprocess.TimeoutExpired:
            if self.identity(process.pid) != identity:
                raise RuntimeError("child identity changed; refusing kill")
            process.kill()
            process.wait(timeout=max(0.05, min(2, budget / 2)))

    def serve(self, port, cert=None):
        server = BoundedHTTPServer(("127.0.0.1", port), MarkerHandler)
        server.fixture_handler_error_types = []
        self.servers.append((server, None))
        if cert:
            context = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
            context.load_cert_chain(self.tls / (cert + ".crt"), self.tls / (cert + ".key"))
            server.fixture_tls_context = context
        thread = threading.Thread(target=server.serve_forever, kwargs={"poll_interval": 0.1}, daemon=True)
        thread.start()
        self.servers[-1] = (server, thread)

    def download(self, asset_source=None):
        # Download only: honor the user's already-configured HTTPS proxy without
        # printing its value. TLS validation stays on. New daemons use self.env,
        # which does not inherit this proxy or other ambient network settings.
        opener = urllib.request.build_opener(urllib.request.ProxyHandler())
        if asset_source:
            asset_source = asset_source.resolve(strict=True)
            if (asset_source.name != "bin" or asset_source.parent.parent != Path("/var/tmp")
                    or not asset_source.parent.name.startswith("hopnode-p01-localhost-")
                    or asset_source.parent.stat().st_uid != os.getuid()
                    or asset_source.parent.stat().st_mode & 0o077):
                raise RuntimeError("asset source must be an owned private previous probe directory")
        for name, (url, expected) in ASSETS.items():
            target = self.bin / name
            digest = hashlib.sha256()
            if asset_source and ((asset_source / name).is_symlink()
                                 or (asset_source / name).resolve().parent != asset_source):
                raise RuntimeError("asset source must contain regular local files")
            source_stream = (asset_source / name).open("rb") if asset_source else opener.open(url, timeout=30)
            with source_stream as source, target.open("xb") as dest:
                while True:
                    block = source.read(1024 * 1024)
                    if not block:
                        break
                    digest.update(block)
                    dest.write(block)
            self.check(name + "_official_sha256", digest.hexdigest() == expected)
        (self.bin / "headscale").chmod(0o700)
        with tarfile.open(self.bin / "tailscale.tgz", "r:gz") as archive:
            for binary in ("tailscale", "tailscaled"):
                member = archive.getmember("tailscale_1.98.4_amd64/" + binary)
                if not member.isfile():
                    raise RuntimeError("expected regular archive member")
                with archive.extractfile(member) as source, (self.bin / binary).open("xb") as target:
                    while True:
                        block = source.read(1024 * 1024)
                        if not block:
                            break
                        target.write(block)
                (self.bin / binary).chmod(0o700)
        for binary, args, expected in (("headscale", ["version"], b"0.29.5"),
                                       ("tailscale", ["version"], b"1.98.4"),
                                       ("tailscaled", ["--version"], b"1.98.4")):
            code, out, err, _ = self.run(binary + "-version", [str(self.bin / binary)] + args)
            self.check(binary + "_version", code == 0 and expected in out + err)

    def openssl(self, label, args):
        code, _out, _err, _ms = self.run(label, ["/usr/bin/openssl"] + args)
        self.check(label, code == 0)

    def pki(self):
        for name in ("ca", "unrelated-ca"):
            self.openssl(name, ["req", "-new", "-x509", "-newkey", "rsa:3072", "-nodes",
                "-days", "2", "-subj", "/CN=HopNode P01 temporary " + name,
                "-addext", "basicConstraints=critical,CA:TRUE",
                "-addext", "keyUsage=critical,keyCertSign,cRLSign",
                "-keyout", str(self.tls / (name + ".key")), "-out", str(self.tls / (name + ".crt"))])
        for name, dns in (("server", "localhost"), ("wrong-name", "wrong-name.invalid")):
            self.openssl(name + "-csr", ["req", "-new", "-newkey", "rsa:3072", "-nodes",
                "-subj", "/CN=HopNode P01 fixture", "-keyout", str(self.tls / (name + ".key")),
                "-out", str(self.tls / (name + ".csr"))])
            ext = self.tls / (name + ".ext")
            ext.write_text("basicConstraints=critical,CA:FALSE\n"
                           "keyUsage=critical,digitalSignature,keyEncipherment\n"
                           "extendedKeyUsage=serverAuth\nsubjectAltName=DNS:" + dns + "\n")
            self.openssl(name + "-sign", ["x509", "-req", "-days", "2", "-sha256",
                "-in", str(self.tls / (name + ".csr")), "-CA", str(self.tls / "ca.crt"),
                "-CAkey", str(self.tls / "ca.key"), "-CAcreateserial", "-extfile", str(ext),
                "-out", str(self.tls / (name + ".crt"))])

    def hs(self, label, args, timeout=15):
        return self.run(label, [str(self.bin / "headscale"), "--config", str(self.root / "config.yaml")]
                        + args, timeout)

    def node(self, label, udp, socks, ca="ca.crt"):
        directory = self.root / label
        directory.mkdir(mode=0o700)
        env = dict(self.env, SSL_CERT_FILE=str(self.tls / ca),
                   SSL_CERT_DIR=str(self.root / "empty-ca-dir"))
        args = [str(self.bin / "tailscaled"), "--tun=userspace-networking",
                "--socket=" + str(directory / "tailscaled.sock"),
                "--state=" + str(directory / "state"), "--statedir=" + str(directory),
                "--port=" + str(udp), "--socks5-server=127.0.0.1:" + str(socks),
                "--no-logs-no-support", "--encrypt-state=false"]
        entry = self.start(label, args, env)
        deadline = time.monotonic() + 5
        while not (directory / "tailscaled.sock").exists():
            if entry[0].poll() is not None or time.monotonic() > deadline:
                raise RuntimeError("new daemon did not create own socket: " + label)
            time.sleep(0.05)
        return directory, entry

    def ts(self, label, directory, args, timeout=15):
        return self.run(label, [str(self.bin / "tailscale"),
                               "--socket=" + str(directory / "tailscaled.sock")] + args, timeout)

    def up(self, label, directory, key, url="https://localhost:18443", seconds=5):
        keyfile = directory / "authkey"
        keyfile.write_text(key)
        keyfile.chmod(0o600)
        return self.ts(label, directory, ["up", "--login-server=" + url,
            "--auth-key=file:" + str(keyfile), "--hostname=" + directory.name,
            "--accept-dns=false", "--accept-routes=false", "--netfilter-mode=off",
            "--ssh=false", "--timeout=" + str(seconds) + "s"], timeout=seconds + 8)

    def tls_tests(self):
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}),
                    urllib.request.HTTPSHandler(context=ssl.create_default_context(cafile=str(self.tls / "ca.crt"))))
        with opener.open("https://localhost:18443/health", timeout=5) as response:
            self.check("tls_python_trusted_localhost", response.status == 200)
        for label, ca, url, needle in (
            ("tls-client-trusted", "ca.crt", "https://localhost:18443", b"invalid pre auth key"),
            ("tls-client-wrong-ca", "unrelated-ca.crt", "https://localhost:18443", b"unknown authority"),
            ("tls-client-wrong-name", "ca.crt", "https://localhost:18444", b"not localhost"),
        ):
            directory, entry = self.node(label, 0, 0, ca)
            _code, out, err, _ms = self.up(label, directory, "p01-intentionally-invalid-key", url)
            log = entry[3].read_bytes()
            check_name = "trusted_control_auth_error_diagnostic" if label == "tls-client-trusted" else label
            self.check(check_name, needle in (out + err + log).lower())
            self.stop(entry)

    def fetch(self, label, socks, peer, port=42881):
        code, out, err, ms = self.run(label, ["/usr/bin/curl", "--silent", "--show-error",
            "--proxy", "socks5h://127.0.0.1:" + str(socks), "--noproxy", "",
            "--connect-timeout", "2", "--max-time", "3", "http://" + peer + ":" + str(port) + "/proof"], 5)
        category = "success" if code == 0 and out == MARKER else "content_mismatch" if code == 0 else "connection_failed"
        if code == 28:
            category = "timeout"
        elif code != 0 and b"SOCKS5" in err:
            category = "socks_connection_failed"
        return {"marker_matches": code == 0 and out == MARKER, "exit_code": code,
                "classification": category, "elapsed_ms": ms}

    def wait_fetch(self, label, socks, peer, expected, seconds=12):
        deadline = time.monotonic() + seconds
        while True:
            observation = self.fetch(label, socks, peer)
            matched = observation["marker_matches"] if expected else observation["exit_code"] != 0
            if matched:
                observation["convergence_ms"] = round((seconds - max(0, deadline - time.monotonic())) * 1000)
                return observation
            if time.monotonic() >= deadline:
                raise RuntimeError("new connection result did not converge: " + label)
            time.sleep(0.4)

    def local_marker(self, port):
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
        with opener.open("http://127.0.0.1:" + str(port) + "/proof", timeout=3) as response:
            return response.read() == MARKER

    def write_policy(self, text):
        pending = self.root / "policy.next"
        with pending.open("x") as stream:
            stream.write(text)
        pending.replace(self.root / "policy.hujson")

    def execute(self, source, asset_source=None):
        for port in TCP_PORTS:
            with socket.socket() as test:
                test.bind(("127.0.0.1", port))
        for port in UDP_PORTS:
            with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as test:
                test.bind(("127.0.0.1", port))
        self.check("candidate_ports_initially_available", True)
        self.download(asset_source)
        self.pki()
        config = (source / "headscale-config.yaml.in").read_text()
        config = config.replace("@P01_ROOT@", str(self.root)).replace("@P01_SERVER_IPV4@", "127.0.0.1")
        config = config.replace("@P01_SERVER_FQDN@", "localhost")
        config = config.replace("level: warn", "level: info")
        # Explicit separate loopback metrics port removes empty-value ambiguity.
        config = config.replace('metrics_listen_addr: ""', 'metrics_listen_addr: "127.0.0.1:19090"')
        self.check("config_no_unexpanded_tokens", "@P01_" not in config)
        (self.root / "config.yaml").write_text(config)
        allow = (source / "policy-port-only.hujson").read_text()
        self.write_policy(allow)
        code, _out, _err, _ms = self.hs("configtest", ["configtest"])
        self.check("headscale_configtest", code == 0)
        headscale = self.start("headscale", [str(self.bin / "headscale"), "--config",
                                             str(self.root / "config.yaml"), "serve"])
        deadline = time.monotonic() + 35
        while not (self.root / "headscale.sock").exists():
            if headscale[0].poll() is not None or time.monotonic() > deadline:
                self.result["headscale_start_exit_code"] = headscale[0].poll()
                raise RuntimeError("new Headscale did not create own socket")
            time.sleep(0.05)
        self.check("headscale_socket_mode_600", (self.root / "headscale.sock").stat().st_mode & 0o777 == 0o600)
        code, _out, _err, _ms = self.hs("initial-policy-check", ["policy", "check", "--file", str(self.root / "policy.hujson")])
        self.check("initial_port_only_policy_check", code == 0)
        self.serve(18444, "wrong-name")
        self.tls_tests()
        code, out, _err, _ms = self.hs("user", ["users", "create", "p01-fixture", "--output", "json"])
        self.check("headscale_user_create", code == 0)
        user = json.loads(out)
        user_id = user["id"]
        nodes = []
        used_key_b = None
        for label, udp, socks in (("p01-a", 42641, 11055), ("p01-b", 42642, 11056)):
            code, out, _err, _ms = self.hs("key-" + label, ["preauthkeys", "create", "--user", str(user_id),
                    "--expiration", "10m", "--output", "json"])
            self.check(label + "_one_time_key_create", code == 0)
            auth = json.loads(out)
            if label == "p01-b":
                used_key_b = auth["key"]
            directory, entry = self.node(label, udp, socks)
            code, _out, _err, ms = self.up("up-" + label, directory, auth["key"], seconds=20)
            self.check(label + "_registered", code == 0)
            self.result[label + "_registration_ms"] = ms
            code, out, _err, _ms = self.ts("status-" + label, directory, ["status", "--json"])
            status = json.loads(out)
            self.check(label + "_backend_running", code == 0 and status["BackendState"] == "Running")
            nodes.append((directory, entry, socks, status["TailscaleIPs"][0]))
        self.check("candidate_tls_trusted_valid_registration", True)
        code, out, _err, _ms = self.hs("nodes", ["nodes", "list", "--output", "json"])
        registered = json.loads(out)
        self.check("headscale_exactly_two_synthetic_nodes", code == 0 and len(registered) == 2)
        code, out, _err, _ms = self.ts("derp-map", nodes[0][0], ["debug", "derp-map"])
        derpmap = json.loads(out)
        regions = derpmap["Regions"]
        self.check("only_private_derp_region", code == 0 and set(regions) == {"999"})
        derpnode = regions["999"]["Nodes"][0]
        self.check("derp_localhost_high_port", derpnode["HostName"] == "localhost" and derpnode["DERPPort"] == 18443)
        self.serve(42881)
        self.serve(42882)
        self.check("both_synthetic_services_locally_alive", self.local_marker(42881) and self.local_marker(42882))
        expression = "( " + " or ".join("sport = :" + str(port) for port in TCP_PORTS + UDP_PORTS) + " )"
        code, out, _err, _ms = self.run("own-candidate-listeners", ["/usr/bin/ss", "-H", "-lntup", expression])
        self.check("candidate_listener_inventory_captured", code == 0)
        owned_pids = {os.getpid()} | {p.pid for p, *_ in self.children if p.poll() is None}
        listeners = []
        for line in out.decode().splitlines():
            fields = line.split()
            if len(fields) < 6:
                continue
            pids = {int(pid) for pid in re.findall(r"pid=(\d+)", line)}
            self.check("candidate_listener_owned_" + str(len(listeners)), bool(pids) and pids <= owned_pids)
            address = fields[4]
            port = int(address.rsplit(":", 1)[1])
            category = "loopback" if address.startswith(("127.", "[::1]:")) else \
                "wildcard" if address.startswith(("0.0.0.0:", "*:", "[::]:")) else "interface"
            listeners.append({"protocol": fields[0], "port": port, "address_class": category})
        self.result["own_candidate_listeners"] = listeners
        for i, node in enumerate(nodes):
            peer = nodes[1 - i][3]
            self.wait_fetch("proof-" + str(i), node[2], peer, True)
            self.check("userspace_peer_proof_" + str(i), True)
            observation = self.fetch("denied-port-" + str(i), node[2], peer, 42882)
            self.result["non_granted_synthetic_port_" + str(i)] = observation
            self.check("non_granted_synthetic_port_new_connection_failed_" + str(i),
                       observation["exit_code"] != 0 and self.local_marker(42882))
            _code, out, _err, _ms = self.ts("path-" + str(i), node[0],
                    ["ping", "--c=3", "--timeout=2s", "--until-direct=false", peer], 12)
            self.result["path_observation_" + str(i)] = {
                "derp_line_seen": b"DERP(" in out,
                "peer_relay_line_seen": b"peer-relay(" in out,
                "non_derp_via_line_seen": any(b"via " in line and b"DERP(" not in line
                                               and b"peer-relay(" not in line for line in out.splitlines()),
                "scope": "synthetic single-host disco observation; not application path proof",
            }
        self.write_policy((source / "policy-deny-all.hujson").read_text())
        code, _out, _err, _ms = self.hs("deny-check", ["policy", "check", "--file", str(self.root / "policy.hujson")])
        self.check("deny_all_policy_check", code == 0)
        self.check("headscale_identity_before_hup", self.identity(headscale[0].pid) == headscale[1])
        headscale[0].send_signal(signal.SIGHUP)
        for i, node in enumerate(nodes):
            self.result["deny_all_new_connection_" + str(i)] = self.wait_fetch(
                "deny-all-" + str(i), node[2], nodes[1 - i][3], False)
            self.check("deny_all_new_connection_failed_" + str(i), True)
        self.check("synthetic_service_still_alive_during_deny", self.local_marker(42881))
        self.write_policy(allow)
        self.check("headscale_identity_before_restore", self.identity(headscale[0].pid) == headscale[1])
        headscale[0].send_signal(signal.SIGHUP)
        for i, node in enumerate(nodes):
            self.wait_fetch("restore-" + str(i), node[2], nodes[1 - i][3], True)
            self.check("allow_restored_" + str(i), True)
        # Only this database's synthetic B is expired. This is not HopNode
        # Directory/Agent revocation and says nothing about established streams.
        matching = [node for node in registered
                    if node.get("given_name", node.get("name", "")) == "p01-b"]
        self.check("exact_synthetic_b_identity_selected", len(matching) == 1)
        expiry_started = time.monotonic()
        code, _out, _err, _ms = self.hs("expire-synthetic-b", ["nodes", "expire", "--identifier", str(matching[0]["id"])])
        self.check("synthetic_b_expire_command", code == 0)
        self.result["expire_b_new_connection"] = self.wait_fetch("expired-b-new-connection", nodes[0][2], nodes[1][3], False, 30)
        self.result["expire_b_new_connection"]["since_expire_command_ms"] = round((time.monotonic() - expiry_started) * 1000)
        self.check("expire_b_new_connection_failed_within_30s",
                   self.result["expire_b_new_connection"]["since_expire_command_ms"] <= 30000 and self.local_marker(42881))
        login_deadline = time.monotonic() + 8
        while True:
            code, out, _err, _ms = self.ts("expired-b-status", nodes[1][0], ["status", "--json"])
            status_b = json.loads(out)
            if status_b.get("BackendState") == "NeedsLogin":
                break
            if time.monotonic() >= login_deadline:
                raise RuntimeError("check failed: expired_b_needs_login")
            time.sleep(0.3)
        self.check("expired_b_needs_login", code == 0)
        third, third_entry = self.node("p01-c-used-key", 0, 0)
        code, used_out, used_err, _ms = self.up("used-key-new-state", third, used_key_b, seconds=6)
        auth_diagnostic = used_key_auth_classification(
            used_out + used_err + third_entry[3].read_bytes())
        self.result["used_key_auth_diagnostic"] = auth_diagnostic
        self.check("used_one_time_key_backend_reuse_rejected",
                   auth_diagnostic["already_used_backend_auth_error"]
                   and not auth_diagnostic["x509_error_seen"])
        code_status, out_status, _err, _ms = self.ts("used-key-new-state-status", third, ["status", "--json"])
        third_status = json.loads(out_status)
        code_nodes, out_nodes, _err, _ms = self.hs("nodes-after-used-key", ["nodes", "list", "--output", "json"])
        self.check("used_one_time_key_new_state_not_registered",
                   code != 0 and code_status == 0 and third_status.get("BackendState") == "NeedsLogin"
                   and code_nodes == 0 and len(json.loads(out_nodes)) == 2)
        self.stop(third_entry)

    def cleanup(self):
        errors = 0
        deadline = time.monotonic() + 30
        for server, thread in reversed(self.servers):
            try:
                if thread and thread.is_alive():
                    closer = threading.Thread(target=server.shutdown, daemon=True)
                    closer.start()
                    closer.join(timeout=max(0.05, min(3, deadline - time.monotonic())))
                    if closer.is_alive():
                        errors += 1
                server.server_close()
                if thread:
                    thread.join(timeout=max(0.05, min(1, deadline - time.monotonic())))
                    if thread.is_alive():
                        errors += 1
            except Exception:
                errors += 1
        for entry in reversed(self.children):
            try:
                self.stop(entry, budget=max(0.1, min(4, deadline - time.monotonic())))
            except Exception:
                errors += 1
        ports_available = True
        for kind, ports in ((socket.SOCK_STREAM, TCP_PORTS), (socket.SOCK_DGRAM, UDP_PORTS)):
            for port in ports:
                try:
                    with socket.socket(socket.AF_INET, kind) as test:
                        if kind == socket.SOCK_STREAM:
                            test.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
                        test.bind(("127.0.0.1", port))
                except OSError:
                    ports_available = False
        self.result["cleanup"] = {"identity_checked_child_stop_errors": errors,
                                  "all_owned_children_exited": all(p.poll() is not None for p, *_ in self.children),
                                  "all_candidate_ports_bindable": ports_available,
                                  "within_30_second_budget": time.monotonic() <= deadline,
                                  "runtime_directory_retained": True}
        clean = errors == 0 and ports_available and self.result["cleanup"]["all_owned_children_exited"] \
            and self.result["cleanup"]["within_30_second_budget"]
        self.result["cleanup"]["passed"] = clean
        if not clean:
            self.result["outcome"] = "INCOMPLETE_CLEANUP"
        self.result["runtime_directory"] = str(self.root)
        (self.root / "sanitized-result.json").write_text(json.dumps(self.result, indent=2) + "\n")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--execute-localhost-only", action="store_true", required=True)
    parser.add_argument("--verified-assets-from", type=Path)
    args = parser.parse_args()
    if not args.execute_localhost_only:
        return 2
    probe = Probe()
    def expired(_signum, _frame):
        raise TimeoutError("bounded probe deadline exceeded")
    signal.signal(signal.SIGALRM, expired)
    signal.signal(signal.SIGTERM, expired)
    signal.signal(signal.SIGINT, expired)
    signal.alarm(240)
    try:
        probe.execute(Path(__file__).resolve().parent, args.verified_assets_from)
        probe.result["outcome"] = "LOCAL_MECHANISM_PASS"
    except Exception as error:
        probe.result["outcome"] = "INCOMPLETE"
        # No raw exception output: command errors may contain credential responses.
        probe.result["failure_type"] = type(error).__name__
        if isinstance(error, RuntimeError) and str(error).startswith(("check failed:", "new daemon", "new Headscale", "new connection")):
            probe.result["failure_stage"] = str(error)
    finally:
        signal.alarm(0)
        # Cleanup has its own deadline/per-operation bounds. A second TERM/INT
        # must not skip later resources after the first cancellation.
        signal.signal(signal.SIGTERM, lambda *_args: None)
        signal.signal(signal.SIGINT, lambda *_args: None)
        try:
            probe.cleanup()
        except Exception as cleanup_error:
            probe.result["outcome"] = "INCOMPLETE_CLEANUP"
            probe.result["cleanup_failure_type"] = type(cleanup_error).__name__
    print(json.dumps(probe.result, indent=2))
    return 0 if probe.result["outcome"] == "LOCAL_MECHANISM_PASS" else 1


if __name__ == "__main__":
    sys.exit(main())
