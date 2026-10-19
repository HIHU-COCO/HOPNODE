//go:build windows

// Purpose: SCM-only service host, without service installation. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/svc"
)

func main() {
	cfg, err := parseConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "SERVICE_POC_INVALID_ARGUMENTS")
		os.Exit(2)
	}
	underSCM, err := svc.IsWindowsService()
	if err != nil || !underSCM {
		fmt.Fprintln(os.Stderr, "SERVICE_POC_REQUIRES_SCM")
		os.Exit(2)
	}
	if err := svc.Run(cfg.serviceName, &serviceHandler{cfg: cfg, start: startOwnedChild}); err != nil {
		fmt.Fprintln(os.Stderr, "SERVICE_POC_SCM_FAILURE")
		os.Exit(1)
	}
}
