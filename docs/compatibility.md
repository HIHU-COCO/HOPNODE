# 兼容性实测台账

日期：2026-10-19。`NOT_RUN` 不表示支持。仅隔离夹具通过不等于两平台正式支持。

| 能力 | Windows | Ubuntu 22.04 WSL | Ubuntu Desktop/Server 22.04/24.04 真机 |
| --- | --- | --- | --- |
| Go 基础 CLI | 待 build-core 实测 | 待 ELF 冒烟 | NOT_RUN |
| Tauri + React/Ant Design | P04 测试中 | 缺 WebKitGTK 开发包 | NOT_RUN |
| Named Pipe / UDS 同身份 IPC | P04 测试中 | P04 测试中 | NOT_RUN |
| SSH Host/HostName/User/Port/IdentityFile | P02 隔离实验 | NOT_RUN | NOT_RUN |
| Host Key 首次确认/变更 | P02 隔离实验 | NOT_RUN | NOT_RUN |
| Include/Match/ProxyJump/IdentityAgent | 未声明支持，待 P02 结论 | NOT_RUN | NOT_RUN |
| Codex 定位与 stdio | P03 测试中 | NOT_RUN | NOT_RUN |
| Tailscale/Headscale/DERP | 不修改已有网络；P01 前置检查 | NOT_RUN | NOT_RUN |

组件固定版本和实际来源见 packaging/manifests/versions.md。SSH 与 MCP 最终兼容矩阵仅在有真实客户端与目标证据时更新。
