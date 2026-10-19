# P02 实测报告（2026-10-19）

创建者：HIHU-COCO。任务：P02；关联 PRD F05、F08–F13、A20、A24。任务状态：IN_PROGRESS；Windows / Ubuntu 系统 SSH 受控端验收尚未完成。

## 环境与命令

本机：Windows 11 专业版 x64，版本 10.0.26200；Go `go1.27.2 windows/amd64`；系统客户端 `OpenSSH_for_Windows_9.5p2, LibreSSL 3.8.2`。Windows `sshd` 服务未发现。另有 WSL Ubuntu 22.04 运行真实 OpenSSH 8.9p1 隔离受控端（解压 `1:8.9p1-3ubuntu0.18` 软件包）；测试未安装主机服务或新增主机账号。测试模块路径：`tests/pocs/ssh`。依赖版本见 `go.mod` / `go.sum`。

执行的 Go 可执行文件为仓库本地 `.tools/go/bin/go.exe`。下列命令在此模块执行，均是真实结果：

| 命令 | 结果 |
| --- | --- |
| `gofmt -w client.go client_test.go main.go` | PASS，退出码 0 |
| `go test -v ./...` | PASS，8 个顶层测试全部通过；`ok hopnode/pocs/ssh 1.574s` |
| `go vet ./...` | PASS，退出码 0，无诊断 |
| Windows `go test -race ./...` | NOT_RUN，运行环境阻塞：`go: -race requires cgo; enable cgo by setting CGO_ENABLED=1`，退出码 1 |
| WSL Linux `go test -race ./...` | PASS，真实 Linux Go / gcc 工具链；`ok hopnode/pocs/ssh 4.443s`，退出码 0 |

## 自动测试证据

| 测试 | 实际覆盖 | 结果 |
| --- | --- | --- |
| `TestPasswordCommandSFTPAndPTY` | 合成密码认证；经回环 SSH 执行本机真实 `whoami`；PTY 请求；stdout / stderr / exit 7；0 / 256 / 10,485,760 字节 SFTP 往返；中文和空格路径；清理 | PASS |
| `TestHostKeyUnknownTrustAndChanged` | 未知主机拒绝且不写记录；显式批准；后续已知主机；即使提供新指纹也拒绝密钥变更，保留原记录 | PASS |
| `TestPrivateKeyAndRejectedPassword` | 错误合成密码拒绝；临时未加密 / 口令加密 Ed25519 私钥认证和命令 | PASS |
| `TestCommandTimeoutDoesNotClaimRemoteTermination` | 已运行 session 超时返回 context deadline / exit -1 / timed_out；后续普通 session 仍可使用 | PASS |
| `TestCommandTimeoutDuringSetup` | 目标未回应 channel-open；目标未回应 exec 请求；本地等待受超时约束 | PASS |
| `TestRoundTripRejectsEscapeAndExistingDirectory` | 拒绝路径逃逸文件名；拒绝已有目录且保留既有文件 | PASS |
| `TestConfigBasicsAndOpenSSHOracle` | 基础配置、否定 Host、首值优先；显式临时 `-F` 与系统 `ssh -G` 对比 | PASS |
| `TestConfigRejectsUnsupportedInsteadOfIgnoring` | Include / Match / ProxyJump / ProxyCommand / IdentityAgent / 多 IdentityFile / tilde / 无效端口 / 引号 / ForwardAgent 明确拒绝 | PASS |

协议夹具使用 Go SSH 服务端与 pkg/sftp，不是 Windows OpenSSH Server。`whoami` 是夹具在主机上执行的真实进程，不能据此标记 Windows 系统 sshd 验收通过。夹具没有运行远端子进程树，取消测试也没有证明远端进程终止。

## Windows 到 WSL 真实 OpenSSH 复验

主代理准备独立 chroot、临时主机/客户端密钥、独立 passwd、专属 `devpts newinstance` 和仅回环监听的 sshd。测试账号 `fixture` 仅存在于该隔离环境；未修改 Windows 或 WSL 用户原有 SSH 配置。子任务复验执行以下命令，其中工具及测试密钥位于根 `.tools/p02-sshd/`，未提交密钥或运行日志：

```powershell
go run . -host 127.0.0.1 -port 22476 -user fixture `
  -key <repo>/.tools/p02-sshd/client_ed25519 `
  -known-hosts <repo>/.tools/p02-sshd/known_hosts `
  -command whoami -timeout 30s -round-trip-parent /workspace
```

首次已由测试管理员独立核验指纹，复验使用该专用 known-hosts。实际输出，退出码 0：

```json
{"stdout":"fixture\n","stderr":"","exit_code":0,"timed_out":false}
{"sftp_round_trip":"PASS","bytes":10485760}
```

相同命令增加 `-pty` 后再次实际执行，退出码 0：

```json
{"stdout":"fixture\r\n","stderr":"","exit_code":0,"timed_out":false}
{"sftp_round_trip":"PASS","bytes":10485760}
```

这证明 Windows Go 客户端到 Ubuntu 真实 OpenSSH 的指定私钥认证、命令、PTY 分配、中文/空格文件路径、目录列举、10 MiB 二进制往返和清理通过。它是同一 Windows 主机上的 WSL / chroot 回环测试，不是独立 Ubuntu Desktop / Server 安装或跨网络验收；尚无真实密码登录或后台任务生命周期结论。

上述成功方法已整理为 [`openssh-fixture.sh`](openssh-fixture.sh) 的 prepare / run 流程，使用固定版本下载及唯一原生 chroot、独立账号文件与 devpts；停止和卸载步骤见 README，保留数据、不自动终止进程。脚本已实际通过 `bash -n`；非 root 执行返回错误；root 对正在使用的 22476 端口执行 prepare 被拒绝，并确认指定的新交换目录没有创建。

### 整合脚本完整复现与清理

2026-10-19 在 WSL Ubuntu 22.04 以 root 使用新交换目录 `<repo>/.tools/p02-repro` 和独立端口 `22477` 完整运行脚本，保留原 22476 夹具：

```bash
bash <worktree>/tests/pocs/ssh/openssh-fixture.sh prepare <repo>/.tools/p02-repro 22477
bash <worktree>/tests/pocs/ssh/openssh-fixture.sh run <repo>/.tools/p02-repro
```

prepare 实际下载固定版本的两个包并成功校验 chroot 配置，退出码 0；唯一原生目录为 `/var/tmp/hopnode-p02-jail-eQoJy5uo`。run 实际启动前台 sshd。Windows Go 探针使用该新交换目录的客户端私钥、已独立核验的指纹和独立 known-hosts，执行 `-host 127.0.0.1 -port 22477 -user fixture -command whoami -pty -timeout 30s -round-trip-parent /workspace`，实际输出如下，退出码 0：

```json
{"stdout":"fixture\r\n","stderr":"","exit_code":0,"timed_out":false}
{"sftp_round_trip":"PASS","bytes":10485760}
```

结束时确认 `/workspace` 为空，并核对新实例 PID `20583`、`/proc/PID/root`、`/proc/PID/exe` 和进程 starttime 均归属本次 chroot；经 pidfd 发送 SIGTERM，仅停止该已核验实例。前台进程退出码 0；再次检查该 chroot 已无进程后，验证专属挂载类型为 devpts 并成功正常卸载。最终 22477 无监听且该 devpts 无挂载，交换目录和 chroot 数据均保留，未递归删除任何目录。原 22476 夹具仍监听且其专属 devpts 仍挂载，未被本次清理触及。

## 真机 / 外部系统矩阵

| 环境 / 项目 | 状态 | 缺失条件或限制 |
| --- | --- | --- |
| Windows 本机协议夹具 | PASS | 真实 SSH / SFTP 字节，单机回环范围 |
| Windows OpenSSH 受控端 | NOT_RUN | 本机未发现 sshd 服务；需已授权独立 Windows sshd 目标 |
| WSL Ubuntu 22.04 真实 OpenSSH 受控端 | PASS（局部） | 指定私钥、whoami、PTY、路径与 10 MiB SFTP；同主机隔离回环，非独立机器 |
| 独立 Ubuntu 22.04 / 24.04 受控机 | NOT_RUN | 尚无独立测试设备证据；WSL 测试不能替代两版本真机矩阵 |
| Ubuntu Desktop GUI | NOT_RUN | 本 POC 无桌面安装或 GUI 测试 |
| Ubuntu Server CLI 安装 | NOT_RUN | 本 POC 未实施正式包安装 |
| ProxyJump / ssh-agent 两平台认证 | NOT_RUN | 尚无已授权跳板与对应 agent 测试会话，POC 明确不支持 |
| Windows / Ubuntu 后台断线任务 | NOT_RUN | 需真实任务与断线、状态再查询及进程归属证据 |
| MCP / 跨网络 / DERP 真调用 | NOT_RUN | 不属于本回环测试；无相应链路证据 |

## 风险与下一依赖

取消关闭 SSH session 不构成远端进程终止保证；未应答 channel-open 的取消会关闭连接。SFTP 超时或断线可能留下探针临时目录；失败不会输出传输 PASS。自动测试不提供续传、后台 task ID、升级、托管撤销或系统服务结论。

后续需补 Windows / Ubuntu 两种系统 sshd 的密码与私钥认证、命令、PTY、路径列举、10 MiB 往返和 Host Key 错误证据，调查两平台后台断线行为，再交由 P05 联合评审。生产 R01 / R06 还需完善凭据库、用户目录和 SSH 配置兼容矩阵。

本报告不修改状态总表，也不把 P02 或 A20 / A24 标为 DONE / PASS。Git commit 由完成该子任务的提交及主线集成记录追溯；未创建或推送 PR。

## 新授权 LAN Ubuntu 22.04 目标测试计划（尚未连接）

2026-10-19 用户已授权主代理通过 SSH 连接一台局域网 Ubuntu 22.04 查看环境。当前子任务仅审阅本机 POC CLI 并准备流程，未连接该目标，以下全部为 NOT_RUN。主代理负责独立 Host Key 核验、确定授权身份与安全认证注入；只有收到其已核验的专用 known-hosts 和认证接入方式后才执行。目标地址、真实用户名、凭据和原始运行日志保留在忽略的本地交换目录，报告仅记录脱敏结果。

### 只读 inventory

先运行一条短命令确认当前身份、系统和现成工具，每项单独标记，允许预期的缺包/无 systemd 状态，不因它们直接推断平台不支持：

```sh
printf 'uid='; id -u
printf 'gid='; id -g
uname -srmo
cat /etc/os-release
getconf LONG_BIT
printf 'pid1='; cat /proc/1/comm
command -v mktemp
command -v stat
command -v rmdir
command -v tty
command -v stty
if command -v dpkg-query >/dev/null 2>&1; then
  dpkg-query -W -f='${binary:Package}\t${Version}\n' openssh-server 2>/dev/null || true
fi
if command -v systemctl >/dev/null 2>&1; then
  systemctl --version | head -n 1
  systemctl is-active ssh 2>/dev/null || true
fi
```

本步骤不读取用户文件、SSH 配置、密钥、网络配置或完整环境变量，不查询私有 IP、主机名、服务日志或其他设备。systemctl 仅查询，不启动/停止/启用服务；不执行 sudo、apt 安装或远端配置修改。

### 命令、PTY 与临时数据

所有调用复用已核验的独立 known-hosts，认证只在进程环境或明确指定的临时私钥中提供；不使用默认 SSH 文件，也不自动批准新指纹。既有 CLI 支持环境密码及口令私钥、指定 Host/端口/用户、`-command`、`-pty`、`-timeout` 与 `-round-trip-parent`，可直接执行以下步骤：

| 顺序 | 操作与范围 | 判据与证据 |
| --- | --- | --- |
| 1 | 普通 `whoami`，只读确认认证后的身份 | stdout 非空、exit 0；报告隐去真实账号 |
| 2 | `printf 'p02-out\n'; printf 'p02-err\n' >&2; exit 7`，无 PTY | stdout / stderr 分离、exit 7；CLI 预期以错误退出，区别于测试失败 |
| 3 | `-pty -command 'tty; stty size'` | 分配真实 PTY，返回终端路径及 `24 80`；PTY 下流合并属于正常终端语义 |
| 4 | 在明确授权的 `/tmp` 执行 `umask 077; mktemp -d /tmp/hopnode-p02-20261019-XXXXXXXX` | 返回唯一新父目录；将路径保留在本地测试状态中；验证路径前缀、非 symlink、归属当前 UID、mode 700 |
| 5 | `-command whoami -round-trip-parent <步骤4新父目录> -timeout 30s` | 探针仅在该父目录下创建随机子目录，10,485,760 字节中文/空格文件往返，原始字节一致、列举及清理成功才输出 PASS |
| 6 | 可选再以 `-pty` 重复步骤 5 | 分别记录普通命令与 PTY 情况，不据此宣称交互输入/resize 完成 |
| 7 | 验证同一唯一父目录仍归属本次测试且为空，执行 `rmdir -- <该父目录>` | 仅删除空目录；若非空或归属/路径不符则保留现场，不递归删除、不清扫其他同前缀目录 |

真实目标不做错误密码探测，避免触发账户锁定；未知/变化 Host Key 的回归测试使用专用本地合成 known-hosts 或原协议夹具，不改变远端 Host Key。该轮不运行后台守护进程、不发起系统注销、不重启、不撤销真实 SSH 密钥；远端进程树终止和持久化任务仍单独 NOT_RUN。

### GUI、systemd 与软件包条件的后续判定

若需补 Ubuntu GUI 测试，可只读查询是否有桌面会话入口、GTK / WebKitGTK 4.1 包及可用显示会话；SSH 内 `$DISPLAY` 为空不能证明机器没有 GUI。实际 GUI 验收仍要求可交互桌面会话、安装包、桌面菜单启动和真实 IPC，须由主代理按对应 P04 范围安排。

PID 1 为 systemd、存在 systemctl 只说明具备查询条件；服务安装/启停、写入 `/etc/systemd/system`、daemon-reload 或 sudo 授权不是本轮测试的一部分。软件包验证先查询架构及依赖版本；未来 `.deb` 可先本地或只读地用 `dpkg-deb --info/--contents` 检查，安装、升级、卸载须有明确授权及隔离数据策略。

这台目标若完成上述流程，只能补充“独立 Ubuntu 22.04 目标、Windows 客户端、LAN SSH/SFTP/PTY”的局部真实证据；不能标记 Ubuntu 24.04、Ubuntu GUI/Server 安装、跨网络、NAT 穿透或 DERP 验收通过。
