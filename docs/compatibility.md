# 兼容性实测台账

日期：2026-10-19。`NOT_RUN` 不表示支持。仅隔离夹具通过不等于两平台正式支持。

| 能力 | Windows | Ubuntu 22.04 WSL | Ubuntu Desktop/Server 22.04/24.04 真机 |
| --- | --- | --- | --- |
| Go 基础 CLI | version / build PASS；仅工程入口 | 无 DISPLAY 静态 ELF version PASS | NOT_RUN |
| Tauri + React/Ant Design | 真实 WebView 刷新/故障恢复 PASS | Rust/前端构建、deb 元数据与 WSL 安装 PASS；完整桌面验证仍缺 | NOT_RUN |
| Named Pipe / UDS 同身份 IPC | Go/CLI/Rust/真实 WebView PASS | Go race 和真实 UDS；补充记录见 P04 Linux 报告 | NOT_RUN |
| SSH Host/HostName/User/Port/IdentityFile | 独立 config 与 ssh -G 临时配置对照 PASS | 隔离 OpenSSH 真实 whoami/PTY/10MiB SFTP PASS | LAN Ubuntu22.04密码/whoami/流与exit7/PTY/10MiB SFTP PASS；24.04 NOT_RUN |
| Host Key 首次确认/变更 | 隔离夹具PASS；多算法单pin/证书与撤销回归PASS；未知需显式确认，变化拒绝 | 新夹具真实首次信任 PASS | LAN已有记录匹配；修复后单pin Ed25519真实协商PASS；未更改远端密钥 |
| Include/Match/ProxyJump/IdentityAgent | 未声明支持，待 P02 结论 | NOT_RUN | NOT_RUN |
| Codex 定位与 stdio | Codex0.153.4真实echo/预期错误/status PASS，协商2025-06-18；CLI协议取消未证实；SDK1.7.0 stdio测试PASS | SDK旧版本/错误/取消/race PASS，非Codex调用 | NOT_RUN |
| Tailscale/Headscale/DERP | 不修改已有网络；P01 前置检查 | NOT_RUN | NOT_RUN |

组件固定版本和实际来源见 packaging/manifests/versions.md。SSH 与 MCP 最终兼容矩阵仅在有真实客户端与目标证据时更新。
