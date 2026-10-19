# K01 独立候选 schema 验证记录 — DRAFT

执行者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。任务：K01 独立草案准备；不是 K01/P05/G1 验收结果。关联范围及剩余决策见 [README](README.md) 与 [候选矩阵](../../../docs/implementation/proposals/K01-contract-draft.md)。

## 首轮环境与范围（后续 Directory/事件扩展见末节）

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

首轮尚未完成 Directory OpenAPI 或事件 schema，后续局部扩展见末节。完整 IPC/事件与工具结果、Go/TS 生成类型、ID 归属/鉴权、时间日历/顺序、分页上下文、持久化幂等与取消实际证据、完整平台矩阵仍未完成。字段名/枚举/候选约束均待总集成评审，不得把本目录当成正式公共契约。

Git：本次仅提交本独立目录及其候选矩阵说明，实际 hash 由 Git 历史提供；未推送，未创建 PR。

## 后续 Directory / IPC 事件扩展（2026-10-19）— DRAFT

允许范围再次核对：Plan P05 要求明确 POC 结论，G1 要求关闭下一步阻断风险或有明确授权替代路线；K01 不因此冻结。Plan §10.2 和当前 ADR-0001 允许独立候选/机器验证。本轮只改本目录及 proposal，不修改 contracts、根依赖、公共入口、CI 或正式台账。

依据 Design §7.2/7.3/7.4、§10.1 与 PRD F04/F05/F13/F15/F17：OpenAPI 3.1 清单覆盖七路径/八操作，三项查询/元数据片段有 schema，其余五项保持 inventory-only TODO；Directory 只读字段、查询/分页/目标路径、显示名/备注 PATCH 与三类建议 IPC 事件均可机器验证。未选择身份/撤销策略，未编写业务实现或访问远端/凭据/session/target 数据。

| 本目录实际命令 | 结果 |
| --- | --- |
| `go fmt ./...` | PASS，格式化新增 directory_events_test.go |
| `go mod verify` | PASS，all modules verified；go.mod/go.sum 未改变 |
| `go test -count=1 -v ./...`（Windows Go 1.27.2） | PASS，ok hopnode.local/pocs/contracts-draft 1.871s |
| `go vet ./...`、`go build ./...`（Windows） | PASS，exit 0 |
| `go test -race -count=1 ./...`（WSL Ubuntu-22.04） | PASS，ok hopnode.local/pocs/contracts-draft 6.488s；无 DATA RACE |

新增 23 个静态正反例（8 正例/15 反例），原 35 个继续通过；总计 58 个固定样例，另外保持原有 26 工具/100 错目标/12 缺 operation_id/10 必需错误循环检查。新增六类受保护 PATCH 字段拒绝检查。OpenAPI 检查核对独立 Design 清单、三项片段/五项 TODO、显式必需路径目标及 15 处 schema 引用实际 resolve；**不是完整 OpenAPI 规范验证器**。所有引用仍只从五个嵌入 JSON Schema ID 加载，不访问网络或任意文件。

事件样例覆盖三类正常 payload、缺 device/operation/task、取消未确认、UNKNOWN_OUTCOME 不可直接重试、类型不匹配及未实现事件拒绝；仅检验声明，未证明内外 ID 真实归属、远端停止、可靠事件投递或幂等。

剩余差距：注册/心跳/停用/healthz 的完整请求响应与授权决策、撤销/重绑路径、完整 OpenAPI 规范校验/客户端生成、所有 IPC 方法/事件与订阅、查询 wire 编码与有界/上下文分页、并发条件、具体 HTTP 错误映射及真实身份/设备/平台验证。DirectoryDevice 与早期 Core Device 子集的最终共用模型仍需收敛；未知字段 null 与 HTTP 200/default 映射仅为候选。

Windows 离线自动检查 PASS；Ubuntu Desktop GUI 真机、Ubuntu Server CLI 真机、MCP/跨网络/目录/事件真调用均 NOT_RUN。WSL race 不替代这些验收，P05/G1/K01 和 A01–A27 状态未改变。
