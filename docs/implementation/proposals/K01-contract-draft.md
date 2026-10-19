# K01 统一契约候选矩阵 — DRAFT

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。状态：**DRAFT，仅供总集成评审；不是公共契约，不代表 K01/P05/G1 完成。**

依据：[PRD V1.4](../../../PRD.md)、[Design V1.3](../../../Design.md)、[Plan V1.3](../../../Plan.md)、[AGENTS V1.1](../../../AGENTS.md)。初稿核对主树 `5e7e07c` 的需求与证据；后续局部候选 schema 验证见 §12。本文件不修改 `contracts/`，不实现业务模块。后续由总集成负责人决定并迁入 K01 产物。

**当前施工状态更新（2026-10-19）**：用户已授权条件路线，G0/G1/P05为CONDITIONAL、K01 READY，见 [ADR-0001](../../decisions/ADR-0001-v1-stack.md)。本文件的候选设计仍是DRAFT；§11–§14保留先前局部验证轮次及当时状态，不能代替当前施工授权或完整契约验收。

标记约定：**依据**表示原文已要求的业务语义或名字；**建议**表示本草案提出的字段、方法名或映射，尚未决定；**待决**表示不可当作既定实现的技术/身份/兼容边界。除 Design 明确写出的 `GetCapabilities`、`ListDevices`、MCP 名称及 Directory 草案路径外，下文 IPC 方法和事件名均为建议。即使来自 Design，其正式 wire schema 仍未冻结。

当前机器草案的完整性缺口见 [K01 审查清单](K01-gap-review.md)，其中十二项仍需补齐；字段计数与局部样例通过不表示完整契约已满足。

## 1. 评审与施工边界

Plan P05 要求四项POC有明确真实结论，G1允许有明确授权的替代路线。当前按Plan §2.4及ADR-0001进入条件施工出口，可依据[十二项缺口](K01-gap-review.md)完成正式K01，不再限于独立草案。冻结仍需完整结构、语义、类型及机器校验和审查，不能用既有局部schema直接宣称K01完成；K01→K04严格串行，G2通过后才进入业务线。

当前局部证据：P01 已有独立 Headscale/Tailscale 单机用户态 TLS、合成流量、策略、节点过期和旧 key 拒绝的机制实证，仍缺不同网络与强制 DERP 业务证据，见 [单机报告](../../../tests/pocs/network-isolated/LOCALHOST-REPORT.md)。P02 已有 Windows 客户端→隔离 Ubuntu 与已授权独立 LAN Ubuntu 22.04 的命令/PTY/中文空格路径/10MiB SFTP 及清理 PASS，仍缺 Windows sshd、异常断线长任务及完整版本矩阵，见 [SSH 报告](../../../tests/pocs/ssh/report.md)。P04 已有 Windows 当前用户 NSIS 安装/真实 GUI IPC/卸载，以及临时 Server 2022 CI 的真实 SCM 两轮启停、ACL 负例和删除；独立 Ubuntu Desktop 22.04 X11 三轮 GUI/内核 peer PID 核验 UDS/独立 Core 通过，含实际工程 deb 安装、系统桌面入口启动及卸载，见 [桌面报告](../test-results/pocs/desktop-runtime.md) 和 [SCM 实证](../../../tests/pocs/service-lifecycle/SCM-CI.md)。这仍不覆盖 Windows 10/11 服务、24.04/全新 Server、注销重启和完整身份矩阵。P03 的 Windows Codex 0.153.4 已真实调用 echo/预期 fail/status，实际协议 `2025-06-18`；客户端协议取消未证实，不能以 SDK 取消替代，见 [P03 报告](../test-results/P03-mcp-2026-10-19.md)。

P05已按下一开发步骤区分契约决策与待补实测，责任、允许范围、失败处理及复测条件见ADR-0001。授权只调整施工顺序，单机/WSL证据仍不替代正式验收。身份、撤销、取消及未知结果等安全语义须先明确，再冻结对应结构；没有因授权自动解决技术缺口。

| 未来 K01 产物 | 本文件提供的候选内容 | 初稿文档轮次不做（后续见 §12） |
| --- | --- | --- |
| `contracts/ipc/core-api.schema.json` | 共用模型、抽象请求/结果、方法/事件和分页 | 决定 wire framing、发布兼容承诺、实现 controller |
| `contracts/ipc/errors.schema.json` | 必需错误及建议扩展、未知结果/重试语义 | 将错误码包装成已经可用的能力 |
| `contracts/mcp/tools.schema.json` | Design §9.3 全部 26 个工具的输入/结果映射 | 将 `poc_probe` 或其全只读属性迁入生产 |
| `contracts/directory/openapi.yaml` | Design §7.3 路径、设备模型和待决身份操作 | 选择未经 P01 验证的 Headscale 私有接口/认证格式 |
| `docs/api.md` | GUI/CLI/MCP 共用语义、正常/错误/取消样例要求 | 宣称已生成类型或运行契约验证器 |

## 2. 共用字段、标识与版本

JSON 对外建议统一 `snake_case`；Go/TypeScript 的属性命名由生成规则映射，避免 `deviceId` 与 `device_id` 在不同入口成为不同字段。Design 部分表格使用 camelCase，因此这项须总集成明确决定。

| 模型/字段 | 依据或建议语义 | 待决/约束 |
| --- | --- | --- |
| `device_id` | 依据：稳定设备身份，不以 IP/主机名为主键；Design §7.2 提出 UUID | 实际格式由 K01 决定；示例 `dev-002`/`poc-local` 不是已冻结格式；手动设备也须稳定本地 ID，重装关联需有效授权 |
| `operation_id` | 依据：同一操作在 GUI/CLI/MCP 统一关联；用于幂等及断线后查询 | 建议变更类调用由客户端在发送前生成并保留；格式、保留期、作用域及持久化由总集成决定；不是“恰好执行一次”的证明 |
| `task_id` | 依据：长任务独立、可查询/取消，始终绑定目标 | 由目标 task-runner 或受管理执行器返回；与 `operation_id` 不混用，不能在客户端断线后另造同一任务 |
| `request_id` | 建议：单次请求/响应相关 ID，与可跨重试的 operation ID 分开 | wire framing 未定；不可把新的 request ID 当作新的实际执行 |
| `session_id` / `backup_id` / `transfer_id` | 建议：终端/备份/传输资源的稳定引用 | 均与调用身份及 `device_id` 绑定；不可只凭 ID 转到其他目标 |
| `created_at` / `updated_at` / `observed_at` | 建议 UTC RFC3339 时间；依据要求保留观测时间 | 区分事件发生、保存与观测时间；缓存不能伪称实时；单调排序及精度待决 |
| `app_version` | 依据 Design §10.4：应用版本 | 不替代 IPC/数据库/工具 schema 版本 |
| `ipc_major` / `ipc_minor` | 依据：本机 API 兼容版本 | 数值、兼容范围和 minor 增量规则在 K04 冻结；POC `poc.v0` 不作正式版本 |
| `store_schema_version` | 依据：本地数据迁移版本 | 与 API 版本分离，迁移/备份/回滚待 K04 与平台验证 |
| `mcp_tool_schema_version` | 依据：工具输入/结果版本 | 与 MCP initialize 的协议版本分离；SDK 版本也不是 wire 版本 |
| `directory_api_version` | 依据：目录 API 版本；Design 路径草案 `/api/v1` | 是否另设版本字段/header、兼容策略待 K04 |

建议抽象请求包含 `request_id`、IPC 兼容版本、`method`、`params`，以及适用的 `device_id`/`operation_id`。这是 schema 设计输入，**尚未选择 HTTP/JSON、JSON-RPC 或长度帧**；MCP 自身使用其标准协议，不受此 IPC framing 待决项影响。

统一业务结果沿用 Design §9.3 的 `device_id,operation_id,status,data,error`；有任务时增加 `task_id`。建议与以下规则一起评审：成功 `error=null`，失败返回统一错误；服务级/全局列表没有单一目标，`device_id` 的省略或 null 须显式定义，不能填当前 UI 所选设备。针对设备的读写/连接/任务/终端工具均要求确定目标，不能沿用隐含“当前设备”。

## 3. 设备、状态与能力模型

| 模型 | 必要字段/值（来源） | 语义及未决点 |
| --- | --- | --- |
| Device | `device_id,display_name,hostname,os_type,os_version,roles`（Design §7.2）；`source=manual_ssh\|managed,capabilities`（§3.4） | roles 允许 controller+agent；首版 Windows/Linux，Design 列出的 macos 不代表首版交付 |
| 地址/版本 | `tailnet_ip,ssh_host,ssh_port,agent_version`（§7.2） | 地址可变；手动设备的 SSH Host 别名与解析后地址建议分字段；主控本地配置不自动发布给 Directory |
| 在线/连接 | `network_status=online\|pending\|offline`，`ssh_status=ready\|unavailable\|auth_failed\|unknown`，`connection_type=direct\|relay\|peer_relay\|unknown`（§7.2） | 三者相互独立；不能以虚拟 IP 推测 direct；手动目标无心跳时如何表达“未检测/不适用”待决，不能假造 online |
| 受控/撤销 | `agent_enabled`；身份状态 `active\|suspended\|revoked\|retired`；授权版本、撤销时间、设备密钥标识（§3.6/7.4） | agent 开关、网络在线、身份状态不能合成一个 bool；密钥标识不是秘密；撤销结果需有落实证据 |
| 观测/展示 | `last_seen_at,observed_at,note`（§7.2） | last_seen 对手动设备可不适用；默认心跳 10s、30s pending、60s offline 是可配置设计值，非实测保证 |
| Operation / Task | `device_id,operation_id,task_id?,status,created_at,updated_at,error?`（Plan K01、Design §8.4/9.3） | 状态建议共用 `queued\|running\|succeeded\|failed\|canceled\|unknown`；只在实际支持的路径返回 task ID |
| 能力结果 | 建议 `name,availability=available\|unavailable\|unknown,reason_code?,constraints?` | 由实际 Core/agent/平台适配器报告；未实现/未验证的恢复、续传、托管撤权不能默认 available；最终形状待决 |

`GetCapabilities` 依据 Design §10.4 必须是首次兼容调用：候选输入为客户端应用版本/可接受 IPC 范围；结果包含服务版本、兼容范围与真实能力。建议 major 不兼容时返回 `VERSION_INCOMPATIBLE` 并禁止高风险调用，minor 差异按能力门控，不进行静默语义降级。最小握手错误的可读格式也应在 K01 明确，以免连错误都需依赖未完成的协商。

## 4. Controller IPC 方法与事件候选矩阵

方法名除上述 Design 明确项外均为建议。GUI、CLI、MCP 调用同一 controller，不在 Rust 或各入口重写 SSH/MCP 业务。系统服务管理/agent 端接口是否与用户 controller 共用端点仍受 P04 身份决策约束。

| 功能 / PRD | 候选方法 | 输入与结果重点 |
| --- | --- | --- |
| 兼容/诊断 F02/F18 | `GetCapabilities`, `GetStatus`, `RunDiagnostics` | 客户端版本；组件状态、逐层诊断、脱敏错误、观测时间；健康不等于有权限操作 |
| 设备视图 F04/F05 | `ListDevices`, `GetDevice`, `AddSSHDevice`, `UpdateDevice`, `RemoveLocalDevice` | 过滤/分页或 device ID；手动 host/port/username/profile；显示名/备注；删除本地记录不等于远程撤销 |
| 连接/身份 F05/F08 | `TestConnection`, `ConnectDevice`, `DisconnectDevice`, `ConfirmHostKey`, `SetConnectionProfile` | 明确目标；受保护的凭据引用、Host Key 首次信任信息；指纹变化必须失败；秘密不进入普通 Device DTO/记录 |
| 模式/托管 F01/F03/F17 | `GetRole`, `SetRole`, `EnrollDevice`, `SetAgentEnabled`, `RevokeDevice`, `RebindDevice` | quick/managed 与 controller/agent/both 分维度；组织注册材料仅指定受保护输入；停用结果分请求已接受与真正落实 |
| 服务 F02/F16 | `GetServiceStatus`, `StartService`, `StopService` | 明确组件及作用域；实际运行身份、权限要求、观察结果；普通 IPC 请求不成为任意提权通道 |
| Codex F06/F07 | `DetectCodex`, `GetMCPStatus`, `InstallMCP`, `TestMCPConnection` | 实际实例/路径/版本及配置所属环境；备份/增量变更；已注册、需重载、真实调用分别报告；不修改既有审批策略 |
| 文件 F09 | `ListFiles`, `SearchFiles`, `ReadFile`, `WriteFile`, `ApplyPatch`, `MakeDirectory`, `MovePath`, `CopyPath`, `DeletePath` | 与下节 MCP 相同的目标、路径、版本条件和备份语义；编码、文件类型、冲突与平台差异不能由 UI 猜测 |
| 传输 F10 | `UploadFile`, `DownloadFile`, `GetTransfer`, `CancelTransfer` | 目标与本地/远端路径或受保护流引用；进度、字节数、校验/结果、是否实际支持续传；具体流协议待决 |
| 命令/任务 F11/F13 | `ExecCommand`, `ListTasks`, `GetTask`, `CancelTask`, `GetOperation`, `CancelOperation` | shell/cwd/env/timeout、前台/后台运行意图；输出/退出码/timeout，task ID；Get/CancelOperation 带原目标和 operation ID，前台/搜索等操作也可关联取消；未知结果先查询，取消不保证撤销副作用 |
| 终端 F12 | `OpenTerminal`, `ReadTerminal`, `WriteTerminal`, `ResizeTerminal`, `CloseTerminal` | device+session、shell/cwd、列/行、数据 offset/序号；断线不假装保留原会话；Resize 是 Design §8.3 要求，MCP 清单未单列 |
| 记录/恢复 F15 | `ListBackups`, `RestoreBackup`, `GetOperationHistory`, `GetLogs` | 目标/operation/task ID、分页、时间过滤；可恢复备份才提供恢复；日志/摘要去除秘密 |

大文件/终端数据按 Design §10.1 走有界流或块，不能在 JSON DTO 中无限展开文件。流引用必须绑定本机实际身份和目标；是否使用文件描述符、临时文件或流端点，生命周期、上限和权限由总集成决定。

建议事件共用 `event_id,event_type,observed_at,device_id?,operation_id?,task_id?,data`，其中涉及目标必须包含相应 ID。候选类型为 `DeviceChanged`、`ConnectionChanged`、`OperationChanged`、`TaskChanged`、`TransferProgress`、`TerminalOutput`、`AgentStateChanged`、`ServiceStateChanged`、`MCPStateChanged`。事件顺序、断线重订阅、offset/游标、重复/丢失、保留上限均待决；未实现可重放事件时断线后通过查询恢复，不能声称一定收到每个事件。日志事件不能代替授权判定或业务状态真值。

## 5. MCP 26 个工具候选输入/结果

唯一服务名 `hopnode`。名称依据 Design §9.3；字段细化是建议，必须由 K01 定稿并以 schema/样例校验。每个目标操作都显式携带 `device_id`，session/task/backup ID 不能隐含改换目标。`list_devices` 是全局发现例外；`list_tasks` 与历史记录是否另允许全局聚合须单独决策，默认候选矩阵采用显式设备范围。

| Tool | 建议输入（除 list_devices 外均加 device_id） | `data` 候选与关键语义 |
| --- | --- | --- |
| `list_devices` | source/roles/status/os/name 过滤、分页 | `items: Device[]`、分页与观测时间；返回手动+托管视图 |
| `get_device` | device ID | `device: Device`；未知 ID 明确错误 |
| `connect_device` | profile 引用、timeout | 真实连接/SSH 状态、Host Key 处理结果；首次信任不可静默接受 |
| `disconnect_device` | 连接/会话范围 | 仅自身管理连接的关闭结果；不承诺停止独立 SSH |
| `list_files` | path、分页、过滤/排序 | 有界 `items`，名称/类型/大小/修改时间/权限信息；字段不可获得时明确未知 |
| `search_files` | path、query、递归/类型条件、分页 | 路径匹配、进度/是否完整；递归可取消 |
| `read_file` | path、offset/length 或文本范围、encoding | 有界内容/流引用、编码/大小/版本标识、是否截断；不能把二进制误当文本 |
| `write_file` | path、content/流引用、encoding、版本条件、备份选项、operation ID | 实际写入/版本、备份引用和差异摘要；并发变化报冲突 |
| `apply_patch` | path、patch、版本条件、备份选项、operation ID | 应用结果/新版本/差异/备份；无部分写入保证时说明实际结果 |
| `make_directory` | path、是否递归、operation ID | 创建结果；已存在且类型不同须区分 |
| `move_path` | source_path、destination_path、覆盖/备份选项、operation ID | 实际移动结果；跨文件系统能力按平台，结果未知不重放 |
| `copy_path` | source_path、destination_path、递归/覆盖/备份选项、operation ID | 实际复制结果；SFTP 不自动保证服务端复制 |
| `delete_path` | path、递归/永久删除意图、备份选项、operation ID | 默认备份区/实际删除结果；显式永久删除不虚构可恢复 |
| `upload_file` | local_path/受保护流引用、remote_path、覆盖/备份选项、operation ID | transfer ID、字节数/校验/状态；local 指 controller 所在机及其授权身份 |
| `download_file` | remote_path、local_path/受保护流引用、覆盖选项、operation ID | transfer ID、字节数/校验/状态；本地写权限需实际判定 |
| `exec_command` | command、shell、cwd、env、timeout、运行模式建议、operation ID | stdout/stderr/exit_code/timed_out；实际支持后台运行才返回 task ID；env 秘密不进入记录 |
| `open_terminal` | shell、cwd、rows/columns | session ID、实际 shell、初始状态；执行身份来自已认证 SSH |
| `read_terminal` | session ID、offset/最大字节数 | 有界输出、next_offset、是否关闭；文本/字节编码待决 |
| `write_terminal` | session ID、输入数据/编码、operation ID | 已接收/已发送量与状态；不能把接收确认当命令完成 |
| `close_terminal` | session ID | 关闭结果；不自动宣称其启动的独立任务也结束 |
| `list_tasks` | 状态/时间过滤、分页 | Task[] 与分页；明确实际数据源和可查询范围 |
| `get_task` | task ID、输出 offset/上限 | 状态、stdout/stderr/退出码可用性、截断/next_offset；断线时允许 unknown |
| `cancel_task` | task ID、operation ID | 请求接收和真正停止分开；未观察到停止不返回 canceled |
| `list_backups` | path/operation/time 过滤、分页 | backup ID、来源操作、保留/可恢复状态；不泄露备份秘密路径 |
| `restore_backup` | backup ID、目标 path/覆盖与版本条件、operation ID | 恢复/冲突/新备份结果；缺备份不能承诺恢复 |
| `get_operation_history` | operation/task/time/status 过滤、分页 | 有界操作摘要、目标、结果、关联 ID；去除凭据与完整秘密环境 |

标准 MCP 参数/schema/协议错误与业务错误须区分：无效工具参数使用 SDK/协议规定错误，业务失败通过工具错误结果携带上述统一业务错误；具体 structuredContent 与文本降级形式、协议版本兼容测试由 K01 决定。不能因 CLI 进程 exit 0 就把工具业务失败当成功。

生产工具按实际行为单独设置能力提示：文件写入/删除/命令/终端写入/恢复绝不能继承 P03 `poc_probe` 的全只读声明。纯查询也须按实际副作用审查。Codex 审批由客户端决定，HopNode 不修改其既有安全设置。

## 6. Directory 路径及身份候选矩阵

以下前八项来自 Design §7.3 草案；原文占位 `{id}` 建议在 OpenAPI 参数中统一命名 `device_id`，路径文字是否调整由总集成决定。Directory 只服务 managed 设备，不转发 SSH 文件/命令，不保存主控用户的私钥。

| 方法/路径 | 请求/结果候选 | 必须保留的身份与状态边界 |
| --- | --- | --- |
| `POST /api/v1/enrollments` | 部署/设备限制、有效期意图 → 一次性登记引用及受保护注册材料 | 管理端部署授权；无个人账号不等于匿名创建材料；算法/TTL/返回渠道待 P01/K04 |
| `POST /api/v1/devices/register` | 注册材料、设备身份公有元数据、Device 基本信息 → 稳定身份/授权版本/注册结果 | 一次性材料使用与销毁、重放拒绝、旧撤销身份不可借旧 token 自动重入 |
| `POST /api/v1/devices/{id}/heartbeat` | 自身 device ID/授权版本/agent 状态与观测 → 接收时间/目录状态/撤销指示候选 | 身份必须绑定同一 device ID；10/30/60s 为可配置设计，不把客户端时间当可信授权时间 |
| `GET /api/v1/devices` | status/os/name/role 过滤、分页 → managed Device[] | 主控需具备部署授权；默认业务视图可只列已注册且开启受控设备，撤销审计不能丢失 |
| `GET /api/v1/devices/{id}` | device ID → Device | 按授权可见性读取，避免向无权调用者泄露设备是否存在；具体错误映射待决 |
| `PATCH /api/v1/devices/{id}` | 仅显示名/备注等允许字段、并发版本条件建议 → 更新后 Device | 不允许借 PATCH 修改身份密钥、授权版本或伪造停用落实状态 |
| `POST /api/v1/devices/{id}/disable` | 停用意图/operation ID → 请求与落实状态 | 最终由 agent/托管入口执行；目录标记与实际凭据/会话撤销结果分别表示 |
| `GET /healthz` | 最小目录健康结果 | 不返回设备清单、凭据、内部配置；健康端点匿名可见范围待部署决策 |
| 建议：设备撤销接口（路径待决） | device ID/operation ID → 身份、网络授权、托管凭据和可归属会话的逐项结果 | 来源 Design §7.4/PRD F17/A26；组织授权管理操作，不能只删除目录记录 |
| 建议：重装关联接口（路径待决） | 新登记证明/待关联旧设备引用 → 关联/拒绝与审计 | 不以主机名/IP 作为授权；身份恢复、历史保留和轮换策略待决 |

设备级认证的凭据格式、HTTPS 信任根/证书策略、主控与管理端授权材料、Headscale 注册与撤销 API、离线撤销传播/缓存时效均未冻结；本草案不提前选择某个 bearer/JWT/mTLS 方案。认证材料只在受保护输入通道出现，不在普通响应、事件、日志或示例公开。错误与分页沿用统一语义；HTTP status 对业务错误码的具体映射列入 K01 待决，不与业务 `status` 混淆。

## 7. 分页、过滤及并发条件

建议设备、文件、搜索、任务、备份、操作记录共用 `page_size` 与不透明 `cursor`；结果为 `items,next_cursor,observed_at`，末页 `next_cursor=null`。`total_count` 仅在能准确获得时可选，不要求遍历全目录才能返回第一页。默认/最大页大小、有界响应字节数/超时均需总集成选定，当前不编造承载承诺。

cursor 必须绑定实际调用身份、目标（适用时）、方法、过滤与排序上下文，不可跨设备复用；变更过滤后应从首页开始。cursor 是分页位置，不是授权凭据。快照一致性、有效期、失效错误及目录变化时是否可能重复/遗漏仍待决；不支持快照时必须说明，不能宣称列表稳定快照。offset 仅用于有序字节/日志流，不能不加说明替代集合分页。

文件写/补丁/覆盖和目录备注更新建议携带 `expected_version` 或 `if_match` 条件；具体取 ETag/hash/mtime+size 以及无版本条件时的行为待决。版本变化返回 `FILE_CONFLICT`（目录冲突使用何码待决），允许按 PRD 读取最新、重试补丁或明确覆盖，不强制全局锁或逐条审批。

## 8. 统一错误与入口映射

沿用 Design §10.1 的 `code,message,device_id,retryable,details`。建议补充关联 `operation_id/task_id/request_id` 或从外层结果继承，并验证内外目标一致；全局错误的 device ID 可空规则须定稿。`message` 可本地化但程序只依赖稳定 code；details 为允许字段，禁止塞原始 stderr、密码、私钥、token、完整秘密 env 或未脱敏配置。

| code | 来源/语义 | 重试/取消限制 |
| --- | --- | --- |
| `DEVICE_OFFLINE` | Design 必含：设备离线 | 查询新状态；不能重放已可能提交的写操作 |
| `NETWORK_UNREACHABLE` | Design：路径不可达 | 与 SSH 服务未就绪、设备离线分开 |
| `SSH_NOT_READY` | Design：SSH 服务/握手未就绪 | 不等同认证失败；不能承诺 Windows 已装 sshd |
| `SSH_AUTH_FAILED` | Design：目标 SSH 认证失败 | 可在修正身份后重试；示例 retryable=true 不等于自动猜测/重复认证 |
| `SSH_HOST_KEY_CHANGED` | Design：已知指纹变化 | 不自动忽略或覆盖信任记录 |
| `PERMISSION_DENIED` | Design：本机 IPC/系统/目标账号权限不足 | 不回退到 agent 系统身份隐式提权 |
| `MCP_NOT_CONNECTED` | Design：MCP 未连通/需接入 | 区分注册存在、客户端未重载和真实工具失败 |
| `TASK_STATUS_UNKNOWN` | Design：暂不能可靠查询任务状态 | 保留 task ID/目标，查询恢复；不新启动任务 |
| `FILE_CONFLICT` | Design：文件已变化或条件写冲突 | 返回安全冲突上下文，读取最新/明确覆盖 |
| `SERVICE_NOT_RUNNING` | Design：目标本机组件未运行 | 与权限不足、版本不兼容、崩溃分开 |
| `UNKNOWN_OUTCOME` | 建议：请求可能提交且结果无法确定 | 最高优先为查询原 operation/task；不自动重放写/删/安装/命令 |
| `INVALID_ARGUMENT`, `NOT_FOUND` | 建议：非法业务输入/未知资源 | MCP 协议参数校验与业务错误的映射需明确 |
| `VERSION_INCOMPATIBLE`, `CAPABILITY_UNAVAILABLE` | 建议：版本/真实能力不满足 | 稳定解释并提供升级或限制说明，不能静默 Mock |
| `HOST_KEY_UNTRUSTED`, `SSH_CONFIG_UNSUPPORTED` | 建议：首次信任待处理/配置指令不支持 | 不新增逐条业务审批；仅履行既有 Host Key/兼容边界 |
| `DEVICE_REVOKED`, `AGENT_DISABLED` | 建议：已撤销身份/托管入口停用 | 新请求拒绝；不自动重新注册绕过撤销 |
| `CANCELED`, `TIMEOUT` | 建议：确认取消/达到期限 | TIMEOUT 不推出远端已停止；不确定时改为 UNKNOWN_OUTCOME/unknown |
| `OPERATION_CONFLICT`, `CURSOR_INVALID` | 建议：同 operation ID 不同请求/分页上下文无效 | 不悄悄生成新 operation ID 或换目标继续 |

`retryable` 仅表示纠正条件后可能重试，不是自动重放许可；建议在统一重试策略中先检查“是否已提交/结果是否已知/是否有持久化去重/操作是否只读”。网络恢复后不得单凭 retryable=true 重发写操作。

CLI 需要稳定非零退出类别（参数、连接、认证、权限、版本、未知结果等），具体整数未定；远端命令非零退出码、CLI 传输错误和工具业务错误需分层表达。GUI 显示同一 code 对应处理入口，MCP 返回同一业务错误，不各自改造为另一套成功状态。

## 9. 取消、结果未知与幂等

1. **收到请求不是完成。** 建议在结果中独立表达取消已请求/已接受/已确认；只有实际执行器确认停止才设 `status=canceled`。具体字段形状待定，不将“取消已接受”塞成 `succeeded` 而让调用方误读为原任务成功。
2. **取消通道与业务目标分离。** MCP 请求取消需映射到同一 controller 操作；`CancelTask(device_id,task_id)` 是显式业务取消，两者不能靠断开 UI/stdio 代替。客户端离线是否取消前台操作、后台任务是否继续须按运行模式明确。
3. **超时/断线不证明失败。** 若请求可能已被提交，返回 `UNKNOWN_OUTCOME` 并保留原 operation/task ID；任务状态可为 `unknown`。查询确认后再转入实际 succeeded/failed/canceled，不能为方便 UI 默认 failed 并自动重试。
4. **取消不会倒转副作用。** 已写的文件/运行的安装命令不能因取消而自动回滚；默认备份只对实际已保留且可恢复内容有效。部分传输/临时文件的清理与结果必须报告。
5. **去重需要服务端证据。** 建议以调用身份+device ID+operation ID 绑定请求摘要；重复同请求查询/返回同结果，不同内容拒绝。保留时长、重启持久化、并发竞争、未知记录和清理策略须明确。task-runner 缺失/未验证时不宣传可靠断线存活或 exactly-once。
6. **证据分层。** P03 SDK 已证明服务端取消计数；真实 Codex Ctrl+C 尝试只证明自建进程退出，无取消协议证据。MCP 工具超时、CLI 中断、SSH session 关闭、系统 Job 清理、远端 task-runner 停止分别测试，互不替代。

## 10. 冻结前的最小决策清单

| 决策 | 当前可写入草案的边界 | 仍依赖的证据/责任 |
| --- | --- | --- |
| D01 G1 与施工授权 | 四项真实结论归档，用户已授权CONDITIONAL施工出口，K01 READY | 总集成按ADR-0001推进完整K01及后续串行K02–K04；安全策略须明确，原实测及正式验收继续补齐 |
| D02 IPC framing/版本/命名 | snake_case、抽象 envelope、GetCapabilities、错误和样例候选 | K01 确定 framing/方法事件名；K04 冻结兼容范围/身份版本；不抄 POC 为生产 |
| D03 用户 controller / 系统 agent | controller 用其登录/专用身份凭据；agent 不读取他人私钥，端点分身份 | P04：服务SID/UID、授权调用者、多用户、注销/重启和Server专用身份仍需完整矩阵；Server2022 CI的SCM局部实证不替代Win10/11 |
| D04 托管身份与撤销 | managed/manual 分开；拒绝新托管请求和逐项撤销结果 | P01/P04：部署授权、一次性注册、Headscale 吊销、缓存失效、专用 SSH 凭据及可归属会话；不能只翻 agent_enabled |
| D05 SSH 兼容/目标绑定 | device ID→profile→session；Host Key 变化拒绝；基础字段要求保留 | P02：Windows sshd、Ubuntu 两版本、Include/Match/ProxyJump/agent；不支持需明确，不扩大已验证范围 |
| D06 Task/取消/幂等 | 同 ID 查询、unknown 状态、无自动重放 | P02/P03/P04：远端断线持久化、实际执行身份、服务重启和 Codex 协议取消；正式工具不能继承夹具只读属性 |
| D07 大文件/分页/并发 | 有界流/分页候选、文件版本条件、备份范围 | K01 决定具体上限、cursor/stream 权限与生命周期；P02 验证各服务器能力；A25 阈值由 K04 审核 |
| D08 Directory 部署/恢复 | HTTPS+设备身份的抽象约束、恢复证明与审计 | P01/K04：TLS/凭据格式、恢复/轮换、旧 token 拒绝、离线撤销策略；不引入个人账号或复杂审批平台 |

## 11. 完整验证清单（初稿文档轮次未执行；后续仅局部验证见 §12）

- schema/OpenAPI 可由固定验证器读取，正例/反例及 Go/TypeScript 生成类型对齐；具体工具链由总集成固定。
- 26 个 MCP 工具逐项包含确定目标（全局发现例外）及统一结果；跨 device/session/task/backup ID 混用必须失败；缺失目标/错误类型均有反例。
- 分页验证末页、无效 cursor、换过滤/换设备、观测时间和截断；文件冲突测试覆盖重新读取、补丁、明确覆盖及实际备份可恢复条件。
- 同 operation ID 重复/并发、不同请求冲突、提交后断线、进程重启后查询、未知结果不自动重放；取消接受与取消确认分别验证。
- 新 GUI/CLI/MCP + 旧 Core、旧客户端 + 新 Core、Agent/Directory 不兼容、不可逆 store 迁移均有明确失败/降级样例，不只比应用版本字符串。
- 身份反例包括不同用户/服务、未授权 IPC、已撤销设备、旧注册材料重放、停用后新请求；Windows Service 与 Ubuntu 正式平台另做实测，WSL/Mock 不替代。
- 原始日志/事件/错误 details 的秘密隔离检查；失败、跳过、NOT_RUN 分别记录，所有正式 A01–A27 状态仍由独立验收台账管理。

初稿文档轮次核对：从 Design §9.3 表格提取的 26/26 工具名、§10.1 的 10/10 必需错误码均已覆盖；检查 Markdown 表格列数一致，`git diff --check` 通过。该轮仅为文档覆盖检查，不是 schema/实现的测试结果。

初稿轮次仅交付此 DRAFT 文档，未运行 schema 验证、代码生成、平台 E2E 或网络验收；后续的局部候选机器验证另记于 §12。以上完整清单尚未全部执行。该轮没有新增生产接口、冻结版本、业务实现或验收通过声明；当前施工状态以文首授权更新、ADR-0001和台账为准。

## 12. 首轮局部候选 schema 验证 — DRAFT（后续见 §13）

独立 [contracts-draft POC](../../../tests/pocs/contracts-draft/README.md) 将少量共用模型、统一错误、六个抽象 CoreRequest 及 26 个 MCP 工具输入转成 JSON Schema 2020-12；仅解析嵌入静态数据，没有生产接口/服务或远程执行。实际结果见 [验证记录](../../../tests/pocs/contracts-draft/report.md)。

- Windows Go 1.27.2 的 module 校验、测试（1.881s）、vet、build PASS；WSL Ubuntu-22.04 的 `go test -race -count=1 ./...` PASS（3.771s）。WSL 结果只证明 Linux schema 自动检查。
- 35 个正反例（16 正例、19 反例）、26 个工具正常输入通过预期验证；25 个目标工具缺失/错类型/空值/camelCase 共 100 个反例被拒绝；12 个易重放候选工具缺 operation_id 被拒绝；10 个 Design 必需错误均可表达。
- UNKNOWN_OUTCOME、取消接受/确认、manual/managed 身份来源和分页末页只验证候选结构声明的一致性，不证明实际执行/鉴权、结果查询、远端停止、cursor 绑定或幂等持久化。
- 该轮尚未创建 Directory OpenAPI 或事件 schema；后续局部进展见 §13。完整 IPC/事件、逐工具 data 结果、Go/TS 类型生成、关系/运行时与兼容安全测试仍 TODO；§11 完整清单未全部执行。候选 ID 格式、manual network unknown 扩展、store 版本类型和其他约束仍待评审。

该轮局部机器验证PASS不构成P05/G1通过、K01冻结、K02或六条业务线启动许可；后续条件施工许可由文首记录的用户授权和ADR-0001单独产生，正式验收仍按Plan判据独立判断。

## 13. Directory / IPC 事件后续片段 — DRAFT

依 Design §7.2/7.3/7.4、§10.1 和 PRD F04/F05/F13/F15/F17，独立 POC 新增 [Directory OpenAPI 候选](../../../tests/pocs/contracts-draft/schemas/directory.openapi.json)、对应 [Directory schema](../../../tests/pocs/contracts-draft/schemas/directory.schema.json) 及 [IPC 事件 schema](../../../tests/pocs/contracts-draft/schemas/ipc-events.schema.json)。依据 Plan §10.2 与 ADR-0001，范围仍限于候选/样例机器验证，没有修改公共 contracts 或启动业务实现。

| 片段 | 本轮候选和验证 | 未决/未实现 |
| --- | --- | --- |
| Directory 清单 | Design 七路径/八操作全部登记；GET 列表、GET 详情、PATCH 名称/备注共三项有参数/body schema；其余五项 inventory-only TODO | 注册材料、心跳身份、停用落实、healthz 可见范围、撤销/重绑路径；不选择 securitySchemes/凭据或 Headscale 策略 |
| Directory 设备 | managed 只读投影补 hostname/系统/地址/SSH端口/版本/备注；分页末页与查询形状；PATCH 拒绝六类受保护字段 | Core 子集与目录完整投影收敛；null/地址格式、HTTP query 编码、cursor 绑定/上限、版本条件、HTTP 错误映射与关联 ID |
| IPC 事件 | DeviceChanged/OperationChanged/TaskChanged 三个建议名，必需目标及适用 operation/task ID，复用既有 payload/unknown/取消声明约束 | 其余六类事件、framing/订阅/重放/顺序/保留/重复语义；内外 ID 相等、授权及实际停止只能由后续运行时证明 |
| 实际检查 | 新增 23 正反例（8/15）；15 处 OpenAPI schema 引用离线 resolve；原 35 例及工具检查继续 PASS；Windows test/vet/build、WSL race PASS | 没有完整 OpenAPI 规范验证器、客户端生成、HTTP/IPC 服务、真实身份/撤销/事件推送或平台验收 |

OpenAPI 的未声明认证方案是明确的未决项，不是匿名访问授权；inventory-only 也不是可执行端点或完成标记。事件 ID 不承诺排序/交付/回放。具体命令与剩余差距见 [POC README](../../../tests/pocs/contracts-draft/README.md) 和 [实际报告](../../../tests/pocs/contracts-draft/report.md)。P05/G1/K01 及正式验收状态不因本轮变化。

## 14. 六方法响应与配对后续片段 — DRAFT

独立 [envelope 草案](../../../tests/pocs/contracts-draft/envelope-cases/README.md) 为现有六个 CoreRequest 补充响应候选：请求 ID/方法回显、逐方法引用已有结果、成功/业务错误/协议错误分支，以及测试专用的有界跨消息配对检查。54 对固定样例包含 21 正例、33 反例；六方法均有明确成功映射。Windows 主树 test/vet/mod verify PASS；Linux race 结果见该片段记录。

查询 GetTask/GetOperation 成功只表示取得原状态，不把原任务 failed/unknown 改为成功。静态配对检查覆盖错误请求 ID、方法、目标及已有资源 ID；UTF-8、重复字段、尾随值、字节/深度边界只针对内存 JSON。未实现 Named Pipe/UDS 传输、真实调用者授权或生产关联检查。

该片段不收敛完整 IPC 或逐工具结果。OperationResult.data、CancelTask operation_id 歧义、精确协议错误分类、实际版本协商、全部流/事件/期限与共享类型仍待补齐。长度前缀和限额仅为候选，无 OS 读写证据，不改变 P05/G1/K01 状态。
