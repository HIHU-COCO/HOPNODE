# ADR-0001：V1 技术 POC 联合评审

日期：2026-10-19；作者/集成负责人：HIHU-COCO。任务：P05；状态：**ACCEPTED — 条件施工路线**。P05/G1 的施工出口为 CONDITIONAL，原有实测缺口仍开放，生产契约尚未冻结。

## 已有实证

| POC | 结论与证据 | 尚未关闭的门槛 |
| --- | --- | --- |
| P01 网络 | 独立Headscale0.29.5/TS1.98.4单机userspace的TLS正负例、双节点/合成TCP/策略切换/精确节点expire/旧key拒绝及清理通过；只读探针另记。见ADR-network.md | 两机/不同网络、公网DERP和独立边界阻断P2P后业务回退未验；单机disco不作业务链路证据 |
| P02 SSH | Windows客户端→隔离Ubuntu与独立LAN Ubuntu22.04：命令、PTY、中文/空格路径、10MiB SFTP及清理通过；修复并复验单pin密钥算法协商。见 tests/pocs/ssh/report.md | Windows系统sshd、完整Ubuntu矩阵、异常断线长任务和高级config兼容仍未验证 |
| P03 MCP | 官方 Go SDK stdio/错误/取消/旧协议及Linux race通过；Codex0.153.4真实echo/预期错误/status通过，实际协商2025-06-18。见 P03-mcp-2026-10-19.md | Codex CLI取消尝试未取得协议通知证据；Ubuntu真实Codex调用尚缺 |
| P04 桌面/IPC | Windows NSIS/真实WebView Named Pipe/卸载通过；Server2022 CI真实SCM启停/ACL负例/删除通过；独立Ubuntu22.04 X11工程deb安装/系统desktop入口/显示刷新/内核peer PID核验UDS/卸载通过，user systemd也通过。见desktop-runtime.md及SCM报告 | Win10/11服务、24.04及Server、完整视觉/托盘/Wayland、注销重启和正向跨身份授权矩阵尚未齐备 |

## 保持的架构与候选界限

继续采用 React / TypeScript / Ant Design / Tauri 2 薄 Rust、Go Controller/CLI/MCP、SSH/SFTP 与现成 Tailscale/Headscale。当前没有证据要求替换技术栈。已验证版本固定于 packaging/manifests/versions.md；Headscale/DERP 兼容版本仍未冻结。

快速 SSH 的候选流程为：显式目标地址与认证 → 检查 Host Key → 用户明确确认未知指纹 → 测试命令；变化的 Host Key 必须失败。POC 未读取默认 SSH 配置，支持的基础字段与明确不支持的 Include/Match/ProxyJump/ssh-agent 见 SSH 报告。托管组网仍需隔离注册、设备身份与撤销 POC，不能借用用户现有网络宣布通过。

用户态 Windows IPC 已验证当前 SID 限权；Linux UDS 已验证目录 0700、socket 0600。系统 agent 与用户 controller 的授权调用者、注销后身份、凭据持有者和多用户隔离尚未冻结。关闭托管受控必须拒绝新请求、撤销 HopNode 管理的凭据并清理可归属会话，不能关闭用户独立 SSH；该要求保留，当前没有产品实现或验收证据。

`poc.v0` 和 `poc_probe` 是隔离试验，不是公共 IPC/MCP 契约。正式唯一 CLI/MCP 名仍是 `hopnodectl` / `hopnode`。官方 SDK 客户端协商结果不能替代 Codex 协商结果。

## 授权与出口决定

2026-10-19，用户在获知原顺序因缺环境而阻塞后续开发、调整会增加后续返工风险、全平台与 27 项验收继续保留的说明后，明确指示：**“那你重新调整，调整完之后告诉我。”** 据此接受[条件施工路线](../implementation/proposals/P05-conditional-development.md)。该授权调整开发顺序；未授予新的测试设备/系统修改权限，也未授权发布。

Plan G1 原已允许“明确授权的替代路线”。现将 **G0/G1 和 P05 的开发出口记为 CONDITIONAL**：已有工程与 CI、四项 POC 的真实结论及缺口清单可追溯，允许进入 K01。T03、P01–P04 仍按实际证据保留 BLOCKED/IN_PROGRESS，原任务未全部满足时不勾选完成；CONDITIONAL 不表示实测 PASS 或 DONE。

### 串行施工及业务入口

1. K01 从 READY 开始，由集成负责人按[十二项缺口](../implementation/proposals/K01-gap-review.md)补齐完整 IPC/事件、26 个 MCP 工具结果、Directory、调用身份、取消/幂等/未知结果和共享或生成类型。已有 `tests/pocs/contracts-draft/` 仍是候选，须经完整校验与审查后迁入正式契约；不能直接宣布冻结。
2. K01 完成并合入集成基线后，依次实施 K02、K03、K04。K04 冻结版本/兼容、身份及双模式约定；性能指标区分已批准阈值与待测候选，A25 的未决项继续 BLOCKED，不虚构数值或性能结论。
3. 只有 K01–K04 的完成判据、共享基线、CI 和目录责任全部满足，G2 才可通过。此后 N/R/M/U/S/C 按各自模块依赖开展实现；缺少 P01–P04 的完整实测不再一概禁止实现，但具体任务的真机完成判据及对应验收仍必须满足。未满足时保持 REVIEW/BLOCKED，不能置 DONE。
4. 开发过程中同步补齐环境和 POC；工程回环、WSL、CI Server OS、辅助 WebSSH、模拟适配器均不替代正式平台或业务链路。发现真实失败推翻架构、身份或撤销假设时，暂停受影响接口/模块，先更新 ADR、契约迁移和复测方案；无关模块可继续。

### 继续开放的验证与责任

| 范围 | 负责人 | 保留的验证和解除条件 |
| --- | --- | --- |
| 测试矩阵 T03 / ENV-01、ENV-03 | 集成与环境管理者；平台 S、远程 R 配合 | Windows 10/11 服务及受控 SSH、两版本 Ubuntu Desktop/Server、完整身份与注销/重启测试；先补授权设备，逐环境记录实测 |
| 组网 P01 / ENV-02 | 网络 N；集成协调环境 | 独立 Headscale/DERP、不同网络及阻断 P2P 后的真实业务回退；正式兼容版本保持候选，I04/A03/A04/A25 等按证据验收 |
| SSH/任务 P02 / TECH-01 | 远程 R、平台 S | Windows sshd、配置兼容、异常断线、task-runner 和可靠取消；manual/managed 分开报告，未知结果不自动重放 |
| MCP P03 / TECH-03 | 接入 M | 配置/进程归属、实际 Codex 协议取消、Ubuntu 真调用；SDK、进程退出和客户端取消分别取证 |
| 安全身份、托管停用与撤销 | 集成冻结策略；N/R/S/M 实现 | OS peer/授权调用者矩阵、设备级身份、专用凭据及可归属会话撤销；策略未定不得冻结对应 K01/K04 内容，能力未实现或未获实际授权时明确拒绝 |
| 性能、恢复与正式包 | 集成；S/C/U 及各模块配合 | K04 指标审批和待测项分列；I11/I12 与 L06 完成实际规模/恢复/升级验证；许可证及发布风险继续记账 |

### 验收、发布与回退

G3–G6 的真实集成和验收门槛保留。G5 通过后构建 L01–L05 候选包，再以同一候选完成安装/升级/卸载验收及 L06；L07 仍需全部 F01–F18 和 A01–A27 通过、缺陷关闭及正式发布条件齐全。当前正式验收为 **0/27 PASS**，没有发布授权。

每个模块报告已验证能力和待验证能力；网络/身份/版本不满足时返回明确错误或不可用状态，生产路径不得静默回落 Mock。安全策略未解决时暂停对应契约冻结；运行时能力缺失时拒绝请求，不关闭用户既有 SSH 或切换其 Tailscale 网络。

回退仅针对本项目的受影响提交/契约及工程构建：标 REWORK、通过后续提交修复并重新运行相关检查；需要不兼容变更时提供版本迁移。不得自动回滚用户系统配置、清理用户数据或扩大测试授权。本轮文档调整不撤销既有测试事实，也不把缺环境造成的 NOT_RUN 改为通过。
