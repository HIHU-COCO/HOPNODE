# P04 已授权 Ubuntu Desktop 的独立暂存验证

作者：HIHU-COCO；日期：2026-10-19。适用：已授权的普通桌面用户、Ubuntu 22.04 x64、已有 X11/GNOME 与所需 GTK/WebKitGTK 运行库。集成负责人负责安全传输与执行；本方案不保存或接触 SSH 密码，不更改 SSH 配置。

## 安全边界与能证明的事实

使用 `dpkg-deb -x` 将已构建 GUI `.deb` 提取到唯一的 0700 `/tmp/hopnode-p04-<uid>-...`。不安装包、不 sudo、不修改 systemd 系统单元、网络、桌面快捷方式或现有程序。Go fixture 为独立纯 CLI/Core 二进制；所有程序/配置/data/cache/state、UDS 都在该唯一目录。GUI 仅使用负责人明确提供的既有 DISPLAY、XAUTHORITY 与用户 DBus，脚本不会猜测或改写这些值。

真实 GUI→UDS 证据使用 Go fixture 可选 `HOPNODE_POC_AUDIT_FILE`：0600、O_EXCL 新文件，以 Linux `SO_PEERCRED` 记录内核返回的 peer PID/UID、请求方法/路径、响应码及 controller PID/实例。只有请求 peer PID **等于实际启动的 GUI PID**、UID 匹配、`GET /v0/status` 返回 200 才算真实 GUI IPC 通过。CLI 的 PID 与 GUI 不同，不会被计入 GUI 成功证据。此只读 POC 审计不是生产认证/授权契约，日志只保留在私有证据目录。

可选 AT-SPI helper 只定位本次 GUI PID 的应用节点，并只遍历该应用树；不读取或控制其他应用的窗口。若该应用暴露「已连接」、实际 Core PID 与「刷新状态」按钮，调用该按钮的真实 accessibility action，并要求新增一条来自同一 GUI PID 的内核验证请求。没有可靠 accessibility 时，`visible_ui_and_refresh=NOT_RUN`，总状态为 PARTIAL；不能把进程存活或单纯内核 IPC 证据宣称为视觉/刷新通过。

退出 GUI 进程后，用明确移除 DISPLAY/XAUTHORITY/DBus 的 CLI 环境再次读取，必须保持相同 Core PID 和实例。退出是进程终止验证，不是托盘关闭/恢复、注销或重启验收。

## 输入与只读前置检测

负责人先在远端创建独有的 0700 上传目录，通过已授权传输放置：

- 当前 Linux GUI `.deb`；
- 从本目录最新源码构建的 Linux `hopnodectl-poc`（必须包含可选 kernel-peer audit）；
- `linux-session-probe.py`、`linux-desktop-stage.py`、`linux-atspi-own-window.py`，三份脚本保持同目录。

以下命令只查询当前用户的 manager 环境和工具可用性；只输出字段存在性，不输出私有环境值：

```sh
python3 linux-session-probe.py
```

确认后在负责人指定的全新文件中保存最少必要环境：

```sh
python3 linux-session-probe.py --write-environment /path/to/private-upload/session.json
```

脚本用 0600/O_EXCL 创建元数据，拒绝覆盖。只保存 uid、DISPLAY、XAUTHORITY、DBUS_SESSION_BUS_ADDRESS、XDG_RUNTIME_DIR、XDG_SESSION_TYPE；不记录完整环境，不保存 SSH 材料。如果 manager 没有明确 DISPLAY/XAUTHORITY，先停止依赖 GUI 的执行，由负责人通过可信的活跃 X11 会话来源补齐私有测试元数据；不根据 `:0`、用户名或文件存在性推测。不得把原始 metadata 输出到聊天或提交 Git。

## 执行与可选 user transient unit

```sh
python3 linux-desktop-stage.py \
  --deb /path/to/private-upload/gui.deb \
  --core /path/to/private-upload/hopnodectl-poc \
  --session-env /path/to/private-upload/session.json
```

可独立复测用户态 systemd Core 生命周期；仅新建带 UUID 的 transient user 单元、RuntimeMaxSec=180、KillMode=control-group，不创建单元文件或启用开机启动：

```sh
python3 linux-desktop-stage.py \
  --deb /path/to/private-upload/gui.deb \
  --core /path/to/private-upload/hopnodectl-poc \
  --session-env /path/to/private-upload/session.json \
  --user-unit
```

user-unit 模式须额外上传同目录 `linux-user-unit.py`。生命周期对象在调用 `systemd-run` 前即登记；即使命令超时、非零或抛异常，也会对该唯一 UUID 单元执行 stop，再独立查询 MainPID/ActiveState/LoadState。无法核验停止时 FAIL 并保留 stage，不以启动命令没有正常返回推断单元不存在。`python3 test-user-unit.py` 注入「单元已创建但父启动命令超时」、停止失败后仍 active、停止后查询超时三种故障，实际 3 测试 PASS；这些故障回归不创建真实单元，也不替代上述真机成功路径。

输出只有检查状态、私有证据目录和清理结果；没有完整会话环境、用户名、UI 文本、地址或凭据。成功会留下 `report.json`、可选 `ui-accessibility.json`、`ipc-audit.jsonl` 和本次私有日志。目录位置是证据索引，不代表 GUI 已通过全部验收。

## 用户授权临时安装后的真实 desktop-entry 模式

用户已明确授权临时安装、测试后卸载，且不安装新依赖。集成负责人另行核对无已有安装、包内容无冲突、没有维护脚本及依赖已满足，使用既有安全通道执行系统安装和最终卸载；本 runner 始终以普通桌面用户运行，不接触 sudo 密码，也不安装、卸载或修改系统包。即使探针失败，负责人仍需核对 owned 资源并完成该次安装的卸载。

安装当前已验证的包后，在原有上传目录额外放置 `linux-desktop-entry.py`，可复用同一接口：

```sh
python3 linux-desktop-stage.py \
  --deb /path/to/private-upload/gui.deb \
  --core /path/to/private-upload/hopnodectl-poc \
  --session-env /path/to/private-upload/session.json \
  --installed-desktop /usr/share/applications/hopnode-desktop-poc.desktop
```

模式只接受该精确系统入口和 `/usr/bin/hopnode-desktop`，二者及父目录必须 root-owned、不可被组/其他用户写入，且无 symlink。它们的 SHA-256 必须与原 `.deb` 解包内容一致；拒绝 Terminal、DBusActivatable、Hidden、NoDisplay 或非预期 Exec。通过真实 `Gio.DesktopAppInfo` 的 `launch_uris_as_manager_with_fds` 启动安装入口，AppLaunchContext 清除继承环境后仅写入明确的 X11 会话及隔离 XDG/endpoint，PATH 限于 `/usr/bin:/bin`。这验证桌面入口启动机制，不自动等同 GNOME 菜单中人工点击。

Gio 的 child-setup 仅对本次新进程执行 `setsid`，PID callback 在启动时记入私有 manifest；先核对新 session/group，再核对 `/proc/<pid>/exe` 确实为已安装二进制。之后复用真实 GUI 内核 PID 的 UDS 审计、AT-SPI 状态/刷新、GUI 退出后独立 Core 与精确清理。报告新增 `gui_launch_method`、`installed_payload_sha256`、`installed_desktop_exec_verified`，保留原五项功能检查；不把未执行的安装/卸载核对写成 runner 的 PASS。

本地回归命令 `python3 test-desktop-entry.py` 只启动唯一私有 `.desktop` 的 sleep fixture，实际核验 Gio callback PID、新 session/group、精确环境、信号退出与进程组消失；不启动用户应用、不等同 GUI 验收。Ubuntu 22.04 WSL 实际运行 1 测试 PASS。初次回归在 exec 完成前读取 `/proc` 环境导致失败，增加等待实际 exec 环境后通过；未将该失败算作产品成功。

## 2026-10-19 真机执行记录

集成负责人已在 Ubuntu Desktop 22.04 x86_64 / kernel 6.8.0-138-generic / GNOME 42.9 / X11 的普通 uid 1000 会话执行上述两个命令，均退出 0、总体 PASS。GTK 3.24.33、WebKitGTK 4.1 2.50.4、AppIndicator 0.5.90 为机器原有依赖，本次没有安装系统包或使用 sudo。

每轮 `headless_cli_before_gui`、`uds_mode_0600`、`gui_to_real_uds`、`visible_ui_and_refresh`、`same_core_after_gui_exit_without_display` 均 PASS。AT-SPI 实际读取本次 GUI 的已连接及 Core PID，再点击刷新；各有恰好 2 条 GUI 内核 PID 的 `GET /v0/status=200` 请求，全部审计身份核验通过。两轮 GUI stderr 均为 0 字节，此次完整 GNOME 环境未复现此前 WSLg GTK 警告，既有 WSLg 失败记录仍然有效。

普通轮 GUI/Core 两个 owned 进程组均已不存在。user-unit 轮 GUI 组已不存在，独有单元停止后 `MainPID=0`、`ActiveState=inactive`、`LoadState=not-found`。两轮 extracted/bin/runtime/config/data/cache/state 七个自建运行子目录均已清理，保留私有证据。

复核使用集成工作树忽略目录 `.tools/authorized-ubuntu-22/` 下两轮报告、审计、UI 布尔结果和 `gui-proof-summary.json`，未读取或提交私有会话环境/目标元数据。生成物哈希和逐项结果见 `docs/implementation/test-results/pocs/desktop-runtime.md`。这里的 PASS 仅指暂存探针列明项目：`.deb` 是解包运行，未系统安装；Wayland、24.04、桌面菜单、托盘、root systemd、Windows SCM 和正式验收仍未覆盖，P04 保持 BLOCKED。

## 清理与限制

脚本仅终止自己通过新 session/process-group 启动的 GUI/Core；发信号前从 `/proc` 确认该组成员 UID/session，TERM/KILL 后必须通过内核组存在性和 `/proc` 成员后验共同确认组已不存在，不能仅凭 leader 退出。user 模式仅停止自己生成 UUID 的 transient 单元，随后核对 MainPID=0 且 inactive/failed，或单元已不存在，并保存脱敏状态。

只有所有 owned 进程组/单元都确认停止后才清理文件，否则保留全部 stage 子目录并标 FAIL。每个将递归清理的目录在删除前必须解析为本次私有 stage 的直接子目录，且不能是 symlink。只删除自己创建的 extracted/bin/runtime/config/data/cache/state，保留私有证据，不删除上传目录、原有用户配置/数据或任何系统文件。`owned-processes.json` 在启动时即记录本次 PID/group/独有单元名，供意外中断后负责人只核查自己的测试资源。SIGTERM/SIGINT 会进入清理；SIGKILL 不能保证自动清理，负责人须按 manifest 验证资源归属后处理，不应直接递归删除仍可能使用的 stage。

`PASS` 仅表示本次全部已列检查通过；`PARTIAL` 表示真实 IPC/独立 Core 等通过但 UI accessibility 未验证；`FAIL` 保留真实错误。runner 自身不安装 `.deb`；默认解包模式与 installed desktop-entry 模式均不能代替负责人独立核对系统安装/卸载、GNOME 菜单点击、Wayland、24.04、系统级 systemd 或 A01–A27 完整验收。
