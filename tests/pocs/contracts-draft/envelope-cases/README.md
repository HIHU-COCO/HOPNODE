# 六方法 IPC envelope 候选 — DRAFT

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联：K01 独立草案、Design §6.3/8.4/9.3/10.1/10.4，PRD F02/F04/F07/F13/F18。

**仅验证静态 JSON、既有结果模型引用和有界请求/响应配对。没有业务服务、Named Pipe/UDS 读写、运行时鉴权或远程调用；不代表 P05/G1/K01 通过。** 本片段只新增自有文件，未修改既有 schema、测试、依赖、公共 contracts 或根文档。

## 消息候选

请求直接引用既有 `core-api.schema.json#/$defs/CoreRequest`，不增改其六个方法或字段。响应候选为：

```json
{
  "message_type": "response",
  "request_id": "draft-request-001",
  "method": "GetTask",
  "outcome": "success",
  "data": {
    "device_id": "draft-device-001",
    "operation_id": "draft-operation-001",
    "task_id": "draft-task-001",
    "status": "unknown",
    "exit_code": null,
    "timed_out": true,
    "observed_at": "2026-10-19T00:00:00Z"
  },
  "error": null
}
```

| outcome | data / error | 含义 |
| --- | --- | --- |
| `success` | 下表具体结果 / null | 此次方法调用成功取得结果；不表示所查询的原操作或任务已经成功 |
| `business_error` | null / 既有 errors.schema.json | 已识别请求的业务拒绝/失败；适用的 request operation/task ID 在响应同名字段关联 |
| `protocol_error` | null / 独立 ProtocolError | IPC JSON/请求/参数/方法/版本问题；不是 MCP JSON-RPC 错误模型，也不复用业务错误码 |

| method | success.data 的实际引用 | 仍存在的边界 |
| --- | --- | --- |
| GetCapabilities | Capabilities | 客户端可接受范围、选定版本和能力语义尚未冻结 |
| ListDevices | DevicePage | 沿用早期 Core Device 子集；不是 Directory 完整设备投影 |
| GetTask | Task | 原任务可为 running/failed/canceled/unknown；输出、执行身份、持久性模型仍不完整 |
| CancelTask | Task + 必需 cancellation.requested=true | accepted 不代表 confirmed；自然完成竞争可返回 succeeded、accepted=false |
| GetOperation | OperationResult | **返回原操作状态，包括 failed/unknown；内部 data 尚未逐工具类型化** |
| CancelOperation | OperationResult + 必需 cancellation.requested=true | 继承既有确认取消约束；不承诺取消撤销副作用 |

所有分支禁止额外响应字段；成功分支 error 必须为 null；错误分支 data 必须为 null。这里只解决六个方法与现有结果的对应，不以任意 object 代替缺失结果：Capabilities、DevicePage、Task 都引用已有具体模型；OperationResult 内部 data 的既有宽泛对象是明确 TODO，未在本片段宣称解决。

协议错误候选码为 `PARSE_ERROR`、`INVALID_REQUEST`、`UNKNOWN_METHOD`、`INVALID_PARAMS`、`IPC_VERSION_UNSUPPORTED`；message 仅为有界安全诊断，禁止回显原始请求、凭据或环境。业务错误复用既有十个必需码及建议码。具体错误映射、文案脱敏和 CLI/MCP 映射仍待定稿。

## 配对规则与范围

标准 JSON Schema 不使用非标准 `$data` 引用来比较两个独立消息。`ipc_envelope_test.go` 中的测试专用检查器先验证 schema，再作以下有界比较：

- 成功/业务错误必须对应有效 CoreRequest，回显完全相同的 request_id 和 method；禁止收到另一请求或另一方法的有效响应。
- 目标结果的 device_id、适用的 task_id 和 Get/CancelOperation 的 operation_id 必须对应请求；原操作内部 error.device_id 也须一致。
- 业务错误目标必须匹配，响应 operation_id/task_id 在本片段只关联请求中已提供的同名字段；不虚构请求没有提供的关联 ID。
- 协议错误中，可从完整、无歧义 JSON 对象识别出的 request_id/method 必须回显。只有不可识别的字段允许 null；即使参数或方法不受支持，也不能抹去已识别关联。
- 损坏、重复字段、超限或不可接受的 JSON 不从部分前缀猜测身份；允许 null/null 的 PARSE_ERROR。可完整解析的请求不能声称 PARSE_ERROR；符合现有请求 schema 的请求仅可在本片段以 IPC_VERSION_UNSUPPORTED 表达协议拒绝。

**CancelTask 的 operation_id 歧义保留：**现有输入尚未明确它是取消动作 ID 还是原任务操作 ID。检查器不会将它与返回 Task.operation_id 强行相等，也不会建立新的业务决定。错误分支同名字段仅原样关联已提供输入。该字段的最终命名、幂等作用域及取消记录查询仍须 K01/K04 决策。

测试专用 JSON 检查器限制为 64 KiB、32 层递归，要求 UTF-8、一个顶层对象、无重复键和无尾随第二值。它处理内存中的固定数据，**不是生产 decoder 或网络读取器**；边界测试不证明网络端在分配内存前正确限长。

## Wire、大小、期限候选与待决

下列仅供集成评审；本片段没有长度头编码/解码或任何 OS 传输证据：

| 事项 | 候选 | 待决/未验证 |
| --- | --- | --- |
| framing | 单条消息采用 4-byte unsigned big-endian 长度前缀，随后恰好该长度的 UTF-8 JSON 对象；请求/响应可多条顺序传输 | 未冻结长度帧、HTTP 或其他方案；拆包、合包、短读、部分写入、并发响应、断连和跨 OS 一律 NOT_RUN |
| 大小 | 长度 1–65536 bytes；超限或无法完成的帧应停止读取并关闭该连接，不能尝试按行重同步 | 这里只验证内存 JSON body 边界；没有实现拒绝前分配上限或 OS reader；事件/大文件另走何种流仍 TODO |
| 建连/握手 | 候选建连和首次协商各 5 秒 | 候选数值未冻结；不同平台调度、慢启动、兼容握手与可读错误尚未实测 |
| 读取 | 候选每个 header 5 秒、每个 body 总计 15 秒，不按每个字节无限续期 | 没有计时器、socket/pipe deadline 或慢读测试；服务端保活/空闲连接策略仍 TODO |
| 方法期限 | 调用期限与上述读取期限分开；长任务不得占用有限 body 读取期限，应有可查询状态 | 默认方法超时、绝对/相对 deadline、取消与自然完成竞争、后台任务存活边界仍 TODO |
| 超时/失联 | 已可能提交的操作保留原 operation/task ID，先查询；不得把超时视为已停止或自动重放 | 持久化去重、重启恢复、真实执行器确认与部分副作用均未实现 |

64 KiB 是此候选检查器上限，并非 A25 容量指标、已冻结产品限制或已经测得的传输吞吐。大文件与终端流的协议、身份保护、生命周期和配额不在本次交付内。

## 文件与离线检查

- `../schemas/ipc-envelope.schema.json`：DRAFT 2020-12 schema；六方法请求引用、三个互斥响应分支及具体结果对应。
- `../ipc_envelope_test.go`：单独的嵌入 schema allowlist、响应 schema 与配对检查、JSON 边界检查；不改既有加载器。
- `pairs.json`：54 个固定配对样例，21 正例、33 反例；19 个反例由响应 schema 拒绝，另外 14 个结构合法但配对/请求语义错误。

在父目录独立 Go module 内执行：

```text
gofmt -l ipc_envelope_test.go
go mod verify
go test -count=1 ./...
go vet ./...
go test -race -count=1 ./...
```

race 在支持的 Linux/Go 工具链执行。测试全部为虚构静态数据，不访问用户设置、秘密、远端或模型服务。加载器只读取确切的嵌入 schema 标识，拒绝任意 HTTP/file/query 引用，不存在外部回退。

2026-10-19 最终实际结果：Windows Go 1.27.2 `go test -count=1 ./...` PASS（6.749s），`go vet ./...`、`go mod verify` PASS；WSL Ubuntu-22.04 / Linux Go 1.27.2 `go test -race -count=1 ./...` PASS（6.228s），无 DATA RACE。`gofmt -l ipc_envelope_test.go` 无输出。既有 Directory、事件、26 工具输入及错误检查在同次 module 测试中继续通过；WSL 仅证明 Linux 离线检查，不是 Ubuntu Desktop/Server 或 OS IPC 验收。

Windows/Ubuntu OS IPC、GUI/SCM、Ubuntu Desktop/Server、MCP/跨网络真调用均 NOT_RUN；没有生产接口或公共契约变更。P05/G1/K01 及正式 A01–A27 状态保持不变。
