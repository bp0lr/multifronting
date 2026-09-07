package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInvalidInputDoesNotCreateOutput(t *testing.T) {
	output := filepath.Join(t.TempDir(), "results.txt")
	cfg := config{workers: 1, frontHost: "lab.test", needle: "marker", output: output}
	err := run(context.Background(), cfg, strings.NewReader("invalid/path"), io.Discard, io.Discard)
	if err == nil {
		t.Fatal("invalid input accepted")
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output created before input validation: %v", err)
	}
}

func TestCancelledRunDoesNotCreateOutput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	output := filepath.Join(t.TempDir(), "results.txt")
	cfg := config{workers: 1, frontHost: "lab.test", needle: "marker", testHost: "localhost", output: output}
	if err := run(ctx, cfg, strings.NewReader(""), io.Discard, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation: %v", err)
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected output: %v", err)
	}
}

func TestCommandExitCodes(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"--help"}, 0},
		{[]string{"--unknown"}, 2},
		{[]string{"-f", "lab.test", "-n", "marker", "--ca-cert", filepath.Join(t.TempDir(), "missing.pem")}, 1},
	} {
		var stdout, stderr bytes.Buffer
		if code := command(tc.args, os.Stdin, &stdout, &stderr); code != tc.code {
			t.Fatalf("args %v: exit %d, want %d; %s", tc.args, code, tc.code, &stderr)
		}
		if stdout.Len() != 0 || stderr.Len() == 0 {
			t.Fatal("unexpected output streams")
		}
	}
}
