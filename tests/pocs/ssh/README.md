# P02：SSH / SFTP 与任务语义探针

用途：验证 Go SSH 栈的基础行为，收集 Windows OpenSSH / Ubuntu sshd 的实际兼容证据。创建者：HIHU-COCO；日期：2026-10-19。

这是隔离的技术 POC 模块，不是正式 `hopnodectl`，不包含设备目录、Agent、后台任务持久化或生产凭据存储。依据 PRD F05、F08–F13、A20、A24，以及 Plan P02；结论见 [report.md](report.md) 与 [ADR-ssh-task.md](../../../docs/decisions/ADR-ssh-task.md)。

## 本地检查

在本目录使用已安装的 Go 1.26+（本次实际使用 Go 1.27.2）：

```powershell
go test -v ./...
go vet ./...
```

自动测试只监听 `127.0.0.1` 的随机端口，使用临时生成的 SSH 主机/客户端密钥与合成密码，退出时关闭监听器和连接。SFTP 文件放在测试临时目录，包含中文、空格、二进制、空文件和 10 MiB 数据校验。`ssh -G -F <临时配置>` 仅用于比对基础配置解析，不连接任何主机。它们属于协议夹具测试，不能替代系统 OpenSSH Server 或跨网络验收。

## 复现 Ubuntu 真实 OpenSSH 夹具

[`openssh-fixture.sh`](openssh-fixture.sh) 整理自本次实际成功的 WSL Ubuntu 22.04 隔离方法。准备阶段须在 Ubuntu 22.04 amd64 以 root 执行，依赖已存在的 `apt-get`、`dpkg-deb`、`ssh-keygen`、`bash`、`ldd`、`mount`、`mknod`、`ss` 等工具，以及允许 chroot / devpts 挂载的环境。脚本仅下载并解压 `openssh-server=1:8.9p1-3ubuntu0.18` 与 `libwrap0=7.6.q-31build2`，保留包版本和 SHA256 证据；固定版本从当前 apt 源不可获取时会明确失败，保留现场。它不安装包、注册主机服务或添加主机账号。

显式指定一个**不存在**的交换目录，父目录须已存在。WSL 下可以使用仓库根 `.tools/` 中的新目录，从 Windows Go 客户端读取临时客户端私钥；它继承所在 Windows 文件系统的 ACL。原生 Linux 交换目录也可用。主机私钥保存在原生 `/var/tmp/hopnode-p02-jail-XXXXXXXX` 的 `0600` 文件中，满足 sshd 权限检查。两者都不应加入 Git；仓库根 `.tools/` 已忽略。

```bash
# Ubuntu / WSL 终端；替换 <repo> 为当前仓库在该 Linux 环境的绝对路径。
sudo bash <repo>/tests/pocs/ssh/openssh-fixture.sh prepare <repo>/.tools/p02-repro-new 22477
sudo bash <repo>/tests/pocs/ssh/openssh-fixture.sh run <repo>/.tools/p02-repro-new
```

默认端口为 `22476`；示例用 `22477` 便于已有夹具继续运行。端口必须空闲且仅绑定 `127.0.0.1`。`run` 在前台执行；保留此终端，在另一个终端运行探针。chroot 内具有独立的 passwd / group / shadow、`fixture` 用户、专用 `/workspace` 和 `devpts newinstance`。它用于受信任的本机协议测试，不是恶意代码的安全沙箱；不在其中运行任意第三方脚本。

交换目录包含 `client_ed25519`、公钥、公开的 `fingerprint`、独立 `known_hosts`、`jail-path`、`port` 及下载包。用管理员刚生成的 `fingerprint` 核验本次本机夹具，在 Windows 探针中传入对应路径：

```powershell
$pocExchange = '<repo>/.tools/p02-repro-new'
$verifiedFingerprint = (Get-Content -LiteralPath "$pocExchange/fingerprint" -Raw).Trim()
go run . -host 127.0.0.1 -port 22477 -user fixture `
  -key "$pocExchange/client_ed25519" -known-hosts "$pocExchange/known_hosts" `
  -trust-host-key $verifiedFingerprint -command whoami -pty `
  -timeout 30s -round-trip-parent /workspace
```

停止时先关闭测试客户端，再在运行该夹具的前台终端按 `Ctrl+C`，核对该终端退出与指定端口监听状态。脚本没有自动 kill / 删除逻辑；`PidFile` 可能残留，不用于直接终止进程。确认该夹具与子会话均已停止后，root 在 Ubuntu 终端使用以下步骤验证路径与归属并卸载专属 devpts，保留全部数据：

```bash
task_exchange=$(realpath -e -- '<repo>/.tools/p02-repro-new')
task_jail=$(cat -- "$task_exchange/jail-path")
[[ "$task_jail" =~ ^/var/tmp/hopnode-p02-jail-[A-Za-z0-9]{8}$ ]] || exit 1
[[ -d "$task_jail" && ! -L "$task_jail" && "$(realpath -e -- "$task_jail")" == "$task_jail" ]] || exit 1
[[ "$(stat -c %u -- "$task_jail")" == 0 ]] || exit 1
[[ "$(cat -- "$task_jail/.hopnode-p02-exchange")" == "$task_exchange" ]] || exit 1
findmnt --mountpoint "$task_jail/dev/pts"
[[ "$(findmnt -rn --mountpoint "$task_jail/dev/pts" -o FSTYPE)" == devpts ]] || exit 1
umount -- "$task_jail/dev/pts"
```

若卸载提示 busy，保留现场并检查该夹具仍存活的子会话；不使用强制或 lazy unmount，也不根据进程名批量终止 sshd。准备失败时同样查看 `jail-path` 与 mount 状态。脚本已通过 `bash -n`，并实测拒绝非 root 运行及已有监听的端口；端口拒绝发生在创建交换目录前，确认未创建目标目录。整合脚本也已完成独立 22477 端口的 prepare → run → Windows PTY / 10 MiB 探针 → 安全停止及 devpts 卸载，详见报告。

## 已授权目标的真实探针

先在目标控制台或可信管理通道核验目标主机的 SHA256 指纹，再运行。以下地址为文档示例，路径及指纹必须替换为已授权测试环境的值。

```powershell
go run . -host 192.0.2.10 -port 2222 -user poc-user `
  -key C:/isolated-poc/client_key `
  -known-hosts C:/isolated-poc/known_hosts `
  -trust-host-key 'SHA256:<已通过独立通道核验的指纹>' `
  -command whoami -pty -timeout 30s `
  -round-trip-parent /tmp
```

`-round-trip-parent` 是明确授权可写的远端父目录。探针在其中创建随机命名的新目录，上传并下载 `中文 binary file.bin` 的 10 MiB 随机二进制数据，比较原始字节、列举目录，最后删除自己创建的文件和目录。已有同名目录会导致失败；清理失败也会导致失败。Windows 受控端须提供其 SFTP 实际可识别的路径，例如经过实测的 `C:/...` 路径，不能直接套用 `/tmp`。

不传 `-round-trip-parent` 时只执行明确指定的远端命令。PTY 仅验证分配和执行，不代表交互终端输入/resize 已完成。非零退出码在 JSON 中返回，同时进程以错误结束。每次连接、命令、传输分别受 `-timeout` 约束；传输超时会关闭本探针的 SSH 连接，远端临时文件可能需要人工检查。

密码认证从 `HOPNODE_POC_SSH_PASSWORD` 读取；加密私钥口令从 `HOPNODE_POC_KEY_PASSPHRASE` 读取。通过会话内安全输入或临时秘密注入提供，并在结束后清除，不写入脚本、命令行参数、仓库、日志。POC 的进程环境并非生产凭据库；Go SSH 密码认证内部可能持有不可擦除的字符串副本，生产实现应接入专用凭据模块。

首次无已知主机记录且未提供正确的 `-trust-host-key`，返回 `SSH_HOST_KEY_UNKNOWN` 并展示待独立核验的指纹；不会写入已知主机文件。批准未知指纹后仅追加到显式指定的独立文件。后续连接省略 `-trust-host-key`。已记录主机的密钥变更始终返回 `SSH_HOST_KEY_CHANGED`，即使提供新的批准指纹也不会覆盖旧记录。探针不读取或修改默认 `~/.ssh/known_hosts`。

对已记录的主机，协商只选择该地址已钉住的密钥类型，避免多密钥服务器优先选择另一种未钉住的类型而造成误报。普通与哈希 known-hosts 名称均由 Go knownhosts 自身匹配；RSA 使用 SHA-2 签名算法，纯 SHA-1 服务器不在这条路径自动降级。协商算法的选择不替代后续公钥字节与撤销校验，真实变更仍拒绝。

## 基础 SSH config

```text
Host isolated-poc
    HostName 192.0.2.10
    User poc-user
    Port 2222
    IdentityFile "C:/isolated-poc/client key"
```

使用 `-config C:/isolated-poc/config -host isolated-poc`；显式传入的 `-user`、`-port`、`-key` 优先于配置，配置内部按首个获取到的标量值生效。支持本 POC 已测试的普通 Host、`*`/`?` 与否定模式及带空格的引号值。配置文件必须显式选择；它不是完整的 OpenSSH 配置解析器。

`Include`、`Match`、`ProxyJump`、`ProxyCommand`、`IdentityAgent`、ssh-agent、多 `IdentityFile`、`~` / `%` 展开及其他未支持指令会给出明确限制或拒绝。真实跳板和 agent 认证尚未测试，不宣称可用。生产版须依据 R06 做完整兼容矩阵、补齐用户目录展开和常见平台差异。

## 取消与后台任务边界

取消普通已建立会话时关闭该 session，返回 `exit_code=-1`；仅超时将 `timed_out=true`。若目标迟迟不响应新 session 的 channel-open，则关闭整个探针连接以释放待处理请求。PTY / exec 请求阶段同样受取消约束。

关闭 SSH session 不保证远端进程及其子孙进程终止，也不保证断线后存活。该探针没有 task ID、恢复查询、进程归属和可靠终止实现。Windows / Linux 的后台任务断线、注销及托管撤销行为必须分别实测后决定 task-runner 策略。
