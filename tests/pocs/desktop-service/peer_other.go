//go:build !linux

// Purpose: avoid claiming peer identity where the Linux-specific POC probe is unavailable. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import "net"

func readPeerIdentity(net.Conn) peerIdentity { return peerIdentity{UID: -1} }
