//go:build windows

// Purpose: SCM handler semantics without installing or starting services. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/sys/windows/svc"
)

type fakeChild struct {
	exited      bool
	stopEntered chan struct{}
	allowStop   chan struct{}
	closed      bool
	stopError   error
}

func (child *fakeChild) Exited() (bool, error) { return child.exited, nil }
func (child *fakeChild) Close() error          { child.closed = true; return nil }
func (child *fakeChild) Stop(ctx context.Context) error {
	close(child.stopEntered)
	select {
	case <-child.allowStop:
		return child.stopError
	case <-ctx.Done():
		return ctx.Err()
	}
}

type serviceResult struct {
	specific bool
	code     uint32
}

func TestSCMStopWaitsForOwnedChild(t *testing.T) {
	for _, command := range []svc.Cmd{svc.Stop, svc.Shutdown} {
		t.Run(map[svc.Cmd]string{svc.Stop: "stop", svc.Shutdown: "shutdown"}[command], func(t *testing.T) {
			child := &fakeChild{stopEntered: make(chan struct{}), allowStop: make(chan struct{})}
			handler := &serviceHandler{cfg: config{stopTimeout: time.Second}, start: func(config) (ownedChild, error) { return child, nil }}
			requests := make(chan svc.ChangeRequest, 1)
			changes := make(chan svc.Status, 8)
			result := make(chan serviceResult, 1)
			go func() {
				specific, code := handler.Execute(nil, requests, changes)
				result <- serviceResult{specific, code}
			}()
			if status := <-changes; status.State != svc.StartPending {
				t.Fatal(status)
			}
			if status := <-changes; status.State != svc.Running {
				t.Fatal(status)
			}
			requests <- svc.ChangeRequest{Cmd: svc.Interrogate}
			if status := <-changes; status.State != svc.Running {
				t.Fatal(status)
			}
			requests <- svc.ChangeRequest{Cmd: command}
			if status := <-changes; status.State != svc.StopPending {
				t.Fatal(status)
			}
			<-child.stopEntered
			select {
			case <-result:
				t.Fatal("SCM handler returned before owned child stop completed")
			default:
			}
			close(child.allowStop)
			outcome := <-result
			if outcome.specific || outcome.code != 0 || !child.closed {
				t.Fatalf("unexpected stop: %+v closed=%v", outcome, child.closed)
			}
		})
	}
}

func TestSCMFailureIsNotSuccessfulStop(t *testing.T) {
	for _, scenario := range []string{"start-fails", "child-exits", "stop-fails", "stop-times-out"} {
		t.Run(scenario, func(t *testing.T) {
			child := &fakeChild{stopEntered: make(chan struct{}), allowStop: make(chan struct{}), exited: scenario == "child-exits"}
			if scenario != "stop-times-out" {
				close(child.allowStop)
			}
			if scenario == "stop-fails" {
				child.stopError = errors.New("fixture stop error")
			}
			handler := &serviceHandler{cfg: config{stopTimeout: 20 * time.Millisecond}, start: func(config) (ownedChild, error) {
				if scenario == "start-fails" {
					return nil, errors.New("fixture start error")
				}
				return child, nil
			}}
			requests := make(chan svc.ChangeRequest, 1)
			if scenario == "stop-fails" || scenario == "stop-times-out" {
				requests <- svc.ChangeRequest{Cmd: svc.Stop}
			}
			changes := make(chan svc.Status, 8)
			specific, code := handler.Execute(nil, requests, changes)
			want := map[string]uint32{"start-fails": 1, "child-exits": 2, "stop-fails": 3, "stop-times-out": 3}[scenario]
			if !specific || code != want {
				t.Fatalf("specific=%v code=%d want=%d", specific, code, want)
			}
			if scenario != "start-fails" && !child.closed {
				t.Fatal("owned resources were not closed")
			}
		})
	}
}
