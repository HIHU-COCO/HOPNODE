// Purpose: isolated Windows SCM POC configuration. Creator: HIHU-COCO. Created: 2026-10-19.
package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type config struct {
	serviceName string
	core        string
	endpoint    string
	stopTimeout time.Duration
}

var servicePattern = regexp.MustCompile(`^HopNode-P04-POC-[a-z0-9-]{1,64}$`)
var pipePattern = regexp.MustCompile(`^\\\\\.\\pipe\\hopnode-p04-service-[a-z0-9-]{1,64}$`)

func parseConfig(args []string) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("hopnode-service-poc", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&cfg.serviceName, "service-name", "", "unique isolated SCM service name")
	flags.StringVar(&cfg.core, "core", "", "absolute path to the desktop-service POC executable")
	flags.StringVar(&cfg.endpoint, "endpoint", "", "unique isolated Named Pipe")
	flags.DurationVar(&cfg.stopTimeout, "stop-timeout", 10*time.Second, "bounded stop wait")
	if err := flags.Parse(args); err != nil || flags.NArg() != 0 {
		return config{}, errors.New("invalid service POC arguments")
	}
	if !servicePattern.MatchString(cfg.serviceName) || !pipePattern.MatchString(cfg.endpoint) {
		return config{}, errors.New("explicit isolated service name and pipe are required")
	}
	if !filepath.IsAbs(cfg.core) || strings.ContainsRune(cfg.core, '\x00') {
		return config{}, errors.New("core must be an absolute executable path")
	}
	info, err := os.Stat(cfg.core)
	if err != nil || !info.Mode().IsRegular() {
		return config{}, errors.New("core must be an existing regular executable file")
	}
	if cfg.stopTimeout < time.Second || cfg.stopTimeout > time.Minute {
		return config{}, errors.New("stop timeout must be between 1s and 1m")
	}
	return cfg, nil
}
