# P03 — Go MCP stdio 协议 POC

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联：F06、F07、F14，A06、A07、A24、A27。

**当前结果：Windows Codex 0.153.4 真实握手及 echo / 预期 fail / status 调用 PASS。** 实际协议为 `2025-06-18`。Codex 协议取消尚未证实，Ubuntu Codex 与正式产品验收仍未完成。首次隔离 CODEX_HOME 未登录的结果仅限该隔离环境；随后使用现有授权认证完成了实际调用。

这是独立 Go module，固定官方 `github.com/modelcontextprotocol/go-sdk v1.7.0`，Go `1.27.2`；不修改生产接口。服务的 MCP implementation name 和 Codex 注册名均为 `hopnode`，版本为 `0.0.0-poc.3`。

## 工具与边界

唯一工具 `poc_probe` 要求明确 `device_id="poc-local"`，这个 ID 表示本地测试夹具，不是真实设备。输入 `action` 为：

| action | 行为 |
| --- | --- |
| `echo` | 返回可选 `message` 和结构化 `device_id` / `operation_id` / `status` |
| `fail` | 返回包含 `POC_EXPECTED_FAILURE` 的工具错误 |
| `wait` | 等待 MCP 请求取消 |
| `status` | 返回当前等待数及已取消数，用于证明服务端收到取消 |

没有 SSH、远端命令、文件操作、设备目录或后台任务能力。工具据此声明 `readOnlyHint=true`、`destructiveHint=false`、`openWorldHint=false`；这些只描述本夹具，不适用于未来生产工具。`operation_id` 仅供夹具关联，不承诺持久化。日志及固定取消观察标记写 stderr，stdout 保留给换行分隔的 JSON-RPC。

## 自动检查

在本目录、Go 1.27.2 可用的终端中执行：

```powershell
go mod verify
gofmt -l main.go main_test.go stdio-relay/main.go stdio-relay/main_test.go
go test -count=1 -v ./...
go vet ./...
go build -trimpath -o "$env:TEMP\hopnode-poc.exe" .
```

`TestSDKSubprocess` 启动真实子进程，验证初始化、tools/list、中文结构化输出、业务错误、缺失/错误目标、schema 错误、服务端取消计数、stderr 和 EOF 退出。SDK 客户端本次实际协商 `2026-07-28`。

`TestLegacyWireAndPureStdout` 通过原始 JSONL 客户端分别发起 `2024-11-05`、`2025-03-26`、`2025-06-18`、`2025-11-25` 初始化，验证协商回显、工具调用、无效参数的 `-32602`、所有 stdout 帧均为 JSON-RPC、EOF 后无多余输出并正常退出。这些客户端是测试夹具，不是对应版本的 Codex。

在具备 C 工具链的 Linux 环境追加 `go test -race -count=1 -v ./...`。2026-10-19 的 WSL Ubuntu-22.04 复测 PASS；此前检测到的测试日志缓冲区 `ReadFrom` 锁绕过已修复。WSL 协议测试不替代 Ubuntu Desktop/Server 真机验收，完整失败及修复记录见测试报告。

## 隔离复现 Codex 注册

使用独立临时 `CODEX_HOME`，保持用户现有 Codex 配置和认证文件原样；不复制已有凭据。先将本 POC 编译到新建临时目录，再在同一 PowerShell 会话执行：

```powershell
$pocDirectory = Join-Path ([IO.Path]::GetTempPath()) ('hopnode-p03-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $pocDirectory | Out-Null
$pocBinary = Join-Path $pocDirectory 'hopnode-poc.exe'
go build -trimpath -o $pocBinary .
if ($LASTEXITCODE -ne 0) { throw 'POC build failed' }
$previousCodexHome = $env:CODEX_HOME
try {
    $env:CODEX_HOME = Join-Path $pocDirectory 'codex-home'
    New-Item -ItemType Directory -Path $env:CODEX_HOME | Out-Null
    codex --version
    codex login status
    codex mcp add hopnode -- $pocBinary
    codex mcp get hopnode --json
    codex mcp list --json
} finally {
    $env:CODEX_HOME = $previousCodexHome
}
```

`Added global MCP server 'hopnode'.` 中的 global 相对于当时的独立 `CODEX_HOME`。`enabled: true` 只证明注册配置可读。stdio 的 `auth_status: unsupported` 不代表协议失败，也不能证明已连接。

以上流程仅验证注册。2026-10-19 已另外使用现有授权认证完成下述实际调用，不复制 token，也不修改用户的 MCP 注册。

## 真实 Codex 调用与握手取证

`stdio-relay/` 是标准库透明中继，按原字节转发两个方向的 JSONL（包括 CRLF 和 EOF 尾帧），stdout 不输出日志。默认 trace 只记录 initialize 的客户端名、客户端版本、请求和返回协议版本；不记录工具参数、输出、环境、配置或认证。可选 `--trace-cancellation` 仅额外记录 `wait_call` 和匹配请求的 `wait_cancel_notification` 事件名，省略 ID、参数及取消原因。trace 文件必须尚不存在。

可在本目录编译后复现以下核心命令。`artifacts/` 已忽略；原始 stdout/stderr 只在本机保留，不直接向模型或提交内容展示。调用需要本机已有授权认证及模型访问能力；不要把认证复制到隔离目录。

```powershell
$evidence = Join-Path $PWD ('artifacts/' + [guid]::NewGuid().ToString('N'))
$emptyCwd = Join-Path ([IO.Path]::GetTempPath()) ('hopnode-p03-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $evidence, $emptyCwd | Out-Null
go build -o (Join-Path $evidence 'hopnode-poc.exe') .
if ($LASTEXITCODE -ne 0) { throw 'POC build failed' }
go build -o (Join-Path $evidence 'hopnode-relay.exe') ./stdio-relay
if ($LASTEXITCODE -ne 0) { throw 'Relay build failed' }
$relay = (Join-Path $evidence 'hopnode-relay.exe').Replace('\','/')
$target = (Join-Path $evidence 'hopnode-poc.exe').Replace('\','/')
$trace = (Join-Path $evidence 'handshake.jsonl').Replace('\','/')
$mcpOverride = "mcp_servers={hopnode={command='$relay',args=['--target','$target','--trace','$trace'],enabled=true,startup_timeout_sec=15,tool_timeout_sec=20}}"
$prompt = 'Call only hopnode poc_probe in order: device_id=poc-local action=echo message=HOPNODE_P03_CODEX_20261019; action=fail with the same device_id (intentional error, do not retry); then action=status with the same device_id. Do not use any other tool or inspect files. Finish with HOPNODE_P03_CODEX_20261019 COMPLETE.'
$prompt | codex -a never -c $mcpOverride -c project_doc_max_bytes=0 exec --ephemeral --json -s read-only --skip-git-repo-check --color never -C $emptyCwd - 1> (Join-Path $evidence 'stdout.raw.jsonl') 2> (Join-Path $evidence 'stderr.raw.txt')
```

本次实测由隐藏 PowerShell 工作进程执行同一 CLI 参数，外层限制 180 秒并仅允许清理本次进程树。未使用 `--ignore-user-config`，保留现有 provider/auth；进程级整表 `mcp_servers` 覆盖只提供 `hopnode`。模型工作目录为空临时目录，`project_doc_max_bytes=0`，提示只允许三个夹具调用。调用前后仅计算 config/auth 文件哈希并比较相等，不向模型读取或输出文件内容。完整原始记录留在主树忽略目录 `.tools/artifacts/p03-codex-f3857df1da0a48c7849476ae1ef31d3d/`。

首轮未声明只读元数据时，真实握手成功，但 echo 返回 `requires approval`，后续调用未运行。修正夹具真实元数据后第二轮 37 秒退出 0，echo marker 匹配、fail 含 `POC_EXPECTED_FAILURE`、status 返回 `active=0,canceled=0`，恰好三个 hopnode 调用，无其他工具事件。两轮 config/auth 哈希均未变。

另一次受限 TTY 测试观察到 wait 调用后发送 Ctrl+C：会话退出 1、夹具/中继无残留，但未观察到取消通知或服务端取消标记。**这不能证明 Codex 协议取消**，该项仍未验收；SDK 取消测试 PASS 单独保留。A06/A24/A27 不能由本 POC 直接设为通过。

完整实测报告：[P03-mcp-2026-10-19.md](../../../docs/implementation/test-results/P03-mcp-2026-10-19.md)。技术结论：[ADR-codex-mcp.md](../../../docs/decisions/ADR-codex-mcp.md)。
