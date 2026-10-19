# HopNode 执行台账

更新日期：2026-10-19（Asia/Shanghai）。依据 PRD V1.4、Design V1.3、Plan V1.3、AGENTS V1.1。

当前为工程初始化与风险 POC，尚未达到正式产品发布条件。POC 本机测试不能替代 A01–A27 真机验收。

**施工顺序已获授权**：2026-10-19 用户要求重新调整，按 [ADR-0001](../decisions/ADR-0001-v1-stack.md) 和[条件路线](proposals/P05-conditional-development.md)进入 K01。G0/G1/P05 的 CONDITIONAL 只表示限定开发出口获准，原有测试缺口继续保留；本轮只调整计划，K01 READY，尚未开工或冻结契约。

## 任务

| ID | 状态 | 负责人 | 证据 / 阻塞 |
| --- | --- | --- | --- |
| T00 | DONE | HIHU-COCO / 集成 | test-environment.md、risks.md、完整 F/A 台账 |
| T01 | DONE | HIHU-COCO / 集成 | pnpm check/build、开发服务、Go build/test/vet、Windows/Linux Rust fmt/clippy/test/build PASS；见 test-results/T00-T01-baseline.md、pocs/desktop-runtime.md |
| T02 | DONE | HIHU-COCO / 集成 | PR #1真实CI；6477909双平台Go/Rust、前端、Ubuntu deb与实际Windows Server2022 SCM全部PASS，Actions run39203998317；后续提交独立复验 |
| T03 | BLOCKED | HIHU-COCO / 集成 | 已有Windows、WSL22.04、授权LAN Ubuntu22.04 Desktop；后者SSH/PTY/SFTP通过；缺完整版本/角色真机矩阵、跨网和隔离部署 |
| P01 | BLOCKED | ssh_finish / 集成 | Headscale0.29.5/便携TS1.98.4单机userspace：TLS正负例、双节点、合成TCP、策略切换、节点expire/旧key拒绝及清理PASS；非跨网/强制DERP；见network-isolated/RESULTS.md及ADR-network.md |
| P02 | BLOCKED | ssh_finish / poc/ssh | Windows→隔离Ubuntu及独立LAN Ubuntu22.04命令/PTY/10MiB SFTP PASS；已修单pin算法协商；第三方WebSSH的PTY/中文/resize与应用退出补测见test-results/P02-webssh-2026-10-19.md，不替Go/Tauri验收；缺Windows sshd与跨平台长任务实测；tests/pocs/ssh/report.md |
| P03 | IN_PROGRESS | mcp_finish / poc/mcp | Codex0.153.4真实echo/预期错误/status与2025-06-18协商PASS；SDK取消/race PASS；app-server三轮预检均未发模型turn，配置清单/进程归属需补证，协议取消仍未证实；test-results/P03-mcp-2026-10-19.md |
| P04 | BLOCKED | desktop_finish、mcp_finish / 集成 | Windows NSIS/真实GUI IPC/卸载、Server2022 CI实际SCM两轮启停/ACL负例/删除PASS；独立Ubuntu22.04 X11工程deb安装/desktop入口/显示刷新/内核PID核验UDS/独立Core/卸载PASS，user systemd亦PASS；缺Win10/11服务、24.04及完整身份矩阵；见P04及SCM报告 |
| P05 | CONDITIONAL | HIHU-COCO / 集成 | 四项真实POC结论/缺口及用户授权已归档ADR-0001；条件开发出口生效，允许K01；原实测未全部通过，未冻结生产契约或正式组网版本 |
| K01 | READY | HIHU-COCO / 集成 | 已授权P05/G1条件前置满足；按proposals/K01-gap-review.md十二项缺口补齐完整契约。已有35例/26工具、Directory/事件23例及envelope54对仅DRAFT局部PASS；本轮未实现或冻结 |
| K02 | TODO | HIHU-COCO / 集成 | 等K01满足完成判据并经CI/审查合入共享基线，再实现公共Core/IPC/桥接 |
| K03 | TODO | HIHU-COCO / 集成 | 等K02完成并合入，再完善夹具、契约/类型检查及CI门禁 |
| K04 | TODO | HIHU-COCO / 集成 | 等K03完成并合入，再冻结兼容/身份/双模式约定与测量标准，逐项保留A25待测指标 |
| N01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| N02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| N03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| N04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| N05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| R01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| R02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| R03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| R04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| R05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| R06 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| M01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| M02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| M03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| M04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| U01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| U02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| U03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| U04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| U05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| U06 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| S01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| S02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| S03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| S04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| S05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| S06 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| C01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| C02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| C03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| C04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| C05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I06 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I07 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I08 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I09 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I10 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I11 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| I12 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L01 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L05 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L06 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| L07 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |

## PRD 功能追踪

| 编号 | 状态 | 说明 |
| --- | --- | --- |
| F01 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F02 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F03 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F04 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F05 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F06 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F07 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F08 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F09 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F10 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F11 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F12 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F13 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F14 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F15 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F16 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F17 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |
| F18 | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |

## 正式验收（0/27 PASS）

| 编号 | 结果 | 环境 / 步骤 / 证据 / 验收日期 |
| --- | --- | --- |
| A01 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A02 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A03 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A04 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A05 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A06 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A07 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A08 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A09 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A10 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A11 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A12 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A13 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A14 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A15 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A16 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A17 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A18 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A19 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A20 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A21 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A22 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A23 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A24 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A25 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A26 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |
| A27 | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |

## 检查点

| 检查点 | 当前决定 | 含义 / 下一门槛 |
| --- | --- | --- |
| G0 | CONDITIONAL | 工程/CI/锁定方式及环境缺口记录已具备；T03全量矩阵仍BLOCKED，已有资源可继续独立测试 |
| G1 | CONDITIONAL | ADR-0001条件施工路线已获用户授权；剩余POC风险继续开放，允许正式K01 |
| G2 | TODO | K01→K02→K03→K04必须串行满足判据、合入共享基线并通过契约/CI；当前未通过 |
| G3–G6 | TODO | 六条业务线尚未启动；真实集成、全平台包及27项验收门槛保留 |

下一项为 **K01**。优先补齐 envelope/共享模型与全部方法结果，再处理 Directory/身份/撤销、取消/幂等、事件/兼容及有界资源规则，最后以Go/TypeScript类型及完整机器校验冻结。G2通过后各模块按Plan依赖实现；原POC缺环境不一概禁止实现，但未满足任务自身真机完成判据不得置DONE。T03、P01–P04和A01–A27不因施工授权改为通过。

每项测试的命令、结果和范围写入 test-results/。Git commit 见 Git 历史；无 PR 时不得填写虚构链接。

## 本次施工顺序调整校验（2026-10-19）

- 任务P05及Plan/AGENTS同步，涉及F01–F18/A01–A27的施工与验收顺序；本轮仅修改13份Markdown文档，没有产品代码、schema、根依赖或工作流变更。
- `git diff --check` PASS；`python scripts/verify-ledger.py` 实际输出：`PASS: all 18 feature and 27 acceptance rows present; acceptance PASS rows require evidence`。
- 只读文档一致性检查：以调整前`387c1b6`为基线，64个任务编号集合完整保留，18行F台账和27行A台账原样保留，A通过数仍0；PRD/Design无差异；50个本地Markdown链接全部存在；169个跟踪文件的真实目标/口令扫描命中0。脱敏细节仅保留在忽略的`.tools/plan-adjustment-proof.json`。
- Windows、Ubuntu Desktop GUI、Ubuntu Server CLI、MCP及跨网络真调用：本轮均NOT_RUN（文档调整）；既有平台实证保持原结果，本轮没有新增平台验收。
- 公共文档先决ADR/授权记录已提交`0064418`，后续对齐提交及CI见当前Git历史和Draft PR #1；完整环境/实测缺口及返工风险继续按ADR-0001与risks.md跟踪。
