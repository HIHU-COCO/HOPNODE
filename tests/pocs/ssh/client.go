// 文件用途：P02 独立 SSH/SFTP 客户端；不读取默认用户配置或自动信任主机。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

var ErrHostKeyUnknown = errors.New("SSH_HOST_KEY_UNKNOWN")
var ErrHostKeyChanged = errors.New("SSH_HOST_KEY_CHANGED")

type Options struct {
	Host, User, KeyPath, KnownHostsPath, TrustFingerprint string
	Port                                                  int
	Password, KeyPassphrase                               []byte
	Timeout                                               time.Duration
}

// algorithmProbe is used only to ask knownhosts for matching records. Its bytes
// cannot be a valid SSH public key, and it can never verify a signature.
type algorithmProbe struct{}

func (algorithmProbe) Type() string    { return "hopnode-poc-algorithm-query" }
func (algorithmProbe) Marshal() []byte { return []byte("hopnode-poc-algorithm-query") }
func (algorithmProbe) Verify([]byte, *ssh.Signature) error {
	return errors.New("host-key algorithm probe cannot verify signatures")
}

// pinnedHostKeyAlgorithms keeps negotiation within the keys already trusted for
// this address. Use knownhosts itself for hashed names, patterns and port rules.
// Actual key bytes (including revocations/certificates) are still checked later.
func pinnedHostKeyAlgorithms(file, address string, remote net.Addr) ([]string, error) {
	if file == "" {
		return nil, errors.New("known-hosts path is required")
	}
	check, err := knownhosts.New(file)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil // Unknown host: retain the explicit fingerprint flow.
	}
	if err != nil {
		return nil, fmt.Errorf("invalid known-hosts: %w", err)
	}
	err = check(address, remote, algorithmProbe{})
	var mismatch *knownhosts.KeyError
	if !errors.As(err, &mismatch) {
		return nil, fmt.Errorf("could not determine pinned host-key algorithms: %w", err)
	}
	if len(mismatch.Want) == 0 {
		return nil, nil
	}
	contents, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(contents), "\n")
	types := map[string]bool{}
	certAuthority := false
	for _, known := range mismatch.Want {
		if known.Line < 1 || known.Line > len(lines) {
			return nil, errors.New("known-hosts changed while reading algorithm metadata")
		}
		if strings.HasPrefix(strings.TrimSpace(lines[known.Line-1]), "@cert-authority ") || strings.HasPrefix(strings.TrimSpace(lines[known.Line-1]), "@cert-authority\t") {
			certAuthority = true // A CA can sign host certificates of other key types.
		} else {
			types[known.Key.Type()] = true
		}
	}
	var algorithms []string
	for _, algorithm := range ssh.SupportedAlgorithms().HostKeys {
		if strings.Contains(algorithm, "-cert-") {
			if certAuthority {
				algorithms = append(algorithms, algorithm)
			}
			continue
		}
		keyType := algorithm
		if algorithm == ssh.KeyAlgoRSASHA256 || algorithm == ssh.KeyAlgoRSASHA512 {
			keyType = ssh.KeyAlgoRSA
		}
		if types[keyType] {
			algorithms = append(algorithms, algorithm)
		}
	}
	if len(algorithms) == 0 {
		return nil, errors.New("pinned host key has no supported secure negotiation algorithm")
	}
	return algorithms, nil
}

// hostVerifier requires explicit out-of-band fingerprint approval only for unknown hosts.
// Changed keys always fail, even when the caller supplies a trust fingerprint.
func hostVerifier(file, approved string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		if file == "" {
			return errors.New("known-hosts path is required")
		}
		var check ssh.HostKeyCallback
		if _, err := os.Stat(file); err == nil {
			var parseErr error
			check, parseErr = knownhosts.New(file)
			if parseErr != nil {
				return fmt.Errorf("invalid known-hosts: %w", parseErr)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if check != nil {
			err := check(hostname, remote, key)
			if err == nil {
				return nil
			}
			var keyErr *knownhosts.KeyError
			if !errors.As(err, &keyErr) {
				return err
			}
			if len(keyErr.Want) > 0 {
				return fmt.Errorf("%w: %s", ErrHostKeyChanged, ssh.FingerprintSHA256(key))
			}
		}
		fingerprint := ssh.FingerprintSHA256(key)
		if approved == "" || subtle.ConstantTimeCompare([]byte(approved), []byte(fingerprint)) != 1 {
			return fmt.Errorf("%w: verify fingerprint out of band: %s", ErrHostKeyUnknown, fingerprint)
		}
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		_, writeErr := fmt.Fprintln(f, knownhosts.Line([]string{hostname}, key))
		closeErr := f.Close()
		return errors.Join(writeErr, closeErr)
	}
}

func connect(ctx context.Context, o Options) (*ssh.Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if o.Host == "" || o.User == "" || o.Port < 1 || o.Port > 65535 {
		return nil, errors.New("host, user and valid port are required")
	}
	auth := []ssh.AuthMethod{}
	if len(o.Password) > 0 {
		auth = append(auth, ssh.Password(string(o.Password)))
	}
	if o.KeyPath != "" {
		key, err := os.ReadFile(o.KeyPath)
		if err != nil {
			return nil, fmt.Errorf("private key unavailable: %w", err)
		}
		var signer ssh.Signer
		if len(o.KeyPassphrase) > 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(key, o.KeyPassphrase)
		} else {
			signer, err = ssh.ParsePrivateKey(key)
		}
		clear(key)
		if err != nil {
			return nil, errors.New("private key could not be parsed or decrypted")
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	if len(auth) == 0 {
		return nil, errors.New("password environment variable or explicit private key required")
	}
	if o.Timeout <= 0 {
		o.Timeout = 15 * time.Second
	}
	address := net.JoinHostPort(o.Host, strconv.Itoa(o.Port))
	conn, err := (&net.Dialer{Timeout: o.Timeout}).DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(o.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	if err := conn.SetDeadline(deadline); err != nil {
		_ = conn.Close()
		return nil, err
	}
	algorithms, err := pinnedHostKeyAlgorithms(o.KnownHostsPath, address, conn.RemoteAddr())
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	cc, channels, requests, err := ssh.NewClientConn(conn, address, &ssh.ClientConfig{User: o.User, Auth: auth, HostKeyCallback: hostVerifier(o.KnownHostsPath, o.TrustFingerprint), HostKeyAlgorithms: algorithms})
	close(done)
	wg.Wait()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := conn.SetDeadline(time.Time{}); err != nil {
		_ = cc.Close()
		return nil, err
	}
	return ssh.NewClient(cc, channels, requests), nil
}

type CommandResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out"`
}

// execCommand closes the SSH session on cancellation. It does not promise remote process termination.
func execCommand(ctx context.Context, client *ssh.Client, command string, pty bool) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{ExitCode: -1, TimedOut: errors.Is(err, context.DeadlineExceeded)}, err
	}
	type openedSession struct {
		session *ssh.Session
		err     error
	}
	opened := make(chan openedSession, 1)
	go func() {
		s, err := client.NewSession()
		opened <- openedSession{s, err}
	}()
	var session *ssh.Session
	var err error
	select {
	case result := <-opened:
		session, err = result.session, result.err
	case <-ctx.Done():
		// Channel-open has no SSH cancellation primitive. Drop an unresponsive
		// transport so its pending request and goroutine are released.
		_ = client.Close()
		result := <-opened
		if result.session != nil {
			_ = result.session.Close()
		}
		return CommandResult{ExitCode: -1, TimedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}, ctx.Err()
	}
	if err != nil {
		return CommandResult{ExitCode: -1}, err
	}
	defer session.Close()
	stopCancel := context.AfterFunc(ctx, func() { _ = session.Close() })
	defer stopCancel()
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr
	if pty {
		if err := session.RequestPty("xterm", 24, 80, ssh.TerminalModes{ssh.ECHO: 0}); err != nil {
			if ctx.Err() != nil {
				return CommandResult{ExitCode: -1, TimedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}, ctx.Err()
			}
			return CommandResult{ExitCode: -1}, err
		}
	}
	if err := session.Start(command); err != nil {
		if ctx.Err() != nil {
			return CommandResult{ExitCode: -1, TimedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}, ctx.Err()
		}
		return CommandResult{ExitCode: -1}, err
	}
	done := make(chan error, 1)
	go func() { done <- session.Wait() }()
	select {
	case err = <-done:
	case <-ctx.Done():
		_ = session.Close()
		<-done
		return CommandResult{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: -1, TimedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}, ctx.Err()
	}
	if ctx.Err() != nil {
		return CommandResult{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: -1, TimedOut: errors.Is(ctx.Err(), context.DeadlineExceeded)}, ctx.Err()
	}
	result := CommandResult{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitStatus()
		} else {
			result.ExitCode = -1
		}
	}
	return result, err
}

// roundTrip only creates a fresh remote directory and removes files it created there.
// This is a fixture/probe operation, not a generic overwrite or backup implementation.
func roundTrip(client *ssh.Client, root, name string, contents []byte) (returnErr error) {
	if name == "" || name == "." || name == ".." || path.Base(name) != name || strings.ContainsAny(name, "\\\x00") {
		return errors.New("probe filename must be a single nonempty basename")
	}
	s, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer s.Close()
	if err = s.Mkdir(root); err != nil {
		return fmt.Errorf("probe requires a new remote directory: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, s.RemoveDirectory(root)) }()
	file := root + "/" + name
	f, err := s.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	if err != nil {
		return err
	}
	_, writeErr := io.Copy(f, bytes.NewReader(contents))
	closeErr := f.Close()
	defer func() { returnErr = errors.Join(returnErr, s.Remove(file)) }()
	if err = errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	r, err := s.Open(file)
	if err != nil {
		return err
	}
	var back bytes.Buffer
	_, readErr := io.Copy(&back, r)
	closeErr = r.Close()
	if err = errors.Join(readErr, closeErr); err != nil {
		return err
	}
	if !bytes.Equal(contents, back.Bytes()) {
		return errors.New("SFTP round-trip byte mismatch")
	}
	entries, err := s.ReadDir(root)
	if err != nil {
		return err
	}
	if len(entries) != 1 || entries[0].Name() != name {
		return errors.New("SFTP directory listing mismatch")
	}
	return nil
}
