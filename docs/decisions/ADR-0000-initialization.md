# ADR-0000：工程启动与 POC 隔离

日期：2026-10-19；状态：实施工程基线，正式业务契约待 P05/K01。

起始目录仅包含四份依据文档，无 Git 历史或代码。用户授权执行计划、创建子智能体，并明确 Git 作者 HIHU-COCO 和 GitHub HOPNODE 远端。

由集成负责人单独创建根 Go 模块、workspace、CI 和 cmd 版本入口；子智能体分别在 .worktrees/ssh、mcp、desktop 独立分支写 P02/P03/P04 隔离 POC。源文件用用户确认的 Git 作者。POC 使用独立模块，禁止把未经风险评审的消息模型发布为正式契约。

T01 的 Core/CLI/Directory 仅声明工程版本；尚未交付的命令明确返回 FEATURE_NOT_IMPLEMENTED 和非零码，不能伪装运行服务或远程成功。桌面空壳仅展示真实 POC 状态。尚未满足 T/P/K 出口门槛，因此 N/R/M/U/S/C 尚不开放。

本 ADR 不改变任何 PRD 范围，不批准用模拟响应替代产品验收。无测试环境时依据 Plan §10.2 继续独立工程、POC 和帮助文档，缺失项维持 BLOCKED/NOT_RUN。
