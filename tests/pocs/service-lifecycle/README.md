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

当前 Windows 会话 `IsAdministrator=False`。默认套件运行无原生进程操作的配置/SCM handler 单元测试，原生 Job 测试明确 Skip/NOT_RUN。额外设置 `HOPNODE_POC_CORE_EXE` 后，`TestOwnedChildNativeLifecycle` 启动本任务隔离构建的 Core，在随机 Named Pipe 上经实际 CLI 校验独立 PID/实例与 Job 成员关系，Stop 后检查主进程退出、整个 Job 活跃进程数为零、CLI 非零退出和句柄关闭。本次显式原生 Job 验证 PASS，无需管理员。

本机 Windows 11 的 **SCM 创建/启停仍 NOT_RUN**，没有提权或创建服务。纯状态机 fake child 和上述普通用户 Job 测试均不计作 SCM 验收；另有下述 Windows Server 2022 CI 局部证据，不能推断任意宿主崩溃、启动中断或注销/重启行为。

默认 Skip 的 `TestSCMNativeLifecycle` 已在显式授权的一次性管理员 GitHub `windows-2022` runner 真实 PASS：提交 `f2961c3e4ff8fcafb754cd0a082dcbb2fd0b7b13`，[run 39202456643 / Windows Go job 117898129568](https://github.com/HIHU-COCO/HOPNODE/actions/runs/39202456643/job/117898129568)。两轮随机 manual/LocalSystem 服务实际启停（host/Core 5844/2464 → 2428/3728），镜像/身份、原 ACL 的跨身份拒绝、零退出码和 handles signaled 均通过；最终 `SCM_OWNED_CLEANUP=PASS`，服务不存在、handles closed，整项测试 PASS 0.26s。二进制哈希、七条 proof、精确命令和限制见 [SCM-CI.md](SCM-CI.md)。这是 Windows Server 2022 正常生命周期的局部结果；本机未提升历史不变，Windows 10/11 完整矩阵、正向 API/产品授权、恢复与注销/重启仍未验，P04 不标 DONE。

本目录下可执行：

```powershell
go mod verify
go test -timeout 20s -count=1 -v ./...
go vet ./...
New-Item -ItemType Directory -Force artifacts | Out-Null
go build -trimpath -o artifacts\hopnode-service-poc.exe .
```

显式选择真实原生 Job 测试时，先在本目录的 ignored artifacts 中构建 desktop-service POC：

```powershell
$nativePocCore = Join-Path (Resolve-Path artifacts).Path 'hopnodectl-poc.exe'
go -C ../desktop-service build -mod=readonly -trimpath -o $nativePocCore .
if ($LASTEXITCODE -ne 0) { throw 'POC build failed' }
$previousPocCore = $env:HOPNODE_POC_CORE_EXE
try {
    $env:HOPNODE_POC_CORE_EXE = $nativePocCore
    go test -timeout 30s -count=1 -run '^TestOwnedChildNativeLifecycle$' -v ./...
} finally {
    $env:HOPNODE_POC_CORE_EXE = $previousPocCore
}
```

测试只清理自身 Job，不安装 SCM、不提权；原始 CLI 身份响应不写日志。Job 可能包含 Windows `conhost.exe` 等辅助进程，因此停止前验证真实 Core PID 属于本 Job，停止后严格要求整个 Job 的 activeProcesses 为零，不假定启动时恰好只有一个进程。

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
