# P01：WSL 单机独立用户态机制实测

日期：2026-10-19。关联 P01、F02/F03/F04；属于 A03/A04/A26 前置证据，不能据此通过任何完整 A 项或将 P01 标 DONE。

## 已实测结论

Ubuntu 22.04 WSL、x86_64、普通用户 UID 1000 下，新建原生 mode 700 目录运行 Headscale 0.29.5 内嵌 DERP、两个便携 Tailscale 1.98.4 用户态节点。**单机共宿主机制局部 PASS**：真实 TLS 正反例、两枚一次性 key 注册、双向合成 TCP、仅允许端口/deny-all/恢复、精确节点 expire 与已用 key 新 state 拒绝均通过。完整执行退出 0，finally 清理通过。

此轮没有连接授权 LAN Ubuntu，没有调用系统 Tailscale CLI/socket，没有读取或改写旧 state，没有执行系统服务、路由、DNS、CA、SSH 或防火墙配置操作。没有通过读取旧网络状态来证明其持续连通；结论限于本轮未发出覆盖旧实例的操作。新 UDP、进程和资源仍与宿主共享。

| 范围 | 实际结果 |
| --- | --- |
| 固定资产 / 版本 | 官方独立资产 SHA-256 与 README 固定摘要一致，版本输出分别 0.29.5 / 1.98.4 |
| Headscale | `configtest`、初始 port-only `policy check`、管理 socket mode 600 通过 |
| TLS 可信正例 | Python 标准 HTTPS 验证 `https://localhost:18443/health` 通过；候选 Tailscale 以有效 key 完成实际注册 |
| 候选 TLS 错 CA | 候选 tailscaled 使用无关 CA，对正确服务报 `unknown authority`；未认证 |
| 候选 TLS 错 DNS 名称 | 独立 18444 TLS 对照服务提供同 CA 签发的 `wrong-name.invalid` 证书，候选客户端连接 localhost 拒绝名称不匹配 |
| 认证诊断 | 无效 key 获服务器 `invalid pre auth key`，只记录“已见认证诊断”；最终 TLS 正例以有效 key 注册为准 |
| 非交互注册 | 新数据库仅合成 A/B 两节点；两端 Running；A/B 该轮 up 耗时 5375/5325ms |
| userspace TCP | 两端各自 SOCKS 对独立 peer 地址的 42881 `/proof` 获取固定 marker；双向通过 |
| 非授权端口 | 同一宿主的独立 42882 合成服务本地可达；经各 peer 的新连接均 curl 28、约 2018ms 超时 |
| deny-all | 仅本轮 Headscale PID 收 SIGHUP，新 42881 连接两方向均 curl 28；对应等待环首次失败观测 4272/3022ms；本地服务继续可达 |
| 恢复策略 | 原子恢复本轮 port-only 文件并向已核验 PID 发送 SIGHUP；两个方向再次得到 marker |
| 节点 B 精确 expire | 从本轮两条记录中按合成名称唯一选择 B；自 expire 命令发起 12795ms 内 A→B 新连接失败，curl 28；该次连接 2059ms，本地服务仍活着 |
| B 客户端状态 | expire 后 `BackendState=NeedsLogin` |
| 已用 key / 新 state | 第三个全新临时 state 复用已成功注册 B 的同一枚一次性 key；本轮输出明确出现同一行 backend error 与 already-used-key 诊断，未见 x509 错误；up 非 0、NeedsLogin，Headscale 仍仅两个节点；未创建替代 key |
| DERP 配置 | map 仅本轮 region 999，HostName localhost，DERPPort 18443；没有加载公共 map |
| 链路观测 | 两方向 disco 输出有非 DERP 的 via 观测；未见 DERP/peer-relay 行。仅该次 disco 观测，不是业务数据路径证明 |
| 正常完成清理 | 零自有进程停止错误，全部自有子进程退出，全部显式候选 TCP/UDP 可重新 bind，独立 30 秒预算内；退出 0 |
| SIGTERM 取消清理 | 外部监督器只对自己启动且身份匹配的 worker 在 wrong-CA 阶段发 TERM；worker 非 0、INCOMPLETE，finally 零错误、全子进程退出、全部候选端口可 bind、预算内 |
| 结束后额外端口复查 | 普通用户逐项测试全部显式候选 TCP/UDP 的 IPv4 与 IPv6 loopback bind 均通过；每次只短暂创建并关闭探测 socket |

curl 28 在这里说明新连接超时，不能解释为已拿到某种明确 ACL 拒绝报文。允许→deny-all→恢复，以及对照服务始终本地可达，为策略行为提供单机局部证据。两个方向串行检查的等待环耗时，也不能直接当作精确 map 传播延迟。expire 的 12795ms 包含该 CLI 命令本身的耗时，明确满足本轮 30 秒观察窗；不推断其他网络/设备上的时限。

## 实际边界与监听

两个逻辑节点在同一 OS 用户和宿主中，最终都转发至同一个 loopback 合成服务。它们不是两台隔离主机，不能证明两台设备拥有独立 loopback、OS 权限、安装形态或真实业务身份。

筛选仅本轮候选端口的 `ss` 原始输出留私有目录，核对每条 PID 均为执行器或其自有子进程。对外只保留端口与地址类别：

- TCP 18443（Headscale/DERP）、18444（错误名 TLS）、15443（TLS gRPC）、19090（metrics）、11055/11056（SOCKS）、42881/42882（合成 HTTP）均 loopback。
- UDP 13478（STUN）为 loopback；UDP 42641/42642（两个节点）实际为 IPv4/IPv6 通配监听。
- TLS 负例和第三新 state 的额外 daemon 用自动 UDP/SOCKS 端口，不复用旧进程；其主进程均已退出。普通用户模式并未对 peer-discovery 做宿主级网络隔离。

本轮 runtime 对 LAN 研究模板作明确的局部参数变体：FQDN=`localhost`，bind/IP=`127.0.0.1`，metrics 显式 `127.0.0.1:19090`（避开空值解释歧义），log level info 仅写私有 raw。服务端证书为 DNS SAN localhost，进程级 `SSL_CERT_FILE` 指向本轮 CA；没有修改系统 CA、hosts 或启用 TLS skip/debug。只有下载公开资产时遵循 WSL 已有 HTTPS 代理，新网络进程的环境不继承该代理、调试变量、认证变量或 SSLKEYLOGFILE。

没有做：不同网络、LAN 双机、阻断 P2P、DERP 承载业务、SSH/PTY/SFTP 经组网、稳定身份重启、活动连接强制回收、Directory/Agent 凭据撤销、Windows 第二实例、GUI/安装/systemd/Service 或规模性能。**DERP map 存在不是 DERP 流量证据；网络层 expire 与旧 key 拒绝也不是 A26 产品撤销。**

## 执行器及复现

主入口：`localhost_probe.py`。在已授权的 WSL 普通用户 shell、仓库目录执行：

```bash
python3 tests/pocs/network-isolated/localhost_probe.py --execute-localhost-only
```

程序拒绝 root，先新建 `/var/tmp/hopnode-p01-localhost-*` mode 700、预检固定候选端口，再从固定官方 URL 下载并核验两个资产。仅抽取 tgz 中精确指定的两个普通文件，不安装包。每个 daemon 指定本轮 socket/state/statedir，所有 Tailscale CLI 必须携带本轮 socket。真实 key 仅文件传入，不进入 argv；raw、私钥、SQLite、state 和明文注册材料均留该受保护目录，不能提交 Git。

重试可显式用 `--verified-assets-from '/var/tmp/hopnode-p01-localhost-REPLACE/bin'` 复用本用户前一轮 mode 700 原生目录中的公开资产，每轮仍重新核验固定摘要并复制至新目录；不复用原 state、用户、数据库、CA 或 key。不接受共享系统安装路径，不自动修改代理。

总执行 deadline 240 秒。启动临界区先登记 Popen，再恢复取消信号；身份检查含 starttime/exe/PPID。长生命周期启动 helper 只接受本轮已核验的两个 Go daemon；其 Go runtime/signal.Notify 处理 TERM/INT，仍保留限时 KILL 兜底，不把该信号屏蔽方案推广到任意程序。

HTTP accept 后的握手/读超时 2 秒，不对监听 socket 做可能无限等待的 TLS 握手。finally 逐资源执行有限等待，独立清理预算 30 秒；只对仍匹配本轮直接子进程身份者发 TERM/KILL，核验全部显式候选端口的 IPv4 bind（TCP 使用 SO_REUSEADDR）及全部自有子进程退出。清理任一错误、子进程未退、候选端口未释放或超预算，outcome 变 `INCOMPLETE_CLEANUP`，进程非 0。目录保留；无递归删除、pkill、系统服务 stop 或无 socket 的 logout。不能承诺 SIGKILL、WSL/宿主崩溃或内核不可中断 I/O 时仍可执行 finally。

## 失败、重试与证据

原始数据都在 WSL 原生私有目录，未复制到仓库。执行器只输出脱敏布尔、合成端口类别、耗时、退出分类及本轮目录名。

| 本轮顺序 | 实际结果 / 保留证据 |
| --- | --- |
| 最初直接下载 | URLError/直连 timeout；尚无子进程；随后只读检测确认已有代理可取官方摘要，未打印代理值 |
| 第二轮 | 资产摘要/版本、PKI、configtest 通过；8 秒未见管理 socket，主动停止自有 Headscale；该次日志为空，原因未确证 |
| 第三轮 | 启动窗放宽至 35 秒后，socket/policy/可信 health 通过；实际服务器返回 `invalid pre auth key`，旧断言误查连写 preauth 而停止；修正为诊断，不把该次 INCOMPLETE 改成历史成功 |
| 首轮完整机制 | 全流程至 deny-all/恢复通过、退出 0、清理通过；尚未加入 expire/旧 key 项 |
| 最终扩展机制 | 上表全部通过、退出 0；`hopnode-p01-localhost-b2i0p1y_/sanitized-result.json`；完整结果只在原生私有目录 |
| 已用 key 诊断复核 | 仅重新分类该最终轮已保留的第三节点输出/日志，未重连或改写历史结果；`used-key-auth-classification.json` 追加已用 key 后端认证诊断=true、x509 错误=false，并用最终代码的同一纯分类函数核验 |
| SIGTERM 取消清理 | 真实发出 TERM，worker 非 0、INCOMPLETE；清理全通过。监督器原生私有证据目录 `hopnode-p01-cancel-ccmu0v9v`，worker 为 `hopnode-p01-localhost-v_1pbf67`，未读取其他进程状态 |

前两次失败不声称已做较后阶段；重试始终创建新的私有目录，保留旧目录和失败记录。初次 8 秒启动窗不足的具体原因未证明，不写成 Headscale 或普通用户不兼容。

## 固定任务汇报

任务编号与关联 PRD：P01；F02/F03/F04；A03/A04/A26 前置。

修改文件与关键实现：独立 localhost 执行器、LAN 与单机范围分离的 README/结果表、本报告；未改公共契约/版本冻结/状态台账。

检查命令及实际结果：官方资产摘要与版本、Headscale configtest/policy check、候选 TLS 正反例、真实注册/TCP/策略/expire/旧 key、正常及 SIGTERM 清理如上；Python AST 解析通过。

Windows 真机：NOT_RUN（Windows 只负责启动 WSL，未运行 Windows 网络 daemon）。

Ubuntu Desktop GUI 真机：NOT_RUN。

Ubuntu Server CLI 真机：NOT_RUN（本次是 Ubuntu 22.04 WSL 普通用户便携 CLI 的局部机制实验，未安装 Server/systemd 产品）。

MCP / 跨网络真调用：NOT_RUN；P01/A03/A04/A25/A26 不标通过。

尚未解决的风险、阻塞、下一依赖任务：LAN 专用 FQDN/两端隔离部署、真正不同网络和可隔离阻断条件、真实 DERP 业务承载、复用 P02 的拨号适配及 SSH Host Key 验证；生产安全和版本冻结仍待 P05。

Git commit / Pull Request：以实际提交为准；没有创建 PR 或推送。
