package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"

	"github.com/scrivy/kagi-agent/internal/client/models"
)

func (c *Client) PostUpload(ctx context.Context, filePath string) (*models.PostUploadResponse, error) {
	header := make(textproto.MIMEHeader)
	header.Add(
		"Content-Disposition",
		multipart.FileContentDisposition(
			"file",
			filePath))

	reqBody := new(bytes.Buffer)
	writer := multipart.NewWriter(reqBody)

	part, err := writer.CreatePart(header)
	if err != nil {
		return nil, fmt.Errorf("writer.CreatePart(): %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("os.Open(filePath): %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("io.Copy(): %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("writer.Close(): %w", err)
	}

	method := http.MethodPost
	path := "/api/upload"
	req, err := http.NewRequestWithContext(
		ctx,
		method,
		baseUri+path,
		reqBody)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest(): %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	c.logger.Debugf("http request: %+v\n", req)

	resp, err := c.c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("c.c.Do(): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll(): %w", err)
	}

	c.logger.Debugf("%s %s response body:\n%s", method, path, body)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("http status: %s, headers: %+v", resp.Status, resp.Header))
	}

	var output models.PostUploadResponse
	err = json.Unmarshal(body, &output)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal(): %w", err)
	}

	c.logger.Debugf("%s %s struct:\n%+v", method, path, output)

	return &output, nil
}
