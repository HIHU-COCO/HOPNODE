# T00/T01 工程基线结果

日期：2026-10-19（Asia/Shanghai）；执行人：HIHU-COCO；版本：0.1.0-dev；基线 commit：32c6aa7，后续修正见 Git 历史。

## 检查事实

| 命令 | 实际结果 | 范围 |
| --- | --- | --- |
| 初始 git status --short / branch | FAIL：not a git repository | 初始目录只含四 md；已初始化 main |
| python scripts/init-ledger.py | PASS：64 tasks, 18 features, 27 acceptances | 初始化台账，不覆盖已有台账 |
| python scripts/verify-ledger.py | PASS：18 F / 27 A 编号完整 | 无验收项目标 PASS |
| go test ./... | PASS：internal/bootstrap；其他入口无测试文件 | Windows Go 1.27.2 |
| go vet ./... | PASS：退出码 0 | Windows |
| scripts/build-core.ps1 | PASS：三个 Windows exe | T01 基础版本入口，非可用产品 |
| artifacts/windows/hopnodectl.exe version | PASS：app_version=0.1.0-dev，os=windows，arch=amd64 | 实际执行 |
| GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ... ./cmd/hopnodectl | PASS：Linux ELF | Windows 交叉编译 |
| WSL env -u DISPLAY -u WAYLAND_DISPLAY hopnodectl version | PASS：app_version=0.1.0-dev，os=linux | Ubuntu 22.04 WSL，非全新 Server 验收 |
| python -m unittest discover -s tests/pocs/network -p test_*.py -v | PASS：4 tests | 脱敏与不推断链路类型 |
| network/probe.py --tailscale ... | PASS：OBSERVED/Running/self_online；connection_type=unknown | 只读状态；跨网/DERP 未验证 |
| git push -u origin main | PASS：32c6aa7 已推送到指定 HOPNODE 仓库 | 初始化提交；CI 见远端实际结果 |

Windows 真机：基础 CLI PASS；GUI/Service 待 P04。Ubuntu Desktop GUI 真机：NOT_RUN。Ubuntu Server 真机：NOT_RUN（只有 WSL ELF 冒烟）。MCP/跨网络：NOT_RUN，不能由上述检查替代。

T00 完成；T01 待桌面工程与 Rust 检查；T02 待远端真实 CI；T03 缺完整平台、不同网络及隔离 Headscale/DERP。
