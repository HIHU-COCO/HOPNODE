# ADR：P02 SSH / SFTP 与任务语义验证边界

- 任务：P02；关联 PRD F05、F08–F13、A20、A24。
- 创建者：HIHU-COCO。
- 日期：2026-10-19。
- 状态：POC 局部结论；已补授权 LAN Ubuntu 22.04 SSH 实测，仍待 Windows OpenSSH、Ubuntu 24.04 及完整任务语义，尚未冻结生产契约。

## 背景

快速 SSH 必须在没有 Headscale、Directory、HopNode Agent 的条件下完成已授权操作，并真实校验 Host Key。Go SSH 库不会自动实现 OpenSSH config。普通 SSH session 的关闭、后台任务生命周期和托管撤销属于不同语义，不能以一次超时单测推断远端进程已经终止。

## 已采用的 POC 方案

使用独立模块 `tests/pocs/ssh`，固定 `golang.org/x/crypto v0.58.0`、`github.com/pkg/sftp v1.13.11`，不修改根 Go 模块或正式接口。客户端显式接收目标、身份、独立 known-hosts 文件、可选基础配置与授权远端临时父目录。测试夹具只监听本机回环，秘密和文件均临时生成。

未知 Host Key 只有提供经独立通道核验的正确指纹才追加记录；变更密钥始终拒绝，不能靠再次传入批准指纹覆盖。密码从临时环境注入，不输出认证秘密；指定私钥支持未加密和口令加密格式。生产版的系统密钥库及文件 ACL 仍由后续 S01 / R01 实施。

多算法服务器可能默认优先提供另一种未被当前文件钉住的主机键；这类协商不匹配不能据此判定原算法公钥变化。POC 已按 knownhosts 的地址匹配结果限制协商为已钉住算法，随后仍严格验证公钥字节与撤销状态；没有以添加未知扫描键绕过校验。普通/哈希名称、多密钥服务器、真正 Ed25519 变化、跨类型 CA 证书及 CA 撤销均有回环回归测试；主代理也在真实多算法 LAN 目标仅使用既有 Ed25519 pin 复验命令成功，未追加信任。

配置只验证 Host / HostName / User / Port / IdentityFile 的受限子集，并与 Windows 系统 `ssh -G -F` 比对；高级配置、agent、跳板、用户目录及 token 展开暂明确拒绝。正式兼容支持仍须 R06 的两平台真实测试。

命令返回 stdout / stderr / exit code。建立 session 后，取消关闭该 session；未回应 channel-open 的目标导致整个连接关闭，以界定本地等待。传输超时关闭探针连接。SFTP 探针只新建随机目录及独占文件；字节比对、目录列举与清理都成功后才输出 PASS。

## 结论与后果

Windows 上的 Go 协议夹具已证明本 POC 的认证、Host Key、PTY 请求、流与退出码、10 MiB 传输和本地取消路径能够工作；另经 Windows 客户端到 WSL Ubuntu 22.04 隔离 chroot 的真实 OpenSSH 8.9p1 复验，指定私钥认证、`whoami`、PTY、目录列举及 10 MiB SFTP 均通过。Linux 上 `go test -race ./...` 通过。

主代理随后在用户授权的独立 LAN Ubuntu 22.04 目标使用与既有记录匹配的主机键与内存认证，完成密码登录、命令流/退出码、PTY、唯一临时目录下中文/空格文件 10 MiB SFTP 往返和空目录清理。只读清单实际执行成功，识别出 X11 / GNOME 与 GUI 运行依赖、活动 systemd 服务及构建依赖缺口；这些条件不能替代桌面菜单启动或安装验收。现有 SSH / Tailscale 服务与配置没有被修改。

约三秒的短 nohup 调查证明本次两秒任务在正常 SSH 连接结束后完成，新连接可以读取完成标记并清理全部本轮文件。它只覆盖显式后台化与正常 session 关闭，不提供异常网络断线、注销、重启、可靠取消或 task-runner 持久化结论。详细证据见 [P02 报告](../../tests/pocs/ssh/report.md)。Windows 系统 sshd、Ubuntu 24.04、GUI/Server 安装、跨网络/DERP 和完整后台任务仍未验收，P02 不标记 DONE。

Quick SSH 的默认任务能力降级为 session 内执行：取消表示本地会话停止等待，不宣称远端进程可靠终止；断线后不声称可以恢复查询。要提供 F13，须在两平台验证专用 task-runner 的 task ID、持久化结果、进程归属和终止策略。不得删除用户原有 SSH 密钥或关闭整个系统 sshd 来模拟托管撤销。

P05 仅可采纳带证据的局部结果及明确的阻塞；生产接口及六条业务线仍按 Plan 的依赖门槛推进。
