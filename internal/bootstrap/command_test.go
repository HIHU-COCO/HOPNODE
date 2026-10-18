// 文件用途：验证 CLI 工程入口的版本输出与未实现命令失败语义。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package bootstrap

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestVersionIsMachineReadable(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run("hopnodectl", []string{"version"}, &out, &errOut); code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut.String())
	}
	var info map[string]any
	if err := json.Unmarshal(out.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info["name"] != "hopnodectl" || info["app_version"] == "" {
		t.Fatalf("invalid version: %v", info)
	}
	if errOut.Len() != 0 {
		t.Fatalf("unexpected stderr: %s", errOut.String())
	}
}

func TestUnsupportedCommandDoesNotReportSuccess(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := Run("hopnodectl", []string{"exec", "--device", "fixture"}, &out, &errOut); code == 0 {
		t.Fatal("unsupported operation succeeded")
	}
	if out.Len() != 0 {
		t.Fatal("failure polluted stdout")
	}
	var failure map[string]any
	if err := json.Unmarshal(errOut.Bytes(), &failure); err != nil {
		t.Fatal(err)
	}
	if failure["code"] != "FEATURE_NOT_IMPLEMENTED" {
		t.Fatalf("wrong error: %v", failure)
	}
}
