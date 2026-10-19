# T01 / P04 桌面和 IPC 测试记录

日期：2026-10-19（Asia/Shanghai）；作者：HIHU-COCO。源码首提交：`0ff6780`，后续修复以 Git 历史为准。Windows 环境：Windows 11 专业版 x64，10.0.26200（Build 26200）的当前用户会话；Go 1.27.2，Rust 1.99.1。下文另记集成负责人实际执行、经脱敏证据核对的 Ubuntu Desktop 22.04 X11 局部验证；其他 Linux 构建/WSLg 记录由集成负责人独立保留。

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

## Ubuntu Desktop 22.04 X11 独立暂存真机验证

新增 `ubuntu-desktop-probe.md` 与三份 `linux-*.py`，只接受明确的当前用户 X11 元数据；所有程序/运行数据放唯一 0700 暂存目录，`.deb` 仅提取不安装。Go fixture 增加可选 0600/O_EXCL 审计，以 Linux `SO_PEERCRED` 验证真实客户端 PID/UID，不能用 CLI 请求或进程存活冒充 GUI 请求。可选 AT-SPI 只读/点击本次 GUI PID 的应用树；无法验证 UI 时明确 NOT_RUN/PARTIAL。清理先核验 owned 进程组确已不存在及独有 user 单元 MainPID/状态，再删除自己的暂存子目录；无法核验则保留。

| 验证 | 结果 |
| --- | --- |
| Windows Go 原有回归 `go test -v ./...` / `go vet ./...` | PASS：3 测试，vet 退出 0 |
| WSL Ubuntu 22.04 `go test -race -v ./...` | PASS：5 测试，包括真实 UDS 内核 PID/UID 审计、拒绝覆盖已有审计文件 |
| WSL `go vet ./...` | PASS：退出 0 |
| 三份 Python `ast.parse` 语法检查 | PASS |
| Linux owned 进程组清理真实冒烟 | PASS：用 runner 同一清理函数终止独立新 session 的 sleep 进程，内核及 `/proc` 后验均确认组不存在 |
| 已授权 Ubuntu 22.04 Desktop 只读 probe | PASS：集成负责人实际执行回传；uid1000，五项必要会话字段存在、Xauthority 可读、GI 与四项工具存在，0600/O_EXCL 元数据创建成功；具体环境值未输出或提交 |
| 远端普通进程轮 | PASS：集成负责人实际执行 runner，无 `--user-unit`，退出码 0；下表五项检查全部通过 |
| 远端 transient user unit 轮 | PASS：集成负责人实际执行同一 runner 加 `--user-unit`，退出码 0；下表五项检查全部通过 |
| 本次报告复核 | PASS：PowerShell 只读解析两轮 report/audit/UI JSON，断言五项 PASS、两条实际 GUI PID 请求、UI action 和清理后验；退出码 0；`git diff --check` 退出码 0 |

真机环境：已授权 Ubuntu Desktop 22.04 x86_64，kernel 6.8.0-138-generic，GNOME 42.9 / gdm3 的活跃本地 X11 会话，普通 uid 1000；GTK 3.24.33、WebKitGTK 4.1 2.50.4、AppIndicator 0.5.90 已安装，用户 systemd 正常运行。执行未 sudo、未安装系统包。GUI `.deb` 只通过 `dpkg-deb -x` 提取到新建私有暂存目录后启动其中的实际二进制，没有执行系统安装或桌面菜单注册。

两轮使用相同生成物，脱敏报告保存的 SHA-256：

- GUI `.deb`：`1890590155d92b2d882b426fd4d8c012b6695ed658f6840e2b82f062a08108ea`。
- 独立 Go Core/CLI fixture：`4aab3e28088ca8a01012496aa6e23811055ca4e17a3ee299a0701b9193daa378`。

| 实际检查 / 证据 | 普通进程轮 | transient user unit 轮 |
| --- | --- | --- |
| `headless_cli_before_gui`：GUI 启动前，无 DISPLAY 的 CLI 返回实际独立 Core PID | PASS | PASS |
| `uds_mode_0600`：本轮独有 UDS 的实际 mode | PASS：0600 | PASS：0600 |
| `gui_to_real_uds`：Linux `SO_PEERCRED` 的 peer PID 等于本次实际 GUI PID、UID 一致，`GET /v0/status` 返回 200 | PASS | PASS |
| `visible_ui_and_refresh`：AT-SPI 只读取本次 GUI 的「已连接」及实际 Core PID，调用「刷新状态」的真实 action，再核对新增的同 GUI 内核 PID 请求 | PASS | PASS |
| `same_core_after_gui_exit_without_display`：终止本次 GUI 后，无 DISPLAY/XAUTHORITY/DBus 的 CLI 仍返回相同 Core PID/实例 | PASS | PASS |
| GUI 审计请求计数 | 2 条：首次连接及刷新；全部审计身份经内核核验，UID 1000 | 2 条：首次连接及刷新；全部审计身份经内核核验，UID 1000 |
| GUI stderr / GTK critical / scale-factor critical | 0 字节 / 0 / 0 | 0 字节 / 0 / 0 |
| owned 进程组停止后验 | GUI、Core 两组均已不存在，`remaining_states=[]` | GUI 组已不存在，`remaining_states=[]` |
| 独有 user 单元停止后验 | 不适用 | `MainPID=0`、`ActiveState=inactive`、`LoadState=not-found` |
| 自建运行子目录清理 | extracted/bin/runtime/config/data/cache/state 七项均 REMOVED | 同左；所有 own 进程组/单元停止核验完成后才删除 |

脱敏证据索引：集成工作树本机忽略目录 `.tools/authorized-ubuntu-22/` 中的 `gui-run.json`、`user-unit-run.json`、`gui-proof-summary.json`，以及 `desktop-run/`、`user-unit-run/` 各自的 `report.json`、`ipc-audit.jsonl`、`ui-accessibility.json`。两轮实际报告、审计和 UI 布尔结果已经核对；原始会话元数据、用户名、目标地址及凭据未写入本报告或提交 Git。私有运行证据保留，原有用户配置、数据和系统设置未清理。

这两轮结果只证明上述 GNOME 22.04 X11 环境中的暂存二进制、真实 GUI→UDS、真实刷新及用户态独立 Core 生命周期。AT-SPI 结果不等于人工目视布局验收。两轮零 stderr 说明完整 GNOME 真机此次未复现先前 WSLg 的 GTK 警告，不覆盖或撤销既有 WSLg 失败/日志。这两轮未验证 `.deb` 系统安装/卸载、桌面菜单、Wayland、Ubuntu 24.04、托盘关闭/恢复、系统级 root systemd、注销/重启或 Windows SCM；后续系统安装轮单独记录如下。

新增可选 `--installed-desktop` 模式，用户授权下由集成负责人独立安装/卸载包，普通用户 runner 仅负责核验安装文件哈希并经真实 Gio desktop-entry 启动、隔离及复用五项检查。`python3 test-desktop-entry.py` 在本地 WSL Ubuntu 22.04 实际 PASS：1 个真实 Gio sleep fixture 生命周期测试，确认新 session/group、PID callback、精确环境与组回收；此测试没有 GUI 验收含义。

审查发现此前 user-unit runner 在 `systemd-run` 正常返回后才登记清理，若单元已创建但父命令异常，可能漏 stop。已改为启动前创建生命周期账本，finally 对独有 UUID 单元 stop 并独立读取状态，未核验停止则 FAIL/保留 stage。`python3 test-user-unit.py` 实际 3 passed：模拟创建成功但父命令超时后仍执行 stop；stop 失败且仍 active 不授权清理；stop 后查询超时也不授权清理。七份 Python `ast.parse` 与 `git diff --check` 退出 0。本次是失败路径回归，不覆盖或撤销此前两轮真实成功路径证据。

## Ubuntu 22.04 系统安装 / desktop-entry / 卸载闭环

用户明确授权「允许临时安装、测试后卸载」，且不安装新依赖。集成负责人在同一已授权 GNOME/X11 真机执行真实系统安装和卸载，密码只经安全执行通道处理，未交给本代理或写入源码。包仍为工程 debug 诊断 POC，`.deb` / Core 哈希与前述两轮相同。

| 步骤 / 命令 | 实际结果 |
| --- | --- |
| 安装前只读 preflight | PASS：包记录不存在；三个 payload 文件均不存在且父目录无 symlink 冲突；包无维护脚本，架构 amd64；不覆盖既有安装 |
| `sudo dpkg --install <已验证 gui.deb>` | PASS：退出 0、stderr 为空；没有安装新依赖；`dpkg-query` 为 `install ok installed` / 0.1.0，`dpkg -V` 无输出（集成负责人实际核对） |
| 普通用户运行 `linux-desktop-stage.py ... --installed-desktop /usr/share/applications/hopnode-desktop-poc.desktop` | PASS：退出 0，`gui_launch_method=Gio.DesktopAppInfo installed desktop entry`；读取实际系统入口，核验 root-owned/权限及与原包一致的哈希后通过 Gio 启动 |
| Gio PID 与实际安装 exe | PASS：新 session/group 归属成立，`installed_desktop_exec_verified=true`；`/proc/<pid>/exe` 为 `/usr/bin/hopnode-desktop` |
| 五项功能检查 | PASS：无 DISPLAY CLI、UDS 0600、真实 GUI→UDS、AT-SPI 连接状态/实际 Core PID/真实刷新、GUI 退出后同一 Core 继续回应，五项全 PASS |
| GUI 请求审计和日志 | PASS：恰好 2 条来自实际 GUI 内核 PID 的成功状态请求，分别为初始连接和刷新；GUI stderr 为 0 字节 |
| 本次 GUI/Core 与运行子目录清理 | PASS：两个 owned 进程组后验均 gone、无剩余状态；七个自建运行子目录全部 REMOVED；原有配置/数据不在清理范围 |
| `sudo dpkg --remove hopnode-desktop-poc` | PASS：退出 0、stderr 为空；包记录与三个 payload 文件均不存在；原有 ssh/tailscaled 服务仍 active |
| 上传目录清理 | PASS：按本次精确 9 文件清单、当前 UID、常规文件与父 0700 校验后逐个 unlink，再 rmdir 空目录；`recursive_delete_used=false`，上传目录已移除；三个仅含证据的独立私有运行目录保留 |

实际安装的 `.desktop` SHA-256 为 `0eb4140b918a538fb7366edc4fb16db412510418a71a477da992aa8caab65ebb`，exe SHA-256 为 `997914054091ef108cd9feec0a8d1644bab54d21edecdb6cff79b3bf286d6653`，均与同一 `.deb` 提取文件一致。

本机忽略证据：`.tools/authorized-ubuntu-22/installed-run.json`、`package-{preflight,install,remove,postflight}.json`、`upload-stage-cleanup.json`，以及 `installed-run/` 内的 report/audit/UI 结果和空 GUI stderr。已只读复核包操作退出码、前后置布尔结果、真实 Gio/安装 exe 标记与两条 GUI 内核请求；未读取私有 session/目标元数据。上传目录已经清理，后续复测必须建立新的授权暂存目录。

本轮证明 Ubuntu 22.04 X11 上这个工程 `.deb` 的真实安装、系统 desktop-entry 启动、只读 IPC 与卸载闭环。Gio 启动不等同在 GNOME 应用抽屉人工点击，AT-SPI 不等同布局目视验收；仍未覆盖托盘、Wayland、24.04、系统级 root 服务、注销/重启、升级恢复或正式 A 项，P04 保持 BLOCKED。

本表只记录实际运行，实际运行用户名未写入报告。随机隔离 Named Pipe 与临时构建产物未提交。

| 平台/验收 | 状态 | 原因 |
| --- | --- | --- |
| Windows 用户态 Core / CLI Named Pipe | PASS | 实际本机独立进程测试 |
| Windows Tauri WebView 渲染 / 刷新 / 真实 IPC / 故障恢复 | PASS | 原生 Tauri 进程内的真实 WebView，通过 CDP 点击刷新验证；没有 mock |
| Windows 11 当前用户 NSIS 安装 / 内嵌前端 / 真实 IPC / 卸载 | PASS | 实际 debug 安装器在唯一 TEMP 目录安装并清理，前端不依赖 devserver；既有应用数据保持不变 |
| Windows 原生窗口可见启动 / 最小尺寸 / 托盘关闭恢复 | NOT_RUN | 隐藏启动下原生窗口句柄不支持可靠交互验证；renderer 截图不替代该验收 |
| Windows Service / SCM、注销、重启 | NOT_RUN | 本记录只有用户态进程证据；SCM 真实启停受当前权限阻塞，不能据此验收 |
| Ubuntu Desktop 22.04 GNOME X11 暂存 GUI / UDS / 真实刷新 / 独立 Core | PASS | 两轮实际真机执行；普通进程与 transient user unit 各自五项检查通过 |
| Ubuntu Desktop 22.04 工程 `.deb` 系统安装 / 实际 desktop-entry / 真实 IPC / 卸载 | PASS | 真实 dpkg 安装、系统 Gio 入口启动、五项检查通过后实际卸载；三个 payload 已移除 |
| GNOME 应用抽屉人工点击 / 布局目视、Wayland、24.04、托盘、系统级 systemd / 注销 / 重启 | NOT_RUN | Gio/AT-SPI 与用户态单元不替代这些验收；两版本完整矩阵仍缺失 |
| Ubuntu Server 22.04/24.04 CLI/systemd | NOT_RUN | WSL/headless 构建不等价两版本真实 Server 与 systemd 生命周期 |
| MCP / 不同网络 / DERP | NOT_RUN | 此 POC 没有这些能力，不属于本地 IPC 检查 |

任务状态建议：T01 的桌面工程子项可以提交集成审查；P04 保持 BLOCKED，不能标 DONE 或认定 A15–A18/A24 已通过。下一个依赖是补齐真实平台、安装和系统服务测试，再进入 P05 联合评审。

2026-10-19 集成复验：Windows `powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/check-pocs.ps1` 与 WSL `bash scripts/check-pocs.sh -race` 均退出 0，覆盖五个独立 Go 模块的 test/vet、网络探针 4 项和 user-unit 故障回归 3 项。WSL `/usr/bin/python3 tests/pocs/desktop-service/test-desktop-entry.py` 真实 Gio/sleep 进程归属与回收测试 1 PASS（0.108s），伴随 Python GI 的 ImportWarning，无测试失败。台账校验与 diff check 通过；受版本控制的文件未匹配本次测试凭据或目标信息。新回归已加入对应平台 CI，等待最新提交的远端结果。
