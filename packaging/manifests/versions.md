# 工程 / POC 版本清单

日期：2026-10-19。正式发布版本冻结需 P05。

| 组件 | 本次固定版本 | 核验来源 / 范围 |
| --- | --- | --- |
| HopNode | 0.1.0-dev | 工程基线，无正式发布声明 |
| Go | 1.27.2 | https://go.dev/dl/?mode=json；Windows ZIP SHA-256 1314008898bd40df77af4b014f777f08873dbdfbcd3d92308728ee03304fe04f |
| Rust | 1.99.1 | rustup 官方 stable 2026-10-15，安装器与官方 .sha256 对比 |
| Node | 26.7.0（本机） | 实际 node --version；CI 使用固定 26.7.0 |
| pnpm | 11.0.9 | packageManager 固定 |
| React / Ant Design / Tauri | 以 desktop package/Cargo 锁文件为准 | npm 官方 registry、crates.io 与真实编译 |
| SSH / SFTP / MCP SDK | 以各 POC go.mod/go.sum 为准 | Go 官方 proxy 与 SDK 源码；未冻结为生产依赖 |
| Headscale / DERP | 未冻结 | 缺隔离部署，不写 latest 为正式约束 |

便携工具位于忽略目录 .tools，不提交工具链、私钥、数据库或构建产物。
