package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"sync"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/pflag"
)

func main() {
	os.Exit(command(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func command(args []string, stdin *os.File, stdout, stderr io.Writer) int {
	cfg, err := parseConfig(args, stderr)
	if errors.Is(err, pflag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// Closing stdin releases a blocked read when the user interrupts the CLI.
	stopClose := context.AfterFunc(ctx, func() { _ = stdin.Close() })
	defer stopClose()
	err = run(ctx, cfg, stdin, stdout, stderr)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		if ctx.Err() != nil {
			return 130
		}
		return 1
	}
	return 0
}

func run(ctx context.Context, cfg config, stdin io.Reader, stdout, stderr io.Writer) (runErr error) {
	client, err := newClient(cfg)
	if err != nil {
		return err
	}
	defer client.CloseIdleConnections()
	hosts, err := readInputs(ctx, stdin, cfg.testHost)
	if err != nil {
		return err
	}
	var file *os.File
	if cfg.output != "" {
		file, err = os.OpenFile(cfg.output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("open output: %w", err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				runErr = errors.Join(runErr, fmt.Errorf("close output: %w", err))
			}
		}()
	}
	report := &reporter{stdout: stdout, stderr: stderr, verbose: cfg.verbose}
	if file != nil {
		report.file = file
	}
	if cfg.progress {
		bar := pb.New(len(hosts)).SetWriter(stderr).SetTemplateString(`Checks: {{counters .}} {{bar .}} {{percent .}}`)
		bar.Start()
		defer bar.Finish()
		report.bar = bar
	}
	front := &url.URL{Scheme: "https", Host: cfg.frontHost}
	tasks := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < cfg.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				if ctx.Err() != nil {
					return
				}
				matched, err := check(ctx, task, front, client, cfg.needle)
				report.record(task, matched, err)
			}
		}()
	}
dispatch:
	for _, host := range hosts {
		select {
		case <-ctx.Done():
			break dispatch
		case tasks <- host:
		}
	}
	close(tasks)
	wg.Wait()
	return errors.Join(ctx.Err(), report.err())
}
