//go:build windows

// Purpose: opt-in disposable-runner SCM lifecycle and unchanged pipe ACL check. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type scmRunSettings struct{ host, core string }

func scmCallerAllowed(elevated, localSystem bool) error {
	if !elevated {
		return errors.New("enabled SCM test requires an already elevated administrator runner; no UAC/self-elevation is attempted")
	}
	if localSystem {
		return errors.New("SCM ACL negative requires an administrator caller distinct from LocalSystem")
	}
	return nil
}

func scmSettings(gate, host, core string) (scmRunSettings, string, error) {
	if gate == "" || gate == "0" {
		return scmRunSettings{}, "NOT_RUN: set HOPNODE_POC_RUN_SCM=1 only on an authorized disposable administrator runner", nil
	}
	if gate != "1" {
		return scmRunSettings{}, "", errors.New("HOPNODE_POC_RUN_SCM must be empty, 0, or 1")
	}
	for _, candidate := range []string{host, core} {
		if candidate == "" || !filepath.IsAbs(candidate) {
			return scmRunSettings{}, "NOT_RUN: absolute HOPNODE_POC_SCM_HOST_EXE and HOPNODE_POC_CORE_EXE are required", nil
		}
		info, err := os.Stat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			return scmRunSettings{}, "NOT_RUN: an explicitly configured SCM fixture binary is missing; build both fixtures first", nil
		}
		if err != nil || !info.Mode().IsRegular() || !strings.EqualFold(filepath.Ext(candidate), ".exe") {
			return scmRunSettings{}, "", errors.New("SCM fixtures must be readable regular .exe files")
		}
	}
	if strings.EqualFold(filepath.Clean(host), filepath.Clean(core)) {
		return scmRunSettings{}, "", errors.New("SCM host and Core must be distinct executables")
	}
	return scmRunSettings{filepath.Clean(host), filepath.Clean(core)}, "", nil
}

type scmProcess struct {
	pid    uint32
	handle windows.Handle
	role   string
}

type scmFixture struct {
	manager   *mgr.Mgr
	service   *mgr.Service // Only the handle returned by this test's successful CreateService.
	name      string
	settings  scmRunSettings
	processes []scmProcess // Stable native handles prevent PID-reuse cleanup mistakes.
}

func processImage(handle windows.Handle) (string, error) {
	buffer := make([]uint16, 32768)
	size := uint32(len(buffer))
	if err := windows.QueryFullProcessImageName(handle, 0, &buffer[0], &size); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buffer[:size]), nil
}

func processIsSystem(handle windows.Handle) (bool, error) {
	var token windows.Token
	if err := windows.OpenProcessToken(handle, windows.TOKEN_QUERY, &token); err != nil {
		return false, err
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		return false, err
	}
	return user.User.Sid.IsWellKnown(windows.WinLocalSystemSid), nil
}

func (fixture *scmFixture) remember(pid uint32, role, expectedImage string) (*scmProcess, error) {
	if pid == 0 || pid == uint32(os.Getpid()) {
		return nil, errors.New("SCM returned an invalid independent process PID")
	}
	for index := range fixture.processes {
		if fixture.processes[index].pid == pid {
			isSystem, err := processIsSystem(fixture.processes[index].handle)
			if err != nil || !isSystem {
				return nil, fmt.Errorf("owned %s process did not verify as LocalSystem", role)
			}
			return &fixture.processes[index], nil
		}
	}
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION|windows.PROCESS_TERMINATE|windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return nil, fmt.Errorf("open owned %s process: %w", role, err)
	}
	image, err := processImage(handle)
	if err != nil || !strings.EqualFold(filepath.Clean(image), expectedImage) {
		windows.CloseHandle(handle)
		return nil, fmt.Errorf("owned %s process image did not match the built fixture", role)
	}
	// Record ownership before further checks, so even an identity-query error is cleaned up.
	fixture.processes = append(fixture.processes, scmProcess{pid, handle, role})
	index := len(fixture.processes) - 1
	isSystem, err := processIsSystem(handle)
	if err != nil || !isSystem {
		return nil, fmt.Errorf("owned %s process did not verify as LocalSystem", role)
	}
	return &fixture.processes[index], nil
}

func directCorePIDs(parent uint32, core string) ([]uint32, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(snapshot)
	entry := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	err = windows.Process32First(snapshot, &entry)
	var matches []uint32
	for err == nil {
		if entry.ParentProcessID == parent && strings.EqualFold(windows.UTF16ToString(entry.ExeFile[:]), filepath.Base(core)) {
			matches = append(matches, entry.ProcessID)
		}
		err = windows.Process32Next(snapshot, &entry)
	}
	if !errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		return nil, err
	}
	return matches, nil
}

func alive(process *scmProcess) bool {
	status, err := windows.WaitForSingleObject(process.handle, 0)
	return err == nil && status == uint32(windows.WAIT_TIMEOUT)
}

func (fixture *scmFixture) capture(status svc.Status) (*scmProcess, *scmProcess, error) {
	host, err := fixture.remember(status.ProcessId, "host", fixture.settings.host)
	if err != nil || !alive(host) {
		return nil, nil, errors.New("SCM host PID/image/identity was not verified alive")
	}
	// Inspect children only while the verified parent handle is alive. Never use an old, reused parent PID.
	pids, err := directCorePIDs(host.pid, fixture.settings.core)
	if err != nil || len(pids) != 1 {
		return host, nil, errors.New("expected exactly one direct Core child of the verified SCM host")
	}
	core, err := fixture.remember(pids[0], "core", fixture.settings.core)
	if err != nil || !alive(core) {
		return host, nil, errors.New("SCM direct Core PID/image/identity was not verified alive")
	}
	return host, core, nil
}

func pipeOpen(endpoint string) error {
	path, err := windows.UTF16PtrFromString(endpoint)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(path, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_OVERLAPPED, 0)
	if err == nil {
		windows.CloseHandle(handle)
	}
	return err
}

func (fixture *scmFixture) waitState(want svc.State, timeout time.Duration) (svc.Status, error) {
	deadline := time.Now().Add(timeout)
	var status svc.Status
	for time.Now().Before(deadline) {
		var err error
		status, err = fixture.service.Query()
		if err != nil {
			return status, err
		}
		if status.State == want {
			return status, nil
		}
		if want == svc.Running && status.State == svc.Stopped {
			return status, fmt.Errorf("SCM stopped before Running (win32=%d service=%d)", status.Win32ExitCode, status.ServiceSpecificExitCode)
		}
		time.Sleep(50 * time.Millisecond)
	}
	return status, fmt.Errorf("SCM did not reach state %d before timeout", want)
}

func scmStoppedSuccess(status svc.Status) error {
	if status.State != svc.Stopped || status.ProcessId != 0 || status.Win32ExitCode != 0 || status.ServiceSpecificExitCode != 0 {
		return fmt.Errorf("SCM stop was not successful (state=%d pid=%d win32=%d service=%d)", status.State, status.ProcessId, status.Win32ExitCode, status.ServiceSpecificExitCode)
	}
	return nil
}

func (fixture *scmFixture) stop() error {
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		status, err := fixture.service.Query()
		if err != nil {
			return err
		}
		if status.State == svc.Stopped {
			return scmStoppedSuccess(status)
		}
		if status.ProcessId != 0 {
			// Best-effort ownership capture even on a failed/partially started test.
			_, _, _ = fixture.capture(status)
		}
		_, err = fixture.service.Control(svc.Stop)
		if err != nil && !errors.Is(err, windows.ERROR_SERVICE_NOT_ACTIVE) && !errors.Is(err, windows.ERROR_SERVICE_CANNOT_ACCEPT_CTRL) {
			return err
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("owned SCM service did not stop within the cleanup bound")
}

func (fixture *scmFixture) processesExited(timeout time.Duration) error {
	for _, process := range fixture.processes {
		result, err := windows.WaitForSingleObject(process.handle, uint32(timeout.Milliseconds()))
		if err != nil || result != windows.WAIT_OBJECT_0 {
			return fmt.Errorf("owned %s process did not exit", process.role)
		}
	}
	return nil
}

func (fixture *scmFixture) cleanup() error {
	var result error
	if fixture.service != nil {
		result = errors.Join(result, fixture.stop())
		if err := fixture.processesExited(3 * time.Second); err != nil {
			result = errors.Join(result, err)
			// Fallback is limited to image-verified native handles created/captured for this service.
			// Killing its host also closes the host's KILL_ON_JOB_CLOSE Job handle.
			for _, process := range fixture.processes {
				if alive(&process) {
					result = errors.Join(result, windows.TerminateProcess(process.handle, 1))
				}
			}
			result = errors.Join(result, fixture.processesExited(8*time.Second))
		}
		if err := fixture.service.Delete(); err != nil && !errors.Is(err, windows.ERROR_SERVICE_MARKED_FOR_DELETE) {
			result = errors.Join(result, fmt.Errorf("delete owned SCM service: %w", err))
		}
		result = errors.Join(result, fixture.service.Close())
		fixture.service = nil
		deadline := time.Now().Add(15 * time.Second)
		absent := false
		for time.Now().Before(deadline) {
			opened, err := fixture.manager.OpenService(fixture.name)
			if errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
				absent = true
				break
			}
			if err == nil {
				opened.Close()
			} else if !errors.Is(err, windows.ERROR_SERVICE_MARKED_FOR_DELETE) {
				result = errors.Join(result, err)
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if !absent {
			result = errors.Join(result, errors.New("owned SCM service still exists after Delete/Close"))
		}
	}
	for _, process := range fixture.processes {
		result = errors.Join(result, windows.CloseHandle(process.handle))
	}
	fixture.processes = nil
	return result
}

func TestSCMNativeLifecycle(t *testing.T) {
	settings, skip, err := scmSettings(os.Getenv("HOPNODE_POC_RUN_SCM"), os.Getenv("HOPNODE_POC_SCM_HOST_EXE"), os.Getenv("HOPNODE_POC_CORE_EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if skip != "" {
		t.Skip(skip)
	}
	token, err := windows.OpenCurrentProcessToken()
	if err != nil {
		t.Fatal("enabled SCM test could not read its token")
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil {
		t.Fatal("enabled SCM test could not read caller identity")
	}
	if err := scmCallerAllowed(token.IsElevated(), user.User.Sid.IsWellKnown(windows.WinLocalSystemSid)); err != nil {
		t.Fatal(err)
	}
	manager, err := mgr.Connect()
	if err != nil {
		t.Fatalf("enabled SCM test could not open local SCM: %v", err)
	}
	// Register first: cleanup callbacks are LIFO, so the SCM stays open until the owned service is gone.
	t.Cleanup(func() {
		if err := manager.Disconnect(); err != nil {
			t.Errorf("close local SCM handle: %v", err)
		}
	})
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		t.Fatal("generate unique SCM fixture names")
	}
	suffix := hex.EncodeToString(nonce[:])
	name := "HopNode-P04-POC-" + suffix
	endpoint := `\\.\pipe\hopnode-p04-service-` + suffix
	if existing, err := manager.OpenService(name); err == nil {
		existing.Close()
		t.Fatal("random SCM name collision; no existing service was changed")
	} else if !errors.Is(err, windows.ERROR_SERVICE_DOES_NOT_EXIST) {
		t.Fatalf("SCM name preflight failed: %v", err)
	}
	if err := pipeOpen(endpoint); !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		t.Fatal("random pipe preflight was not absent; no existing endpoint was changed")
	}
	service, err := manager.CreateService(name, settings.host, mgr.Config{
		DisplayName: name, ServiceType: windows.SERVICE_WIN32_OWN_PROCESS,
		StartType: mgr.StartManual, ErrorControl: mgr.ErrorNormal, ServiceStartName: "LocalSystem",
	}, "--service-name", name, "--core", settings.core, "--endpoint", endpoint, "--stop-timeout", "8s")
	if err != nil {
		t.Fatalf("create unique manual LocalSystem fixture: %v", err)
	}
	fixture := &scmFixture{manager: manager, service: service, name: name, settings: settings}
	t.Cleanup(func() {
		if err := fixture.cleanup(); err != nil {
			t.Errorf("SCM_OWNED_CLEANUP=FAIL: %v", err)
		} else {
			t.Log("SCM_OWNED_CLEANUP=PASS; recorded host/Core handles exited; service absent; handles closed")
		}
	})
	actualConfig, err := service.Config()
	if err != nil || actualConfig.StartType != mgr.StartManual || actualConfig.ServiceType != windows.SERVICE_WIN32_OWN_PROCESS || !strings.EqualFold(actualConfig.ServiceStartName, "LocalSystem") || len(actualConfig.Dependencies) != 0 {
		t.Fatal("actual SCM configuration did not match isolated manual LocalSystem fixture")
	}
	for _, binary := range []struct{ name, path string }{{"host", settings.host}, {"core", settings.core}} {
		contents, err := os.ReadFile(binary.path)
		if err != nil {
			t.Fatal("read fixture hash")
		}
		t.Logf("SCM_FIXTURE_%s_SHA256=%x", strings.ToUpper(binary.name), sha256.Sum256(contents))
	}
	var previousHost, previousCore uint32
	for round := 1; round <= 2; round++ {
		if err := service.Start(); err != nil {
			t.Fatalf("SCM Start round %d: %v", round, err)
		}
		status, err := fixture.waitState(svc.Running, 20*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		var host, core *scmProcess
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			host, core, err = fixture.capture(status)
			if err == nil {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		if err != nil || host == nil || core == nil {
			t.Fatal("SCM Running did not produce image/token-verified host and direct Core")
		}
		if host.pid == previousHost || core.pid == previousCore {
			t.Fatal("SCM restart did not create new host/Core PIDs")
		}
		deadline = time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			err = pipeOpen(endpoint)
			if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
				break
			}
			if err == nil {
				t.Fatal("cross-identity administrator unexpectedly accessed the service-identity-only pipe")
			}
			if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || !alive(host) || !alive(core) {
				t.Fatal("service pipe readiness/ACL produced an unexpected result")
			}
			time.Sleep(50 * time.Millisecond)
		}
		if !errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			t.Fatal("LocalSystem-only pipe was not observed denying the distinct caller")
		}
		if err := fixture.stop(); err != nil {
			t.Fatal(err)
		}
		if err := fixture.processesExited(8 * time.Second); err != nil {
			t.Fatal(err)
		}
		if err := pipeOpen(endpoint); !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
			t.Fatal("stopped SCM fixture still exposed its pipe")
		}
		t.Logf("SCM_ROUND_%d=PASS host_pid=%d core_pid=%d; LocalSystem verified; Running->Stopped; unchanged ACL denied distinct caller; both handles signaled", round, host.pid, core.pid)
		previousHost, previousCore = host.pid, core.pid
	}
	t.Log("SCM_LIFECYCLE=PASS Start/Running/Stop/Stopped/restart; positive cross-identity API authorization NOT_RUN; Delete/absence checked by cleanup")
}

func TestSCMSettingsGateAndPaths(t *testing.T) {
	host := filepath.Join(t.TempDir(), "host.exe")
	core := filepath.Join(t.TempDir(), "core.exe")
	for _, path := range []string{host, core} {
		if err := os.WriteFile(path, []byte("configuration fixture; never executed"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, test := range []struct {
		name, gate, host, core string
		wantSkip, wantError    bool
	}{
		{"default", "", host, core, true, false},
		{"disabled", "0", host, core, true, false},
		{"invalid-gate", "true", host, core, false, true},
		{"missing-host", "1", "", core, true, false},
		{"relative-core", "1", host, "core.exe", true, false},
		{"missing-file", "1", host, filepath.Join(t.TempDir(), "missing.exe"), true, false},
		{"same-file", "1", host, host, false, true},
		{"directory", "1", t.TempDir(), core, false, true},
		{"valid", "1", host, core, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, skip, err := scmSettings(test.gate, test.host, test.core)
			if (skip != "") != test.wantSkip || (err != nil) != test.wantError {
				t.Fatalf("gate/path decision skip=%t error=%t", skip != "", err != nil)
			}
		})
	}
}

func TestSCMCallerPermissionIsFailure(t *testing.T) {
	for _, test := range []struct {
		name                   string
		elevated, system, fail bool
	}{
		{"ordinary-user", false, false, true},
		{"distinct-admin", true, false, false},
		{"same-system-identity", true, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := scmCallerAllowed(test.elevated, test.system); (err != nil) != test.fail {
				t.Fatal("enabled SCM caller permission must return an error, never a Skip decision")
			}
		})
	}
}

func TestSCMStoppedRequiresSuccessfulExit(t *testing.T) {
	for _, test := range []struct {
		name   string
		status svc.Status
		fail   bool
	}{
		{"clean-stop", svc.Status{State: svc.Stopped}, false},
		{"stopped-with-pid", svc.Status{State: svc.Stopped, ProcessId: 123}, true},
		{"win32-failure", svc.Status{State: svc.Stopped, Win32ExitCode: 1066}, true},
		{"specific-failure", svc.Status{State: svc.Stopped, ServiceSpecificExitCode: 3}, true},
		{"not-stopped", svc.Status{State: svc.Running}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := scmStoppedSuccess(test.status); (err != nil) != test.fail {
				t.Fatal("Stopped state must require zero PID and zero Win32/service exit codes")
			}
		})
	}
}
