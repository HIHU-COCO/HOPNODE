# 测试环境（脱敏）

日期：2026-10-19，Asia/Shanghai。执行人：HIHU-COCO。

| 项目 | 观察到的事实 | 限制 |
| --- | --- | --- |
| 仓库 | 起始仅四份 md，非 Git 仓库；现初始化 main，用户指定 Git 作者和 GitHub 远端 | 起始没有可保护的已有代码；四份依据保留 |
| Windows | Windows 11 x64，build 26200；OpenSSH 客户端存在，WebView2 存在 | 未提供远程 Windows sshd 测试机；不改既有 SSH |
| 工具链 | Node 26.7.0、pnpm 11.0.9、Git 2.55；隔离安装 Go 1.27.2、Rust 1.99.1 | PowerShell .ps1 受策略限制，使用 .cmd 或显式进程级执行策略，不修改系统策略 |
| MSVC | VS BuildTools 18.10，VC.Tools.x86.x64 组件存在；Tauri Windows 编译和当前用户 NSIS 安装/IPC/卸载通过 | 当前进程无管理员身份，Windows SCM 未安装实测 |
| Ubuntu | 已有 WSL Ubuntu 22.04.5 x64；systemd running、WSLg present | 不等同于全新 Ubuntu Desktop/Server；无 24.04；无免密 sudo |
| Linux 构建依赖 | 初查缺开发包；已通过 WSL root 安装 GCC/pkg-config/GTK3/WebKitGTK4.1/AppIndicator；Rust/前端构建、deb安装和WSLg desktop entry启动完成 | WSLg有GTK scale-factor critical，未验证完整视觉/托盘；不等同两版UbuntuDesktop验收 |
| 组网 | Windows Tailscale 可执行文件存在 | 未提供隔离 Headscale/DERP、公网 TLS 或不同网络测试机；不执行 logout/up |
| Codex | 0.153.4；隔离CODEX_HOME内注册hopnode、get/list通过；隔离登录状态Not logged in；默认CLI只读认证检查可用 | 不读取/复制认证令牌，不改写现有MCP配置；模型真调用正在以临时配置验证 |
| GitHub | 已推送main工程基线及feat/poc-integration；PR #1已建立，双平台CI真实通过记录已存档 | 最新提交继续复验；未设置分支保护，不存在正式Release |

尚未验证：Windows 10、Ubuntu Desktop 22.04/24.04 菜单启动和 Wayland/X11、Ubuntu Server 22.04/24.04 全新安装、用户注销/系统重启、1 GB 跨网传输、10 台目标、100 逻辑设备与 DERP。

WSL transient systemd 本机服务生命周期已通过；停止后本次unit被回收、专属运行目录消失。测试不安装持久系统服务或改开机配置。

网络隔离夹具只绑定 loopback，密钥仅在忽略的私有测试目录生成；sshd与devpts已停止/卸载，测试目录保留，不擅自递归删除。没有向现有 sshd/known_hosts/Tailscale 写配置。桌面安装测试完成后已卸载测试包，原有Windows应用数据保留。
