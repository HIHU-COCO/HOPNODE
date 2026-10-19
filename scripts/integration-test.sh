#!/usr/bin/env bash
# 文件用途：真实环境集成测试未具备时返回明确失败，防止冒充验收。
# 创建者：HIHU-COCO；创建日期：2026-10-19
set -euo pipefail
printf '%s\n' 'NOT_RUN: real-machine E2E drivers and authorized platform/network fixtures are not ready; see docs/implementation/risks.md.' >&2
exit 2
