package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"
)

func readInputs(ctx context.Context, input io.Reader, single string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if single != "" {
		if err := validateHost(single); err != nil {
			return nil, err
		}
		return []string{single}, nil
	}
	var hosts []string
	scanner := bufio.NewScanner(input)
	line := 0
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line++
		host := strings.TrimSpace(scanner.Text())
		if host == "" {
			continue
		}
		if err := validateHost(host); err != nil {
			return nil, fmt.Errorf("input line %d: %w", line, err)
		}
		hosts = append(hosts, host)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read input: %w", err)
	}
	if len(hosts) == 0 {
		return nil, fmt.Errorf("no input hosts")
	}
	return hosts, nil
}
