#!/usr/bin/env bash
# 文件用途：分别验证独立 POC module，可选 race 检测。
# 创建者：HIHU-COCO；创建日期：2026-10-19
set -euo pipefail
cd "$(dirname "$0")/.."
for module in tests/pocs/*/go.mod; do
  test -f "$module" || continue
  printf 'Checking POC: %s\n' "$(dirname "$module")"
  (cd "$(dirname "$module")" && go test "${@}" ./... && go vet ./...)
done
python3 -m unittest discover -s tests/pocs/network -p 'test_*.py' -v
python3 tests/pocs/desktop-service/test-user-unit.py
