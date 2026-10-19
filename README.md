# 跃点 HopNode

跨越边界，连接无限。

**当前状态：工程初始化与技术 POC，尚未交付可用产品或正式安装包。** Windows/Ubuntu GUI、CLI 和后台服务均保留为首版目标；不将工程空壳、隔离夹具或 WSL 结果标为正式验收。

依据：[PRD](PRD.md)、[技术设计](Design.md)、[实施计划](Plan.md)、[执行规则](AGENTS.md)。实际进度与缺失条件见 [执行台账](docs/implementation/status.md)、[测试环境](docs/implementation/test-environment.md)、[风险](docs/implementation/risks.md)。

## 开发环境

- Go 1.27.2；Rust 1.99.1（rustfmt/clippy）；Node 26.7.0；pnpm 11.0.9。
- Windows Tauri 需要 MSVC C++ Build Tools 和 WebView2；Ubuntu Tauri 需要 GTK3/WebKitGTK 4.1 开发依赖。纯 Go CLI 无图形依赖。
- `.tools/` 为本机隔离工具目录，不提交；使用系统安装的等版本工具也可。不要上传凭据、数据库、日志或构建产物。

```powershell
# PowerShell：.cmd 避开当前机器对 npm/pnpm.ps1 的脚本限制
pnpm.cmd install --frozen-lockfile
pnpm.cmd check
pnpm.cmd build
pnpm.cmd dev
go test ./...
go vet ./...
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/build-core.ps1
./artifacts/windows/hopnodectl.exe version
```

```bash
# Ubuntu 纯 Go 工程入口
go test ./...
go vet ./...
bash scripts/build-core.sh
./artifacts/linux/hopnodectl version
```

本阶段三个 Go 程序仅提供 `version` / `--version` / `--help`。其他命令返回 `FEATURE_NOT_IMPLEMENTED` 与非零码；它们不会安装服务、执行远程操作或伪造成功状态。隔离 POC 使用独立模块，执行方法见 `tests/pocs/*/README.md`。

## 当前验证边界

工程 CI 检查 Go、前端、Rust 和隔离 POC。真实网络集成与发布工作流设为手动门禁，在缺少驱动/验收时明确失败。不存在正式 Release。完整产品完成要求为 F01–F18、A01–A27 真实验证，并交付 Windows GUI/CLI/Service、Ubuntu GUI `.deb`、Ubuntu Core/CLI `.deb`。

本机现有 SSH、Tailscale 和 Codex 配置不会由 POC 自动改写。测试密钥只在临时夹具内生成。许可证 [待所有者决定](docs/LICENSE-PENDING.md)。
