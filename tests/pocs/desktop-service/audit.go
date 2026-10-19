// Purpose: optional private POC evidence binding real IPC requests to OS peer credentials. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

type peerIdentity struct {
	PID      int  `json:"peer_pid"`
	UID      int  `json:"peer_uid"`
	Verified bool `json:"peer_identity_verified"`
}
type peerContextKey struct{}
type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (writer *statusRecorder) WriteHeader(code int) {
	writer.code = code
	writer.ResponseWriter.WriteHeader(code)
}

func newServer(status Status, audit *os.File) *http.Server {
	var auditLock sync.Mutex
	endpointHandler := handler(status)
	var wrapped http.Handler = endpointHandler
	if audit != nil {
		wrapped = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
			endpointHandler.ServeHTTP(recorder, r)
			peer, _ := r.Context().Value(peerContextKey{}).(peerIdentity)
			event := struct {
				peerIdentity
				Method        string `json:"method"`
				Path          string `json:"path"`
				HTTPStatus    int    `json:"http_status"`
				ControllerPID int    `json:"controller_pid"`
				InstanceID    string `json:"instance_id"`
			}{peer, r.Method, r.URL.Path, recorder.code, status.PID, status.InstanceID}
			auditLock.Lock()
			defer auditLock.Unlock()
			if err := json.NewEncoder(audit).Encode(event); err != nil {
				fmtAuditFailure(err)
			}
		})
	}
	return &http.Server{Handler: wrapped, ReadHeaderTimeout: 3 * time.Second, WriteTimeout: 3 * time.Second, IdleTimeout: 3 * time.Second, MaxHeaderBytes: 4096, ConnContext: func(ctx context.Context, conn net.Conn) context.Context {
		return context.WithValue(ctx, peerContextKey{}, readPeerIdentity(conn))
	}}
}

func fmtAuditFailure(err error) {
	_, _ = os.Stderr.WriteString("POC audit write failed: " + err.Error() + "\n")
}
