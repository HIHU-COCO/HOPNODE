# ADR：P04 桌面运行时与本机 IPC 验证

日期：2026-10-19；作者：HIHU-COCO。状态：POC 局部验证通过，P04 保持 BLOCKED；平台全矩阵未通过，生产契约尚未冻结。

关联任务：T01、P04；PRD F01、F16、F18；验收 A15–A18、A24。依据：Design 的 Tauri 2 薄桥、独立 Go controller、Windows Named Pipe / Ubuntu Unix Socket 和服务身份要求。

## 决策

保留 React / TypeScript / Ant Design 6 / Tauri 2，不引入替代技术栈。工程预览只做真实只读状态诊断与最小托盘；设备、SSH、受控开关等业务等待 P05 / K 阶段。Tauri 2.12.3 对应 tauri-build 2.7.3（初选 2.6.3 在实际 Cargo 解析时与 tauri 的 ^2.7.2 依赖冲突，已经修正）。最低 Rust 1.90，实测工具链 1.99.1。Cargo.lock 固定传递依赖。

独立 Go module `tests/pocs/desktop-service` 使用隔离的 `poc.v0`，不修改公共 `contracts/`。通过本机管道上的 HTTP/1.1 只读端点验证真实双向调用。没有监听 TCP，没有 Go 业务复制到 Rust；Rust 只发送请求、限时/限长和验证版本，前端验证状态结构。

Windows 使用当前进程 SID 的受保护 DACL，go-winio 默认拒绝远程客户端；Ubuntu 检查私有用户目录并将 UDS 设置为 0600。不覆盖已有 socket，不读取 SSH/Tailscale 配置。当前默认端点只面向用户态 controller；正式系统 agent 需要单独定义服务 SID、受授权调用用户、权限和多用户端点，此 POC 不能证明系统态身份隔离已完成。

GUI 不持有 Core 子进程。关闭 GUI 隐藏至托盘，托盘退出仅退出 GUI。由此得到进程架构上的解耦；注销/重启是否继续运行需要 Windows Service/systemd 的真实实现与测试，不能由结构推断验收通过。

## Ubuntu 22.04 X11 的局部实测依据

2026-10-19，集成负责人在已授权的普通用户 GNOME 42.9 / X11 真机上分别以普通独立进程和独有 transient `systemd --user` 单元运行 Core。GUI 来自 `.deb` 的 `dpkg-deb -x` 私有暂存解包，不是系统安装。两轮均通过：无 DISPLAY 的 CLI、UDS 0600、实际 GUI PID 的内核身份审计、AT-SPI 读取连接状态与实际 Core PID 并真实点击刷新、GUI 退出后同一 Core 仍可由无 DISPLAY CLI 查询。

每轮记录 2 条来自本次 GUI 内核 PID 的成功状态请求（首次连接、刷新），没有用 CLI 请求替代 GUI。owned 进程组已核验不存在；user 单元停止后 MainPID=0、inactive/not-found；七个自建运行子目录均已清理，原有配置和数据不在清理范围。两轮 GUI stderr 都为 0 字节，只说明此次完整 GNOME 真机未复现已有 WSLg GTK 警告，既有 WSLg 记录仍保留。

这些证据支持当前用户态 IPC 与进程解耦方案。它们不证明 `.deb` 安装/卸载或菜单启动、Wayland、24.04、托盘、系统级 root 服务、Windows SCM 或注销/重启验收通过；AT-SPI 操作也不代替布局目视验收。脱敏证据索引与实际生成物哈希见桌面运行时测试记录。

随后用户明确授权临时安装并在测试后卸载。集成负责人完成同一工程 `.deb` 的真实 dpkg 安装，普通用户 runner 经实际系统 `hopnode-desktop-poc.desktop` 的 Gio manager 启动入口，确认 `/proc` 中实际安装二进制、同包哈希、真实 GUI PID 请求和 AT-SPI 刷新，五项检查全部通过；退出 GUI 后同一独立 Core 继续工作。随后实际 dpkg 卸载，包记录与三个 payload 均已不存在，owned 进程组、运行子目录及独有上传文件已清理，私有证据保留。该结果补充系统安装/入口机制的局部依据；不等同 GNOME 应用抽屉人工点击、布局目视或全部 A 项通过，P04 仍 BLOCKED。

## 取舍与后续

Linux 工程预览使用 `tauri.linux.conf.json` 将打包产品名设为 ASCII `hopnode-desktop-poc`，通过 desktop 模板保留中文桌面名称。实际首轮 Tauri 打包返回成功，但 `dpkg-deb -f` 拒绝中文 `Package` 字段，因此必须用 Debian 工具检查元数据及安装，不能以打包命令退出 0 作为可安装证据。此包只覆盖 P04 桌面壳，正式 GUI/Core/CLI 分包仍由 S04 实现。

- POC 的 `hopnodectl-poc` 是测试 fixture；正式 CLI 的唯一名称仍为 `hopnodectl`。
- 用户态 Core 不授予远端操作能力，不把只读 IPC 成功解释为受控、SSH 或 MCP 可用。
- 生产 IPC 的错误模型、身份、版本协商由集成负责人在 K01/K04 冻结；现有 `poc.v0` 无兼容承诺。
- Ubuntu 22.04 X11 工程 `.deb` 的安装/Gio 系统入口/卸载已有真机证据；必须补 GNOME 菜单人工操作、完整功能及 Ubuntu 24.04 / Wayland 等矩阵，WSL 不等价。
- 已取得两轮 Ubuntu 22.04 X11 GUI 终止后用户态 Core 独立运行证据；必须继续补 Windows Service、系统级 systemd、用户注销/系统重启、跨身份拒绝访问和原生托盘关闭/恢复。
- P04 当前不能标 DONE；证据见 `docs/implementation/test-results/pocs/desktop-runtime.md`。

集成验证方案：在现有双平台 POC 检查和 Go CI 中执行纯 Python 的 user-unit 异常清理回归；在已有 GTK/Gio 依赖的 Ubuntu Rust CI 中执行真实 Gio `/usr/bin/sleep` 生命周期夹具。前者无需 systemd，后者只创建临时 desktop 文件和自有 sleep 进程，均不安装包或服务；它们不能替代上述真机 GUI 证据。
