# P01：托管网络 POC

当前只有只读前置检查，不是完成的托管组网实现。

```powershell
python tests/pocs/network/probe.py --tailscale 'C:/Program Files/Tailscale/tailscale.exe'
python -m unittest discover -s tests/pocs/network -p 'test_*.py' -v
```

该工具只执行 `status --json`，只保留后端状态、在线布尔值、节点数量；不输出用户、主机名、地址、密钥、认证 URL 或原始 stderr。现有节点数量不能证明跨网可用；连接类型始终为 unknown，除非后续有明确的端点实测。

完整 P01 复现需要管理方提供隔离的 Headscale/DERP（固定版本）、有效 TLS 域名、两个不同网络的授权测试端点和一次性注册材料。凭据仅放受保护存储，不提交仓库。不要对已有网络执行 logout/up，不改用户防火墙来测试 DERP。

实测顺序：记录现有网络 → 在隔离设备注册 → 检查稳定节点身份 → SSH/SFTP 基础操作 → 记录明确的直连观测 → 只在隔离网络阻断端间直连 → 用底层实际观测证明 DERP → 重复命令与文件校验 → 恢复隔离网络并确认不影响原有配置。使用 tests/pocs/ssh 的命令工具时仍须显式确认 Host Key。

证据必须记录 OS/版本、底层组件固定版本、两个网络条件、脱敏连接方式、文件校验、命令退出码和日期。当前跨网络/DERP 均 NOT_RUN。
