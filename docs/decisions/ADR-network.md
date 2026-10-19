# ADR-network：保留已有网络，独立用户态机制验证

日期：2026-10-19；状态：P01 BLOCKED。

本机已有 Tailscale 的只读 status 曾确认后端 Running；已有 peers 不作为授权测试端点。保留现网，不调用旧实例的入网、登出、切换或防火墙操作。原有只读探针与四项单测继续保留。

2026-10-19，WSL Ubuntu 22.04 普通用户在新建原生 0700 目录运行独立 Headscale 0.29.5 / 内嵌 DERP 与两个便携 Tailscale 1.98.4 userspace 实例。资产与官方 SHA-256 逐字节一致；客户端使用独有 socket/state、私有 CA 和 `https://localhost` 的 DNS SAN，不安装系统 CA、不改 hosts/DNS/路由或系统服务。候选版本依据官方兼容窗口选择，不升级或降级现有安装。

实测通过：候选客户端可信证书下有效 key 注册、错误 CA/同 CA 错域名拒绝；双节点 Running；受限合成端口双向 TCP；非授权端口新连接失败且本地服务存活；deny-all 后新连接失败、恢复策略后再次成功；精确合成 B 节点 expire 后 30 秒窗口内新连接失败且 B NeedsLogin；已用一次性 key 在第三个新 state 中不能注册。新进程退出与候选 TCP/UDP 端口释放均核验通过。详细耗时、失败修复和运行边界见 `tests/pocs/network-isolated/RESULTS.md`。

该结果仅是单机共宿主机制实证。协调端 HTTPS 在 localhost；tailscaled UDP 包含通配绑定，不能称所有监听均 loopback。私有 DERP map 存在不证明业务流量经过 DERP；观察到的 disco 路径不能替代命令/文件传输链路证据。未完成两机 LAN、不同网络、独立边界阻断 P2P 后的 DERP 回退或 P02 SSH/SFTP 经组网验证。Headscale 节点过期不等于 HopNode Directory/Agent 凭据、活动会话和重装身份撤销。

保留 Tailscale/Headscale/DERP 技术路线与所有 PRD 范围。P01 整体保持 BLOCKED，A03/A04/A25/A26 不能据此标 PASS；版本尚未正式冻结。2026-10-19用户已授权 [ADR-0001](ADR-0001-v1-stack.md) 条件施工路线，P05/G1的开发出口为CONDITIONAL，允许先K01–K04、G2后模块实现；不同网络/DERP等实测义务仍开放，本ADR不据授权扩大网络能力结论。
