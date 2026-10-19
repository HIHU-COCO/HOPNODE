# P04 Windows SCM 一次性管理员 CI 验证

作者：HIHU-COCO。日期：2026-10-19（Asia/Shanghai）。关联 P04、F16/F18、A16/A22。此处准备真实 SCM 集成测试；本机普通用户仅编译并明确 Skip，**尚无 SCM 实际运行 PASS**。

## 最小设计与边界

`TestSCMNativeLifecycle` 复用本目录 SCM-only 宿主及 `../desktop-service` 独立 fixture。仅当 `HOPNODE_POC_RUN_SCM=1`，且提供现存的绝对 `HOPNODE_POC_SCM_HOST_EXE` / `HOPNODE_POC_CORE_EXE` 时启用。默认、显式 0、路径缺失/相对/文件不存在均明确 Skip；无效 gate、非 exe/目录、同一文件报错。启用且路径齐备后，未提升的调用者、LocalSystem 调用者或 SCM 权限失败均 **Fail，不能 Skip**。测试不调用 UAC、自我提权或密码工具。

每次生成 128-bit 随机服务名及管道名，先确认均不存在。只通过新建成功返回的服务 handle 管理 `HopNode-P04-POC-<随机值>`，配置 `SERVICE_WIN32_OWN_PROCESS`、manual/demand、LocalSystem、无依赖；无自动启动、恢复策略、SCM 全局或既有服务修改。不改管道 ACL、用户配置，不重启系统、不外联。

两轮执行真实 `Start → Running → Stop → Stopped`，第二轮是 Windows 服务的 Stop/Start 重启；不是模拟 handler。每轮从 SCM PID 核对宿主镜像和 LocalSystem token，以 Toolhelp 直接 parent PID 找 Core，再核对精确镜像和 LocalSystem token；持有同步/终止句柄以避免 PID 重用清理错误，第二轮必须是新 host/Core PID。Stopped 必须同时满足 PID=0、Win32ExitCode=0、ServiceSpecificExitCode=0；非零退出不能冒充成功停止，仍继续清理/删除自建服务。只记录 PID、固定身份结论及二进制 SHA-256，未读写原始 CLI 身份响应。

Core 使用现有身份专属 Named Pipe DACL。独立管理员调用者必须从该随机端点得到实际 `ERROR_ACCESS_DENIED`；不存在端点、超时或其他错误不能算成功，意外允许访问直接 Fail。此负例说明当前身份 ACL 拒绝另一用户；没有放宽到 Administrators/Everyone，也没有宣称正式产品调用者授权已完成。**SCM Running 加管道拒绝不证明正向 status API 或业务健康**；LocalSystem 正向请求、业务、注销/重启、崩溃恢复均不在此测试中。

成功 CreateService 后立即登记 cleanup，SCM manager 关闭回调先登记，使其 LIFO 顺序晚于服务清理。任何断言失败仍先停止本次服务，等待所有已核验 host/Core 原生句柄 signaled；必要时只终止已核验镜像的精确原生句柄，再核验退出。随后 Delete/Close 自建服务，循环查询直至该服务确实不存在，最后关闭进程和 manager 句柄。清理失败使测试 Fail。测试不会按进程名全局 kill、修改其他服务或递归删除目录。

不能保证 Go 进程被外部强杀、机器中断时执行 cleanup，因此只在已授权的一次性 runner 上开启，给本例 5 分钟总预算。成功标准为整项测试 PASS **且** `SCM_OWNED_CLEANUP=PASS`；不能只引用清理前打印的阶段标记。

## GitHub `windows-2022` 精确命令

仅由集成负责人在一次性 GitHub-hosted `windows-2022` 管理员 job 加入下列 PowerShell 步骤。Go 1.27.2 已在 PATH，工作目录为仓库根；本目录不修改 `.github/`。该 CI OS 是 Windows Server 2022，不替代 Windows 10/11 真机矩阵。

```powershell
$ErrorActionPreference = 'Stop'
$scmPocBuildDir = Join-Path $env:RUNNER_TEMP ('hopnode-p04-scm-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $scmPocBuildDir | Out-Null
$scmPocCore = Join-Path $scmPocBuildDir 'hopnodectl-poc.exe'
$scmPocHost = Join-Path $scmPocBuildDir 'hopnode-service-poc.exe'

go -C tests/pocs/desktop-service build -mod=readonly -trimpath -o $scmPocCore .
if ($LASTEXITCODE -ne 0) { throw 'Core POC build failed' }
go -C tests/pocs/service-lifecycle build -mod=readonly -trimpath -o $scmPocHost .
if ($LASTEXITCODE -ne 0) { throw 'SCM host build failed' }
foreach ($scmPocBinary in @($scmPocCore, $scmPocHost)) {
    if (-not [IO.Path]::IsPathFullyQualified($scmPocBinary) -or
        -not (Test-Path -LiteralPath $scmPocBinary -PathType Leaf)) {
        throw 'SCM CI requires both built absolute executable paths'
    }
}

$previousScmGate = $env:HOPNODE_POC_RUN_SCM
$previousScmHost = $env:HOPNODE_POC_SCM_HOST_EXE
$previousScmCore = $env:HOPNODE_POC_CORE_EXE
try {
    $env:HOPNODE_POC_RUN_SCM = '1'
    $env:HOPNODE_POC_SCM_HOST_EXE = $scmPocHost
    $env:HOPNODE_POC_CORE_EXE = $scmPocCore
    Push-Location tests/pocs/service-lifecycle
    try {
        go test -mod=readonly -timeout 5m -count=1 -run '^TestSCMNativeLifecycle$' -v ./...
        if ($LASTEXITCODE -ne 0) { throw 'SCM lifecycle or owned cleanup failed' }
    } finally {
        Pop-Location
    }
} finally {
    $env:HOPNODE_POC_RUN_SCM = $previousScmGate
    $env:HOPNODE_POC_SCM_HOST_EXE = $previousScmHost
    $env:HOPNODE_POC_CORE_EXE = $previousScmCore
}
```

前置构建与绝对路径检查必须保留，避免因 fixture 缺失产生 Skip 被当 CI PASS。CI 日志必须看到两条 `SCM_ROUND_*`、`SCM_OWNED_CLEANUP=PASS` 和整项 `--- PASS: TestSCMNativeLifecycle`；任何 Skip 均记 NOT_RUN，不是 SCM 通过。二进制目录留作本次 runner 证据，不删除用户路径。

## 本机仅编译 / 默认 Skip

本目录执行；普通用户不要设置 SCM 开门值，不创建本机服务：

```powershell
$env:HOPNODE_POC_RUN_SCM = '0'
$env:HOPNODE_POC_CORE_EXE = ''
go mod verify
go test -mod=readonly -timeout 30s -count=1 -v ./...
go vet ./...
New-Item -ItemType Directory -Path artifacts -Force | Out-Null
go build -mod=readonly -trimpath -o artifacts/hopnode-service-poc.exe .
go test -c -mod=readonly -o artifacts/service-lifecycle.test.exe .
```

2026-10-19 实际本机结果：默认套件 PASS，SCM 与需显式 Core 的原生 Job 测试各明确 Skip；新门控/路径 9 个子项、权限错误 3 个子项、停止成功判定 5 个子项 PASS；mod verify、vet、宿主 build、test -c 均退出 0。格式已 gofmt。本次未开启 SCM、未提权、未创建服务，P04 与正式验收保持未完成。
