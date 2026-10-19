# P04 系统服务局部 POC 实测报告

作者/执行者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联任务 P04，F16/F18，A16/A22。结论：Windows 宿主编译/纯状态机、显式原生 Job、Windows Server 2022 CI 的真实 SCM 正常生命周期及清理局部 PASS；WSL systemd transient 生命周期 PASS。Windows 10/11 完整矩阵、P04 及产品服务正式验收仍未完成。

## 范围与环境

- 仅新增 `tests/pocs/service-lifecycle/`。独立 Go module 使用 Go 1.27.2、`golang.org/x/sys v0.49.0`；未修改 desktop-service、公共 contracts 或根依赖。
- 参考本地 `x/sys@v0.49.0/windows/svc/service.go`、`svc/security.go`、Windows Job Object 类型和 syscall 声明；SCM 宿主通过 `windows/svc` 接入。
- 本机 Windows 11 amd64；管理员组有效身份检查返回 `False`，没有提权、创建/安装/启动 Windows 服务。下述独立 Windows Server 2022 CI 不覆盖此历史。
- 另在一次性 GitHub-hosted `windows-2022` 管理员 runner 执行真实 SCM；证据对应提交 `f2961c3e4ff8fcafb754cd0a082dcbb2fd0b7b13`、run `39202456643`、job `117898129568`。
- WSL 发行版 Ubuntu-22.04；Linux 内核 `6.18.33.2-microsoft-standard-WSL2`；`systemctl is-system-running` 为 `running`；systemd `249 (249.11-0ubuntu3.22)`；Python 3.10.12。
- Linux 测试使用已授权 WSL root 会话；具体运行身份字段不输出、不提交。没有使用真实网络设备或读取用户配置。

## 实际检查与结果

Go 命令在本目录执行；Windows `go` 是仓库 `.tools/go/bin/go.exe`，Linux `go` 是 `.tools/go-linux/go/bin/go`。

| 实际命令/步骤 | 结果 |
| --- | --- |
| Windows `gofmt -w` 本目录 Go 源码 | PASS |
| Windows `go mod tidy` / `go mod verify` | PASS；`all modules verified` |
| Windows `go test -timeout 20s -count=1 -v ./...` | PASS；`ok .../tests/pocs/service-lifecycle 6.614s` |
| Windows `go vet ./...` | PASS，exit 0，无输出 |
| Windows `go build -trimpath -o artifacts/hopnode-service-poc.exe .` | PASS，exit 0 |
| WSL `go test -timeout 20s -race -count=1 ./...` | PASS；`ok .../tests/pocs/service-lifecycle 1.019s`；Linux 只编译配置测试，Windows handler 文件不参与 |
| WSL `CGO_ENABLED=0 go build -mod=readonly -trimpath -o <本目录>/artifacts/hopnodectl-poc .`，工作目录为根仓库 `tests/pocs/desktop-service/` | PASS，exit 0；`file` 确认 x86-64 静态 ELF，无 GUI 依赖 |
| WSL `bash -n test-systemd.sh` | PASS，exit 0 |
| WSL `python3 -m py_compile verify-systemd.py` | PASS，exit 0 |
| `wsl.exe -d Ubuntu-22.04 -u root --exec bash <本目录>/test-systemd.sh <本目录>/artifacts/hopnodectl-poc` | PASS，exit 0；真实 systemd/UDS/CLI/停止/重启/最终清理 |

Windows 默认单元测试覆盖配置缺失、生产/远端管道、相对/不存在 Core、超出范围停止超时以及额外命令的拒绝；SCM handler 以 fake child 验证 Stop/Shutdown 等待清理完成、Interrogate、启动失败、主子进程退出、停止失败和超时不返回成功。默认不调用原生创建进程/Job API；显式原生测试见下节。二者均不记作 Windows Service 安装或运行成功。

Linux 使用现有独立 desktop-service POC，协议 `poc.v0`、版本 `0.1.0-poc`，二进制 SHA-256：`3c1e0cc89b2fafe52164dff2ecc18492f26f6a16a5fc478a242a74ebf55bd875`。没有误用根目录生产骨架的 CLI。服务直接由 systemd ExecStart 启动，没有额外 daemon 壳。

## 显式 Windows 原生 Job 生命周期

新增 `TestOwnedChildNativeLifecycle`，仅在 `HOPNODE_POC_CORE_EXE` 非空时执行。默认完整套件实际输出：

```text
=== RUN   TestOwnedChildNativeLifecycle
NOT_RUN: set HOPNODE_POC_CORE_EXE to an isolated desktop-service POC executable; no SCM installation is performed
--- SKIP: TestOwnedChildNativeLifecycle (0.00s)
PASS
```

默认整套 `go test -timeout 20s -count=1 -v ./...` 为 PASS，`ok .../tests/pocs/service-lifecycle 6.198s`，但该结果包含上述明确 Skip，不表示原生 Job 已运行。随后从根 `tests/pocs/desktop-service/` 执行 `go build -mod=readonly -trimpath -o <本目录>/artifacts/hopnodectl-poc.exe .`，exit 0。按 README 的 try/finally 临时设置 `HOPNODE_POC_CORE_EXE` 为该独立可执行文件绝对路径，执行 `go test -timeout 30s -count=1 -run '^TestOwnedChildNativeLifecycle$' -v ./...`。

初次原生测试 **FAIL**：我额外加入的“停止前 Job 必须恰好只有 1 个活跃进程”断言过强；API 实际无错误地返回 2。为确认归属，仅查询本任务 Job 的成员 PID/可执行文件 basename，观察到 `hopnodectl-poc.exe` 和 Windows `conhost.exe`，未输出命令行、环境、身份或完整镜像路径。失败测试的 defer 清理成功。此处是测试假设修订，不是已发现的宿主实现缺陷。

测试修订为停止前 activeProcesses 至少为 1，并从原生 Job PID 列表确认精确 Core PID 属于本 Job。保留停止后严格要求主进程 Exited、整个 Job activeProcesses=0、实际 CLI 非零退出且非测试超时、CloseHandle 成功的检查。宿主原生实现未修改。

修订后同一命令真实复测，exit 0：

```text
=== RUN   TestOwnedChildNativeLifecycle
NATIVE_JOB_LIFECYCLE=PASS child_pid=34212; stable IPC instance; activeProcesses=2->0; stopped CLI rejected; handles closed; SCM NOT_RUN
--- PASS: TestOwnedChildNativeLifecycle (0.21s)
PASS
ok  github.com/HIHU-COCO/HOPNODE/tests/pocs/service-lifecycle  6.101s
```

命名管道使用随机本任务端点；两次实际 `status --json` 返回的 PID/实例相同，PID 与 native process handle 的 GetProcessId 一致，区别于测试进程自身。只解析允许的协议/PID/实例/OS 字段，不打印含身份的原始 JSON。测试退出后恢复原环境变量；Core binary 留在 ignored artifacts。此结果证明普通当前用户下的一次真实 Job/Named Pipe/停止清理场景，不涉及 SCM 安装、SCM 启停、系统服务身份、注销或重启。

## Windows SCM 管理员 CI：本机准备与实际运行

新增 `scm_integration_windows_test.go` 和 [SCM-CI.md](SCM-CI.md)，只在本目录施工。独立环境门控 `HOPNODE_POC_RUN_SCM=1` 加两份现存绝对 fixture 路径才启用；默认与路径缺失明确 Skip。开门后权限不够必须 Fail，不提权/UAC、不创建本机普通用户服务。

测试设计：一次性 runner 上创建唯一随机 manual/demand、OwnProcess、LocalSystem 服务，复用既有宿主与 desktop-service；两轮真实 Start/Running/Stop/Stopped，核对 SCM host PID、直接 Core parent PID、两份镜像、LocalSystem token 及持有原生句柄退出。当前用户访问服务身份专属管道须被实际拒绝，不放宽 ACL；清理从 CreateService 成功即登记，仅停止/删除自建服务，失败时仅对已验证精确镜像句柄后备终止，再检查进程退出及 Delete/Close 后服务不存在。manager 的 LIFO 生命周期保留至服务清理完成。

| 本次实际检查 | 结果 |
| --- | --- |
| Windows `gofmt -w scm_integration_windows_test.go` | PASS |
| `HOPNODE_POC_RUN_SCM=0`、Core env 为空；`go test -mod=readonly -timeout 30s -count=1 -v ./...` | PASS，6.199s；SCM Native 与需显式 Core 的 Job 测试明确 Skip，未创建本机服务 |
| `TestSCMSettingsGateAndPaths` | PASS：默认/关闭/非法 gate、缺失/相对/不存在路径、同文件、目录和有效配置，9 子项 |
| `TestSCMCallerPermissionIsFailure` | PASS：普通用户错误、不同身份管理员允许、LocalSystem 调用者错误，3 子项；纯决策函数不调用 SCM |
| `TestSCMStoppedRequiresSuccessfulExit` | PASS：Stopped 仅在 PID/Win32ExitCode/ServiceSpecificExitCode 均为 0 时成功；5 子项覆盖非零错误码和非 Stopped 状态，失败仍不跳过清理/删除 |
| 补充退出码断言后 `go test -mod=readonly -timeout 30s -count=1 -run '^TestSCM' -v ./...`，gate=0 | PASS，7.504s；含上述 17 个新子项和原 handler 测试；Native SCM 仍明确 Skip |
| `go mod verify` / `go vet ./...` | PASS：all modules verified；vet 退出 0 |
| `go build -mod=readonly -trimpath -o artifacts/hopnode-service-poc.exe .` | PASS，退出 0 |
| `go test -c -mod=readonly -o artifacts/service-lifecycle.test.exe .` | PASS，退出 0，含 Windows SCM 原生 API 测试代码 |
| `TestSCMNativeLifecycle` 本机管理员实际启用 | NOT_RUN：本机未提权，没有创建服务；此历史结果保留 |

CI 实际运行命令为本 module 下 `go test -mod=readonly -timeout 5m -count=1 -run '^TestSCMNativeLifecycle$' -v ./...`，build/env/绝对路径前检及恢复步骤见 [SCM-CI.md](SCM-CI.md)。已只读核对提交 `f2961c3e4ff8fcafb754cd0a082dcbb2fd0b7b13` 的 workflow 使用 `windows-2022`、独立双 fixture build、显式门控和此命令；未修改 workflow 或公共台账。

[Windows Go job 117898129568](https://github.com/HIHU-COCO/HOPNODE/actions/runs/39202456643/job/117898129568)（PR run `39202456643`）真实 PASS。集成负责人通过安全 GitHub API 回传七条脱敏 proof，未向日志存储转发 token：

| proof | 实际结果 |
| --- | --- |
| Host SHA-256 | `163490a321a1af14a89276a506a42ffce0099a3419f25afcb5d5671060ac0231` |
| Core SHA-256 | `9f3e74423c881034660656f937b4b297244045c679223187697c7b8ab439771d` |
| Round 1 | host 5844 / 直接 Core 2464；镜像/LocalSystem 身份、实际 pipe `ERROR_ACCESS_DENIED`、Running→Stopped、PID及两退出码0、原生 handles signaled 全部 PASS |
| Round 2（Stop/Start 重启） | 新 host 2428 / Core 3728；同样的身份/镜像/ACL拒绝/零退出码/handles 退出后验全部 PASS |
| 生命周期 | `SCM_LIFECYCLE=PASS` |
| 最终清理 | `SCM_OWNED_CLEANUP=PASS`；自建服务不存在，已记录进程退出、handles closed |
| 整项结果 | `TestSCMNativeLifecycle` PASS，0.26s，没有 Skip |

集成负责人另已核验同一 SHA 的 [完整 run 39202456643](https://github.com/HIHU-COCO/HOPNODE/actions/runs/39202456643)：五个 job 全部 `completed/success`，包括上述实际 SCM 和 Ubuntu `.deb` 构建检查。此处记录 CI 状态，不由构建成功推断新增 GUI 真机或正式 A 项验收。

该结果是 Windows Server 2022 的真实正常生命周期与成功清理局部证据，未将本机普通用户 Skip 改成 PASS，也不代替 Windows 10/11 真机矩阵。SCM Running 和管道 ACL 拒绝负例不冒充 LocalSystem 正向 API 或业务健康；正式产品调用者授权、真实失败/外部取消后的完整清理、恢复、注销/系统重启均未覆盖，P04/正式 A 项保持未完成。

## 实际 systemd 生命周期证据

脚本第一次和补充元数据/最终清理检查后复测均 PASS。每次创建新随机 unit，使用本次独占 RuntimeDirectory 0700 和 Core UDS 0600。`verify-systemd.py` 调用 `/usr/bin/env -u DISPLAY <core> status|doctor --json`，核对服务 MainPID 和实例一致、状态协议/OS 正确、身份字段非空，但仅保存 REDACTED。

最终脚本输出顺序为：

```text
HEADLESS_STATUS_DOCTOR_AND_PRIVATE_IPC=PASS
TRANSIENT_DIRECT_EXEC_START=PASS
STOP_PROCESS_ABSENT_AND_REQUESTS_REJECTED=PASS
HEADLESS_STATUS_DOCTOR_AND_PRIVATE_IPC=PASS
RESTART_NEW_PID_AND_INSTANCE=PASS
HEADLESS_STATUS_DOCTOR_AND_PRIVATE_IPC=PASS
RESTART_NEW_PID_AND_INSTANCE=PASS
SYSTEMCTL_RESTART=PASS
STOP_PROCESS_ABSENT_AND_REQUESTS_REJECTED=PASS
SANITIZED_ARTIFACTS=<本目录>/artifacts/systemd.<本次随机后缀>
SYSTEMD_LIFECYCLE_POC=PASS
```

最终复测的本地忽略证据目录为 `artifacts/systemd.SPVBL3mQ/`：

| 阶段 | systemd MainPID = status PID = doctor PID | 运行目录 / UDS | 保存身份 |
| --- | --- | --- | --- |
| 首次启动 | 521 | 0700 / 0600 | REDACTED |
| 停止后再次启动 | 631 | 0700 / 0600 | REDACTED |
| 原生 systemctl restart | 661 | 0700 / 0600 | REDACTED |

忽略目录保留 metadata.json（仅 unit/端点/二进制哈希）、first.json、stopped.json、second.json、restarted.json、cleanup.json。运行时原始身份、环境变量、认证文件及日志均未保存/提交。最终额外只读检查输出 `MainPID=0`、`LoadState=not-found`，运行目录不存在检查 exit 0；不停止其他 unit，不删除用户目录。提交前凭据模式扫描未命中，源码及脱敏证据人工复核未见秘密。

## 未覆盖与阻塞

| 范围 | 结果 | 原因/限制 |
| --- | --- | --- |
| Windows SCM 宿主编译、纯配置/状态机测试 | PASS | 固定 x/sys v0.49.0；fake child 测试仅作单元证据 |
| Windows 当前用户原生 Job/Named Pipe/Stop/句柄关闭 | PASS | 显式门控测试；实际 2 个本 Job 成员停止后归零；不包含任意宿主崩溃 |
| Windows Server 2022 CI 的 SCM 创建/两轮启停/重启/删除、原 ACL 跨身份拒绝 | PASS | 上述 run/job；真实 LocalSystem 和镜像/直接子进程核验、零退出码、handles 退出和最终服务不存在 |
| 本机 Windows 11 SCM / Windows 10/11 完整服务矩阵 | NOT_RUN | 本机仍未提权或创建服务；Server 2022 CI 不替代产品目标系统矩阵 |
| WSL Ubuntu 22.04 transient systemd + headless CLI | PASS | 真实 systemd/UDS/status/doctor/stop/restart；身份脱敏 |
| Ubuntu Desktop 22.04/24.04 完整 GUI 矩阵 | NOT_RUN | 本目录不覆盖 GUI；22.04 局部安装/Gio/IPC 另见桌面运行时报告，不能替代两版本完整矩阵 |
| Ubuntu Server 22.04/24.04 CLI/systemd 真机 | NOT_RUN | WSL 不等同正式 Server 环境，未安装生产 Core/CLI .deb |
| 注销/系统重启、自启、崩溃恢复、升级/卸载 | NOT_RUN | transient unit 不持久安装，未改变主机启动配置 |
| MCP / 跨网络真实调用 | NOT_RUN | 此目录只验证本机只读 IPC 生命周期 |
| A16/A22/P04 正式验收 | NOT_RUN | 缺上述正式平台完整矩阵和产品身份策略 |

Windows 宿主采用强制结束自身 job 的 POC 策略，不宣称已支持产品任务的优雅停止/事务恢复。SCM Running 不表示 IPC 已健康。既有 desktop-service 管道 DACL 仅授予进程当前身份；本 POC 未解决不同登录用户访问系统服务的最终权限设计，也未读取任何用户 SSH 凭据。Windows 启动中断/任意崩溃的清理边界仍需真实故障 SCM/Job 测试，不能由本次正常生命周期或 kill-on-close 标志推断全部场景通过。

下一步由集成负责人将这些局部结果纳入 P04/P05；Windows 10/11 完整服务矩阵、Ubuntu Desktop/Server 真机、产品服务身份与调用者授权、持久安装/恢复仍为依赖。本目录未新增公共契约或自行推送/创建 PR，CI run 与提交按上文索引追溯；文档提交以 Git 历史为准。
