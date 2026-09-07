package main

import (
	"context"
	"encoding/pem"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResponseLimitAndReadErrors(t *testing.T) {
	for _, size := range []int64{0, maxResponseBytes, maxResponseBytes + 1} {
		data, err := readResponse(strings.NewReader(strings.Repeat("x", int(size))))
		if size > maxResponseBytes {
			if err == nil || data != nil {
				t.Fatal("oversized response was accepted")
			}
		} else if err != nil || int64(len(data)) != size {
			t.Fatalf("size %d: got %d bytes, %v", size, len(data), err)
		}
	}
	want := errors.New("read failed")
	if _, err := readResponse(failingReader{want}); !errors.Is(err, want) {
		t.Fatalf("lost response read error: %v", err)
	}
}

func TestTLSRequiresTrustedCertificate(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	server.Config.ErrorLog = log.New(io.Discard, "", 0)
	server.StartTLS()
	defer server.Close()
	client, err := newClient(config{})
	if err != nil {
		t.Fatal(err)
	}
	defer client.CloseIdleConnections()
	if resp, err := client.Get(server.URL); err == nil {
		resp.Body.Close()
		t.Fatal("untrusted TLS certificate accepted")
	}
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(caFile, cert, 0600); err != nil {
		t.Fatal(err)
	}
	trusted, err := newClient(config{caCert: caFile})
	if err != nil {
		t.Fatal(err)
	}
	defer trusted.CloseIdleConnections()
	resp, err := trusted.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}

func TestClientRejectsInvalidCAAndProxy(t *testing.T) {
	badCA := filepath.Join(t.TempDir(), "bad.pem")
	if err := os.WriteFile(badCA, []byte("not a certificate"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, cfg := range []config{{caCert: badCA}, {caCert: badCA + ".missing"}, {proxy: "not-a-proxy"}} {
		if _, err := newClient(cfg); err == nil {
			t.Fatal("invalid client configuration accepted")
		}
	}
}

func TestClientDoesNotFollowRedirectsAndHonorsCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			t.Error("redirect followed")
		}
		http.Redirect(w, r, "/unexpected", http.StatusFound)
	}))
	defer server.Close()
	client, err := newClient(config{})
	if err != nil {
		t.Fatal(err)
	}
	defer client.CloseIdleConnections()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp, err := client.Do(req); !errors.Is(err, context.Canceled) {
		if resp != nil {
			resp.Body.Close()
		}
		t.Fatalf("expected cancellation, got %v", err)
	}
}
