# GitHub 工程基线

CI 使用固定 Action commit、Go/Rust/Node/pnpm；Windows 与 Ubuntu 分开检查。P01 网络集成仅手动触发，挂到 `hopnode-integration` Environment。管理方需要给该 Environment 配置保护与隔离设备，普通 PR 不接触网络凭据。

建议 main 启用禁止 force push、要求 PR 审查、要求 Engineering baseline 的所有作业成功。仓库是否支持/启用这些设置以 GitHub 实际配置为准，本文件不冒充已设置。

初始化的 integration/release 工作流故意在未具备驱动或验收时明确失败，不发布空壳产物，不将 NOT_RUN 隐藏为绿色成功。其当前任务状态为部分实现，后续任务依赖未满足。正式发布需 L01–L06 与 A01–A27 全部通过，再由 L07 创建受保护标签。

读取远端 tags 固定 Action commit 的日期：2026-10-19。Actions Runner 镜像由 GitHub 管理，不能宣称固定到不可变镜像；正式候选构建需保存实际镜像和 SDK 信息。
