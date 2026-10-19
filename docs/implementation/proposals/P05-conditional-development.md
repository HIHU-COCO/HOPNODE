# P05 替代施工顺序提案（待用户决定）

日期：2026-10-19。状态：**PROPOSED；未授权、未生效，不改变 P05/G1/K01 状态。**

## 提案缘由

Windows/Ubuntu 22.04 GUI 与独立 Core IPC、工程包安装/卸载、真实 Codex MCP 调用、SSH/PTY/SFTP 已有局部实证。Windows Server 2022 临时 CI 也已实际完成 SCM 两轮启停、服务身份管道负例及删除清理，见 `tests/pocs/service-lifecycle/SCM-CI.md`。独立 Headscale/Tailscale 的单机机制验证见 `tests/pocs/network-isolated/RESULTS.md`；单机结果不能代替不同网络或强制 DERP 回退。

尚缺 Windows 10/11 系统服务及受控 SSH、Ubuntu 24.04/全新 Server、不同网络/公网 DERP、完整服务调用者身份、注销重启和升级恢复等证据。当前工程是诊断预览，业务功能未交付，正式验收仍为 0/27。

`AGENTS.md` §3 要求“T → P → K 阶段通过且证据齐全后，才并行 N、R、M、U、S、C 六条开发线”；`Plan.md` §4 的 G1 要求“阻断后续开发的未决风险已关闭或有明确授权的替代路线”。因此不能仅凭继续施工意愿把这些缺口登记为通过。

## 请求授权的具体替代路线

允许以已实测的 React/TypeScript/Ant Design/Tauri 2、Go、SSH/SFTP、Tailscale/Headscale 架构进入**有条件的 P05 出口**，将缺环境的实机验证保留为独立阻塞项，同时继续 K01–K04 和随后满足依赖的业务开发。此授权只调整施工顺序，不删减任何平台、角色、GUI、CLI、服务、托管组网或验收要求，不代表批准发布。

1. 先完成 K01：补齐 Directory 注册/心跳/停用的候选协议、完整 IPC 方法与事件、26 个 MCP 工具结果、版本协商、身份与取消/未知结果语义，以及 Go/TypeScript 共享或生成类型。已有离线片段仍是草案，不能直接宣布 K01 完成。
2. 每个未决安全策略须先写清决策和拒绝行为，再由集成负责人收敛契约。网络不可用、身份不匹配、能力未验证或结果未知时返回明确错误/能力状态；不伪造成功，不自动重放可能产生副作用的操作。
3. K01 机器校验和审查通过后，再串行推进 K02–K04。只有 G2 证据齐全，才按目录责任并行六条业务线，继续共享同一 Go Controller/IPC 实现。
4. Windows 10/11、Ubuntu Desktop/Server 两版本、不同网络/DERP、服务身份、恢复/升级等未运行项保留 `NOT_RUN/BLOCKED`，持续补测。现有用户 SSH/Tailscale/Codex 配置不得被覆盖。
5. 候选包依 Plan 在 G5 通过后执行 L01–L05 构建，再以候选包完成依赖它的真机验收及 L06；L07 正式发布仍需全部 F/A 追踪和 27/27 PASS。任何测试替代品、CI Server OS、WSL 或单机组网均不得冒充对应真机验收。

## 授权后的记录方式

若用户明确同意，更新 `ADR-0001-v1-stack.md` 和状态台账，记录用户授权日期、此替代路线及剩余阻塞；随后从最小 K01 契约交付开始。不得把现有局部 POC 状态或 A01–A27 改成 PASS 来适配新顺序。

若用户不同意或尚未回答，继续现有独立 POC、草案与工程检查；P05/G1 保持 BLOCKED，公共契约不冻结，不启动依赖它们的业务线。
