# HopNode 执行台账

更新日期：2026-10-19（Asia/Shanghai）。依据 PRD V1.4、Design V1.3、Plan V1.2、AGENTS V1.0。

当前为工程初始化与风险 POC，尚未达到正式产品发布条件。POC 本机测试不能替代 A01–A27 真机验收。

## 任务

| ID | 状态 | 负责人 | 证据 / 阻塞 |
| --- | --- | --- | --- |
| T00 | DONE | HIHU-COCO / 集成 | test-environment.md、risks.md、完整 F/A 台账 |
| T01 | DONE | HIHU-COCO / 集成 | pnpm check/build、开发服务、Go build/test/vet、Windows/Linux Rust fmt/clippy/test/build PASS；见 test-results/T00-T01-baseline.md、pocs/desktop-runtime.md |
| T02 | DONE | HIHU-COCO / 集成 | PR #1已真实触发CI；7046b91的双平台Go/Rust、前端与Ubuntu deb smoke全部PASS，Actions run 39198868346；后续提交独立复验 |
| T03 | BLOCKED | HIHU-COCO / 集成 | 已有Windows、WSL22.04、授权LAN Ubuntu22.04 Desktop；后者SSH/PTY/SFTP通过；缺完整版本/角色真机矩阵、跨网和隔离部署 |
| P01 | BLOCKED | HIHU-COCO / 集成 | 只读前置检查和4项单测PASS；缺隔离 Headscale/DERP 与不同网络端点；ADR-network.md |
| P02 | BLOCKED | ssh_finish / poc/ssh | Windows→隔离Ubuntu及独立LAN Ubuntu22.04命令/PTY/10MiB SFTP PASS；已修单pin算法协商；缺Windows sshd与跨平台长任务实测；tests/pocs/ssh/report.md |
| P03 | IN_PROGRESS | mcp_finish / poc/mcp | Codex0.153.4真实echo/预期错误/status与2025-06-18协商PASS；SDK取消/race PASS；Codex协议取消尚未证实；test-results/P03-mcp-2026-10-19.md |
| P04 | BLOCKED | desktop_finish、mcp_finish / 集成 | Windows NSIS安装/真实GUI IPC/卸载PASS；独立Ubuntu22.04 X11工程deb安装/系统desktop入口启动/显示刷新/内核PID核验UDS/退出后Core/卸载复查PASS，暂存与user systemd亦PASS；异常unit启动清理回归3PASS；缺SCM、24.04及完整版本/身份矩阵；见P04报告 |
| P05 | BLOCKED | HIHU-COCO / 集成 | 已形成 ADR-0001-v1-stack.md 联合评审；真实Codex调用已有证据，跨网络/完整平台/服务身份风险未关闭，G1未通过 |
| K01 | TODO | HIHU-COCO / 集成 | 独立DRAFT矩阵及contracts-draft三schema/35正反例/26工具校验PASS；Directory/完整事件/类型生成待补；未冻结，正式实施依赖P05 |
| K02 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| K03 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
| K04 | TODO | HIHU-COCO / 集成 | 依赖按 Plan 校验，尚未验收 |
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

G0 未通过：已有远端 CI 成功记录，T03 测试矩阵仍不齐。G1/G2 未通过：POC 阻断风险未关闭，未冻结正式契约，六条业务线尚未启动。

每项测试的命令、结果和范围写入 test-results/。Git commit 见 Git 历史；无 PR 时不得填写虚构链接。
