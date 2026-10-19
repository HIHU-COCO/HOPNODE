//go:build windows

// Purpose: explicitly opted-in native Job/Named Pipe lifecycle check, without SCM. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

type nativeStatus struct {
	APIVersion string `json:"api_version"`
	PID        uint32 `json:"pid"`
	InstanceID string `json:"instance_id"`
	OS         string `json:"os"`
}

func nativeStatusCommand(ctx context.Context, cfg config) *exec.Cmd {
	command := exec.CommandContext(ctx, cfg.core, "status", "--json")
	command.Env = []string{"HOPNODE_POC_ENDPOINT=" + cfg.endpoint, "SystemRoot=" + os.Getenv("SystemRoot")}
	command.Stderr = io.Discard
	// Stderr is discarded. Output is parsed into an allowlist struct; raw identity
	// fields from the fixture are never logged, including on assertion failures.
	return command
}

func readNativeStatus(ctx context.Context, cfg config) (nativeStatus, error) {
	output, err := nativeStatusCommand(ctx, cfg).Output()
	if err != nil {
		return nativeStatus{}, err
	}
	var status nativeStatus
	if err := json.Unmarshal(output, &status); err != nil {
		return nativeStatus{}, errors.New("invalid fixture JSON")
	}
	return status, nil
}

func nativeJobCount(child *windowsChild) (uint32, error) {
	var accounting jobAccounting
	err := windows.QueryInformationJobObject(child.job, windows.JobObjectBasicAccountingInformation, uintptr(unsafe.Pointer(&accounting)), uint32(unsafe.Sizeof(accounting)), nil)
	return accounting.activeProcesses, err
}

func nativeJobContainsPID(child *windowsChild, expected uint32) (bool, error) {
	var members struct {
		assigned uint32
		listed   uint32
		pids     [64]uintptr
	}
	if err := windows.QueryInformationJobObject(child.job, windows.JobObjectBasicProcessIdList, uintptr(unsafe.Pointer(&members)), uint32(unsafe.Sizeof(members)), nil); err != nil {
		return false, err
	}
	if members.listed > uint32(len(members.pids)) {
		return false, errors.New("job member list exceeded fixture bound")
	}
	for _, member := range members.pids[:members.listed] {
		if member == uintptr(expected) {
			return true, nil
		}
	}
	return false, nil
}

func TestOwnedChildNativeLifecycle(t *testing.T) {
	core := os.Getenv("HOPNODE_POC_CORE_EXE")
	if core == "" {
		t.Skip("NOT_RUN: set HOPNODE_POC_CORE_EXE to an isolated desktop-service POC executable; no SCM installation is performed")
	}
	var nonce [12]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal("generate isolated native fixture name")
	}
	suffix := hex.EncodeToString(nonce[:])
	cfg, err := parseConfig([]string{
		"--service-name", "HopNode-P04-POC-" + suffix,
		"--core", core,
		"--endpoint", `\\.\pipe\hopnode-p04-service-` + suffix,
		"--stop-timeout", "5s",
	})
	if err != nil {
		t.Fatalf("invalid explicitly opted-in native fixture: %v", err)
	}
	owned, err := startOwnedChild(cfg)
	if err != nil {
		t.Fatalf("native suspended-child/job startup failed: %v", err)
	}
	child := owned.(*windowsChild)
	defer func() {
		if child.job != 0 {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := child.Stop(ctx); err != nil {
				t.Errorf("cleanup of owned native job failed: %v", err)
			}
		}
		if err := child.Close(); err != nil {
			t.Errorf("close owned native handles: %v", err)
		}
	}()
	expectedPID, err := windows.GetProcessId(child.process)
	if err != nil || expectedPID == uint32(os.Getpid()) || expectedPID == 0 {
		t.Fatal("native process handle did not identify an independent child")
	}
	readyCtx, cancelReady := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancelReady()
	var first nativeStatus
	for {
		queryCtx, cancel := context.WithTimeout(readyCtx, time.Second)
		first, err = readNativeStatus(queryCtx, cfg)
		cancel()
		if err == nil {
			break
		}
		if exited, waitErr := child.Exited(); waitErr != nil || exited {
			t.Fatal("owned native child exited before IPC readiness")
		}
		select {
		case <-readyCtx.Done():
			t.Fatal("native Named Pipe readiness timed out")
		case <-time.After(25 * time.Millisecond):
		}
	}
	if first.PID != expectedPID || first.InstanceID == "" || first.APIVersion != "poc.v0" || first.OS != "windows" {
		t.Fatal("actual CLI did not identify the owned Windows fixture")
	}
	second, err := readNativeStatus(readyCtx, cfg)
	if err != nil || second != first {
		t.Fatal("repeated actual CLI request changed the target PID or instance")
	}
	activeBefore, err := nativeJobCount(child)
	if err != nil || activeBefore < 1 {
		t.Fatalf("native job was not active before stop: active=%d error=%v", activeBefore, err)
	}
	// Windows may also attach a console helper to this job. Require membership of
	// our exact Core PID instead of incorrectly assuming there is only one member.
	if contains, err := nativeJobContainsPID(child, expectedPID); err != nil || !contains {
		t.Fatal("owned Core PID was not a member of the native job")
	}
	stopCtx, cancelStop := context.WithTimeout(context.Background(), cfg.stopTimeout)
	defer cancelStop()
	if err := child.Stop(stopCtx); err != nil {
		t.Fatalf("native owned-job stop failed: %v", err)
	}
	if exited, err := child.Exited(); err != nil || !exited {
		t.Fatal("native child was not signaled after stop")
	}
	if active, err := nativeJobCount(child); err != nil || active != 0 {
		t.Fatal("native job remained active after stop")
	}
	rejectedCtx, cancelRejected := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelRejected()
	err = nativeStatusCommand(rejectedCtx, cfg).Run()
	var exitError *exec.ExitError
	if rejectedCtx.Err() != nil || !errors.As(err, &exitError) || exitError.ExitCode() == 0 {
		t.Fatal("stopped fixture must reject actual CLI status with a nonzero exit, not a test timeout")
	}
	if err := child.Close(); err != nil || child.job != 0 || child.process != 0 {
		t.Fatal("native process/job handles were not closed")
	}
	t.Logf("NATIVE_JOB_LIFECYCLE=PASS child_pid=%d; stable IPC instance; activeProcesses=%d->0; stopped CLI rejected; handles closed; SCM NOT_RUN", expectedPID, activeBefore)
}
