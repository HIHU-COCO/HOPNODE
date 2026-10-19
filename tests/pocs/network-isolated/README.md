# P01：独立用户态组网的 LAN 可行性方案

日期：2026-10-19。任务：P01；关联 F02/F03/F04、A03/A04/A26 的前置研究。

本目录包含 LAN 方案、配置模板及一个已在 WSL 普通用户下执行的 **localhost 协调端、单机共宿主** 执行器。TLS、两用户态节点注册、合成 TCP 与策略切换、网络层 expire/旧一次性 key 负例取得局部机制结果，详见 [单机报告](LOCALHOST-REPORT.md)。LAN 方案本身未执行，没有连接授权远端或修改现有网络配置；P01 的跨网络要求和原 ADR 的阻塞结论保持不变。

单机执行入口是 `localhost_probe.py --execute-localhost-only`，代码拒绝 root。它创建原生 mode 700 临时目录、核验固定官方资产、最多执行 240 秒，并在 finally 用独立 30 秒清理预算核对自有子进程/候选端口。协调/TLS/管理/SOCKS/合成 HTTP 是 loopback；tailscaled 的 peer-discovery UDP 实际为通配监听，不能把参数名解释为所有 socket 都只绑定 loopback。不要在另一轮仍运行、端口占用或未授权的主机上执行。LAN 阶段仍需下面的 FQDN 与网络条件，单机 localhost 的结果不能替代它。

## 结论与推荐拓扑

具备一个两端已能解析、指向授权服务端的专用 DNS 名称后，可推进受限的普通用户试验：在授权 Ubuntu 22.04 上用独立目录运行 Headscale 和内嵌 DERP，在该机及 WSL Ubuntu 22.04 各运行一个便携版 `tailscaled --tun=userspace-networking`。每个实例显式指定新 socket、state、statedir、UDP 端口、loopback SOCKS 端口和私有 CA。所有端口大于 1024，不使用系统服务、TUN、现有 Tailscale socket/state，也不修改系统 DNS 或防火墙规则。目前尚无这个 DNS 条件的实证，TLS/注册阶段仍受该条件阻塞。

这是依据官方文档及固定标签源码的可行性判断，尚非运行结果。推荐把协调和中继服务放在 LAN Ubuntu 上，WSL 主动访问它；避免为 WSL 入站增加 Windows 端口转发或改 NAT。若现有防火墙不允许新增的高端口，本阶段记录 BLOCKED，不能擅自添加规则。

| 角色 | 进程 / 候选监听 | 存储和权限 |
| --- | --- | --- |
| LAN Ubuntu 普通用户 | Headscale HTTPS + 内嵌 DERP：指定 LAN 地址 TCP 18443；STUN：同地址 UDP 13478 | 新建原生 Linux `/var/tmp/hopnode-p01.XXXXXXXX`，mode 700 |
| 同一 Headscale | 管理 Unix socket；TLS gRPC 仅 `127.0.0.1:15443`；metrics 禁用 | Unix socket mode 600；新 SQLite、Noise、DERP 私钥 |
| LAN Ubuntu 普通用户 | 独立 tailscaled：UDP 42641；SOCKS `127.0.0.1:11055` | 该机新建目录的 `node/`；不用系统默认路径 |
| WSL 普通用户 | 独立 tailscaled：UDP 42642；SOCKS `127.0.0.1:11056` | 该机另一个新建原生 Linux 目录；不用 `/mnt/c` 放运行私钥/state |
| 两个合成服务 | 各自 `127.0.0.1:42881`，仅提供临时 proof 文件；可选拒绝对照端口 42882 | 新目录、固定 marker、短时运行；不提供宿主目录内容 |

18443/13478/15443/42641/42642/11055/11056/42881/42882 都只是待检查候选；执行人必须先确认未占用。模板中为新 tailnet 分配的地址段是合成参数，不是任何真实设备地址；执行前确认不与已有环境混淆。用户态实例不会向宿主安装这些路由。

## 固定版本候选与下载核验

下表是研究候选，不修改公共 `packaging/manifests/versions.md`，不代表 P05 已冻结。执行前仍需核对发布说明和安全公告；不自动升级或降级现有安装。

| 组件 | 候选 | 依据 / 尚需验证 |
| --- | --- | --- |
| Headscale / 内嵌 DERP | `v0.29.5`、官方 Linux amd64 独立二进制 | 同一二进制内嵌 DERP，无独立 derper 版本；Ubuntu 22.04 启动、配置和客户端兼容待测 |
| 两端便携 Tailscale | `1.98.4` Linux amd64 tgz | v0.29.5 的 capability 表包含 v1.80–v1.98，声明支持最近 10 个 minor；1.98.4 是该表末端的固定 patch 候选 |
| 已有 Tailscale | 保持原版本及进程 | 已观察到的 1.104.1 不进入本候选兼容结论；不能因连接成功就假定属于 v0.29.5 的声明窗口 |
| TLS | OpenSSL 生成两天有效的私有 CA 与 DNS SAN 叶证书 | 仅新进程信任，不安装系统 CA、不改 hosts；名称须两端已有可用解析，实际 tailscaled TLS 验证待测 |

官方资产及本次读到的 SHA-256（2026-10-19，仅下载摘要文本，未安装/运行二进制）：

- [Headscale 独立二进制](https://github.com/juanfont/headscale/releases/download/v0.29.5/headscale_0.29.5_linux_amd64)，[官方 checksums.txt](https://github.com/juanfont/headscale/releases/download/v0.29.5/checksums.txt)：`e979d27fe17d924d33bb6ab7504f6338d0635989a0cfdcfbc0ceb637dbcbe980`。
- [Tailscale tgz](https://pkgs.tailscale.com/stable/tailscale_1.98.4_amd64.tgz)，[官方摘要](https://pkgs.tailscale.com/stable/tailscale_1.98.4_amd64.tgz.sha256)：`e6c08a8ee7e63e69aaf1b62ecd12672b3883fbcd2a176bf6cfa42a15fdce0b6b`。

由执行人另行下载到忽略目录并逐字节核验 SHA-256，解压便携文件到本轮 `bin/`。不安装 `.deb`，不运行官方 install.sh，不复用 `/usr/sbin/tailscaled`。摘要一致证明所取文件与官方摘要一致，不证明所有安全/兼容性问题已解决。若摘要、版本或 release 来源不符，停止。

## 执行前只读检查与权限边界

执行人通过既有已核验 SSH Host Key 的渠道访问目标；本方案不包含认证信息和真实地址。先收集脱敏基线：OS/架构、普通用户布尔、`openssl/python3/curl/ss/timeout` 是否存在、既有 tailscaled service 状态与进程可执行文件身份、默认 socket 的元数据，以及有关候选端口是否空闲。原始 `ss`、路由、status JSON、环境、日志可能含设备名或地址，只保留在受保护的忽略目录，禁止直接粘贴或提交。

可用的只读命令类别：`id -u`、`uname -m`、`command -v ...`、`systemctl is-active tailscaled`、`ss -H -lnt`、`ss -H -lnu`、`stat`；可由已有只读 P01 探针得到旧网络后端状态和节点数量。无需读取系统 state 内容或现有 Tailscale 密钥。不要为基线调用旧实例的 `up/down/logout/set`。

| 操作 | 普通用户能否做 / 处理 |
| --- | --- |
| 原生临时目录、便携二进制、高端口、Unix socket、私有 CA 文件 | 预期可以；实际权限需验证 |
| userspace tailscaled、SOCKS、loopback 合成服务 | 官方支持无 TUN 的模式；两台当前主机启动仍 NOT_RUN |
| 系统 TUN、iptables/nftables、系统路由/DNS、系统 CA、systemd service | 不在本轮范围；通常需要 root/CAP_NET_ADMIN 或管理授权 |
| 私有 CA 进程级信任 | Linux Go `SSL_CERT_FILE` 支持；候选 Tailscale 的 FQDN 名称验证与链验证仍需正负实测 |
| 已有可用的专用 DNS 名称 | 当前未确认；缺少则阻塞 TLS/注册，不以 IP URL、改 hosts 或 debug DNS 替代 |
| 强制阻断 P2P 的独立网络 namespace / VM 防火墙 | 通常需 root 或已许可的 user namespace；当前未具备该试验边界 |
| 公网 ACME、443、不同网络、可公开到达 DERP | 本轮无资源承诺；LAN 高端口/私有 CA 不等同生产部署 |

不尝试 `sudo` 安装、赋 capability、修改 sysctl、创建 host firewall 规则或调试开关绕过 TLS。若缺普通用户启动条件，记录原错误；执行人再决定是否另开隔离 VM/netns 试验。

## 分阶段命令草案（均未执行）

下列命令只供执行人审查和逐步执行，不是可一键部署的脚本。尖括号表示需要运行时提供的值，不能把真实值写回本目录。每一阶段失败即停止依赖它的阶段。所有 shell 都禁用 `set -x`，不得打印 preauth key、state、私钥或原始 status；不把 token 放命令参数。新网络进程/CLI 使用最小进程环境，防止继承 `HEADSCALE_*` 配置覆盖、代理、`TS_DEBUG_*` 或 `SSLKEYLOGFILE`；不修改父会话环境，也不读取/打印其中的秘密。

### 1. 每台分别建目录并核验便携工具

```bash
set -eu
umask 077
test "$(id -u)" -ne 0
P01_ROOT=$(mktemp -d /var/tmp/hopnode-p01.XXXXXXXX)
chmod 700 "$P01_ROOT"
mkdir -m 700 "$P01_ROOT/bin" "$P01_ROOT/node" "$P01_ROOT/tls" \
  "$P01_ROOT/raw" "$P01_ROOT/www" "$P01_ROOT/empty-ca-dir"
# 执行人安全保留该主机的绝对目录；不要将其作为另一台的路径复用。
# 将已核验资产置于 bin/，仅 chmod 本轮文件为可执行；不移动系统文件。
"$P01_ROOT/bin/tailscale" version >"$P01_ROOT/raw/tailscale-version.txt"
"$P01_ROOT/bin/tailscaled" --version >"$P01_ROOT/raw/tailscaled-version.txt"
# 仅服务端需要：
"$P01_ROOT/bin/headscale" version >"$P01_ROOT/raw/headscale-version.txt"
```

先核验 CLI `--help` 与本方案一致；不因名称相同就使用 PATH 中已有程序。执行前保存已核验二进制的绝对路径、摘要。版本命令不启动 daemon。

### 2. 服务端私有 TLS；客户端只收 CA 公钥

由执行人通过受保护运行变量提供授权 LAN IPv4 `P01_SERVER_IP` 和专用 FQDN `P01_SERVER_NAME`；两端须已能通过正常解析得到该授权地址。不能为本轮改 hosts、resolv.conf 或系统 DNS。先验证变量格式及实际解析归属；变量只用于本轮配置，不提交。不要使用 `http://` 或 IP literal login server。

```bash
: "${P01_SERVER_IP:?provide the authorized LAN IPv4 in the protected session}"
: "${P01_SERVER_NAME:?provide an already resolvable authorized FQDN}"
export P01_SERVER_IP P01_SERVER_NAME
python3 -c 'import ipaddress,os; ipaddress.IPv4Address(os.environ["P01_SERVER_IP"])'
# 执行人先验证 NAME 是普通 DNS FQDN（不是 IP/URL、无换行/扩展语法），
# 再用 socket.getaddrinfo 检查两个主机的正常解析均只指向授权服务端。
openssl req -new -x509 -newkey rsa:3072 -nodes -days 2 \
  -subj '/CN=HopNode P01 temporary CA' \
  -addext 'basicConstraints=critical,CA:TRUE' \
  -addext 'keyUsage=critical,keyCertSign,cRLSign' \
  -keyout "$P01_ROOT/tls/ca.key" -out "$P01_ROOT/tls/ca.crt" \
  >"$P01_ROOT/raw/ca.stdout" 2>"$P01_ROOT/raw/ca.stderr"
openssl req -new -newkey rsa:3072 -nodes \
  -subj '/CN=HopNode P01 server' \
  -keyout "$P01_ROOT/tls/server.key" -out "$P01_ROOT/tls/server.csr" \
  >"$P01_ROOT/raw/csr.stdout" 2>"$P01_ROOT/raw/csr.stderr"
cat >"$P01_ROOT/tls/server.ext" <<EOF
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:${P01_SERVER_NAME}
EOF
openssl x509 -req -days 2 -sha256 -in "$P01_ROOT/tls/server.csr" \
  -CA "$P01_ROOT/tls/ca.crt" -CAkey "$P01_ROOT/tls/ca.key" -CAcreateserial \
  -extfile "$P01_ROOT/tls/server.ext" -out "$P01_ROOT/tls/server.crt" \
  >"$P01_ROOT/raw/sign.stdout" 2>"$P01_ROOT/raw/sign.stderr"
openssl verify -CAfile "$P01_ROOT/tls/ca.crt" \
  -verify_hostname "$P01_SERVER_NAME" "$P01_ROOT/tls/server.crt" \
  >"$P01_ROOT/raw/cert-verify.txt" 2>&1
```

CA 私钥、服务端私钥始终留服务端新目录。只把 `ca.crt` 经已核验 SSH 通道传到 WSL 自有 `tls/ca.crt`，比较 SHA-256；绝不复制 CA 私钥。新进程设置 `SSL_CERT_FILE` 指向此 CA，`SSL_CERT_DIR` 指向新建空目录；不运行 `update-ca-certificates`，不设置 `TS_DEBUG_*`、TLS skip/insecure 或修改客户端源码。

这里信任的是专用 CA，服务端叶证书仍须链验证、有效期验证及 DNS SAN 名称验证。Go 与 Tailscale 固定源码支持这一推断，但实际候选二进制必须先通过 TLS 正负例。若失败，只记录失败；不绕过验证。普通公网 CA 的证书/DNS 是另一条以后可选择的正式路径。

保守排除 IP URL 的原因：候选 `tlsdial.Config` 使用 `ConnectionState.ServerName` 做 x509 名称验证，而 Go 客户端用 `hostnameInSNI` 构造该字段，IP literal 会被移除。因此不能只根据“配置 IP SAN + curl 成功”推断 tailscaled 对 IP 名称验证可靠。这是源码审查发现的待确认边界，不声称已经复现上游漏洞。使用 FQDN 并测试同 CA 错误名称的拒绝，避免依赖该不确定行为。

### 3. 最小策略与 Headscale 模板

把 `headscale-config.yaml.in` 复制到本轮目录，精确替换三个 token：`@P01_ROOT@` 为当前服务端原生绝对目录，`@P01_SERVER_IPV4@` 为已验证地址，`@P01_SERVER_FQDN@` 为上述已能解析的专用名称；不得保留 token 或写入 `/etc/headscale`。从本目录复制 `policy-port-only.hujson` 到该目录 `policy.hujson`，**在首个节点入网前就配置它**。

该策略只允许本轮新 tailnet 内 TCP 42881，没有 SSH、subnet/exit routes、任意宿主端口或公网访问授权。只有本轮单一用户、两枚一次性注册 key，不能导入现有 tailnet 用户/节点。Headscale 无策略、空对象或漏掉 grants/acls 会默认 allow-all；不可用空对象代替拒绝策略。`policy-deny-all.hujson` 用于后续拒绝对照。

```bash
# configtest 也可能创建本轮数据库/密钥，只能指向这个新目录。
env -i PATH=/usr/bin:/bin LANG=C.UTF-8 HOME="$HOME" \
  "$P01_ROOT/bin/headscale" --config "$P01_ROOT/config.yaml" configtest \
  >"$P01_ROOT/raw/configtest.stdout" 2>"$P01_ROOT/raw/configtest.stderr"
env -i PATH=/usr/bin:/bin LANG=C.UTF-8 HOME="$HOME" \
  "$P01_ROOT/bin/headscale" --config "$P01_ROOT/config.yaml" serve \
  >"$P01_ROOT/raw/headscale.log" 2>&1 &
P01_HEADSCALE_PID=$!
hs() { env -i PATH=/usr/bin:/bin LANG=C.UTF-8 HOME="$HOME" \
  "$P01_ROOT/bin/headscale" --config "$P01_ROOT/config.yaml" "$@"; }
hs policy check --file "$P01_ROOT/policy.hujson" \
  >"$P01_ROOT/raw/policy-check.stdout" 2>"$P01_ROOT/raw/policy-check.stderr"
```

启动人立即记录 PID、`/proc/PID/stat` 的 starttime、`/proc/PID/exe` 及新目录归属，核对仅预期端口；不以 `pidof`/`pkill` 选择进程。确认 Unix socket 600、新 SQLite/私钥均在新目录，无系统服务变化。所有策略检查必须显式传本轮 `--config`；不使用直接数据库 bypass 修改运行中服务。

两机各用 `curl --cacert "$P01_ROOT/tls/ca.crt" --noproxy '*' --connect-timeout 3 --max-time 5 "https://${P01_SERVER_NAME}:18443/health"`，保存原始响应到私有 `raw/`；须校验证书成功且服务返回健康结果。再对同一服务使用一个无关临时 CA，要求 TLS 拒绝。用同 CA 签发错误 DNS SAN 叶证书的独立短时 TLS 对照实例，要求候选新客户端因名称不符拒绝；不能只拿 curl 负例代替 tailscaled 负例。证书过期或链不可信也须拒绝。不得用 `curl -k`，不得对正式目标替换证书，不得用 DNS 覆盖掩盖校验问题。

### 4. 一次性注册材料和独立客户端

在新 Headscale 数据库创建合成用户名 `p01-fixture`；其数字 ID 由 JSON 结果获取。两端分别创建一枚 `--expiration 10m`、不可复用、非 ephemeral 的 key，以观察稳定节点记录。创建命令的 JSON 全量仅保存到新目录受保护文件；用执行人审查过的解析器把 key 写 mode 600 文件，另保留 key ID，不在工具输出显示 JSON 或 key。

```bash
hs users create p01-fixture --output json >"$P01_ROOT/raw/user.json"
# P01_USER_ID 来自上述本轮用户；不是旧部署的 ID。
hs preauthkeys create --user "$P01_USER_ID" --expiration 10m --output json \
  >"$P01_ROOT/raw/auth-a.json"
hs preauthkeys create --user "$P01_USER_ID" --expiration 10m --output json \
  >"$P01_ROOT/raw/auth-b.json"
```

执行人安全传递 WSL 的单枚 key 到其本轮 `node/authkey`，只传一次，不复用给第二节点。CLI `--auth-key=file:...` 只暴露文件路径。所有状态/认证输出重定向到私有文件；不导出通用 `TS_AUTHKEY` 到共享会话。

以下在两端各执行一次；按表为各主机设置 `P01_UDP_PORT`、`P01_SOCKS_PORT` 及合成 `P01_NODE_NAME`（例如 `p01-a`、`p01-b`）。`P01_CONTROL_URL` 为运行时 HTTPS 地址；每台 `P01_ROOT` 均为自己的目录。

```bash
env -i PATH=/usr/bin:/bin LANG=C.UTF-8 HOME="$HOME" \
  SSL_CERT_FILE="$P01_ROOT/tls/ca.crt" SSL_CERT_DIR="$P01_ROOT/empty-ca-dir" \
  "$P01_ROOT/bin/tailscaled" --tun=userspace-networking \
  --socket="$P01_ROOT/node/tailscaled.sock" \
  --state="$P01_ROOT/node/tailscaled.state" --statedir="$P01_ROOT/node" \
  --port="$P01_UDP_PORT" --socks5-server="127.0.0.1:$P01_SOCKS_PORT" \
  --no-logs-no-support --encrypt-state=false \
  >"$P01_ROOT/raw/tailscaled.log" 2>&1 &
P01_TAILSCALED_PID=$!
ts() { env -i PATH=/usr/bin:/bin LANG=C.UTF-8 HOME="$HOME" \
  "$P01_ROOT/bin/tailscale" --socket="$P01_ROOT/node/tailscaled.sock" "$@"; }
ts up --login-server="$P01_CONTROL_URL" --auth-key="file:$P01_ROOT/node/authkey" \
  --hostname="$P01_NODE_NAME" --accept-dns=false --accept-routes=false \
  --netfilter-mode=off --ssh=false --timeout=30s \
  >"$P01_ROOT/raw/up.stdout" 2>"$P01_ROOT/raw/up.stderr"
ts status --json >"$P01_ROOT/raw/status.json"
```

`--encrypt-state=false` 仅使本轮临时文件不探测/使用 TPM，不改变旧 state；目录和 state 必须严格保密。它不是生产秘密存储方案。如果候选二进制不支持该 flag，先复核构建，不默默退到系统实例。每次 `ts` 调用前都核对 socket 绝对路径归属，不执行无 socket 参数的 Tailscale 命令。成功后删除本轮明文 key 文件前应由执行人确认它正是本轮新文件；不递归清理目录。

预期 Headscale 恰有两个本轮节点，两个客户端只见本轮 peer；记录 backend Running、node ID 稳定布尔和 peer count。不保存地址/节点 key 到仓库。停止并重启本轮一个 daemon、复用它自己的 state，可测 ID 稳定；不可把两端 state 互换。

### 5. 用户态连通、策略对照及链路读取

userspace 模式不会让普通 `ssh <tailnet-IP>` 或当前 P02 Go 工具的直接 `net.Dial` 自动进入新网络。当前阶段使用独立 SOCKS 或 `ts nc <peer-IP> 42881`；后续生产适配需正式设计，不能复制一套 SSH 实现来规避共用 controller。

在两端 `www/proof` 写本轮固定无隐私 marker，分别启动最多 90 秒的 `python3 -m http.server 42881 --bind 127.0.0.1 --directory "$P01_ROOT/www"`；监听日志留私有目录。执行人保存该新进程身份，并由 `timeout --signal=TERM --kill-after=5s 90s` 限时。服务目录只含合成文件，绝不能指定仓库根、home 或 `/`。

```bash
# P01_PEER_IP 仅从独立实例 status 中解析，绝不从旧网络取得。
curl --proxy "socks5h://127.0.0.1:$P01_SOCKS_PORT" --noproxy '' \
  --connect-timeout 3 --max-time 8 "http://${P01_PEER_IP}:42881/proof" \
  >"$P01_ROOT/raw/proof.received" 2>"$P01_ROOT/raw/proof.stderr"
ts debug derp-map >"$P01_ROOT/raw/derp-map.json"
ts ping --c=3 --timeout=3s --until-direct=false "$P01_PEER_IP" \
  >"$P01_ROOT/raw/path.txt" 2>"$P01_ROOT/raw/path.stderr"
ts status --json >"$P01_ROOT/raw/status-after-proof.json"
```

两方向 marker 比较通过才记录 LAN userspace TCP PASS。`ts nc` 可作为不依赖 SOCKS 的单次 TCP备选；它使用该 socket 的 LocalAPI `DialTCP`，不改宿主路由。此测试没有 SSH 身份认证、命令语义或 SFTP，因此也不能据此标记 SSH/SFTP 经组网完成。

必须核对 DERP map 只有本轮 region 999、预期 HTTPS 高端口和 STUN 端口，没有加载 Tailscale 公共 DERP map。map 存在仅说明配置，不证明流量经过 DERP。`ping` 默认 disco 结果的 Endpoint 表示该次直连观测，DERPRegionID/Code 表示该次中继观测；status 的 PeerRelay 另记 `peer-relay`，不能当 DERP。短时 relay 后转 direct 很常见，分别记录时间、结果；无充分证据写 unknown。不能把 home DERP 字段当某次业务连接的实际路径。

如有独立 loopback 42882 合成对照服务，要求通过新 tailnet 对该端口的新 TCP 连接被策略拒绝，且本机访问对照 marker 正常。更强的对照：把本轮策略原子替换成 `policy-deny-all.hujson`，验证后仅向已核验本轮 Headscale PID 发 SIGHUP，等待 map 下发，再要求 **新建** 42881 连接失败；恢复 port-only 策略并再次成功。不得使用 `kill -HUP $(pidof headscale)`、修改旧策略或对系统 SSH 端口作业务操作。策略拒绝不等同网络断线；已建立连接是否继续必须另外记录。

### 6. 网络授权撤销的有限测试

先确认目标 ID 精确对应本轮节点 B；保存新连接成功基线。然后 `hs nodes expire --identifier "$P01_NODE_B_ID"`，观察 B 需要重新认证、A 的 peer/map 更新，以及从 A **新建** proof 连接在限定 30 秒观察窗口内是否失败。不要事先 `ts down` 掩盖服务端撤销是否生效。记录实际传播时间；超时未拒绝即 FAIL/UNKNOWN，不能擅自解释为成功。

仅把一次性 preauth key 设过期不能撤销已注册设备。必要时再由执行人针对已核验 B ID 执行 `hs nodes delete --identifier ...` 的交互确认，或精确 expire 本轮 key ID；不使用全局无范围批量操作。单独检查旧一次性 key 不能在第三个全新临时 state 重新注册；不得为验证而覆盖 B 的原 state，也不允许自动创建新 key 修复负例。所有负例日志仍需脱敏。

这只测试 Headscale 网络授权及旧 key 行为。它不实现 HopNode Directory/Agent 身份撤销、应用凭据作废、活动 SSH 会话回收或防重装冒用，所以 A26 继续 NOT_RUN。已建立 TCP、缓存 map、服务端失联时的失效时间也应分别记录，不能与新连接拒绝混为一谈。

### 7. 结束与现有网络复查

执行人只停止自己本轮启动并核对了 PID、starttime、exe、目录及所属 UID 的进程。推荐先停两个合成服务，再两个新 tailscaled，再新 Headscale；逐个 SIGTERM、等待退出、核对新端口已消失。PID 已退出或身份不一致时不发送信号；必要时保留进程信息等待人工处理。禁止 `pkill tailscaled`、`killall headscale`、`tailscaled --cleanup`、`systemctl stop tailscaled`、无 socket 的 `logout` 和递归删目录。原始证据目录保留，凭据按执行人的明确范围处理。

复查原 daemon 的 PID/starttime/服务状态/后端状态、系统默认 socket 和原网络概览，与基线比较；无意外变更才记录隔离保护 PASS。新的 peer discovery UDP、STUN 流量和进程资源仍会占用宿主网络/CPU，这是一轮有限授权试验，不能承诺零资源影响。

## 不能由此证明的内容与下一阶段

- WSL 与授权 Ubuntu 即使地址或 NAT 不同，仍属于本次 LAN 试验条件；不能写家庭/公司不同网络 PASS。
- 有 DERP map、DERP health 或初始 disco relay 观测，不等同“阻断 P2P 后通过 DERP 完成命令与文件传输”。强制回退需要独立 VM/netns 或明确只影响新实例的网络边界；当前不改主机防火墙、不用 debug 开关强制结果。
- 缺专用 FQDN 时，本主方案阻塞而不是改 IP URL。另可评估 `https://localhost` + DNS SAN localhost + 独立、已验证 SSH TCP 隧道的受限方案，无需修改系统解析；但那会使协调/DERP 经过 SSH 隧道，STUN/DERP map 也需独立设计，不能冒充 LAN 直接 DERP 或公网中继。该替代不包含在当前模板内，未执行。
- 生产要求仍需公网可达 HTTPS/DERP、合理证书和 DNS、维护升级、安全配置及不同网络授权端点。官方推荐 443；v0.29.5 源码能广告高端口不保证所有客户端/代理环境都接受。
- 当前 P02 工具需要受审查的 SOCKS/LocalAPI 拨号适配，才可复用其 Host Key、命令/PTY、10MiB SFTP 验证；单纯 curl marker 不能代替它。
- 本轮不启动 Windows 第二个 daemon，不证明 Windows 系统服务隔离、GUI 链路展示、100 设备、并发/1GB 性能或 A25。

## 官方依据（2026-10-19 只读核验）

1. [Tailscale userspace networking](https://tailscale.com/kb/1112/userspace-networking)：无 TUN 的 SOCKS/HTTP 模式。
2. [tailscaled 参数](https://tailscale.com/kb/1278/tailscaled)，[1.98.4 tailscaled 源码](https://github.com/tailscale/tailscale/blob/v1.98.4/cmd/tailscaled/tailscaled.go)，[up 参数及 file: key](https://github.com/tailscale/tailscale/blob/v1.98.4/cmd/tailscale/cli/up.go)。
3. [Headscale v0.29.5](https://github.com/juanfont/headscale/releases/tag/v0.29.5)，[客户端政策](https://github.com/juanfont/headscale/blob/v0.29.5/docs/about/clients.md)，[固定 capability 表](https://github.com/juanfont/headscale/blob/v0.29.5/hscontrol/capver/capver_generated.go)。此表是候选选择依据，不是此环境兼容测试。
4. [Headscale 要求](https://headscale.net/stable/setup/requirements/)、[TLS](https://headscale.net/stable/ref/tls/)、[内嵌 DERP](https://headscale.net/stable/ref/derp/)、[策略默认行为](https://headscale.net/stable/ref/policy/)。
5. [固定 config-example](https://github.com/juanfont/headscale/blob/v0.29.5/config-example.yaml)、[DERP 高端口映射实现](https://github.com/juanfont/headscale/blob/v0.29.5/hscontrol/derp/server/derp_server.go)、[gRPC 监听实现](https://github.com/juanfont/headscale/blob/v0.29.5/hscontrol/app.go)。
6. [Go 1.26.3 Linux CA 环境变量](https://github.com/golang/go/blob/go1.26.3/src/crypto/x509/root_unix.go)、[Tailscale 1.98.4 TLS 验证](https://github.com/tailscale/tailscale/blob/v1.98.4/net/tlsdial/tlsdial.go)、[Go 客户端 SNI 与连接状态](https://github.com/golang/go/blob/go1.26.3/src/crypto/tls/handshake_client.go)。私有 CA 信任来自标准证书验证，不以忽略验证的方式替代。
7. [用户态入站 loopback 转发实现](https://github.com/tailscale/tailscale/blob/v1.98.4/wgengine/netstack/netstack.go)、[LocalAPI nc](https://github.com/tailscale/tailscale/blob/v1.98.4/cmd/tailscale/cli/nc.go)、[链路观测 ping](https://github.com/tailscale/tailscale/blob/v1.98.4/cmd/tailscale/cli/ping.go)。
8. [Headscale 节点 expire/delete](https://github.com/juanfont/headscale/blob/v0.29.5/cmd/headscale/cli/nodes.go)、[preauth key 语义](https://github.com/juanfont/headscale/blob/v0.29.5/cmd/headscale/cli/preauthkeys.go)、[policy check](https://github.com/juanfont/headscale/blob/v0.29.5/cmd/headscale/cli/policy.go)。

运行证据填写 [RESULTS.md](RESULTS.md)；此方案本身不能改变共享台账状态。
