//go:build windows

// Purpose: SCM lifecycle with bounded owned-child shutdown. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"time"

	"golang.org/x/sys/windows/svc"
)

type ownedChild interface {
	Exited() (bool, error)
	Stop(context.Context) error
	Close() error
}

type serviceHandler struct {
	cfg   config
	start func(config) (ownedChild, error)
}

func (handler *serviceHandler) Execute(_ []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending, WaitHint: 5000}
	child, err := handler.start(handler.cfg)
	if err != nil {
		return true, 1
	}
	defer child.Close()
	// Running means process creation succeeded, not an IPC readiness/health claim.
	status := svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	changes <- status
	poll := time.NewTicker(250 * time.Millisecond)
	defer poll.Stop()
	stop := func() error {
		changes <- svc.Status{State: svc.StopPending, WaitHint: uint32(handler.cfg.stopTimeout.Milliseconds())}
		ctx, cancel := context.WithTimeout(context.Background(), handler.cfg.stopTimeout)
		defer cancel()
		return child.Stop(ctx)
	}
	for {
		select {
		case request, open := <-requests:
			if !open || request.Cmd == svc.Stop || request.Cmd == svc.Shutdown {
				if err := stop(); err != nil {
					return true, 3
				}
				return false, 0
			}
			if request.Cmd == svc.Interrogate {
				changes <- status
			}
		case <-poll.C:
			exited, err := child.Exited()
			if err != nil || exited {
				// Also terminate remaining descendants if the primary child exits first.
				if err := stop(); err != nil {
					return true, 3
				}
				return true, 2
			}
		}
	}
}
