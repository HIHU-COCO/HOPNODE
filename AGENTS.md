# AGENTS.md — 跃点 HopNode Codex 仓库执行约束 V1.1

**文档用途**：提供可被 Codex 及其他代码智能体直接读取的仓库级施工规则。它不是功能需求或产品验收结果。  
**创建者**：ChatGPT（本地 Git 实际提交人以 `git config user.name` 为准）  
**创建日期**：2026-10-18  
**依据**：`PRD.md` V1.4、`Design.md` V1.3、`Plan.md` V1.3<br/>
**修订日期**：2026-10-19；V1.1 按用户授权与 ADR-0001 同步条件施工路线，保留交付及验收要求。<br/>
**状态**：执行中；实际代码、实测和任务状态以 `docs/implementation/status.md` 为准。

## 1. 工作开始顺序

1. 检查 `git status --short`、当前分支、操作系统、文件列表、可用工具链及测试设备；保护已有未提交修改。
2. 依次读取 `PRD.md`、`Design.md`、`Plan.md`，再阅读本文件及当前模块的 `contracts/`；必要时读取 `docs/implementation/status.md`。
3. 遵守优先级：**PRD 定义业务范围和验收；Design 定义技术实现；Plan 定义任务依赖；AGENTS 约束每次执行和证据要求**。发现冲突时生成 ADR/阻塞报告，不能自行删除需求或违反冻结契约。
4. 从 Plan 中选择开发前置已满足、当前未被其他分支占用、与本会话权限相符的最小可交付任务；条件前置按 Plan §2.4 和 ADR-0001 解释，原有实测义务不得删减。不得宣称能在聊天会话结束后自动继续工作。

## 2. 不可删减的首版交付

- Windows 10/11 x64：Tauri GUI、`hopnodectl.exe`、`hopnode-core`、Windows Service。
- Ubuntu Desktop 22.04/24.04 x64：**真正的 GUI `.deb`**、CLI、systemd/Core；从桌面菜单真实可启动。
- Ubuntu Server 22.04/24.04 x64：无需 GUI/GTK/WebKitGTK 的 `hopnodectl`、Core/agent/systemd。
- Windows/Ubuntu 均提供主控、受控及双角色；`hopnode` MCP stdio 在已安装本机 Codex 的环境能真实调用。
- 「快速 SSH」无需部署 Headscale/目录/Agent；「托管组网」使用 Tailscale/Headscale/DERP 和设备目录。两模式共用 Go Controller/IPC，不复制 SSH/MCP 实现。

## 3. 施工与并行规则

- 原路线为 T → P → K 后业务开发；2026-10-19 用户已授权 Plan §2.4/ADR-0001 条件路线：G0/G1/P05 开发出口为 CONDITIONAL，T03 与各 POC 未决实测按事实保留，允许正式 K01。K01→K02→K03→K04 必须串行完成并合入共享基线、G2 完整通过后，才按模块依赖并行 N、R、M、U、S、C。每个分支/工作树只写负责目录。
- CONDITIONAL 仅是限定开发许可，不表示 DONE、实测 PASS 或发布许可；Pxx 在条件路线中的实现依赖与原有验收义务分列，未满足任务自身完成判据不得置 DONE。身份/撤销等策略先决策明确拒绝行为，再冻结接口；真实失败推翻共享假设时暂停相关部分，更新 ADR/迁移并复测。同一路线授权持续有效，无需重复确认。
- G2后模块的后继实现前置按Plan §2.4.4核对经审查合入的真实实现、稳定冻结接口和自动检查；仅缺列明的环境实测可继续后继实现。前后任务的DONE及影响结论的待测依赖仍须闭环；空实现、已知阻断失败、安全策略未决或不稳定接口不能满足该前置。K串行、真实E2E及G2–G6不适用此放行规则。
- `contracts/`、根 `go.mod`/`go.sum`、前端锁文件、`cmd/` 共用入口、`.github/`、PRD/Design/Plan/AGENTS 均属于集成负责人控制的公共区。修改公共文件要先提交接口方案/ADR，不允许多个工作树互相覆盖。
- 对每项任务记录：任务 ID、对应 F/A 编号、修改路径、自动测试命令及实际输出、平台实测结果、证据、阻塞、风险和 PR/commit。
- 不覆盖用户已有的 SSH/Tailscale 配置，不提交真实密钥、密码、设备私有 IP、运行数据库和日志；测试只对已授权设备执行。
- 遇到缺测试机、权限、Codex 或公网部署时把对应测试写为 `NOT_RUN` 或 `BLOCKED`，继续推进独立工作，不把 Mock 或静态截图标为验收完成。

## 4. 技术及接口硬规则

- UI：React + TypeScript + Ant Design；桌面壳：Tauri 2 + 薄 Rust；核心/CLI/MCP：Go；网络采用 SSH/SFTP 与现成 Tailscale/Headscale。除非 ADR 明确批准，不更换技术栈。
- **唯一 MCP 名**：`hopnode`。**唯一 CLI 主命令**：`hopnodectl`。组织设备注册用 `enroll`（旧 `login` 只可明确兼容）。Tool 对目标设备操作必须携带确定的 `device_id`。
- Windows 用 Named Pipe，Ubuntu 用 Unix Domain Socket，本机 IPC 按实际用户/服务身份做权限检查；外部目录 API 使用 HTTPS 与设备级身份。
- 关闭托管受控不能只修改 UI 标记；必须拒绝新的 HopNode 请求并按设计撤销托管凭据/可归属会话，且不破坏用户既有系统 SSH。
- GUI、CLI、Core、MCP、Directory 通过 `contracts/version-policy.md` 进行兼容协商；升级要验证备份、迁移、恢复，发布前检验 Ubuntu GUI 与 CLI 真实安装。
- SSH 基础 config/Host Key 和 Codex GUI 路径检测按 `docs/compatibility.md` 实测；MCP SDK 和 Codex 协商的协议版本不靠推断。

## 5. 测试、完成与停止条件

- 先运行与当前模块相关的格式、静态检查、单测、契约测试；再做 Windows/Ubuntu 真实 E2E；失败、跳过、未运行必须分别写清。
- 未提供实际环境/凭据不能声称完成跨公网、DERP、Windows Service、Ubuntu GUI 安装、无头 Server、MCP 真调用、设备撤销或升级恢复。
- `docs/implementation/status.md` 只在有证据时将任务设为 `DONE`；A01–A27 每项需要环境/步骤/结果/证据/验收日期，27/27 PASS 后才可按 Plan L07 发布正式版。
- 用户未要求删除或覆盖时不要清理其数据；遇到破坏性操作先确认该操作属于已授权测试范围，并优先在临时目录与隔离网络验证。
- 若某任务不可执行：报告具体阻碍和缺失条件，不用伪实现跳过；若仍有独立工作可以进行，遵守任务依赖继续实施。

## 6. 每次任务结束的固定汇报结构

```text
任务编号与关联 PRD：
修改文件与关键实现：
检查命令及实际结果：
Windows 真机：PASS / FAIL / NOT_RUN（原因）
Ubuntu Desktop GUI 真机：PASS / FAIL / NOT_RUN（原因）
Ubuntu Server CLI 真机：PASS / FAIL / NOT_RUN（原因）
MCP / 跨网络真调用：PASS / FAIL / NOT_RUN（原因）
尚未解决的风险、阻塞、下一依赖任务：
Git commit / Pull Request（若已实际生成）：
```

**注**：不存在的构建脚本不得假装已经运行；本文件不意味着代码、安装包或验收已完成。
