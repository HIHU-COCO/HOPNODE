# K01 contracts-draft POC — DRAFT

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联：K01 草案准备，PRD F01–F18 的共用接口；尤其 F04/F07/F09–F15/F18 与 A06/A10/A13/A17/A27。

**仅为可执行的候选 schema 证据，不是已冻结的 contracts，不代表 P05/G1/K01 通过，也没有实现任何设备、SSH、目录或 MCP 业务服务。** 正式 A01–A27 状态不因本测试改变。

依据 [K01 候选矩阵](../../../docs/implementation/proposals/K01-contract-draft.md)、[Design §6–10](../../../Design.md)、[PRD](../../../PRD.md)。独立 module 使用 Go 1.27.2 和已有 P03 依赖 `github.com/google/jsonschema-go v0.4.3`；根 go.mod/锁文件、contracts、公共入口及 CI 未改。

## 已实现的候选范围

| 文件 | 可机器读取的范围 |
| --- | --- |
| `schemas/core-api.schema.json` | 2020-12；ID/UTC 时间文本形状、Device/DevicePage、PageRequest、版本/Capabilities、OperationResult、Cancellation、Task，六个 CoreRequest 方法候选 |
| `schemas/errors.schema.json` | Design §10.1 十个必需错误 + 十三个建议错误；统一 message/device/retryable/details；未知结果禁止直接重试标记，details 只允许明确字段 |
| `schemas/mcp-tools.schema.json` | Design §9.3 全部 26 个工具的候选输入与通用结果引用；25 个目标工具必须明确非空字符串 device_id；list_devices 为全局发现 |
| `schemas/directory.openapi.json` | OpenAPI 3.1 DRAFT；Design §7.3 七条路径/八项操作清单；仅列表、详情、显示名/备注 PATCH 三项提供 body/参数候选，其余五项 inventory-only TODO |
| `schemas/directory.schema.json` | Directory managed 设备只读投影、分页/查询/路径参数及只允许显示名/备注的 PATCH；不定义注册材料或撤销协议 |
| `schemas/ipc-events.schema.json` | DeviceChanged/OperationChanged/TaskChanged 三种建议事件，显式目标/关联 ID、类型化 payload，复用既有 unknown/取消约束 |
| `examples/cases.json` | 35 个正常/错误样例（16 正例、19 反例），包括分页、身份来源、版本、请求、取消接受/确认、UNKNOWN_OUTCOME |
| `examples/directory-events-cases.json` | 后续新增 23 个样例（8 正例、15 反例）；静态 Directory 数据和事件声明，不发送 HTTP/IPC |
| `examples/valid/tool-calls.json` | 每个工具一个纯数据样例；示例命令、路径、ID 全部虚构，测试绝不执行它们 |
| `schema_test.go`、`directory_events_test.go` | 嵌入数据，实际解析/resolve/验证；没有外部 HTTP 或文件引用回退 |

JSON Schema 标识使用保留域 `https://hopnode.invalid/draft/`；它是标识符，不是要部署或请求的网站。加载器仅接受五个准确的嵌入 schema ID，其他 host、file URL、路径与 query 均拒绝。OpenAPI 的 `directory.invalid` 同样是保留占位地址，测试不连接它。MCP 文件的根是离线候选 `{name,arguments}`，不是完整 JSON-RPC 帧、Tool 元数据列表或一个 MCP 服务。

六个 CoreRequest 方法为 `GetCapabilities`、`ListDevices`、`GetTask`、`CancelTask`、`GetOperation`、`CancelOperation`；方法输入/结果尚未组成完整请求响应协议，不代表已选定 IPC framing。新增三类事件也仅是结构候选，其余方法/事件仍 TODO。

## 本轮候选决策

- snake_case、非空 ASCII 不透明 ID（最长 128 字符）、store_schema_version 整数和版本形状均为草案选择，未决定生产 UUID/兼容规则；例子中的版本 0/DRAFT 不是正式版本分配。
- 手动设备不冒充 agent 心跳：本草案将 manual_ssh 的 network_status 固定 unknown、agent/identity/last_seen 置 null；其 ssh_status 可独立 ready。为此增加的 network unknown 是候选扩展，需集成负责人确认，不能误写成 Design 原文枚举。
- 结果 unknown 必须携带 operation_id 与 UNKNOWN_OUTCOME/TASK_STATUS_UNKNOWN 错误，retryable=false；不能把同一错误的状态写成 failed 并自动重放。
- 取消 accepted 仍可为 running；只有 confirmed=true、accepted/requested=true 且有 confirmed_at 的候选结果才允许 canceled。Task 同样要求确认字段，且确认取消必须对应 canceled 状态。这只验证声明的一致性，不证明远端实际停止。
- 12 个易重放候选工具输入要求 operation_id：write_file/apply_patch/make_directory/move_path/copy_path/delete_path/upload_file/download_file/exec_command/write_terminal/cancel_task/restore_backup。其他调用的前置关联与幂等规则仍需收敛，不以本测试声称所有副作用已经幂等。
- 成功 error 必须 null；错误结果需统一错误。details 限制仅防止未声明字段进入结构，无法证明 message/data/字符串内容已脱敏，运行时秘密处理仍待实现和审查。

## Directory / IPC 事件后续片段

依据 Design §7.2/7.3/7.4、§10.1 及 PRD F04/F05/F13/F15/F17。按 Plan §10.2 与总集成 ADR-0001 的允许范围，本轮只补独立候选/样例/验证，不推进 K01 冻结或实现业务服务。

- Directory 保留 Design 原始路径占位 `{id}`，映射到稳定 `device_id`；列表候选过滤 `status` 指 network_status，另有 os/name/page_size/cursor。查询对象的 null cursor 是离线 DTO 表示；HTTP query 编码/省略规则、上限、过滤上下文与快照仍待定。
- DirectoryDevice 是 managed 目录只读投影，补有 Design 设备表的 hostname、系统、SSH/网络地址、端口、版本、备注等。未知可选观测使用 null 是草案选择；地址仍为不透明文本，没有声称验证 IP/主机名格式。它与早期 Core Device 子集的最终共用模型仍需收敛。
- `PATCH` 候选仅允许 display_name/note 且至少一项，拒绝 agent_enabled、身份状态、密钥标识和授权版本。这不代替鉴权，也不定义停用或撤销步骤。成功 200 和 default 错误体只是当前草案映射，完整 HTTP status/并发条件/关联 ID 待决。
- 注册、创建材料、心跳、停用和 healthz 仅登记 Design 名称/路径及各自 TODO，不提供虚构请求格式。认证是必需边界；OpenAPI 未选择 securitySchemes 表示未决，**不能理解为匿名访问许可**。撤销/重绑路径仍未指定。
- 三个事件名来自 proposal 建议，Design 仅要求事件 schema：DeviceChanged 载 Core Device 子集；OperationChanged 载 OperationResult；TaskChanged 载 Task。目标/operation/task ID 按事件种类必需，payload 继承 unknown/取消声明约束；内外 ID 相等与实际授权/停止证据仍需运行时校验。
- event_id 仅为候选关联标识，未定义排序、可靠交付、重复消除、重放游标、保留时间或订阅协议；不可把事件当授权或业务状态真值。其余六类事件尚未支持，当前 schema 拒绝它们。

`TestDirectoryOpenAPICandidateInventoryAndReferences` 核对七条路径/八项操作、三项片段/五项 TODO、目标路径必填及 15 处 JSON Schema 引用实际 resolve；**没有使用完整 OpenAPI 规范验证器，也没有生成客户端或发出 HTTP 请求**。`TestDirectoryMetadataPatchRejectsProtectedFields` 另验证六类受保护字段被拒绝。共 58 个固定正反例（原 35 + 新 23），不含目标变体及专门循环检查。

## 检查

在本目录、Go 1.27.2 环境执行：

```text
go mod verify
gofmt -l doc.go schema_test.go directory_events_test.go
go test -count=1 -v ./...
go vet ./...
go build ./...
```

可在 Linux 加 `go test -race -count=1 ./...`。测试全部读取嵌入的静态示例，不访问用户配置/认证，不运行例子里的命令，不连接设备或网络。

`TestAll26ToolsAndExplicitTargetRejections` 对 25 个目标工具分别删除、改错类型、置空、改为 deviceId，实际拒绝 100 个反例；`TestDesignRequiredBusinessErrors` 独立检查 Design 十个要求码；其余测试验证真实 schema 解析、35 个正反例和 12 个缺 operation_id 反例。测试通过不能替代授权判断、ID 归属或业务正确性。

## 明确 TODO / 未覆盖

| 项目 | 当前状态与后续条件 |
| --- | --- |
| 完整 Directory | 三项只读/元数据片段及八项清单已有局部验证；其余五项请求响应、完整 OpenAPI 规范验证、撤销/重绑路径 TODO。HTTPS、注册/管理/设备授权、Headscale 吊销仍依 P01/K04 |
| 完整 IPC 与事件 | 三类事件结构已有局部验证；ConnectionChanged/TransferProgress/TerminalOutput/AgentStateChanged/ServiceStateChanged/MCPStateChanged、其余方法、wire framing、流传输、上限、重订阅、CLI 整数退出码及响应绑定仍 TODO |
| 精确工具结果 | TODO：26 个工具各自 data schema、文本降级与 structuredContent、参数版本/迁移、各工具能力提示；现在仅共用结果引用 |
| 关系与时间校验 | TODO：device/session/task/backup/operation ID 真实归属、内外目标一致、取消与执行证据、兼容范围 min<=max；时间正则仅 UTC 文本形状，不校验日历/时间先后 |
| 分页与文件语义 | TODO：最大页/字节/超时、cursor 的上下文绑定/签名/失效、快照、条件写入及备份可恢复性；本草案不宣称生产资源上限 |
| 幂等与未知结果 | TODO：请求摘要绑定、持久化去重、重启恢复、部分副作用、原操作查询；schema 拒绝不安全声明不等于已经实现 exactly-once |
| 真实能力与身份 | TODO：P01 组网/撤销，P02 Windows sshd/断线任务，P03 Codex 协议取消，P04 SCM/完整平台/服务身份；capabilities 示例全部为虚构 DRAFT 数据 |
| 类型生成/正式门槛 | TODO：Go/TS 共享或生成类型、兼容矩阵和 K04 冻结；本 module 不导出公共业务接口，不改变 P05/G1/K01 状态 |

实际命令与结果见 [report.md](report.md)。
