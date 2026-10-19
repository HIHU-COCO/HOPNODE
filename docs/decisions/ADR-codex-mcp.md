# ADR — 本机 Codex 与 Go MCP stdio P03

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。状态：**部分验证；Windows Codex 握手/实际调用 PASS，Codex 取消未证实；非生产契约冻结**。

关联任务 P03；PRD F06/F07/F14、A06/A07/A24/A27；依据 Plan §4 P03 与 Design 的本机 MCP stdio 方案。

## 方案与已验证事实

采用官方 `github.com/modelcontextprotocol/go-sdk v1.7.0`，在 `tests/pocs/mcp/` 的独立 Go module 中实现 stdio 服务。Go 固定 `1.27.2`。沿用唯一 MCP 名 `hopnode`；夹具唯一工具 `poc_probe` 明确要求 `device_id=poc-local`。夹具不引入生产 SSH、文件或任务接口，公共 contracts 仍由集成负责人冻结。

2026-10-19 Windows amd64 实测：

| 客户端 | 服务/SDK | 实际协商协议 | 结果 |
| --- | --- | --- | --- |
| 官方 SDK Go 客户端 v1.7.0 / `hopnode-poc-test` v1 | `hopnode` v0.0.0-poc.3 / SDK v1.7.0 | `2026-07-28` | 子进程 initialize/list/call/error/cancel/EOF PASS |
| 原始 JSONL 测试客户端 `wire-test` v1 | 同上 | 分别为 `2024-11-05`、`2025-03-26`、`2025-06-18`、`2025-11-25` | 每版本 initialize/list/call/invalid params/EOF PASS |
| 本机 `codex-cli 0.153.4` / `codex-mcp-client` 0.153.4 | 同一夹具及透明 stdio 中继 | `2025-06-18`（请求与应答一致） | 真实模型 echo / 预期 fail / status PASS；Codex 协议取消未证实 |

SDK 测试通过独立 status 请求观察 wait 处理器已开始，再取消调用，最后观察 `active=0,canceled=1`，证明服务端收到取消。旧协议测试逐行解码全部 stdout，验证 stderr 诊断不污染协议；故意错误参数产生 JSON-RPC `-32602`，工具业务失败则作为 `isError` 返回，两者不同。

## Codex 验收边界

最初在新建临时 CODEX_HOME 中注册。该隔离目录的 `codex login status` 输出 `Not logged in`；注册命令输出 `Added global MCP server 'hopnode'.`，get/list 返回 stdio 传输和 `enabled: true`。这只证明隔离注册，也只描述该目录的认证状态，不能推断现有用户环境未登录。

随后在已有授权认证的环境中实际运行 Codex。保留既有 provider/auth，通过 `-c mcp_servers={hopnode={...}}` 为当前进程整体覆盖 MCP 表，不持久化注册；设置 `-a never`、`-s read-only`、`--ephemeral`、空临时工作目录和 `project_doc_max_bytes=0`。原始输出只留在忽略目录，模型只检查允许的事件字段与 marker。config/auth 文件仅做哈希相等检查，三次运行前后均未变；未读取、复制或输出 token。

首轮真实初始化协商成功，echo 因 `requires approval` 被拒绝。夹具原先缺少能力提示；依据其无外部写入/网络/设备行为补充正确的 `readOnlyHint=true`、`destructiveHint=false`、`openWorldHint=false`，不修改既有审批策略或使用 bypass 参数。第二轮实际 echo marker 匹配、fail 返回 `POC_EXPECTED_FAILURE`、status 返回 `active=0,canceled=0`，37 秒退出 0；只有三次 hopnode 工具调用，没有其他工具事件。透明 relay 原字节转发并仅记录握手元数据，直接证明客户端实际请求及服务实际返回均为 `2025-06-18`，不以 SDK 的 `2026-07-28` 代替。

另一次 TTY Ctrl+C 尝试在 wait 已发出后退出 1，自建进程无残留，但没有观察到 `notifications/cancelled` 或服务端取消标记。Codex 协议取消仍未证实，不用 SDK 取消或进程退出代替。中继可选取消 trace 仅记录 wait 与匹配取消的事件名，不记录参数、请求 ID、原因或环境。

Windows 菜单启动路径发现、Ubuntu Desktop 菜单启动路径发现、Ubuntu Codex stdio、真实远程设备、多设备防串目标、生产升级兼容均未验收。保留这些必交付需求。

## 后续约束

1. 继续验证可靠的 Codex 协议取消及 Ubuntu Codex stdio；Windows 与 Ubuntu 分别记录客户端/SDK/协议版本。Windows echo/fail/status 与握手已实证，不再以缺少认证描述为阻塞。
2. P03 记录为部分验证；A06/A24/A27 维持未验收。其他 POC 继续独立推进，P05 依据已验证和剩余阻塞分别评审，不能用本 ADR 宣称 G1/G2 已通过。
3. 通过 P05 后由 K01/K04 冻结生产工具 schema、统一错误与版本策略；M01–M04 通过公共 controller 实现，不复制 SSH/MCP 核心。

证据及命令：[P03-mcp-2026-10-19.md](../implementation/test-results/P03-mcp-2026-10-19.md)。复现方法：[README](../../tests/pocs/mcp/README.md)。
