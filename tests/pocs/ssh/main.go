// 文件用途：P02 显式目标 SSH/SFTP 探针命令；不用默认用户配置。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func run() error {
	var o Options
	flag.StringVar(&o.Host, "host", "", "explicit authorized SSH host or config alias")
	flag.StringVar(&o.User, "user", "", "SSH user")
	flag.IntVar(&o.Port, "port", 22, "SSH port")
	flag.StringVar(&o.KeyPath, "key", "", "explicit private key path")
	flag.StringVar(&o.KnownHostsPath, "known-hosts", "", "dedicated POC known-hosts file (required)")
	flag.StringVar(&o.TrustFingerprint, "trust-host-key", "", "out-of-band verified SHA256 fingerprint for unknown host only")
	config := flag.String("config", "", "explicit POC SSH config file")
	command := flag.String("command", "whoami", "authorized remote command (target shell syntax)")
	timeout := flag.Duration("timeout", 15*time.Second, "per-operation connect, command and transfer timeout")
	pty := flag.Bool("pty", false, "request PTY before executing command")
	remoteParent := flag.String("round-trip-parent", "", "explicit authorized remote parent for new temporary directory and 10 MiB probe")
	flag.Parse()
	o.Timeout = *timeout
	if *config != "" {
		f, err := os.Open(*config)
		if err != nil {
			return err
		}
		p, err := parseConfig(f, o.Host)
		_ = f.Close()
		if err != nil {
			return err
		}
		set := map[string]bool{}
		flag.Visit(func(f *flag.Flag) { set[f.Name] = true })
		o.Host = p.Host
		if !set["user"] {
			o.User = p.User
		}
		if !set["port"] {
			o.Port = p.Port
		}
		if !set["key"] {
			o.KeyPath = p.IdentityFile
		}
	}
	o.Password = []byte(os.Getenv("HOPNODE_POC_SSH_PASSWORD"))
	o.KeyPassphrase = []byte(os.Getenv("HOPNODE_POC_KEY_PASSPHRASE"))
	defer clear(o.Password)
	defer clear(o.KeyPassphrase)
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	c, err := connect(ctx, o)
	if err != nil {
		return err
	}
	defer c.Close()
	cmdCtx, cmdCancel := context.WithTimeout(context.Background(), *timeout)
	defer cmdCancel()
	result, cmdErr := execCommand(cmdCtx, c, *command, *pty)
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		return err
	}
	if cmdErr != nil {
		return cmdErr
	}
	if *remoteParent != "" {
		transferCtx, transferCancel := context.WithTimeout(context.Background(), *timeout)
		defer transferCancel()
		stopTransfer := context.AfterFunc(transferCtx, func() { _ = c.Close() })
		defer stopTransfer()
		var suffix [8]byte
		if _, err := rand.Read(suffix[:]); err != nil {
			return err
		}
		payload := make([]byte, 10*1024*1024)
		if _, err := rand.Read(payload); err != nil {
			return err
		}
		root := strings.TrimRight(*remoteParent, "/") + "/hopnode-poc-" + hex.EncodeToString(suffix[:])
		if err := roundTrip(c, root, "中文 binary file.bin", payload); err != nil {
			if transferCtx.Err() != nil {
				return transferCtx.Err()
			}
			return err
		}
		fmt.Fprintln(os.Stdout, `{"sftp_round_trip":"PASS","bytes":10485760}`)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
