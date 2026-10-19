//go:build windows

// Purpose: create suspended child, bind to own job, terminate and wait. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type windowsChild struct {
	job     windows.Handle
	process windows.Handle
}

// Windows JOBOBJECT_BASIC_ACCOUNTING_INFORMATION layout; x/sys exposes its information class.
type jobAccounting struct {
	totalUserTime, totalKernelTime, periodUserTime, periodKernelTime int64
	pageFaults, totalProcesses, activeProcesses, terminatedProcesses uint32
}

func childEnvironment(endpoint string) ([]uint16, error) {
	var block []uint16
	// Minimal explicit service environment; do not forward the host's tokens or personal PATH.
	for _, entry := range []string{"HOPNODE_POC_ENDPOINT=" + endpoint, "SystemRoot=" + os.Getenv("SystemRoot")} {
		encoded, err := windows.UTF16FromString(entry)
		if err != nil {
			return nil, err
		}
		block = append(block, encoded...)
	}
	return append(block, 0), nil
}

func startOwnedChild(cfg config) (ownedChild, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil, err
	}
	child := &windowsChild{job: job}
	complete := false
	defer func() {
		if !complete {
			_ = child.Close()
		}
	}()
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		return nil, err
	}
	application, err := windows.UTF16PtrFromString(cfg.core)
	if err != nil {
		return nil, err
	}
	command, err := windows.UTF16PtrFromString(windows.ComposeCommandLine([]string{cfg.core, "serve"}))
	if err != nil {
		return nil, err
	}
	cwd, err := windows.UTF16PtrFromString(filepath.Dir(cfg.core))
	if err != nil {
		return nil, err
	}
	environment, err := childEnvironment(cfg.endpoint)
	if err != nil {
		return nil, err
	}
	startup := windows.StartupInfo{}
	startup.Cb = uint32(unsafe.Sizeof(startup))
	var process windows.ProcessInformation
	flags := uint32(windows.CREATE_SUSPENDED | windows.CREATE_UNICODE_ENVIRONMENT | windows.CREATE_NO_WINDOW)
	if err := windows.CreateProcess(application, command, nil, nil, false, flags, &environment[0], cwd, &startup, &process); err != nil {
		return nil, err
	}
	child.process = process.Process
	defer windows.CloseHandle(process.Thread)
	if err := windows.AssignProcessToJobObject(job, process.Process); err != nil {
		// Assignment failed before resume: terminate exactly the child we created.
		_ = windows.TerminateProcess(process.Process, 1)
		_, _ = windows.WaitForSingleObject(process.Process, 5000)
		return nil, err
	}
	if _, err := windows.ResumeThread(process.Thread); err != nil {
		_ = windows.TerminateJobObject(job, 1)
		_, _ = windows.WaitForSingleObject(process.Process, 5000)
		return nil, err
	}
	complete = true
	return child, nil
}

func (child *windowsChild) Exited() (bool, error) {
	result, err := windows.WaitForSingleObject(child.process, 0)
	return result == windows.WAIT_OBJECT_0, err
}

func (child *windowsChild) Stop(ctx context.Context) error {
	if err := windows.TerminateJobObject(child.job, 1); err != nil {
		return err
	}
	for {
		exited, err := child.Exited()
		if err != nil {
			return err
		}
		var accounting jobAccounting
		if err := windows.QueryInformationJobObject(child.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil); err != nil {
			return err
		}
		if exited && accounting.activeProcesses == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}

func (child *windowsChild) Close() error {
	// No other goroutine waits on these handles. Closing this non-inherited job
	// also kills owned descendants if the bounded Stop path failed.
	var result error
	if child.job != 0 {
		result = errors.Join(result, windows.CloseHandle(child.job))
		child.job = 0
	}
	if child.process != 0 {
		result = errors.Join(result, windows.CloseHandle(child.process))
		child.process = 0
	}
	return result
}
