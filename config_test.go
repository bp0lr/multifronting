package main

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestConfigRejectsInvalidArguments(t *testing.T) {
	for name, args := range map[string][]string{
		"missing required flags": {},
		"empty marker":           {"-f", "lab.test"},
		"whitespace marker":      {"-f", "lab.test", "-n", " "},
		"URL instead of host":    {"-f", "https://lab.test", "-n", "marker"},
		"zero workers":           {"-f", "lab.test", "-n", "marker", "-w", "0"},
		"too many workers":       {"-f", "lab.test", "-n", "marker", "-w", "100"},
		"positional argument":    {"-f", "lab.test", "-n", "marker", "extra"},
		"unknown flag":           {"--unknown"},
		"invalid input host":     {"-f", "lab.test", "-n", "marker", "-u", "host.test/path"},
		"proxy without scheme":   {"-f", "lab.test", "-n", "marker", "-p", "localhost:8080"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseConfig(args, io.Discard); err == nil {
				t.Fatal("expected argument error")
			}
		})
	}
}

func TestConfigDefaultsAndHelp(t *testing.T) {
	cfg, err := parseConfig([]string{"-f", "lab.test", "-n", "marker"}, io.Discard)
	if err != nil || cfg.workers != 50 || cfg.progress || cfg.output != "" {
		t.Fatalf("unexpected defaults: %+v, %v", cfg, err)
	}
	var help bytes.Buffer
	_, err = parseConfig([]string{"--help"}, &help)
	if !errors.Is(err, pflag.ErrHelp) || !strings.Contains(help.String(), "--ca-cert") {
		t.Fatalf("help: %v, %s", err, help.String())
	}
}

func TestHostValidation(t *testing.T) {
	for _, host := range []string{"localhost", "lab.test", "lab.test:443", "127.0.0.1", "[::1]", "[::1]:8443"} {
		if err := validateHost(host); err != nil {
			t.Errorf("valid host %q: %v", host, err)
		}
	}
	for _, host := range []string{"", "lab.test/", "lab.test?x", "user@lab.test", "lab.test:0", "lab.test:65536", "lab.test:", "lab.test:abc", "::1", "[bad]", "a b", "https://lab.test"} {
		if err := validateHost(host); err == nil {
			t.Errorf("accepted invalid host %q", host)
		}
	}
}

func TestProxyErrorsDoNotExposeCredentials(t *testing.T) {
	for _, proxy := range []string{"http://user:secret@", "http://user:secret@localhost:0", "http://user:secret@localhost/path", "socks5://user:secret@localhost:1080"} {
		_, err := parseProxy(proxy)
		if err == nil || strings.Contains(err.Error(), "secret") {
			t.Errorf("unsafe or missing error for invalid proxy: %v", err)
		}
	}
}
