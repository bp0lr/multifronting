package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func check(ctx context.Context, remoteServer string, frontingServer *url.URL, client *http.Client, needle string) (bool, error) {

	var urlLocal string
	var req *http.Request
	var err error

	form := url.Values{}
	form.Add("op", "d3bug")

	urlLocal = "https://" + remoteServer

	req, err = http.NewRequestWithContext(ctx, "POST", urlLocal, strings.NewReader(form.Encode()))
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}
	req.Host = frontingServer.Host
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	body, err := readResponse(resp.Body)
	if err != nil {
		return false, err
	}

	if bytes.Contains(body, []byte(needle)) {
		return true, nil
	}

	return false, nil
}
