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
| `examples/cases.json` | 35 个正常/错误样例（16 正例、19 反例），包括分页、身份来源、版本、请求、取消接受/确认、UNKNOWN_OUTCOME |
| `examples/valid/tool-calls.json` | 每个工具一个纯数据样例；示例命令、路径、ID 全部虚构，测试绝不执行它们 |
| `schema_test.go` | 嵌入数据，实际解析/resolve/验证；没有外部 HTTP 或文件引用回退 |

JSON Schema 标识使用保留域 `https://hopnode.invalid/draft/`；它是标识符，不是要部署或请求的网站。加载器仅接受三个准确的嵌入 schema ID，其他 host、file URL、路径与 query 均拒绝。MCP 文件的根是离线候选 `{name,arguments}`，不是完整 JSON-RPC 帧、Tool 元数据列表或一个 MCP 服务。

六个 CoreRequest 方法为 `GetCapabilities`、`ListDevices`、`GetTask`、`CancelTask`、`GetOperation`、`CancelOperation`；方法输入/结果尚未组成完整请求响应协议，不代表已选定 IPC framing。其余候选 IPC 方法/事件仍 TODO。

## 本轮候选决策

- snake_case、非空 ASCII 不透明 ID（最长 128 字符）、store_schema_version 整数和版本形状均为草案选择，未决定生产 UUID/兼容规则；例子中的版本 0/DRAFT 不是正式版本分配。
- 手动设备不冒充 agent 心跳：本草案将 manual_ssh 的 network_status 固定 unknown、agent/identity/last_seen 置 null；其 ssh_status 可独立 ready。为此增加的 network unknown 是候选扩展，需集成负责人确认，不能误写成 Design 原文枚举。
- 结果 unknown 必须携带 operation_id 与 UNKNOWN_OUTCOME/TASK_STATUS_UNKNOWN 错误，retryable=false；不能把同一错误的状态写成 failed 并自动重放。
- 取消 accepted 仍可为 running；只有 confirmed=true、accepted/requested=true 且有 confirmed_at 的候选结果才允许 canceled。Task 同样要求确认字段，且确认取消必须对应 canceled 状态。这只验证声明的一致性，不证明远端实际停止。
- 12 个易重放候选工具输入要求 operation_id：write_file/apply_patch/make_directory/move_path/copy_path/delete_path/upload_file/download_file/exec_command/write_terminal/cancel_task/restore_backup。其他调用的前置关联与幂等规则仍需收敛，不以本测试声称所有副作用已经幂等。
- 成功 error 必须 null；错误结果需统一错误。details 限制仅防止未声明字段进入结构，无法证明 message/data/字符串内容已脱敏，运行时秘密处理仍待实现和审查。

## 检查

在本目录、Go 1.27.2 环境执行：

```text
go mod verify
gofmt -l doc.go schema_test.go
go test -count=1 -v ./...
go vet ./...
go build ./...
```

可在 Linux 加 `go test -race -count=1 ./...`。测试全部读取嵌入的静态示例，不访问用户配置/认证，不运行例子里的命令，不连接设备或网络。

`TestAll26ToolsAndExplicitTargetRejections` 对 25 个目标工具分别删除、改错类型、置空、改为 deviceId，实际拒绝 100 个反例；`TestDesignRequiredBusinessErrors` 独立检查 Design 十个要求码；其余测试验证真实 schema 解析、35 个正反例和 12 个缺 operation_id 反例。测试通过不能替代授权判断、ID 归属或业务正确性。

## 明确 TODO / 未覆盖

| 项目 | 当前状态与后续条件 |
| --- | --- |
| Directory OpenAPI | **TODO，未创建/未验证。** 需将 Design §7.3 八端点及撤销/重绑候选转为 OpenAPI；HTTPS、注册/管理/设备授权、Headscale 吊销仍依 P01/K04 |
| 完整 IPC 与事件 | TODO：其余方法、事件/schema、wire framing、流传输、上限、重订阅、CLI 整数退出码及响应与具体方法绑定 |
| 精确工具结果 | TODO：26 个工具各自 data schema、文本降级与 structuredContent、参数版本/迁移、各工具能力提示；现在仅共用结果引用 |
| 关系与时间校验 | TODO：device/session/task/backup/operation ID 真实归属、内外目标一致、取消与执行证据、兼容范围 min<=max；时间正则仅 UTC 文本形状，不校验日历/时间先后 |
| 分页与文件语义 | TODO：最大页/字节/超时、cursor 的上下文绑定/签名/失效、快照、条件写入及备份可恢复性；本草案不宣称生产资源上限 |
| 幂等与未知结果 | TODO：请求摘要绑定、持久化去重、重启恢复、部分副作用、原操作查询；schema 拒绝不安全声明不等于已经实现 exactly-once |
| 真实能力与身份 | TODO：P01 组网/撤销，P02 Windows sshd/断线任务，P03 Codex 协议取消，P04 SCM/完整平台/服务身份；capabilities 示例全部为虚构 DRAFT 数据 |
| 类型生成/正式门槛 | TODO：Go/TS 共享或生成类型、兼容矩阵和 K04 冻结；本 module 不导出公共业务接口，不改变 P05/G1/K01 状态 |

实际命令与结果见 [report.md](report.md)。
