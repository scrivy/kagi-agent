package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// do returns the http body, status code, and error
func (c *Client) do(ctx context.Context, method, path string, input, output any) ([]byte, int, error) {
	var reqBody io.Reader

	if input != nil {
		payload, err := json.Marshal(input)
		if err != nil {
			return nil, 0, fmt.Errorf("json.Marshal(): %w", err)
		}

		c.logger.Debugf("%s %s request payload:\n%s", method, path, payload)

		reqBody = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		method,
		baseUri+path,
		reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("http.NewRequest(): %w", err)
	}

	if input != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("c.c.Do(): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("io.ReadAll(): %w", err)
	}

	c.logger.Debugf("%s %s %d response body:\n%s", method, path, resp.StatusCode, body)

	if err := json.Unmarshal(body, output); err != nil {
		return body, resp.StatusCode, fmt.Errorf("json.Unmarshal(): %w", err)
	}

	c.logger.Debugf("%s %s struct:\n%+v", method, path, output)

	return body, resp.StatusCode, nil
}
