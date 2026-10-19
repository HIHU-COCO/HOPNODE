# ADR-0001：V1 技术 POC 联合评审

日期：2026-10-19；作者/集成负责人：HIHU-COCO。任务：P05；状态：评审进行中，G1 BLOCKED，未冻结生产契约。

## 已有实证

| POC | 结论与证据 | 尚未关闭的门槛 |
| --- | --- | --- |
| P01 网络 | 只读 Tailscale 前置检查和 4 项脱敏单测通过；链路类型保持 unknown。见 ADR-network.md | 隔离 Headscale/DERP、不同网络的授权端点、阻断 P2P 后回退均未测试 |
| P02 SSH | Windows 客户端连接隔离 Ubuntu 22.04 OpenSSH，命令、PTY、中文/空格路径、10 MiB SFTP 校验和夹具清理通过；协议测试通过。见 tests/pocs/ssh/report.md | Windows 系统 sshd、完整 Ubuntu 矩阵、断线长任务和高级 config 兼容仍未验证 |
| P03 MCP | 官方 Go SDK 真实 stdio 进程、工具、错误、取消、旧协议客户端测试通过；Linux race 问题已修复。见 P03-mcp-2026-10-19.md | 隔离 Codex 已注册，但没有登录，缺本机 Codex 真实模型调用证据 |
| P04 桌面/IPC | Windows Tauri WebView 调用独立 Go Named Pipe、故障与恢复通过；Ubuntu 22.04 WSL Rust/前端编译和 GUI deb 元数据检查通过。见 desktop-runtime.md 及 Linux 补充报告 | Windows SCM、Ubuntu 22.04/24.04 完整 Desktop/Server、菜单/托盘、注销重启和跨身份矩阵尚未齐备 |

## 保持的架构与候选界限

继续采用 React / TypeScript / Ant Design / Tauri 2 薄 Rust、Go Controller/CLI/MCP、SSH/SFTP 与现成 Tailscale/Headscale。当前没有证据要求替换技术栈。已验证版本固定于 packaging/manifests/versions.md；Headscale/DERP 兼容版本仍未冻结。

快速 SSH 的候选流程为：显式目标地址与认证 → 检查 Host Key → 用户明确确认未知指纹 → 测试命令；变化的 Host Key 必须失败。POC 未读取默认 SSH 配置，支持的基础字段与明确不支持的 Include/Match/ProxyJump/ssh-agent 见 SSH 报告。托管组网仍需隔离注册、设备身份与撤销 POC，不能借用用户现有网络宣布通过。

用户态 Windows IPC 已验证当前 SID 限权；Linux UDS 已验证目录 0700、socket 0600。系统 agent 与用户 controller 的授权调用者、注销后身份、凭据持有者和多用户隔离尚未冻结。关闭托管受控必须拒绝新请求、撤销 HopNode 管理的凭据并清理可归属会话，不能关闭用户独立 SSH；该要求保留，当前没有产品实现或验收证据。

`poc.v0` 和 `poc_probe` 是隔离试验，不是公共 IPC/MCP 契约。正式唯一 CLI/MCP 名仍是 `hopnodectl` / `hopnode`。官方 SDK 客户端协商结果不能替代 Codex 协商结果。

## 出口决定

Plan 允许在四项 POC 均有明确结论后启动评审，但 G1 要求阻断后续开发的风险关闭或存在明确授权的替代路线。上述风险仍开放，未获得以单机/WSL 代替正式平台与跨网络验证的授权。因此 P05 保持 BLOCKED；K01–K04 不冻结，N/R/M/U/S/C 六条生产开发线不启动。

资源补齐后依次复测原 POC，更新实证及风险，再决定 G1。不得删除 Ubuntu GUI、系统服务、托管组网或 Codex 真调用要求来通过门槛。性能测量草案见 docs/implementation/performance-baseline.md，尚不是 A25 验收结果。
