// Package postgrest implements domain.TodoRepository on top of the
// PostgREST REST API via the shared httpclient package.
package postgrest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/ductran999/shared-pkg/transport/httpclient/v2"
)

// ErrPostgrestRequest is returned when PostgREST answers with a non-2xx status.
var ErrPostgrestRequest = errors.New("postgrest request failed")

// Client is a minimal PostgREST HTTP client.
type Client struct {
	baseURL    string
	httpClient *httpclient.Client
}

// NewClient builds a Client targeting the given PostgREST base URL,
// e.g. http://localhost:3000. A non-empty token is sent as a bearer
// Authorization header on every request (empty = anonymous).
func NewClient(baseURL, token string) *Client {
	clientOpts := []httpclient.Option{}

	if token != "" {
		clientOpts = append(clientOpts,
			httpclient.WithDefaultHeader("Authorization", "Bearer "+token))
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpclient.New(clientOpts...),
	}
}

// Do sends one JSON request and returns the raw response body.
// A Prefer header value is attached when non-empty.
func (c *Client) Do(
	ctx context.Context,
	method, path string,
	query url.Values,
	body any,
	prefer string,
) ([]byte, error) {
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

	opts := []httpclient.RequestOption{
		httpclient.WithContentType("application/json"),
	}

	if prefer != "" {
		opts = append(opts, httpclient.WithHeader("Prefer", prefer))
	}

	resp, err := c.httpClient.Do(ctx, method, u, r, opts...)
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
