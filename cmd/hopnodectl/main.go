// 文件用途：Windows/Ubuntu 共用 hopnodectl 的 T01 版本入口。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"github.com/HIHU-COCO/HOPNODE/internal/bootstrap"
	"os"
)

func main() { os.Exit(bootstrap.Run("hopnodectl", os.Args[1:], os.Stdout, os.Stderr)) }
