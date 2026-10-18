// 文件用途：设备目录服务的 T01 版本入口。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"github.com/HIHU-COCO/HOPNODE/internal/bootstrap"
	"os"
)

func main() { os.Exit(bootstrap.Run("hopnode-directory", os.Args[1:], os.Stdout, os.Stderr)) }
