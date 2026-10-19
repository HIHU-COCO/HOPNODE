# 工程 / POC 版本清单

日期：2026-10-19。正式发布版本冻结需 P05。

| 组件 | 本次固定版本 | 核验来源 / 范围 |
| --- | --- | --- |
| HopNode | 0.1.0-dev | 工程基线，无正式发布声明 |
| Go | 1.27.2 | https://go.dev/dl/?mode=json；Windows ZIP SHA-256 1314008898bd40df77af4b014f777f08873dbdfbcd3d92308728ee03304fe04f |
| Rust | 1.99.1 | rustup 官方 stable 2026-10-15，安装器与官方 .sha256 对比 |
| Node | 26.7.0（本机） | 实际 node --version；CI 使用固定 26.7.0 |
| pnpm | 11.0.9 | packageManager 固定 |
| React / Ant Design | 19.3.1 / 6.6.6 | npm 官方 registry、pnpm-lock.yaml；Windows/Linux 前端检查和真实 Windows WebView |
| Tauri / tauri-build | 2.12.3 / 2.7.3 | crates.io、Cargo.lock；Windows/Linux Rust、NSIS/deb构建；独立Ubuntu22.04 X11工程deb安装/Gio系统desktop入口/IPC/卸载PASS；API JS 2.12.4 / CLI 2.12.3 |
| SSH / SFTP | x/crypto 0.58.0 / pkg/sftp 1.13.11 | 独立P02 go.mod/go.sum；Windows客户端→隔离与独立LAN Ubuntu OpenSSH8.9p1命令/PTY/10MiB；已pin算法协商回归 |
| MCP Go SDK | 1.7.0 | P03 go.mod/go.sum与stdio测试；Codex0.153.4真实echo/错误/status PASS，实际协议2025-06-18；客户端取消未证实 |
| JSON Schema 草案验证器 | google/jsonschema-go 0.4.3 | contracts-draft独立模块，2020-12候选schema与正反例；未冻结生产契约 |
| go-winio / x/sys | 0.6.4 / 0.49.0 | 本机IPC/SCM POC独立模块；不表示系统服务产品验收 |
| Headscale / 内嵌DERP / 便携Tailscale | 候选0.29.5 / 同二进制内嵌 / 1.98.4，未冻结 | 官方资产摘要逐字节核验；WSL单机userspace机制PASS；非公网/不同网络/强制DERP结果，既有Tailscale版本不变 |

便携工具位于忽略目录 .tools，不提交工具链、私钥、数据库或构建产物。
