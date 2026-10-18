// 文件用途：所有 Go 程序共用可追溯的开发构建标识。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package buildinfo

import "runtime"

var (
	Version = "0.1.0-dev"
	Commit  = "uncommitted"
	BuiltAt = "unknown"
)

type Info struct {
	Name      string `json:"name"`
	Version   string `json:"app_version"`
	Commit    string `json:"commit"`
	BuiltAt   string `json:"built_at"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	GoVersion string `json:"go_version"`
}

func Get(name string) Info {
	return Info{name, Version, Commit, BuiltAt, runtime.GOOS, runtime.GOARCH, runtime.Version()}
}
