//go:build linux

// Purpose: use kernel SO_PEERCRED rather than trusting a client-provided PID. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"net"
	"syscall"
)

func readPeerIdentity(conn net.Conn) peerIdentity {
	unknown := peerIdentity{UID: -1}
	unixConn, ok := conn.(*net.UnixConn)
	if !ok {
		return unknown
	}
	raw, err := unixConn.SyscallConn()
	if err != nil {
		return unknown
	}
	var credentials *syscall.Ucred
	var credentialErr error
	err = raw.Control(func(fd uintptr) {
		credentials, credentialErr = syscall.GetsockoptUcred(int(fd), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	})
	if err != nil || credentialErr != nil || credentials == nil {
		return unknown
	}
	return peerIdentity{PID: int(credentials.Pid), UID: int(credentials.Uid), Verified: true}
}
