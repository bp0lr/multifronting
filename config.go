package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
)

type config struct {
	workers   int
	frontHost string
	testHost  string
	needle    string
	proxy     string
	output    string
	caCert    string
	verbose   bool
	progress  bool
}

func parseConfig(args []string, stderr io.Writer) (config, error) {
	var cfg config
	flags := pflag.NewFlagSet("multifronting", pflag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.IntVarP(&cfg.workers, "workers", "w", 50, "Number of workers (1-99)")
	flags.StringVarP(&cfg.frontHost, "fronturl", "f", "", "Your host, without a scheme or path (required)")
	flags.StringVarP(&cfg.testHost, "testurl", "u", "", "Host to check; defaults to standard input")
	flags.StringVarP(&cfg.needle, "needle", "n", "", "Expected response text (required)")
	flags.StringVarP(&cfg.proxy, "proxy", "p", "", "HTTP or HTTPS proxy URL")
	flags.StringVarP(&cfg.output, "output", "o", "", "Append results to this file")
	flags.StringVar(&cfg.caCert, "ca-cert", "", "PEM CA certificates to trust in addition to system roots")
	flags.BoolVarP(&cfg.verbose, "verbose", "v", false, "Write diagnostics to standard error")
	flags.BoolVar(&cfg.progress, "use-pb", false, "Show progress on standard error")
	if err := flags.Parse(args); err != nil {
		return cfg, err
	}
	if flags.NArg() != 0 {
		return cfg, fmt.Errorf("unexpected positional arguments; use --help for usage")
	}
	if cfg.workers < 1 || cfg.workers > 99 {
		return cfg, fmt.Errorf("--workers must be between 1 and 99")
	}
	if err := validateHost(cfg.frontHost); err != nil {
		return cfg, fmt.Errorf("--fronturl: %w", err)
	}
	if cfg.testHost != "" {
		if err := validateHost(cfg.testHost); err != nil {
			return cfg, fmt.Errorf("--testurl: %w", err)
		}
	}
	if strings.TrimSpace(cfg.needle) == "" {
		return cfg, fmt.Errorf("--needle must not be empty or whitespace")
	}
	if cfg.proxy != "" {
		if _, err := parseProxy(cfg.proxy); err != nil {
			return cfg, err
		}
	}
	return cfg, nil
}

func validateHost(host string) error {
	if host == "" || strings.ContainsAny(host, "/?#@\\ \t\r\n") {
		return fmt.Errorf("expected a hostname or IP address, optionally with a port, without a scheme or path")
	}
	u, err := url.Parse("https://" + host)
	if err != nil || u.Hostname() == "" {
		return fmt.Errorf("invalid hostname or port")
	}
	if strings.HasPrefix(host, "[") && !strings.Contains(u.Hostname(), ":") {
		return fmt.Errorf("brackets are only valid for IPv6 addresses")
	}
	if strings.ContainsAny(u.Hostname(), ":[]") && net.ParseIP(u.Hostname()) == nil {
		return fmt.Errorf("invalid IP address; enclose IPv6 addresses in brackets")
	}
	if strings.Contains(u.Hostname(), ":") && !strings.HasPrefix(host, "[") {
		return fmt.Errorf("enclose IPv6 addresses in brackets")
	}
	if strings.HasSuffix(host, ":") {
		return fmt.Errorf("port must not be empty")
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
	}
	return nil
}

func parseProxy(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("--proxy requires a valid HTTP or HTTPS URL")
	}
	if err := validateHost(u.Host); err != nil {
		return nil, fmt.Errorf("--proxy: %w", err)
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" || u.ForceQuery {
		return nil, fmt.Errorf("--proxy must not contain a path, query, or fragment")
	}
	return u, nil
}
