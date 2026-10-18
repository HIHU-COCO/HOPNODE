# 文件用途：CI 防止遗漏 PRD 功能、验收编号或将无证据结果标为通过。
# 创建者：HIHU-COCO；创建日期：2026-10-19
from pathlib import Path
import re

root = Path(__file__).resolve().parents[1]
ledger = (root / "docs/implementation/status.md").read_text(encoding="utf-8")
for prefix, count in (("F", 18), ("A", 27)):
    for number in range(1, count + 1):
        identifier = f"{prefix}{number:02}"
        if not re.search(rf"^\| {identifier} \|", ledger, re.M):
            raise SystemExit(f"Missing ledger row: {identifier}")
for row in ledger.splitlines():
    if re.match(r"\| A\d{2} \| PASS \|", row):
        if not re.search(r"test-results/.+\.md", row) or not re.search(r"\d{4}-\d{2}-\d{2}", row):
            raise SystemExit(f"PASS requires dated evidence: {row}")
print("PASS: all 18 feature and 27 acceptance rows present; acceptance PASS rows require evidence")
