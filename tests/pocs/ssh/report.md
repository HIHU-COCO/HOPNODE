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
