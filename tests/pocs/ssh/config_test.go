// 文件用途：验证基础配置优先级、限制提示，并以独立 -F 文件对比系统 ssh -G。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigBasicsAndOpenSSHOracle(t *testing.T) {
	input := "Host test-* !test-denied\n HostName 127.0.0.1\n User poc-user\n Port 2222\n IdentityFile \"C:/temporary fixture/key\"\nHost *\n User fallback\n Port 22\n"
	p, err := parseConfig(strings.NewReader(input), "test-good")
	if err != nil {
		t.Fatal(err)
	}
	if p.Host != "127.0.0.1" || p.User != "poc-user" || p.Port != 2222 || p.IdentityFile != "C:/temporary fixture/key" {
		t.Fatalf("unexpected profile %+v", p)
	}
	denied, err := parseConfig(strings.NewReader(input), "test-denied")
	if err != nil {
		t.Fatal(err)
	}
	if denied.Host != "test-denied" || denied.User != "fallback" || denied.Port != 22 {
		t.Fatalf("negative Host pattern ignored: %+v", denied)
	}
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("NOT_RUN: system OpenSSH client not available")
	}
	file := filepath.Join(t.TempDir(), "poc-config")
	if err := os.WriteFile(file, []byte(input), 0600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(sshPath, "-G", "-F", file, "test-good").Output()
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"hostname 127.0.0.1", "user poc-user", "port 2222", "identityfile C:/temporary fixture/key"} {
		if !strings.Contains(string(out), expected) {
			t.Fatalf("ssh -G disagrees with basic parser on %q", expected)
		}
	}
}

func TestConfigRejectsUnsupportedInsteadOfIgnoring(t *testing.T) {
	for _, input := range []string{"Include child", "Match all", "ProxyJump bastion", "ProxyCommand anything", "IdentityAgent socket", "IdentityFile a\nIdentityFile b", "IdentityFile ~/.ssh/key", "Port 70000", "HostName \"unterminated", "ForwardAgent yes"} {
		t.Run(input, func(t *testing.T) {
			if _, err := parseConfig(strings.NewReader(input), "alias"); err == nil {
				t.Fatal("unsupported/invalid config accepted")
			}
		})
	}
}
