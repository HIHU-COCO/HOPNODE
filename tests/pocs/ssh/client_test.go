// 文件用途：P02 loopback SSH/SFTP 协议夹具及真实字节传输测试；不改系统 sshd。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type fixture struct {
	listener    net.Listener
	key         ssh.Signer
	clientKey   ed25519.PrivateKey
	pty         atomic.Bool
	stallOpen   atomic.Bool
	wg          sync.WaitGroup
	mu          sync.Mutex
	connections []net.Conn
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	_, hostKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := ssh.NewSignerFromKey(hostKey)
	if err != nil {
		t.Fatal(err)
	}
	_, clientKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	clientSigner, err := ssh.NewSignerFromKey(clientKey)
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{listener: l, key: signer, clientKey: clientKey}
	config := &ssh.ServerConfig{PasswordCallback: func(c ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
		if c.User() == "poc-user" && string(password) == "synthetic-poc-password" {
			return nil, nil
		}
		return nil, errors.New("fixture authentication rejected")
	}, PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
		if c.User() == "poc-user" && bytes.Equal(key.Marshal(), clientSigner.PublicKey().Marshal()) {
			return nil, nil
		}
		return nil, errors.New("fixture key rejected")
	}}
	config.AddHostKey(signer)
	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			f.mu.Lock()
			f.connections = append(f.connections, conn)
			f.mu.Unlock()
			f.wg.Add(1)
			go func() {
				defer f.wg.Done()
				defer conn.Close()
				_, channels, requests, err := ssh.NewServerConn(conn, config)
				if err != nil {
					return
				}
				go ssh.DiscardRequests(requests)
				for channel := range channels {
					if f.stallOpen.Load() {
						continue // Deliberately leave channel-open unanswered until transport closes.
					}
					if channel.ChannelType() != "session" {
						_ = channel.Reject(ssh.UnknownChannelType, "session only")
						continue
					}
					ch, reqs, err := channel.Accept()
					if err != nil {
						continue
					}
					f.wg.Add(1)
					go func() { defer f.wg.Done(); defer ch.Close(); f.serve(ch, reqs) }()
				}
			}()
		}
	}()
	t.Cleanup(func() {
		_ = l.Close()
		f.mu.Lock()
		for _, c := range f.connections {
			_ = c.Close()
		}
		f.mu.Unlock()
		f.wg.Wait()
	})
	return f
}

func (f *fixture) serve(ch ssh.Channel, requests <-chan *ssh.Request) {
	for req := range requests {
		switch req.Type {
		case "subsystem":
			var msg struct{ Name string }
			if ssh.Unmarshal(req.Payload, &msg) != nil || msg.Name != "sftp" {
				_ = req.Reply(false, nil)
				continue
			}
			_ = req.Reply(true, nil)
			server, err := sftp.NewServer(ch)
			if err != nil {
				return
			}
			_ = server.Serve()
			_ = server.Close()
			return
		case "pty-req":
			f.pty.Store(true)
			_ = req.Reply(true, nil)
		case "exec":
			var msg struct{ Command string }
			if ssh.Unmarshal(req.Payload, &msg) != nil {
				return
			}
			if msg.Command == "fixture-stall-start" {
				for range requests {
				}
				return
			}
			_ = req.Reply(true, nil)
			code := uint32(0)
			switch msg.Command {
			case "whoami":
				out, err := exec.Command("whoami").Output()
				if err != nil {
					code = 1
				}
				_, _ = ch.Write(out)
			case "fixture-fail":
				_, _ = io.WriteString(ch, "fixture stdout\n")
				_, _ = io.WriteString(ch.Stderr(), "fixture stderr\n")
				code = 7
			case "fixture-wait":
				for range requests {
				}
				return
			default:
				_, _ = io.WriteString(ch.Stderr(), "fixture command unsupported\n")
				code = 127
			}
			_, _ = ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{code}))
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func (f *fixture) options(t *testing.T) Options {
	t.Helper()
	host, port, err := net.SplitHostPort(f.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	n, _ := strconv.Atoi(port)
	return Options{Host: host, Port: n, User: "poc-user", Password: []byte("synthetic-poc-password"), KnownHostsPath: filepath.Join(t.TempDir(), "known_hosts"), TrustFingerprint: ssh.FingerprintSHA256(f.key.PublicKey()), Timeout: 3 * time.Second}
}

func TestPasswordCommandSFTPAndPTY(t *testing.T) {
	f := newFixture(t)
	o := f.options(t)
	c, err := connect(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	result, err := execCommand(context.Background(), c, "whoami", true)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(result.Stdout) == "" || result.ExitCode != 0 {
		t.Fatalf("whoami result: %+v", result)
	}
	if !f.pty.Load() {
		t.Fatal("PTY request not observed")
	}
	result, err = execCommand(context.Background(), c, "fixture-fail", false)
	if err == nil || result.ExitCode != 7 || result.Stdout != "fixture stdout\n" || result.Stderr != "fixture stderr\n" {
		t.Fatalf("exit status/streams mismatch: %+v %v", result, err)
	}
	for _, size := range []int{0, 256, 10 * 1024 * 1024} {
		t.Run(fmt.Sprint(size), func(t *testing.T) {
			data := make([]byte, size)
			if _, err := rand.Read(data); err != nil {
				t.Fatal(err)
			}
			root := filepath.ToSlash(filepath.Join(t.TempDir(), "probe 中文 space"))
			if err := roundTrip(c, root, "二进制 file.bin", data); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("probe directory not removed")
			}
		})
	}
}

func TestHostKeyUnknownTrustAndChanged(t *testing.T) {
	f := newFixture(t)
	o := f.options(t)
	fingerprint := o.TrustFingerprint
	o.TrustFingerprint = ""
	if c, err := connect(context.Background(), o); err == nil {
		c.Close()
		t.Fatal("unknown host was trusted")
	} else if !strings.Contains(err.Error(), ErrHostKeyUnknown.Error()) {
		t.Fatal(err)
	}
	if _, err := os.Stat(o.KnownHostsPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unknown host modified known-hosts")
	}
	o.TrustFingerprint = "SHA256:incorrect-out-of-band-fingerprint"
	if c, err := connect(context.Background(), o); err == nil {
		c.Close()
		t.Fatal("incorrect fingerprint accepted")
	}
	if _, err := os.Stat(o.KnownHostsPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("incorrect fingerprint modified known-hosts")
	}
	o.TrustFingerprint = fingerprint
	c, err := connect(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	o.TrustFingerprint = ""
	c, err = connect(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	before, err := os.ReadFile(o.KnownHostsPath)
	if err != nil {
		t.Fatal(err)
	}
	_, otherKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	other, err := ssh.NewSignerFromKey(otherKey)
	if err != nil {
		t.Fatal(err)
	}
	err = hostVerifier(o.KnownHostsPath, ssh.FingerprintSHA256(other.PublicKey()))(f.listener.Addr().String(), f.listener.Addr(), other.PublicKey())
	if !errors.Is(err, ErrHostKeyChanged) {
		t.Fatalf("changed key not rejected: %v", err)
	}
	after, _ := os.ReadFile(o.KnownHostsPath)
	if !bytes.Equal(before, after) {
		t.Fatal("changed key overwrote known-hosts")
	}
}

func TestPrivateKeyAndRejectedPassword(t *testing.T) {
	f := newFixture(t)
	o := f.options(t)
	o.Password = []byte("wrong-synthetic-password")
	if c, err := connect(context.Background(), o); err == nil {
		c.Close()
		t.Fatal("wrong password accepted")
	}
	for _, encrypted := range []bool{false, true} {
		t.Run(fmt.Sprint(encrypted), func(t *testing.T) {
			o.Password = nil
			var block *pem.Block
			var err error
			if encrypted {
				o.KeyPassphrase = []byte("synthetic-key-passphrase")
				block, err = ssh.MarshalPrivateKeyWithPassphrase(f.clientKey, "P02 test fixture", o.KeyPassphrase)
			} else {
				o.KeyPassphrase = nil
				block, err = ssh.MarshalPrivateKey(f.clientKey, "P02 test fixture")
			}
			if err != nil {
				t.Fatal(err)
			}
			o.KeyPath = filepath.Join(t.TempDir(), "fixture_key")
			if err := os.WriteFile(o.KeyPath, pem.EncodeToMemory(block), 0600); err != nil {
				t.Fatal(err)
			}
			c, err := connect(context.Background(), o)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			if _, err := execCommand(context.Background(), c, "whoami", false); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCommandTimeoutDoesNotClaimRemoteTermination(t *testing.T) {
	f := newFixture(t)
	c, err := connect(context.Background(), f.options(t))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	result, err := execCommand(ctx, c, "fixture-wait", false)
	if !errors.Is(err, context.DeadlineExceeded) || !result.TimedOut || result.ExitCode != -1 {
		t.Fatalf("timeout result %+v %v", result, err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("cancellation did not bound local wait")
	}
	if _, err := execCommand(context.Background(), c, "whoami", false); err != nil {
		t.Fatalf("canceled session broke shared SSH client: %v", err)
	}
}

func TestCommandTimeoutDuringSetup(t *testing.T) {
	for _, stallOpen := range []bool{false, true} {
		t.Run(fmt.Sprint(stallOpen), func(t *testing.T) {
			f := newFixture(t)
			c, err := connect(context.Background(), f.options(t))
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			f.stallOpen.Store(stallOpen)
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			start := time.Now()
			result, err := execCommand(ctx, c, "fixture-stall-start", false)
			if !errors.Is(err, context.DeadlineExceeded) || !result.TimedOut || result.ExitCode != -1 {
				t.Fatalf("setup timeout: %+v %v", result, err)
			}
			if time.Since(start) > 2*time.Second {
				t.Fatal("setup timeout did not bound local wait")
			}
		})
	}
}

func TestRoundTripRejectsEscapeAndExistingDirectory(t *testing.T) {
	f := newFixture(t)
	c, err := connect(context.Background(), f.options(t))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	root := filepath.ToSlash(filepath.Join(t.TempDir(), "new-probe"))
	for _, name := range []string{"../escape", "subdir/file", "subdir\\file", "", ".", ".."} {
		if err := roundTrip(c, root, name, nil); err == nil {
			t.Fatalf("unsafe name accepted: %q", name)
		}
	}
	if _, err := os.Stat(root); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unsafe filename created a directory")
	}
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(filepath.FromSlash(root), "keep")
	if err := os.WriteFile(keep, []byte("preserve-existing-data"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := roundTrip(c, root, "probe.bin", nil); err == nil {
		t.Fatal("existing probe directory accepted")
	}
	data, err := os.ReadFile(keep)
	if err != nil || string(data) != "preserve-existing-data" {
		t.Fatalf("existing data changed: %q %v", data, err)
	}
}
