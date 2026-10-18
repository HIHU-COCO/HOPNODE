# 测试环境（脱敏）

日期：2026-10-19，Asia/Shanghai。执行人：HIHU-COCO。

| 项目 | 观察到的事实 | 限制 |
| --- | --- | --- |
| 仓库 | 起始仅四份 md，非 Git 仓库；现初始化 main，用户指定 Git 作者和 GitHub 远端 | 起始没有可保护的已有代码；四份依据保留 |
| Windows | Windows 11 x64，build 26200；OpenSSH 客户端存在，WebView2 存在 | 未提供远程 Windows sshd 测试机；不改既有 SSH |
| 工具链 | Node 26.7.0、pnpm 11.0.9、Git 2.55；隔离安装 Go 1.27.2、Rust 1.99.1 | PowerShell .ps1 受策略限制，使用 .cmd 或显式进程级执行策略，不修改系统策略 |
| MSVC | VS BuildTools 18.10，VC.Tools.x86.x64 组件存在 | 桌面 POC 仍需实际编译 |
| Ubuntu | 已有 WSL Ubuntu 22.04.5 x64；systemd running、WSLg present | 不等同于全新 Ubuntu Desktop/Server；无 24.04；无免密 sudo |
| Linux 构建依赖 | 初查没有 Go/Rust/gcc/pkg-config/WebKitGTK 4.1 开发包 | 可从 Windows 交叉编译纯 Go；不能将其视为 Linux GUI 构建通过 |
| 组网 | Windows Tailscale 可执行文件存在 | 未提供隔离 Headscale/DERP、公网 TLS 或不同网络测试机；不执行 logout/up |
| Codex | 命令可发现；版本、stdio 能力由 P03 测试记录 | 不读取或改写用户现有秘密配置；MCP 真调用需独立证据 |
| GitHub | 用户指定 https://github.com/HIHU-COCO/HOPNODE；初查远端无 refs | 推送、CI、分支保护与 Release 以实际结果记录 |

尚未验证：Windows 10、Ubuntu Desktop 22.04/24.04 菜单启动和 Wayland/X11、Ubuntu Server 22.04/24.04 全新安装、用户注销/系统重启、1 GB 跨网传输、10 台目标、100 逻辑设备与 DERP。

所有隔离夹具只绑定 loopback，密钥在临时目录生成和销毁，不向现有 sshd/known_hosts/Tailscale 写配置。
