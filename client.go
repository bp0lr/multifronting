package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

const maxResponseBytes int64 = 1 << 20

func newClient(cfg config) (*http.Client, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if cfg.caCert != "" {
		pem, err := os.ReadFile(cfg.caCert)
		if err != nil {
			return nil, fmt.Errorf("read CA certificates: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("load system CA certificates: %w", err)
		}
		if !roots.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("CA file contains no valid PEM certificates")
		}
		tlsConfig.RootCAs = roots
	}
	tr := &http.Transport{
		MaxIdleConns:    30,
		IdleConnTimeout: time.Second,
		TLSClientConfig: tlsConfig,
		DialContext:     (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}
	if cfg.proxy != "" {
		proxy, err := parseProxy(cfg.proxy)
		if err != nil {
			return nil, err
		}
		tr.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

func readResponse(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if int64(len(data)) > maxResponseBytes {
		return nil, fmt.Errorf("response exceeds the %d-byte limit", maxResponseBytes)
	}
	return data, nil
}
