package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestInputWhitespaceAndSingleHost(t *testing.T) {
	hosts, err := readInputs(context.Background(), strings.NewReader("  lab.test \r\n\r\nlocalhost:443\n"), "")
	if err != nil || len(hosts) != 2 || hosts[0] != "lab.test" || hosts[1] != "localhost:443" {
		t.Fatalf("hosts: %v, error: %v", hosts, err)
	}
	hosts, err = readInputs(context.Background(), failingReader{errors.New("stdin must not be read")}, "lab.test")
	if err != nil || len(hosts) != 1 || hosts[0] != "lab.test" {
		t.Fatalf("single host: %v, %v", hosts, err)
	}
}

func TestInputFailures(t *testing.T) {
	for _, input := range []string{"\n ", "lab.test\nbad/path\n", strings.Repeat("a", 70_000)} {
		if _, err := readInputs(context.Background(), strings.NewReader(input), ""); err == nil {
			t.Fatal("expected input error")
		}
	}
	want := errors.New("input failure")
	if _, err := readInputs(context.Background(), failingReader{want}, ""); !errors.Is(err, want) {
		t.Fatalf("lost reader error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readInputs(ctx, strings.NewReader("lab.test"), ""); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}
