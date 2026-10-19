# P03 — Go MCP stdio 协议 POC

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联：F06、F07、F14，A06、A07、A24、A27。

**当前结果：协议夹具及 Windows 子进程测试 PASS；P03 整体验收 BLOCKED。** 本机 Codex CLI 注册已在独立临时配置目录复现，但该目录未登录，尚未取得 Codex 模型发起工具调用的证据。SDK 客户端成功不等于 Codex 实调用成功。

这是独立 Go module，固定官方 `github.com/modelcontextprotocol/go-sdk v1.7.0`，Go `1.27.2`；不修改生产接口。服务的 MCP implementation name 和 Codex 注册名均为 `hopnode`，版本为 `0.0.0-poc.3`。

## 工具与边界

唯一工具 `poc_probe` 要求明确 `device_id="poc-local"`，这个 ID 表示本地测试夹具，不是真实设备。输入 `action` 为：

| action | 行为 |
| --- | --- |
| `echo` | 返回可选 `message` 和结构化 `device_id` / `operation_id` / `status` |
| `fail` | 返回包含 `POC_EXPECTED_FAILURE` 的工具错误 |
| `wait` | 等待 MCP 请求取消 |
| `status` | 返回当前等待数及已取消数，用于证明服务端收到取消 |

没有 SSH、远端命令、文件操作、设备目录或后台任务能力。`operation_id` 仅供夹具关联，不承诺持久化。日志写 stderr，stdout 保留给换行分隔的 JSON-RPC。

## 自动检查

在本目录、Go 1.27.2 可用的终端中执行：

```powershell
go mod verify
gofmt -l main.go main_test.go
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

后续必须在有权使用且具备模型访问能力的隔离本机 Codex 中，请求调用 `poc_probe`，例如参数 `{"device_id":"poc-local","action":"echo","message":"P03 Codex"}`。保存实际工具调用和返回以及客户端协商协议的脱敏证据；继续验证 fail 与取消。缺这些证据时保持 P03 BLOCKED，A06/A24/A27 不验收通过。

完整实测报告：[P03-mcp-2026-10-19.md](../../../docs/implementation/test-results/P03-mcp-2026-10-19.md)。技术结论：[ADR-codex-mcp.md](../../../docs/decisions/ADR-codex-mcp.md)。
