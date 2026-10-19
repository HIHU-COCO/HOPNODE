# T01 / P04 桌面和 IPC 测试记录

日期：2026-10-19（Asia/Shanghai）；作者：HIHU-COCO。源码首提交：`0ff6780`，后续修复以 Git 历史为准。环境：Windows 11 专业版 x64，10.0.26200（Build 26200）的当前用户会话；Go 1.27.2，Rust 1.99.1。Linux 记录由集成负责人补充。

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

## Windows 当前用户 NSIS 安装与卸载冒烟

集成负责人实际执行 `pnpm --dir apps/desktop tauri build --debug --bundles nsis`，本次独立验证该生成物的安装、内嵌前端与清理行为；不是正式发布验收。

- 安装器：`apps/desktop/src-tauri/target/debug/bundle/nsis/跃点 HopNode_0.1.0_x64-setup.exe`，2,864,489 字节。
- 安装器 SHA-256：`5D1F7C91A1AFA60AB6946A4A97C3164397433D56E0C2A861033FE0240E829B22`。
- 实际安装 exe：14,073,344 字节，SHA-256 `75D443D13E8292439C3BB3F6329C32AF1FB4401E145AC54185D4BAA2E6F2F6F2`。其哈希与后来已重新编译的 `target/debug/hopnode-desktop.exe` 不同，故不以当前开发二进制冒充打包源；本次实际启动的是安装目录内的 exe。

| 步骤 / 命令 | 实际结果 |
| --- | --- |
| 安装前只读检查 HKCU、HKLM、WOW6432Node 卸载项与同名快捷方式 | PASS：无已有 HopNode 安装或卸载注册，无同名桌面/开始菜单快捷方式；已有 `LOCALAPPDATA/io.hopnode.desktop/EBWebView` 数据目录，作为受保护的既有数据保留 |
| 检查生成的 `installer.nsi` | PASS：`INSTALLMODE=currentUser`、`RequestExecutionLevel user`；删除 AppData 只在显式选择删除数据复选框时触发 |
| `Start-Process <installer> -ArgumentList /S,/D=<唯一TEMP/app> -WindowStyle Hidden -Wait` | PASS：退出码 0；未提权，安装至唯一隔离目录 |
| 核对实际 exe、uninstall.exe、HKCU 卸载入口和快捷方式目标 | PASS：两份可执行文件存在；HKCU 的版本为 0.1.0，InstallLocation/UninstallString 和桌面/开始菜单快捷方式均指向本次隔离目录 |
| 停止前端 devserver并检查 1420 无监听，启动安装版 exe 与独立 Go Core | PASS：未运行 Vite；安装版真实 WebView URL 为 `http://tauri.localhost/`，Tauri 内部桥存在，证明前端来自安装包内嵌资源 |
| Playwright 连接本次安装版 WebView 的本机 CDP，实际点击「刷新状态」 | PASS：显示已连接；刷新前后页面 PID 都与本次独立 Go Core 实际 PID 一致；1120×760 无横纵溢出，截图目视通过 |
| 停止本次 CDP/GUI/Core，复核注册仍指向隔离目录后执行其生成的 `uninstall.exe /S` | PASS：退出码 0；安装 exe、安装目录、卸载注册、桌面及开始菜单快捷方式均已移除 |
| 清理本次安装新建且仅指向该 TEMP 路径的厂商安装位置记忆项 | PASS：生成卸载器默认保留此位置记忆项；确认安装前不存在、值等于本次路径且无子键后，只移除本次新建项及其空父键，未触碰既有用户数据 |
| 比较原有应用数据文件清单/大小/UTC 修改时间 | PASS：前后均为 234 个文件，元数据完全一致；原有目录保留 |
| 最终进程与端口清理 | PASS：本次 GUI/Core/调试连接已停止，9224 无监听；没有修改 Windows Service、SSH 或 Tailscale |

安装与卸载证据保留于本机 `%TEMP%/hopnode-p04-install-d40b95a7eef44659872de2f213a3a9a3/`，含 `installed-gui-connected.png` 和隔离运行日志；包含真实运行身份的原始证据不提交 Git。本次安装/卸载冒烟仅覆盖当前 Windows 11 用户、debug 工程包和只读 IPC，不证明 Windows 10、Windows Service、原生托盘交互、升级恢复或 A01–A27 通过。

## Ubuntu Desktop 独立暂存探针准备

新增 `ubuntu-desktop-probe.md` 与三份 `linux-*.py`，只接受明确的当前用户 X11 元数据；所有程序/运行数据放唯一 0700 暂存目录，`.deb` 仅提取不安装。Go fixture 增加可选 0600/O_EXCL 审计，以 Linux `SO_PEERCRED` 验证真实客户端 PID/UID，不能用 CLI 请求或进程存活冒充 GUI 请求。可选 AT-SPI 只读/点击本次 GUI PID 的应用树；无法验证 UI 时明确 NOT_RUN/PARTIAL。清理先核验 owned 进程组确已不存在及独有 user 单元 MainPID/状态，再删除自己的暂存子目录；无法核验则保留。

| 验证 | 结果 |
| --- | --- |
| Windows Go 原有回归 `go test -v ./...` / `go vet ./...` | PASS：3 测试，vet 退出 0 |
| WSL Ubuntu 22.04 `go test -race -v ./...` | PASS：5 测试，包括真实 UDS 内核 PID/UID 审计、拒绝覆盖已有审计文件 |
| WSL `go vet ./...` | PASS：退出 0 |
| 三份 Python `ast.parse` 语法检查 | PASS |
| Linux owned 进程组清理真实冒烟 | PASS：用 runner 同一清理函数终止独立新 session 的 sleep 进程，内核及 `/proc` 后验均确认组不存在 |
| 已授权 Ubuntu 22.04 Desktop 只读 probe | PASS：集成负责人实际执行回传；uid1000，五项必要会话字段存在、Xauthority 可读、GI 与四项工具存在，0600/O_EXCL 元数据创建成功；具体环境值未输出或提交 |
| 远端 GUI→UDS、AT-SPI 刷新、退出后独立 Core、user transient unit | NOT_RUN：脚本准备完成，等待集成负责人通过已授权通道实际传输/执行；不由本地 WSL 结果推断 |

本表只记录实际运行，测试进程身份未写入报告。随机隔离 Named Pipe 与临时构建产物未提交。

| 平台/验收 | 状态 | 原因 |
| --- | --- | --- |
| Windows 用户态 Core / CLI Named Pipe | PASS | 实际本机独立进程测试 |
| Windows Tauri WebView 渲染 / 刷新 / 真实 IPC / 故障恢复 | PASS | 原生 Tauri 进程内的真实 WebView，通过 CDP 点击刷新验证；没有 mock |
| Windows 11 当前用户 NSIS 安装 / 内嵌前端 / 真实 IPC / 卸载 | PASS | 实际 debug 安装器在唯一 TEMP 目录安装并清理，前端不依赖 devserver；既有应用数据保持不变 |
| Windows 原生窗口可见启动 / 最小尺寸 / 托盘关闭恢复 | NOT_RUN | 隐藏启动下原生窗口句柄不支持可靠交互验证；renderer 截图不替代该验收 |
| Windows Service、注销、重启 | NOT_RUN | 尚未实现系统服务 POC，不能用用户态进程替代 |
| Ubuntu Desktop 22.04/24.04 GUI、.deb、GNOME/Wayland/X11 | NOT_RUN | 没有已验收的完整 Ubuntu Desktop 两版本实机矩阵 |
| Ubuntu Server 22.04/24.04 CLI/systemd | NOT_RUN | WSL/headless 构建不等价两版本真实 Server 与 systemd 生命周期 |
| MCP / 不同网络 / DERP | NOT_RUN | 此 POC 没有这些能力，不属于本地 IPC 检查 |

任务状态建议：T01 的桌面工程子项可以提交集成审查；P04 仍 IN_PROGRESS / 部分 BLOCKED，不能标 DONE 或认定 A15–A18/A24 已通过。下一个依赖是补齐真实平台和系统服务测试，再进入 P05 联合评审。
