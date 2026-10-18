# 文件用途：从 Plan 建立覆盖任务、18 项功能、27 项验收的初始台账。
# 创建者：HIHU-COCO；创建日期：2026-10-19
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
plan = (ROOT / "Plan.md").read_text(encoding="utf-8")
ids = sorted(set(re.findall(r"\b(?:T\d{2}|P0[1-5]|K0[1-4]|N0[1-5]|R0[1-6]|M0[1-4]|U0[1-6]|S0[1-6]|C0[1-5]|I\d{2}|L0[1-7])\b", plan)), key=lambda x: ("TPKNRMUSCIL".index(x[0]), x))
target = ROOT / "docs/implementation/status.md"
if target.exists():
    raise SystemExit("Refusing to overwrite an existing execution ledger")
target.parent.mkdir(parents=True, exist_ok=True)
lines = ["# HopNode 执行台账", "", "更新日期：2026-10-19（Asia/Shanghai）。依据 PRD V1.4、Design V1.3、Plan V1.2、AGENTS V1.0。", "", "当前为工程初始化与风险 POC，尚未达到正式产品发布条件。POC 本机测试不能替代 A01–A27 真机验收。", "", "## 任务", "", "| ID | 状态 | 负责人 | 证据 / 阻塞 |", "| --- | --- | --- | --- |"]
for task in ids:
    state = "DONE" if task == "T00" else "IN_PROGRESS" if task == "T01" else "TODO"
    evidence = "test-environment.md、risks.md、完整 F/A 台账" if task == "T00" else "依赖按 Plan 校验，尚未验收"
    lines.append(f"| {task} | {state} | HIHU-COCO / 集成 | {evidence} |")
lines += ["", "## PRD 功能追踪", "", "| 编号 | 状态 | 说明 |", "| --- | --- | --- |"]
lines += [f"| F{i:02} | TODO | 以 Plan §8.1 的对应任务和真实结果验收 |" for i in range(1, 19)]
lines += ["", "## 正式验收（0/27 PASS）", "", "| 编号 | 结果 | 环境 / 步骤 / 证据 / 验收日期 |", "| --- | --- | --- |"]
lines += [f"| A{i:02} | NOT_RUN | 尚未执行；真机条件见 test-environment.md；无验收日期 |" for i in range(1, 28)]
lines += ["", "## 检查点", "", "G0 未通过：CI 尚未在远端运行，T03 测试矩阵不齐。G1/G2 未通过：未冻结正式契约，六条业务线尚未启动。", "", "每项测试的命令、结果和范围写入 test-results/。Git commit 见 Git 历史；无 PR 时不得填写虚构链接。", ""]
target.write_text("\n".join(lines), encoding="utf-8")
print(f"Created ledger: {len(ids)} tasks, 18 features, 27 acceptances")
