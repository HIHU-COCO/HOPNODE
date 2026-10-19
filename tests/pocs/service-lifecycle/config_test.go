// Purpose: reject ambiguous or non-isolated SCM inputs. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func fixtureConfigArgs(t *testing.T) []string {
	t.Helper()
	core := filepath.Join(t.TempDir(), "core fixture.exe")
	if err := os.WriteFile(core, nil, 0600); err != nil {
		t.Fatal(err)
	}
	return []string{"--service-name", "HopNode-P04-POC-test", "--core", core, "--endpoint", `\\.\pipe\hopnode-p04-service-test`}
}

func TestIsolatedConfiguration(t *testing.T) {
	args := fixtureConfigArgs(t)
	if _, err := parseConfig(args); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		args []string
	}{
		{"missing", nil},
		{"production-name", append(append([]string{}, args...), "--service-name", "HopNode")},
		{"default-pipe", append(append([]string{}, args...), "--endpoint", `\\.\pipe\hopnode-poc-existing`)},
		{"remote-pipe", append(append([]string{}, args...), "--endpoint", `\\other\pipe\hopnode-p04-service-test`)},
		{"relative-core", append(append([]string{}, args...), "--core", "core.exe")},
		{"missing-core", append(append([]string{}, args...), "--core", filepath.Join(t.TempDir(), "missing.exe"))},
		{"short-stop", append(append([]string{}, args...), "--stop-timeout", "0s")},
		{"long-stop", append(append([]string{}, args...), "--stop-timeout", "2m")},
		{"extra-argument", append(append([]string{}, args...), "install")},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := parseConfig(test.args); err == nil {
				t.Fatal("unsafe or ambiguous configuration accepted")
			}
		})
	}
}
