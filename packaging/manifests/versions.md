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
| Tauri / tauri-build | 2.12.3 / 2.7.3 | crates.io、Cargo.lock；Windows/Linux Rust、NSIS/deb 构建；API JS 2.12.4 / CLI 2.12.3 |
| SSH / SFTP | x/crypto 0.58.0 / pkg/sftp 1.13.11 | 独立 P02 go.mod/go.sum；Windows客户端→隔离Ubuntu OpenSSH8.9p1真实命令/PTY/10MiB |
| MCP Go SDK | 1.7.0 | P03 go.mod/go.sum与真实stdio测试；隔离Codex0.153.4已注册但未登录，模型真调用未验证 |
| go-winio / x/sys | 0.6.4 / 0.49.0 | 本机IPC/SCM POC独立模块；不表示系统服务产品验收 |
| Headscale / DERP | 未冻结 | 缺隔离部署，不写 latest 为正式约束 |

便携工具位于忽略目录 .tools，不提交工具链、私钥、数据库或构建产物。
