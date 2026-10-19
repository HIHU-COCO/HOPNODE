# P04 Linux IPC 与工程预览打包补充记录

日期：2026-10-19（Asia/Shanghai）；执行人：HIHU-COCO。关联 T01/P04、F01/F16/F18、A15–A19/A22/A24。环境仅为本机 WSL Ubuntu 22.04.5 x64 + WSLg；不是完整 Desktop/Server 验收环境。

## 构建与真实 IPC

工具：Go 1.27.2、Rust 1.99.1、Node 26.7.0、pnpm 11.0.9。Linux 工具与 frontend node_modules 独立于 Windows 安装。已安装 GCC11.4、pkg-config、GTK3、WebKitGTK4.1（pkg-config 2.50.4）、AppIndicator 等开发依赖；未改现有 SSH/Tailscale 配置。

| 实际命令/操作 | 结果 |
| --- | --- |
| 根模块与独立 Go POC 的 `go test -race ./...` | PASS；MCP 首次 race FAIL 及修复后复测另见 P03 记录 |
| `cargo clippy --locked --manifest-path apps/desktop/src-tauri/Cargo.toml --all-targets -- -D warnings` | PASS，首次 Linux 编译 5m19s |
| `cargo test --locked --manifest-path apps/desktop/src-tauri/Cargo.toml` | PASS，1 passed / 1 ignored；ignored 不是未启动 Core 时的通过声明 |
| `cargo build --locked --manifest-path apps/desktop/src-tauri/Cargo.toml` | PASS，生成 Linux GUI ELF |
| 全新 Linux frontend 安装 `pnpm install --frozen-lockfile` / `pnpm check` / `pnpm build` | PASS；格式/类型/lint/4 项 Vitest/构建；617 KiB 单 chunk 提示仍在，非构建错误 |
| 独立 Go Core + `env -u DISPLAY -u WAYLAND_DISPLAY hopnodectl-poc status --json` | PASS，真实 PID 与本次独立进程匹配；未打印实际身份 |
| 同端点显式 `cargo test --locked --manifest-path apps/desktop/src-tauri/Cargo.toml real_controller_roundtrip -- --ignored --nocapture` | PASS，1 passed，Rust → 真实 Unix Socket → 独立 Go Core；`distinct process=true` |

## GUI deb 的失败、修复与安装

首次 `pnpm --dir apps/desktop tauri build --debug --bundles deb` 退出 0，生成中文名文件；但 `dpkg-deb -f` 实际 FAIL，退出 2：`invalid package name in 'Package' field: must start with an alphanumeric character`。这轮产物不可安装，不计成功。

修复使用平台配置 `tauri.linux.conf.json` 的 ASCII 打包名 `hopnode-desktop-poc`，自定义 desktop 模板保留「跃点 HopNode（工程预览）」和真实 executable/icon。参考当前已安装 Tauri CLI schema 与官方 [Configuration reference](https://v2.tauri.app/reference/config/)。Windows 品牌显示不受该 Linux 配置影响。

| 实际操作 | 结果 |
| --- | --- |
| 修复后再次 `tauri build --debug --bundles deb` | PASS，52.88 MiB 工程预览包 |
| `python3 tests/pocs/desktop-service/validate-deb.py <package.deb>` | PASS：Debian Package/amd64/WebKitGTK4.1 依赖、ELF 可执行权限、中文 desktop Exec 和 icon 均有效；已加入 Ubuntu CI |
| 检查 dpkg 中不存在同名既有安装、control 无 pre/post install/remove 脚本 | PASS，不覆盖既有包 |
| WSL root `dpkg --install <package.deb>` | PASS：`Setting up hopnode-desktop-poc (0.1.0)`，实际 `/usr/bin/hopnode-desktop` 与 `/usr/share/applications/hopnode-desktop-poc.desktop` 存在 |
| 普通用户经 Gio.DesktopAppInfo 启动已安装的中文 desktop entry | PASS（仅启动）：连接 WSLg display、真实 GUI ELF 运行 5 秒；独立 Core 同时运行；缓存/配置/数据重定向至本次私有临时目录 |
| 启动诊断 | 观察到两条 GTK `gtk_widget_get_scale_factor` critical；未验证此 WSLg 窗口视觉、托盘或 Linux GUI 按钮/IPC 状态，不计完整 GUI PASS |
| 结束测试 | 本次 GUI/Core 已停止；WSL `dpkg --remove hopnode-desktop-poc` 已实际执行；本地临时证据保留，未删除用户数据 |

Windows NSIS 当前用户安装→内嵌前端真实 IPC→卸载记录见 `pocs/desktop-runtime.md`。WSL systemd 真正 start/stop/restart、0700/0600 与无 DISPLAY status/doctor 记录见 `tests/pocs/service-lifecycle/report.md`；不能与 Ubuntu Server 正式验收混同。

## 本机工程预览产物

产物是 debug POC，不是候选版或 Release，不含正式服务/Directory/SSH UI/MCP 集成功能。文件位于忽略目录 `artifacts/poc/`，未上传 GitHub Releases。

| 文件 | 字节 | SHA-256 |
| --- | --- | --- |
| linux/hopnode-desktop-poc_0.1.0_amd64.deb | 55449492 | 1890590155d92b2d882b426fd4d8c012b6695ed658f6840e2b82f062a08108ea |
| windows/hopnode-desktop-poc_0.1.0_x64-setup.exe | 2864489 | 5d1f7c91a1afa60ab6946a4a97c3164397433d56e0c2a861033fe0240e829b22 |

后续独立 Ubuntu Desktop 22.04 GNOME/X11 已补齐工程 deb 实际安装、Gio 系统 desktop 入口启动、真实 GUI 显示/刷新/IPC 及卸载，见 [真机补充记录](pocs/desktop-runtime.md) 与 [运行步骤](../../../../tests/pocs/desktop-service/ubuntu-desktop-probe.md)。此处保留历史 WSLg 警告及其原始边界。

Ubuntu Desktop 24.04、两版本 Wayland/完整视觉/托盘/应用抽屉人工点击、Ubuntu Server 22.04/24.04 全新安装、系统服务身份/注销/重启/升级恢复均 NOT_RUN。P04 保持 BLOCKED，A01–A27 仍 0/27 PASS。
