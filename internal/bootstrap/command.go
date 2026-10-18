// 文件用途：T01 工程启动命令；未交付的业务入口返回明确失败。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package bootstrap

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/HIHU-COCO/HOPNODE/internal/buildinfo"
)

func Run(name string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "version") {
		if err := json.NewEncoder(stdout).Encode(buildinfo.Get(name)); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}
	if len(args) == 0 || (len(args) == 1 && (args[0] == "--help" || args[0] == "help")) {
		fmt.Fprintf(stdout, "%s — 跃点 HopNode engineering bootstrap\nUsage: %s version | --version | --help\n", name, name)
		fmt.Fprintln(stdout, "Development baseline only. Remote operations and services are not installed by this command.")
		return 0
	}
	_ = json.NewEncoder(stderr).Encode(map[string]any{
		"code": "FEATURE_NOT_IMPLEMENTED", "message": "工程初始化阶段尚未交付此命令；查看 docs/implementation/status.md", "retryable": false,
	})
	return 2
}
