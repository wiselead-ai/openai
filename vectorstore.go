package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/wiselead-ai/httpclient"
)

func (c *Client) CreateVectorStore(ctx context.Context, in *CreateVectorStoreInput) (*VectorStore, error) {
	if in == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	if in.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if len(in.FileIDs) == 0 {
		return nil, fmt.Errorf("fileIDs is required")
	}

	// Validate file types before creating vector store
	for _, fileID := range in.FileIDs {
		fileInfo, err := c.GetFileMetadata(ctx, fileID)
		if err != nil {
			return nil, fmt.Errorf("failed to get file metadata for %s: %w", fileID, err)
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileInfo.Filename), "."))
		if !supportedFileTypes[ext] {
			return nil, fmt.Errorf(
				"file %s has unsupported extension '.%s'. Supported types: .pdf, .txt, .json, .md",
				fileInfo.Filename, ext,
			)
		}
	}

	c.logger.Info("Creating vector store",
		slog.String("name", in.Name),
		slog.Any("fileIDs", in.FileIDs))

	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/vector_stores", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := httpclient.DoWithRetry(c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create vector store, status: '%s', body: '%s'", resp.Status, b)
	}

	var out VectorStore
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

func (c *Client) WaitForVectorStoreCompletion(ctx context.Context, vectorStoreID string, timeout, maxDelay time.Duration) error {
	startTime := time.Now()
	delay := 1 * time.Second // initial delay for exponential backoff

	for {
		c.logger.Info("Checking vector store status", slog.String("vectorStoreID", vectorStoreID))

		req, err := http.NewRequest(http.MethodGet, c.baseURL+"/vector_stores/"+vectorStoreID, nil)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("OpenAI-Beta", "assistants=v2")

		resp, err := httpclient.DoWithRetry(c.httpClient, req)
		if err != nil {
			return fmt.Errorf("failed to send HTTP request: %w", err)
		}
		defer resp.Body.Close()

		var response VectorStore
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		c.logger.Info("Vector store response", slog.Any("response", response))

		if response.Status == "completed" {
			c.logger.Info("Vector store creation completed successfully")
			return nil
		}

		if response.Status == "failed" {
			return fmt.Errorf("vector store creation failed")
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout reached while waiting for vector store completion")
		}

		if delay < maxDelay {
			delay *= 2 // Double the delay for the next attempt
		}
		c.logger.Info("Waiting for delay before retrying", slog.Any("delay", delay))
		time.Sleep(delay)
	}
}

// Add new helper method to get file metadata
func (c *Client) GetFileMetadata(ctx context.Context, fileID string) (*FileDetails, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/files/%s", c.baseURL, fileID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}
	defer resp.Body.Close()

	var fileInfo FileDetails
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, fmt.Errorf("failed to decode file metadata: %w", err)
	}

	return &fileInfo, nil
}
