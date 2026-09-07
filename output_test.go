package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/cheggaaa/pb/v3"
)

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestReporterSeparatesResultsAndDiagnostics(t *testing.T) {
	var stdout, stderr, file bytes.Buffer
	r := &reporter{stdout: &stdout, stderr: &stderr, file: &file, verbose: true}
	r.record("one.test", true, nil)
	r.record("two.test", false, nil)
	r.record("three.test", false, errors.New("request failed"))
	if stdout.String() != "one.test\n" || file.String() != stdout.String() {
		t.Fatalf("unexpected results: stdout=%q, file=%q", &stdout, &file)
	}
	if !strings.Contains(stderr.String(), "two.test") || !strings.Contains(stderr.String(), "three.test") || r.err() == nil {
		t.Fatalf("missing diagnostics: %q, %v", &stderr, r.err())
	}
}

func TestReporterKeepsResultsWithProgress(t *testing.T) {
	var stdout, stderr bytes.Buffer
	r := &reporter{stdout: &stdout, stderr: &stderr, bar: pb.New(2)}
	r.record("one.test", true, nil)
	r.record("two.test", false, errors.New("request failed"))
	if stdout.String() != "one.test\n" || stderr.Len() != 0 || r.err() == nil {
		t.Fatalf("unexpected progress output: %q, %q, %v", &stdout, &stderr, r.err())
	}
}

func TestReporterPropagatesWriteFailures(t *testing.T) {
	want := errors.New("disk full")
	for _, r := range []*reporter{
		{stdout: failingWriter{want}, stderr: io.Discard},
		{stdout: io.Discard, stderr: io.Discard, file: failingWriter{want}},
	} {
		r.record("one.test", true, nil)
		if !errors.Is(r.err(), want) {
			t.Fatalf("lost write failure: %v", r.err())
		}
	}
	r := &reporter{stdout: io.Discard, stderr: failingWriter{want}}
	r.record("one.test", false, errors.New("request failed"))
	if !errors.Is(r.err(), want) {
		t.Fatalf("lost stderr failure: %v", r.err())
	}
}

func TestReporterConcurrentWritesPreserveLines(t *testing.T) {
	var stdout bytes.Buffer
	r := &reporter{stdout: &stdout, stderr: io.Discard}
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.record(fmt.Sprintf("host%d.test", i), true, nil)
		}()
	}
	wg.Wait()
	lines := strings.Fields(stdout.String())
	seen := make(map[string]bool)
	for _, line := range lines {
		seen[line] = true
	}
	if len(lines) != 25 || len(seen) != 25 || r.err() != nil {
		t.Fatalf("lost or interleaved lines: %q, %v", &stdout, r.err())
	}
}
