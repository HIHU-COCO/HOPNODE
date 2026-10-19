# T01 / P04 桌面和 IPC 测试记录

日期：2026-10-19（Asia/Shanghai）；作者：HIHU-COCO。源码首提交：`0ff6780`，后续修复以 Git 历史为准。环境：当前 Windows x64 用户会话；Go 1.27.2，Rust 1.99.1。Linux 记录由集成负责人补充。

| 检查 | 命令 | 实际结果 |
| --- | --- | --- |
| Windows Go 本机 IPC | 在 tests/pocs/desktop-service 执行 `go test -v ./...` | PASS：3 个测试，真实 Named Pipe 往返、只读 HTTP、离线错误 |
| Windows Go 静态检查 | `go vet ./...` | PASS：退出码 0，无诊断 |
| 独立 Core/CLI 生命周期 | `test-runtime.ps1 -GoExecutable <portable go>` | PASS：独立进程 PID/实际身份与 CLI 状态一致；停止 Core 后 CLI 非零退出 |
| Rust 格式 | `cargo fmt --manifest-path apps/desktop/src-tauri/Cargo.toml -- --check` | PASS：退出码 0 |
| Rust 静态检查 | `cargo clippy --locked --manifest-path apps/desktop/src-tauri/Cargo.toml --all-targets -- -D warnings` | PASS：退出码 0，57.69 秒 |
| Rust 编译/测试 | `cargo test --locked --manifest-path apps/desktop/src-tauri/Cargo.toml` | PASS：1 passed、1 ignored（需独立 Core，已另行显式执行）；初次缺 icon.ico 已修复 |
| Rust 真实 IPC | `test-runtime.ps1 -GoExecutable <portable go> -CargoExecutable <portable cargo>` | PASS：显式 `real_controller_roundtrip` 1 passed；Rust → Named Pipe → 独立 Go Core，实际 PID 与测试进程不同 |
| Windows GUI 二进制 | `cargo build --locked --manifest-path apps/desktop/src-tauri/Cargo.toml` | PASS：退出码 0，生成本机 debug 桌面可执行文件 |
| 前端工程 | 主集成工作树的 typecheck / ESLint / 4 Vitest / Vite build | PASS：集成负责人实际运行并回报，供本次 GUI 复用 |

## Windows Tauri WebView 真实功能与视觉 QA

采用已读取的 `playwright-interactive` 工作流，先列检查清单，再通过本次单独启动的 Tauri WebView2 调试端口访问真实桌面 renderer。Core 为独立 Go 进程，没有 mock、网络 API 替代或手工注入状态。

检查清单包括：真实身份/版本/PID、刷新按钮、服务停止后的错误/清空旧数据、服务重启恢复、1120×760 默认布局、640×480 响应式渲染、重复刷新及离线异常、GUI 退出后 Core 独立运行。

启动方式：主树真实 `pnpm -C apps/desktop dev` 提供 1420 页面；本工作树 Cargo 构建的原生 `hopnode-desktop.exe` 以隐藏启动模式运行；对该进程设置 `WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS=--remote-debugging-port=9223 --remote-debugging-address=127.0.0.1`、独立 `WEBVIEW2_USER_DATA_FOLDER` 和随机隔离 `HOPNODE_POC_ENDPOINT`。Playwright 通过本机 CDP 连接到该 WebView。缺少独立 profile 的最初调试尝试因 shared_worker 目标缺 browserContextId 触发工具内部断言，隔离 profile 后解决；未将失败的调试尝试算作产品成功。

| 操作 / 观察 | 结果 |
| --- | --- |
| 检查真实 Tauri `__TAURI_INTERNALS__`，页面 PID 与独立 Go controller PID 比较 | PASS：真实桌面桥接、PID 一致；身份、版本、实例标识由 Go 实际返回 |
| 点击「刷新状态」3 次 | PASS：均保持同一 Core 实例和实际 PID |
| 停止隔离 Core，点击刷新 | PASS：显示「未连接」「无法连接本机 Core」及真实 Named Pipe os error 2，旧 PID/身份数据清空 |
| 重新启动 Core，点击刷新 | PASS：显示「已连接」与新进程 PID |
| 页面重载与脚本错误观察 | PASS：恢复真实连接，无 pageerror |
| 1120×760 默认 WebView 渲染 | PASS：无横纵溢出，状态与提示可见；视觉检查无数据截断、控件重叠或不可读文本 |
| 640×480 WebView renderer 响应式尺寸 | PASS：无横向溢出，状态与刷新按钮可见，提示说明可正常纵向滚动查看；原生 HWND 调整并未可靠验证，不将 renderer 尺寸等同原生窗口验收 |
| 视觉修复 | 初次品牌 H 继承 header 行高导致偏下；`line-height:1` 修复并在两种 renderer 尺寸重新目视通过 |
| 退出本次 GUI 进程后 CLI 查询 | PASS：独立 Core PID 不变、CLI 可调用；这是进程隔离验证，未替代托盘菜单或用户注销测试 |
| 清理 | PASS：关闭 CDP 连接，停止本次 GUI、Core、1420 devserver；未改系统服务或用户网络配置 |

截图保留于本机 `%TEMP%/hopnode-p04-gui-8f2907a3a7994420af3683922b02a146/` 的 `connected-default-final.png`、`connected-640x480-final.png`、`core-stopped.png`、`small-notice.png`。截图包含本机实际运行身份，只作为本地证据，未提交仓库。原生窗口可见启动、原生最小窗口、托盘点击/关闭恢复仍需真正桌面交互验证。

本表只记录实际运行，测试进程身份未写入报告。随机隔离 Named Pipe 与临时构建产物未提交。

| 平台/验收 | 状态 | 原因 |
| --- | --- | --- |
| Windows 用户态 Core / CLI Named Pipe | PASS | 实际本机独立进程测试 |
| Windows Tauri WebView 渲染 / 刷新 / 真实 IPC / 故障恢复 | PASS | 原生 Tauri 进程内的真实 WebView，通过 CDP 点击刷新验证；没有 mock |
| Windows 原生窗口可见启动 / 最小尺寸 / 托盘关闭恢复 / 安装器 | NOT_RUN | 隐藏启动下原生窗口句柄不支持可靠交互验证；renderer 截图不替代该验收 |
| Windows Service、注销、重启 | NOT_RUN | 尚未实现系统服务 POC，不能用用户态进程替代 |
| Ubuntu Desktop 22.04/24.04 GUI、.deb、GNOME/Wayland/X11 | NOT_RUN | 没有已验收的完整 Ubuntu Desktop 两版本实机矩阵 |
| Ubuntu Server 22.04/24.04 CLI/systemd | NOT_RUN | WSL/headless 构建不等价两版本真实 Server 与 systemd 生命周期 |
| MCP / 不同网络 / DERP | NOT_RUN | 此 POC 没有这些能力，不属于本地 IPC 检查 |

任务状态建议：T01 的桌面工程子项可以提交集成审查；P04 仍 IN_PROGRESS / 部分 BLOCKED，不能标 DONE 或认定 A15–A18/A24 已通过。下一个依赖是补齐真实平台和系统服务测试，再进入 P05 联合评审。
