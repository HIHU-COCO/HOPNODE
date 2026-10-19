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
