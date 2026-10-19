# 风险与阻塞

| 编号 | 影响任务 / 验收 | 现状 | 解除条件 |
| --- | --- | --- | --- |
| ENV-01 | T03/P01、A02–A05/A14/A25 | 已授权LAN Ubuntu22.04 Desktop，真实密码SSH/PTY/10MiB SFTP通过；BLOCKED：缺Windows受控、Ubuntu24.04及不同网络/隔离组网环境 | 继续使用已授权Ubuntu推进独立测试；补齐其他目标和隔离Headscale/DERP，凭据不入库 |
| ENV-02 | P01、A03/A04 | BLOCKED：没有隔离 Headscale/DERP 的域名、TLS 和部署入口 | 管理方提供可访问的测试部署；不得借用用户现网做破坏测试 |
| ENV-03 | P04、A15–A19/A22/A23 | BLOCKED：缺 Ubuntu 24.04、全新 Desktop/Server；Linux构建依赖已补齐，WSL deb安装/启动及systemd局部测试通过；Windows SCM缺管理员环境 | 配齐测试 VM/机器及安装权限；WSL 结果单独列出 |
| ENV-04 | T02、L04/L07 | PR #1与真实CI已有PASS记录，最新提交继续检查；分支保护未设置，27项验收未通过 | 保持当前提交CI通过；正式发布前完成所需保护/审查与27项验收 |
| TECH-01 | P02/R05、A10/A21 | SSH 断线不保证远端任务存活；撤销 authorized_keys 不终止既有会话 | Windows/Linux 任务与可归属会话 POC，明确降级边界 |
| TECH-02 | K01–K04 | 正式 IPC/MCP/身份/性能契约未冻结 | 收集 P01–P04 的真实结论后 P05 评审；不得以 POC 代替产品 |
| TECH-03 | P03、A06/A27 | Codex 配置存在不能证明工具已连接 | stdio 协商/调用/取消记录与本机 Codex 实调用证据 |
| LEGAL-01 | 正式分发 | 产品许可证尚未由所有者选择 | 选择许可证并完成第三方许可清单 |
| GUI-01 | P04/U05、Ubuntu GUI | WSLg已安装GUI经desktop entry启动，但出现GTK scale-factor critical；视觉、托盘及GNOME/Wayland/X11未验收 | 在完整UbuntuDesktop两版本复现/排除并保存实际窗口操作证据 |

按 Plan §10.2 持续完成独立工程与 POC；阻塞不取消 Windows/Ubuntu GUI、CLI、系统服务的交付范围。
