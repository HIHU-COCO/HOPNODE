// 文件用途：P02 基础 OpenSSH Host 配置解析；显式拒绝未经验证的复杂指令。
// 创建者：HIHU-COCO
// 创建日期：2026-10-19
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

type Profile struct {
	Host, User, IdentityFile string
	Port                     int
}

func fields(line string) ([]string, error) {
	var out []string
	var b strings.Builder
	var quote rune
	flush := func() {
		if b.Len() > 0 {
			out = append(out, b.String())
			b.Reset()
		}
	}
	for _, r := range line {
		if quote != 0 {
			if r == quote {
				quote = 0
			} else {
				b.WriteRune(r)
			}
			continue
		}
		switch r {
		case '\'', '"':
			quote = r
		case '#':
			flush()
			return out, nil
		case ' ', '\t', '=':
			flush()
		default:
			b.WriteRune(r)
		}
	}
	if quote != 0 {
		return nil, errors.New("unterminated SSH config quote")
	}
	flush()
	return out, nil
}

func matches(alias string, patterns []string) bool {
	matched := false
	for _, pattern := range patterns {
		negative := strings.HasPrefix(pattern, "!")
		pattern = strings.TrimPrefix(pattern, "!")
		ok, err := path.Match(strings.ToLower(pattern), strings.ToLower(alias))
		if err != nil {
			return false
		}
		if ok && negative {
			return false
		}
		if ok {
			matched = true
		}
	}
	return matched
}

// parseConfig follows first-obtained scalar values. It never reads a default user file.
// This POC intentionally rejects multiple IdentityFile values and advanced directives.
func parseConfig(r io.Reader, alias string) (Profile, error) {
	p := Profile{}
	active := true
	seen := map[string]bool{}
	s := bufio.NewScanner(r)
	line := 0
	for s.Scan() {
		line++
		f, err := fields(s.Text())
		if err != nil {
			return p, fmt.Errorf("config line %d: %w", line, err)
		}
		if len(f) == 0 {
			continue
		}
		key := strings.ToLower(f[0])
		if len(f) < 2 {
			return p, fmt.Errorf("config line %d: missing value", line)
		}
		if key == "host" {
			active = matches(alias, f[1:])
			continue
		}
		switch key {
		case "include", "match", "proxyjump", "proxycommand", "identityagent":
			return p, fmt.Errorf("unsupported SSH config directive: %s", f[0])
		}
		if !active {
			continue
		}
		if len(f) != 2 {
			return p, fmt.Errorf("config line %d: expected one value", line)
		}
		switch key {
		case "hostname", "user", "port", "identityfile":
			if seen[key] {
				if key == "identityfile" {
					return p, errors.New("multiple IdentityFile values unsupported in POC")
				}
				continue
			}
			seen[key] = true
			value := f[1]
			if strings.Contains(value, "%") || strings.Contains(value, "~") {
				return p, errors.New("SSH config token/tilde expansion unsupported in POC; use absolute paths")
			}
			switch key {
			case "hostname":
				p.Host = value
			case "user":
				p.User = value
			case "identityfile":
				p.IdentityFile = value
			case "port":
				p.Port, err = strconv.Atoi(value)
				if err != nil || p.Port < 1 || p.Port > 65535 {
					return p, errors.New("invalid SSH port")
				}
			}
		default:
			return p, fmt.Errorf("unsupported SSH config directive: %s", f[0])
		}
	}
	if err := s.Err(); err != nil {
		return p, err
	}
	if p.Host == "" {
		p.Host = alias
	}
	if p.Port == 0 {
		p.Port = 22
	}
	return p, nil
}
