# P04 桌面 / CLI / 独立 Core IPC fixture

创建者：HIHU-COCO；日期：2026-10-19。此独立 Go module 验证本地运输与进程隔离，不冻结生产 IPC，也不替代 Windows Service/systemd 实现。

## 接口与边界

- HTTP/1.1 `GET /v0/status`：`api_version=poc.v0`、版本、真实 PID、随机进程实例标识、OS、实际进程用户身份、UTC 启动时间。
- `GET /v0/version`：版本；只读请求，没有设备、任意命令或文件接口。
- Windows：`\\.\pipe\hopnode-poc-%USERNAME%`，go-winio 0.6.4；受保护 DACL 仅授予当前 SID，go-winio 拒绝远程 Named Pipe 客户端。
- Linux：`$XDG_RUNTIME_DIR/hopnode-poc.sock`。父目录必须由当前用户拥有且权限不向组/其他用户开放，socket 为 `0600`；不删除或替换已有路径。
- 可通过环境变量 `HOPNODE_POC_ENDPOINT` 指定隔离测试端点，GUI 与 CLI 必须使用同一环境值。GUI 不接受 Web 前端传入任意端点。
- 超时 3 秒、64 KiB 响应上限；服务断开不会返回演示数据。

## Windows 复现

```powershell
go build -o "$env:TEMP\hopnodectl-poc.exe" .
# 终端 A
& "$env:TEMP\hopnodectl-poc.exe" serve
# 终端 B
& "$env:TEMP\hopnodectl-poc.exe" status --json
& "$env:TEMP\hopnodectl-poc.exe" doctor
```

从仓库根启动 Tauri，刷新状态并核对 CLI / GUI 的 PID 与实例标识相同。关闭 GUI 后在 CLI 再次读取相同 PID/实例；退出桌面程序后同样验证；停止 Core 后 GUI/CLI 均应报告真实连接失败。

自动隔离冒烟从本目录执行：

```powershell
.\test-runtime.ps1 -GoExecutable (Get-Command go).Source -CargoExecutable (Get-Command cargo).Source
```

它在临时目录构建二进制、启动隐藏的独立 controller、验证 CLI 和 Rust 薄桥的真实 Named Pipe 通信，然后停止该进程并验证错误。不更改用户服务或网络配置。`CargoExecutable` 可省略，此时只检查独立 Core/CLI。

## Ubuntu / 无 DISPLAY 复现

```sh
go test -race ./...
go vet ./...
go build -o /tmp/hopnodectl-poc .
poc_runtime=$(mktemp -d)
chmod 700 "$poc_runtime"
export HOPNODE_POC_ENDPOINT="$poc_runtime/core.sock"
/tmp/hopnodectl-poc serve &
poc_pid=$!
env -u DISPLAY /tmp/hopnodectl-poc status --json
env -u DISPLAY /tmp/hopnodectl-poc doctor
stat -c '%a' "$HOPNODE_POC_ENDPOINT"
# 独立启动 GUI，比较 PID/实例，然后关闭 GUI 再读 CLI
kill "$poc_pid"
```

Rust 真实传输测试默认 ignored，必须先启动独立 Core、设置相同端点，再显式执行：

```text
cargo test --locked --manifest-path apps/desktop/src-tauri/Cargo.toml real_controller_roundtrip -- --ignored --nocapture
```

报告身份时脱敏，勿将个人运行日志提交仓库。WSL 编译/UDS 测试不能充当 Ubuntu Desktop 22.04/24.04 菜单启动、GNOME/Wayland/X11、`.deb` 安装或 Ubuntu Server 实体机验收。当前不支持自动服务安装、注销恢复、系统重启恢复或 agent 身份划分，这些保持 NOT_RUN / 后续实现。
