package openaicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/wiselead-ai/httpclient"
)

func (c *Client) CreateVectorStore(ctx context.Context, in *CreateVectorStoreInput) (*VectorStore, error) {
	c.logger.Info("Creating vector store", slog.Any("input", in))

	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/vector_stores", bytes.NewBuffer(body))
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
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

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

		fmt.Printf("Vector store response: %+v\n", response)

		if response.Status == "completed" {
			fmt.Println("Vector store creation completed successfully.")
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

		fmt.Printf("Waiting for %v before retrying...\n", delay)
		time.Sleep(delay)
	}
}
