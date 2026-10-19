# ADR — 本机 Codex 与 Go MCP stdio P03

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。状态：**部分验证；Codex 实调用 BLOCKED；非生产契约冻结**。

关联任务 P03；PRD F06/F07/F14、A06/A07/A24/A27；依据 Plan §4 P03 与 Design 的本机 MCP stdio 方案。

## 方案与已验证事实

采用官方 `github.com/modelcontextprotocol/go-sdk v1.7.0`，在 `tests/pocs/mcp/` 的独立 Go module 中实现 stdio 服务。Go 固定 `1.27.2`。沿用唯一 MCP 名 `hopnode`；夹具唯一工具 `poc_probe` 明确要求 `device_id=poc-local`。夹具不引入生产 SSH、文件或任务接口，公共 contracts 仍由集成负责人冻结。

2026-10-19 Windows amd64 实测：

| 客户端 | 服务/SDK | 实际协商协议 | 结果 |
| --- | --- | --- | --- |
| 官方 SDK Go 客户端 v1.7.0 / `hopnode-poc-test` v1 | `hopnode` v0.0.0-poc.3 / SDK v1.7.0 | `2026-07-28` | 子进程 initialize/list/call/error/cancel/EOF PASS |
| 原始 JSONL 测试客户端 `wire-test` v1 | 同上 | 分别为 `2024-11-05`、`2025-03-26`、`2025-06-18`、`2025-11-25` | 每版本 initialize/list/call/invalid params/EOF PASS |
| 本机 `codex-cli 0.153.4` | 同一编译服务 | **未取得 Codex 协商记录** | 临时 CODEX_HOME 的 add/get/list PASS；模型调用 BLOCKED |

SDK 测试通过独立 status 请求观察 wait 处理器已开始，再取消调用，最后观察 `active=0,canceled=1`，证明服务端收到取消。旧协议测试逐行解码全部 stdout，验证 stderr 诊断不污染协议；故意错误参数产生 JSON-RPC `-32602`，工具业务失败则作为 `isError` 返回，两者不同。

## Codex 验收边界

本次仅在新建临时 CODEX_HOME 中注册。`codex login status` 实际输出 `Not logged in`；没有读取、修改或借用用户现有 Codex 配置及认证。注册命令输出 `Added global MCP server 'hopnode'.`，get/list 返回 stdio 传输和 `enabled: true`。

该证据证明 Codex CLI 可写入和读取隔离注册配置，不能证明 Codex 加载、初始化或模型调用工具。当前 Codex 的协议版本不从 SDK、配置文本或桌面界面推断。隔离环境缺少可用的模型访问认证，因此真实 Codex 工具调用 BLOCKED；这也不代表用户其他 Codex 环境未登录。

Windows 菜单启动路径发现、Ubuntu Desktop 菜单启动路径发现、Ubuntu Codex stdio、真实远程设备、多设备防串目标、生产升级兼容均未验收。保留这些必交付需求。

## 后续约束

1. 在具备授权认证的隔离本机 Codex 中取得工具发现、实际 echo/fail/取消及握手协议证据；Windows 与 Ubuntu 分别记录客户端/SDK/协议版本。
2. P03 维持 BLOCKED；A06/A24/A27 维持未验收。其他 POC 继续独立推进，P05 依据明确阻塞结果评审，不能用本 ADR 宣称 G1/G2 已通过。
3. 通过 P05 后由 K01/K04 冻结生产工具 schema、统一错误与版本策略；M01–M04 通过公共 controller 实现，不复制 SSH/MCP 核心。

证据及命令：[P03-mcp-2026-10-19.md](../implementation/test-results/P03-mcp-2026-10-19.md)。复现方法：[README](../../tests/pocs/mcp/README.md)。
