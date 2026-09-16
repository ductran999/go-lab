// Package postgrest implements domain.TodoRepository on top of the
// PostgREST REST API using only the standard library.
package postgrest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// ErrPostgrestRequest is returned when PostgREST answers with a non-2xx status.
var ErrPostgrestRequest = errors.New("postgrest request failed")

// Client is a minimal PostgREST HTTP client.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient builds a Client targeting the given PostgREST base URL,
// e.g. http://localhost:3000.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: http.DefaultClient,
	}
}

// Do sends one JSON request and returns the raw response body.
// A Prefer header value is attached when non-empty.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any, prefer string) ([]byte, error) {
	var r io.Reader

	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}

		r = bytes.NewReader(buf)
	}

	u := c.baseURL + path

	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	if prefer != "" {
		req.Header.Set("Prefer", prefer)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s -> %d: %s: %w",
			method, u, resp.StatusCode, string(data), ErrPostgrestRequest)
	}

	return data, nil
}
