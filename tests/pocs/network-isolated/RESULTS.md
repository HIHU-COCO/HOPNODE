# P01 独立实例证据表

日期：2026-10-19。下表保留 **LAN 方案** 的执行状态；随后追加的单机 localhost 机制已取得局部结果，见 [单机报告](LOCALHOST-REPORT.md)。这两个范围分别记录，不能用同宿主结果覆盖 LAN/不同网络。P01 整体、A03/A04/A25/A26 均未通过；共享状态与版本清单不由本目录更改。

| 检查 | 状态 | 证据 / 执行时应填字段 |
| --- | --- | --- |
| 官方文档与固定标签源码研究 | 已完成研究 | README 的直接官方链接；无进程/监听/远端操作 |
| 固定资产实际 SHA-256 / version | NOT_RUN | 两端实际资产摘要、版本，仅下载摘要文本不算运行验证 |
| 旧网络基线与新端口空闲 | NOT_RUN | 脱敏旧后端/服务状态、PID/starttime 相同布尔、端口空闲布尔 |
| 普通用户 Headscale / 内嵌 DERP 启动 | NOT_RUN | OS、UID 非 root 布尔、configtest/serve 结果、预期监听匹配布尔 |
| 两端已有专用 FQDN 解析 | NOT_RUN | 两端正常 DNS 解析归属正确布尔；缺少则阻塞后续 TLS/注册 |
| 两端 TLS 正例、不可信 CA/错误 DNS SAN 负例 | NOT_RUN | 链、名称、期限检查；实际新进程是否保持证书校验，不能用 IP URL 绕过 |
| 独立 userspace tailscaled 启动 | NOT_RUN | socket/state 与旧实例分离布尔；无 TUN/路由/DNS/服务变更 |
| 两端一次性 key 非交互注册 | NOT_RUN | 合成节点数 2、backend Running、只有本轮 peers；不提交 key/ID/地址 |
| 新 state 重启稳定节点身份 | NOT_RUN | 重启相同实例后 ID 相同布尔；不提交原始 ID/state |
| 最小策略及拒绝端口对照 | NOT_RUN | 本机合成服务可用，允许端口成功、拒绝端口失败；传播时间 |
| LAN 双向 userspace TCP proof | NOT_RUN | 双向 marker 校验、退出码、限时、完整清理；不是 SSH/SFTP |
| 私有 DERP map / 高端口广告 | NOT_RUN | 只有本轮 region、端口正确、无公共 DERP map；map 不是流量证明 |
| 单次 direct / DERP / peer-relay / unknown 观测 | NOT_RUN | 时间、观测 API/命令、类型、依据；不提交端点地址 |
| 服务端 expire 与新连接拒绝 | NOT_RUN | 只操作本轮 B、实际传播时间、客户端认证状态、新连接结果 |
| 旧一次性 key 不能在新 state 入网 | NOT_RUN | 独立第三 state 的拒绝证据；不可先自动换 key 掩盖负例 |
| 本轮进程停止及旧网络保持 | NOT_RUN | 本轮进程身份核验/退出，新监听消失，旧网络基线比较 |
| P02 SSH/PTY/SFTP 经该新网络 | NOT_RUN | 需要受审查的拨号适配及独立 SSH Host Key 验证 |
| 独立边界阻断 P2P 后 DERP 承载命令/文件 | NOT_RUN | 需要新网络隔离/权限；不得以初始 relay 或 debug 强制偏好代替 |
| 不同网络 / 公网 DERP | NOT_RUN | 本方案仅 LAN 设计，缺两个不同网络授权端点与公网服务实测 |
| Windows daemon / GUI / Service 隔离 | NOT_RUN | 两个 Linux userspace 实例不能代替 Windows 产品验收 |
| A25 性能与 A26 产品撤销 | NOT_RUN | 本轮不覆盖规模/并发/1GB、Directory/Agent 凭据和重装身份 |

原始日志/JSON/证书/state/数据库/认证材料仅位于运行时受保护目录，不复制到本目录。以后只提交必要的脱敏布尔、合成 marker、状态、耗时及实际日期。FAIL、NOT_RUN、BLOCKED 分别填写，不能将连通失败重新标为预期负例。

最初只读研究的静态检查：README 的 6 段 bash 代码仅经 WSL `bash -n` 语法检查，退出 0；两份 `.hujson` 是严格 JSON 子集，PowerShell `ConvertFrom-Json` 均通过；模板占位符清点恰为 ROOT/FQDN/IPV4 三项；`git diff --cached --check` 无输出、退出 0。随后单机运行实际下载、configtest、policy check 及运行结果单列于报告；LAN 模板仍未在独立主机间验证。
