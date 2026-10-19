# 风险与阻塞

| 编号 | 影响任务 / 验收 | 现状 | 解除条件 |
| --- | --- | --- | --- |
| ENV-01 | T03/P01、A02–A05/A14/A25 | 已授权LAN Ubuntu22.04 Desktop，真实密码SSH/PTY/10MiB SFTP通过；BLOCKED：缺Windows受控、Ubuntu24.04及不同网络/隔离组网环境 | 继续使用已授权Ubuntu推进独立测试；补齐其他目标和隔离Headscale/DERP，凭据不入库 |
| ENV-02 | P01、A03/A04 | 单机隔离Headscale/TS与localhost私有TLS已有机制PASS；BLOCKED：不同网络端点、公网可达域名/TLS/DERP和独立P2P阻断边界仍缺 | 补可访问的隔离部署与不同网络授权设备；单机结果不替代跨网；不借用用户现网 |
| ENV-03 | P04、A15–A19/A22/A23 | BLOCKED：缺 Ubuntu 24.04、全新 Desktop/Server；22.04 X11安装/desktop入口/IPC/卸载PASS；Server2022临时CI SCM局部PASS，本机Win11未提权，Win10/11服务及完整身份未验 | 配齐剩余测试 VM/机器与授权身份；CI Server OS与工程包均不替代产品完整矩阵 |
| ENV-04 | T02、L04/L07 | PR #1与真实CI已有PASS记录，最新提交继续检查；分支保护未设置，27项验收未通过 | 保持当前提交CI通过；正式发布前完成所需保护/审查与27项验收 |
| TECH-01 | P02/R05、A10/A21 | SSH 断线不保证远端任务存活；撤销 authorized_keys 不终止既有会话 | Windows/Linux 任务与可归属会话 POC，明确降级边界 |
| TECH-02 | K01–K04/G2 | 已授权条件施工，K01 READY；正式IPC/MCP/Directory/身份/兼容契约未冻结，旧DRAFT仅局部验证 | 集成负责人按K01十二项缺口补齐完整结构/策略/共享类型及正反例，经审查合入后串行K02–K04；G2完整通过才启动业务 |
| TECH-03 | P03、A06/A27 | Codex0.153.4实际协商/echo/预期错误/status已PASS；CLI协议取消未证实；app-server实际清单未满足仅自有MCP守卫，三轮未发模型turn；Ubuntu Codex未验 | 先证明临时配置隔离与原生进程归属，再补客户端取消及Ubuntu调用；不得用参数、握手或进程退出替代取消 |
| LEGAL-01 | 正式分发 | 产品许可证尚未由所有者选择 | 选择许可证并完成第三方许可清单 |
| GUI-01 | P04/U05、Ubuntu GUI | WSLg旧测试出现GTK scale-factor critical；独立GNOME22.04 X11三轮显示/刷新/真实IPC PASS且stderr空，含工程deb真实安装与系统desktop入口启动；人工应用抽屉/布局/托盘/Wayland/24.04未验 | 保留WSLg原日志与环境差异；补24.04安装及两版本托盘、Wayland与完整视觉证据 |
| PLAN-01 | 条件路线/K01–K04、六业务线 | 用户已授权先开发、同步补齐剩余实测；后续环境差异可能推翻共享假设并产生返工，安全策略仍需明确决策 | 集成负责人维护ADR-0001、开发版本及风险责任；按G2/模块依赖实施，有环境即复测；发现失败先暂停受影响内容、修订ADR/契约迁移并回归；保留全平台及27项正式验收 |

按 Plan §2.4/§10.2 和 [ADR-0001](../decisions/ADR-0001-v1-stack.md) 的已授权条件路线推进 K01–K04，G2后按模块依赖实现，同时补齐上述环境和实测。CONDITIONAL只表示开发出口获准，不关闭ENV/TECH/GUI风险，不取消全平台交付，不授予新的测试权限或发布权限。
