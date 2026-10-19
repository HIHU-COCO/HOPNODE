# P04 独立系统服务生命周期 POC

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联 P04、F16/F18、A16/A22；仅追加系统服务局部证据，**不表示产品服务或正式验收完成**。

本目录是 Go 1.27.2 独立 module，固定已验证 `golang.org/x/sys v0.49.0`。复用 `../desktop-service/` 的只读 `serve/status/doctor` 夹具，不复制其 IPC 或 Go Controller 实现，不修改公共契约、桌面壳、系统现有服务、SSH/Tailscale 或 Codex 配置。

## Windows SCM 宿主

`hopnode-service-poc.exe` 仅接受 SCM 启动，不包含 install/delete、自我提权、开机自启配置或产品服务注册功能。必须显式指定：

| 参数 | 限定 |
| --- | --- |
| `--service-name` | `HopNode-P04-POC-<唯一小写字母/数字/短横线>` |
| `--core` | 已构建 desktop-service POC 可执行文件的绝对路径 |
| `--endpoint` | `\\.\pipe\hopnode-p04-service-<唯一小写字母/数字/短横线>` |
| `--stop-timeout` | 1 秒至 1 分钟，默认 10 秒 |

宿主创建自身未命名 Job Object，配置 `KILL_ON_JOB_CLOSE`；创建 suspended Core 子进程，先加入 job 再 resume，避免子进程执行后才建立归属。它不启动 shell，只执行指定二进制的 `serve`；仅传入端点与 SystemRoot，不转发用户 PATH/token 环境。

SCM 状态按 StartPending → Running → StopPending 转换。Stop/Shutdown 终止自身 job，等待主子进程退出且 job 活跃进程数归零后才成功返回。启动失败、主子进程意外退出、停止失败/超时返回不同服务错误码；关闭 job 作为剩余自身后代的清理后备。SCM Running 只表示子进程已创建，未承诺 IPC readiness。

当前 Windows 会话 `IsAdministrator=False`。只运行无原生进程操作的配置/SCM handler 单元测试、vet 和编译；**Windows 服务创建、启停、Job Object 原生行为、服务账号 ACL、注销/重启恢复均 NOT_RUN**。纯状态机测试中的 fake child 不计作 Windows Service 真机验收。正常 Stop 的 job 清理设计不能用来推断任意宿主崩溃、启动中断的行为。

本目录下可执行：

```powershell
go mod verify
go test -timeout 20s -count=1 -v ./...
go vet ./...
New-Item -ItemType Directory -Force artifacts | Out-Null
go build -trimpath -o artifacts\hopnode-service-poc.exe .
```

## Linux transient systemd 实测

先从 `../desktop-service/` 构建 Go POC，而非根目录产品骨架的 `hopnodectl`：

```bash
# 在 tests/pocs/service-lifecycle 下；Go 1.27.2 已在 PATH。
mkdir -p artifacts
poc_artifacts=$(realpath artifacts)
CGO_ENABLED=0 go -C ../desktop-service build -mod=readonly -trimpath -o "$poc_artifacts/hopnodectl-poc" .
bash -n test-systemd.sh
python3 -m py_compile verify-systemd.py
# 仅在已授权 root 会话执行；本脚本不自行 sudo/提权。
bash test-systemd.sh "$poc_artifacts/hopnodectl-poc"
```

脚本每次用随机 UUID 生成唯一 `hopnode-p04-lifecycle-*.service`，拒绝与已有 unit 碰撞；使用 transient unit 直接 ExecStart `Core serve`，不修改 `/etc/systemd/system`，不部署开机自启。systemd 创建本次独占 `/run/` 目录，模式 0700；Core 创建 0600 UDS。KillMode 为 control-group，KillSignal 为 SIGINT 以适配该 POC 的中断处理，停止有 10 秒上限及 systemd SIGKILL 后备。

验证过程为首次启动、无 DISPLAY 的真实 CLI status/doctor 与 systemd MainPID 比对、目录和 socket 权限断言、停止后原 PID 消失且请求失败、再次启动新 PID/实例、原生 `systemctl restart` 再产生新 PID/实例。停止后的 transient unit 可能被 systemd 回收，脚本必要时按同一任务名称重建，再单独验证原生 restart。

EXIT/INT/TERM trap 仅停止并 reset 本次随机名称的 unit，检查 MainPID 归零、已记录进程消失和请求失败。运行目录由本 unit 的 RuntimeDirectory 生命周期清理；脚本不递归删除任何目录。脱敏 JSON、二进制留在本目录 ignored `artifacts/`；原始 CLI 身份字段仅在验证进程内存中使用，保存前替换为 `REDACTED`，不保存原始响应或服务日志。

2026-10-19 的 WSL Ubuntu 22.04/systemd 249 实测 PASS，包括最终清理；Go 夹具为 CGO_ENABLED=0 的静态 ELF。**这不等于 Ubuntu Server 正式真机安装/自启验收，也不等于 Ubuntu Desktop GUI 验收**。完整命令和范围见 [report.md](report.md)。
