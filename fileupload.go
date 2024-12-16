package openaicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

func (c *Client) UploadFile(ctx context.Context, data io.Reader, purpose string) (*FileUploadResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filename := fmt.Sprintf("data_%d.json", time.Now().UnixNano())

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %w", err)
	}

	if _, err := io.Copy(part, data); err != nil {
		return nil, fmt.Errorf("error copying data to form file: %w", err)
	}

	if err := writer.WriteField("purpose", purpose); err != nil {
		return nil, fmt.Errorf("error writing purpose field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/files", &body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, responseBody)
	}

	var uploadResp FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &uploadResp, nil
}

func (c *Client) ListFiles(ctx context.Context) (*ListResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/files", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, body)
	}

	var fileList ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileList); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &fileList, nil
}

func (c *Client) GetFileContent(ctx context.Context, fileID string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/files/%s", c.baseURL, fileID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file metadata: %w", err)
	}
	defer resp.Body.Close()

	var fileInfo FileDetails
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, fmt.Errorf("error decoding file metadata: %w", err)
	}

	if fileInfo.Purpose == "assistants" {
		return nil, fmt.Errorf("cannot download files with purpose: assistants")
	}

	contentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/files/%s/content", c.baseURL, fileID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating content request: %w", err)
	}

	contentReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	contentResp, err := c.doWithRetry(contentReq)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file content: %w", err)
	}
	defer contentResp.Body.Close()

	content, err := io.ReadAll(contentResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return content, nil
}
