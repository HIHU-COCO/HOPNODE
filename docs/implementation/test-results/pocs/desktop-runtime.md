# T01 / P04 桌面和 IPC 测试记录

日期：2026-10-19（Asia/Shanghai）；作者：HIHU-COCO。源码首提交：`0ff6780`，后续修复以 Git 历史为准。环境：当前 Windows x64 用户会话；Go 1.27.2，Rust 1.99.1。Linux 记录由集成负责人补充。

| 检查 | 命令 | 实际结果 |
| --- | --- | --- |
| Windows Go 本机 IPC | 在 tests/pocs/desktop-service 执行 `go test -v ./...` | PASS：3 个测试，真实 Named Pipe 往返、只读 HTTP、离线错误 |
| Windows Go 静态检查 | `go vet ./...` | 执行中，最终结果需更新 |
| 独立 Core/CLI 生命周期 | `test-runtime.ps1 -GoExecutable <portable go>` | PASS：独立进程 PID/实际身份与 CLI 状态一致；停止 Core 后 CLI 非零退出 |
| Rust 格式 | `cargo fmt --manifest-path apps/desktop/src-tauri/Cargo.toml` | PASS：格式化已执行，待最终 --check |
| Rust 编译/测试 | `cargo test --locked --manifest-path apps/desktop/src-tauri/Cargo.toml` | 初次 FAIL：缺 Windows icon.ico，已生成修复；复测进行中 |

本表只记录实际运行，测试进程身份未写入报告。随机隔离 Named Pipe 与临时构建产物未提交。

| 平台/验收 | 状态 | 原因 |
| --- | --- | --- |
| Windows 用户态 Core / CLI Named Pipe | PASS | 实际本机独立进程测试 |
| Windows GUI 显示/托盘/退出/IPC | NOT_RUN | 待完成 Rust 构建后实际窗口交互 |
| Windows Service、注销、重启 | NOT_RUN | 尚未实现系统服务 POC，不能用用户态进程替代 |
| Ubuntu Desktop 22.04/24.04 GUI、.deb、GNOME/Wayland/X11 | NOT_RUN | 没有已验收的完整 Ubuntu Desktop 两版本实机矩阵 |
| Ubuntu Server 22.04/24.04 CLI/systemd | NOT_RUN | WSL/headless 构建不等价两版本真实 Server 与 systemd 生命周期 |
| MCP / 不同网络 / DERP | NOT_RUN | 此 POC 没有这些能力，不属于本地 IPC 检查 |

任务状态建议：T01 的桌面工程子项可以提交集成审查；P04 仍 IN_PROGRESS / 部分 BLOCKED，不能标 DONE 或认定 A15–A18/A24 已通过。下一个依赖是补齐真实平台和系统服务测试，再进入 P05 联合评审。
