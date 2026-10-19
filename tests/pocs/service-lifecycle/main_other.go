//go:build !windows

// Purpose: avoid implying SCM support on other systems. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "Windows SCM POC only; use test-systemd.sh on Linux")
	os.Exit(2)
}
