package main

import (
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cheggaaa/pb/v3"
)

// reporter serializes writes and keeps output failures visible to the caller.
type reporter struct {
	mu       sync.Mutex
	stdout   io.Writer
	stderr   io.Writer
	file     io.Writer
	verbose  bool
	bar      *pb.ProgressBar
	failures int
	writeErr error
}

func (r *reporter) record(host string, matched bool, checkErr error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.bar != nil {
		defer r.bar.Increment()
	}
	if checkErr != nil {
		r.failures++
		// The final error remains visible even if progress suppresses per-host logs.
		if r.bar == nil {
			r.write(r.stderr, "error: %s: %v\n", host, checkErr)
		}
		return
	}
	if matched {
		if r.file != nil {
			r.write(r.file, "%s\n", host)
		}
		r.write(r.stdout, "%s\n", host)
	} else if r.verbose && r.bar == nil {
		r.write(r.stderr, "no match: %s\n", host)
	}
}

func (r *reporter) write(dst io.Writer, format string, args ...any) {
	if _, err := fmt.Fprintf(dst, format, args...); err != nil && r.writeErr == nil {
		r.writeErr = fmt.Errorf("write output: %w", err)
	}
}

func (r *reporter) err() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var checkErr error
	if r.failures != 0 {
		checkErr = fmt.Errorf("%d request(s) failed; results may be incomplete", r.failures)
	}
	return errors.Join(checkErr, r.writeErr)
}
