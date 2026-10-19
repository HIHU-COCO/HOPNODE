# HopNode 桌面工程壳

创建者：HIHU-COCO；日期：2026-10-19。任务：T01 / P04；关联 F01、F16、F18，A15、A16、A17、A18、A24。此目录是工程预览，尚未实现完整业务 UI 或正式安装器。

Windows 与 Ubuntu 共用 React / TypeScript / Ant Design 6，Tauri 2 仅提供只读 IPC 薄桥和托盘。桌面程序不启动或持有 Go Core 子进程。Core 不存在、拒绝连接、超时或版本不匹配时显示真实错误。

从仓库根安装锁定依赖，再执行：

```text
pnpm -C apps/desktop dev
pnpm -C apps/desktop typecheck
pnpm -C apps/desktop lint
pnpm -C apps/desktop test
pnpm -C apps/desktop build
pnpm -C apps/desktop tauri dev
```

浏览器预览明确显示无法访问本机 IPC。要读取真实状态，先按 `tests/pocs/desktop-service/README.md` 独立启动 Go controller，再启动 Tauri。

```text
cargo fmt --manifest-path apps/desktop/src-tauri/Cargo.toml -- --check
cargo clippy --locked --manifest-path apps/desktop/src-tauri/Cargo.toml --all-targets -- -D warnings
cargo test --locked --manifest-path apps/desktop/src-tauri/Cargo.toml
```

Rust 固定 Tauri 2.12.3 / tauri-build 2.7.3，Cargo.lock 锁定传递依赖；Tauri 的最低 Rust 为 1.90，当前验证工具链为 1.99.1。Linux GUI 需要 WebKitGTK 4.1、GTK3 和托盘开发依赖。前端依赖由根 pnpm 锁文件管理。

关闭窗口会隐藏至托盘，托盘「打开跃点 HopNode」恢复窗口，「退出桌面程序」退出 GUI。这些行为与独立 Go 进程生命周期无耦合。系统注销/重启与 Windows Service/systemd 仍需独立实测；当前 controller 是手动启动的用户态进程。

图标是代码生成的工程占位标记，不代表最终品牌设计。正式 Ubuntu GUI `.deb`、Windows 安装器与无 GUI Core/CLI 包必须按 Plan 后续任务分别生成和验收。
