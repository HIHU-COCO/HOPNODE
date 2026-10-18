#!/usr/bin/env bash
# 文件用途：可重复构建 Linux 无图形依赖的 Go 基础程序。
# 创建者：HIHU-COCO；创建日期：2026-10-19
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p artifacts/linux
for name in hopnode-core hopnodectl hopnode-directory; do
  CGO_ENABLED=0 go build -trimpath -o "artifacts/linux/$name" "./cmd/$name"
done
