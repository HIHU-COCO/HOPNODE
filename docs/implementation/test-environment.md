# 测试环境（脱敏）

日期：2026-10-19，Asia/Shanghai。执行人：HIHU-COCO。

| 项目 | 观察到的事实 | 限制 |
| --- | --- | --- |
| 仓库 | 起始仅四份 md，非 Git 仓库；现初始化 main，用户指定 Git 作者和 GitHub 远端 | 起始没有可保护的已有代码；四份依据保留 |
| Windows | Windows 11 x64，build 26200；OpenSSH 客户端存在，WebView2 存在 | 未提供远程 Windows sshd 测试机；不改既有 SSH |
| 工具链 | Node 26.7.0、pnpm 11.0.9、Git 2.55；隔离安装 Go 1.27.2、Rust 1.99.1 | PowerShell .ps1 受策略限制，使用 .cmd 或显式进程级执行策略，不修改系统策略 |
| MSVC | VS BuildTools 18.10，VC.Tools.x86.x64 组件存在；Tauri Windows 编译和当前用户 NSIS 安装/IPC/卸载通过 | 当前进程无管理员身份，Windows SCM 未安装实测 |
| Ubuntu | 已有 WSL Ubuntu 22.04.5 x64；systemd running、WSLg present | 不等同于全新 Ubuntu Desktop/Server；无 24.04；无免密 sudo |
| 独立 Ubuntu 真机 | 用户授权LAN Ubuntu22.04 x86_64，kernel6.8.0-138-generic；GNOME42.9、活动X11会话、systemd及用户manager running；普通uid1000；SSH/PTY/10MiB SFTP通过 | 用户另行授权sudo临时安装/卸载工程deb，已完成且未装新依赖；既有SSH/tailscaled仍active，未改网络；目标/账号/凭据不入库 |
| 独立 Ubuntu GUI | GTK3.24.33、WebKitGTK4.1 2.50.4、AppIndicator0.5.90均installed；X11下两轮暂存与一轮实际安装GUI均真实显示/刷新/内核peer PID核验UDS通过，三轮GUI stderr均空；实际安装轮由Gio启动系统desktop入口并核验/usr/bin执行文件 | 工程debug包，随后卸载且包记录与3个新增文件均不存在；未人工点击应用抽屉或目视验证布局；Wayland/24.04/托盘未验；WSLg旧warning单独保留 |
| 独立 Ubuntu 用户服务 | 普通独立Core与UUID transient user systemd各一轮，加一轮实际安装GUI连接独立Core：GUI退出后无DISPLAY CLI仍读相同实例；三轮自建运行子目录/进程已清理；user unit最终MainPID0且not-found | 非系统root单元、非开机启动/注销/重启；上传暂存目录已精确逐文件清理，3个私有证据目录保留；异常启动清理故障回归另测3PASS |
| Linux 构建依赖 | 初查缺开发包；已通过 WSL root 安装 GCC/pkg-config/GTK3/WebKitGTK4.1/AppIndicator；Rust/前端构建、deb安装和WSLg desktop entry启动完成 | WSLg有GTK scale-factor critical，未验证完整视觉/托盘；不等同两版UbuntuDesktop验收 |
| 组网 | Windows Tailscale 可执行文件存在 | 未提供隔离 Headscale/DERP、公网 TLS 或不同网络测试机；不执行 logout/up |
| Codex | 0.153.4；隔离注册通过；使用既有认证与临时配置真实echo/预期错误/status通过，协商2025-06-18 | 配置/认证哈希未变；CLI取消尝试未取得协议取消证据；SDK取消独立记录 |
| GitHub | 已推送main工程基线及feat/poc-integration；PR #1已建立，双平台CI真实通过记录已存档 | 最新提交继续复验；未设置分支保护，不存在正式Release |

尚未验证：Windows 10、Ubuntu Desktop 24.04、两版 Wayland/托盘与应用抽屉人工点击、Ubuntu Server 22.04/24.04 全新安装、用户注销/系统重启、1 GB 跨网传输、10 台目标、100 逻辑设备与 DERP。独立22.04 X11的实际安装及desktop入口启动结果已单独记录。

WSL transient systemd及独立Ubuntu22.04 transient user systemd局部生命周期已通过；本次unit停止后被回收。测试不安装持久系统服务或改开机配置。

网络隔离夹具只绑定 loopback，密钥仅在忽略的私有测试目录生成；sshd与devpts已停止/卸载，测试目录保留，不擅自递归删除。没有向现有 sshd/known_hosts/Tailscale 写配置。桌面安装测试完成后已卸载测试包，原有Windows应用数据保留。
