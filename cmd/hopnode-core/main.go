// 文件用途：HopNode Core 的 T01 版本入口；业务子命令由后续 POC 和契约决定。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"github.com/HIHU-COCO/HOPNODE/internal/bootstrap"
	"os"
)

func main() { os.Exit(bootstrap.Run("hopnode-core", os.Args[1:], os.Stdout, os.Stderr)) }
