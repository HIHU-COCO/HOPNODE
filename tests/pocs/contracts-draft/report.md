# K01 独立候选 schema 验证记录 — DRAFT

执行者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。任务：K01 独立草案准备；不是 K01/P05/G1 验收结果。关联范围及剩余决策见 [README](README.md) 与 [候选矩阵](../../../docs/implementation/proposals/K01-contract-draft.md)。

## 环境与范围

- Windows amd64，Go 1.27.2，根 `.tools/go/bin/go.exe`。
- WSL Ubuntu-22.04 linux/amd64，根 `.tools/go-linux/go/bin/go`；仅 Linux 解析/并发检查，不视为 Ubuntu Desktop/Server 正式环境。
- 独立 `hopnode.local/pocs/contracts-draft` module；固定 google/jsonschema-go v0.4.3。go.sum 包含其传递测试依赖 google/go-cmp v0.7.0 的校验值。
- schema 实现仅在 `tests/pocs/contracts-draft/`，同时补充原候选矩阵 §12 的后续验证链接和历史语境。JSON Schema 为 2020-12，全部标 DRAFT，嵌入加载器没有网络或任意本机路径回退。

## 实际结果

| 本目录实际命令 | 结果 |
| --- | --- |
| `go fmt ./...` | PASS，Go 源格式化 |
| `go mod tidy` | PASS，仅本独立 module 的 go.sum 更新 |
| `go mod verify` | PASS，all modules verified |
| `go test -count=1 ./...`（Windows） | PASS，ok hopnode.local/pocs/contracts-draft 1.881s |
| `go vet ./...`、`go build ./...`（Windows） | PASS，exit 0；仅 schema 测试包，没有可运行业务服务 |
| `go test -race -count=1 ./...`（WSL） | PASS，exit 0；ok hopnode.local/pocs/contracts-draft 3.771s，无 DATA RACE |

文档卫生检查首次 `git diff --cached --check` 报告两个经 PowerShell 更新的 JSON 文件混合 CRLF 及生成文件 EOF 空行，已统一为 LF/单尾换行；复核 exit 0、无输出。该问题仅涉及文本格式，JSON 的已验证内容未改变。

首轮 Windows `go test -count=1 -v ./...` 也实际 PASS（1.906s），当时 32 个正反例；随后补充 Task 取消确认的正反例和十个必需错误的独立测试，34 个样例复测 PASS（Windows 1.852s、WSL race 3.717s）。最终审查再补 Task 已确认取消不能仍声明 running 的约束/反例，最终 35 个样例复测结果见上表。

机器检查范围：三个 schema 均可 resolve 且跨文件引用只从嵌入内容加载；35 个业务/状态/分页/版本/请求正反例符合预期；Design 全部 26 工具正常输入有效；25 个目标工具的 100 个错误目标变体被拒绝；12 个易重放候选输入缺 operation_id 被拒绝；十个 Design 必需业务错误均被接受。P03 的 poc_probe 不在候选 26 工具中，被拒绝。

## 平台与正式验收

| 范围 | 结果/边界 |
| --- | --- |
| Windows 本机离线 schema 自动检查 | PASS，以上解析/验证/静态命令 |
| Windows GUI/SCM/真实设备 | NOT_RUN，本任务没有业务/安装实现 |
| Ubuntu Desktop GUI 真机 | NOT_RUN，WSL schema 测试不替代 |
| Ubuntu Server CLI 真机 | NOT_RUN，本任务没有 CLI/部署实现 |
| MCP / 跨网络真调用 | NOT_RUN，仅验证静态候选 JSON；未运行 MCP 服务或执行示例命令 |
| P05/G1/K01 与 A01–A27 | 未验收/未改变状态，不能由候选 schema PASS 推导通过 |

尚未完成：Directory OpenAPI、全部 IPC/事件与工具结果、Go/TS 生成类型、ID 归属/鉴权、时间日历/顺序、分页上下文、持久化幂等与取消实际证据、完整平台矩阵。字段名/枚举/本轮候选约束均待总集成评审，不得把本目录当成正式公共契约。

Git：本次仅提交本独立目录及其候选矩阵说明，实际 hash 由 Git 历史提供；未推送，未创建 PR。
