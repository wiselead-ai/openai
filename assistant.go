package openaicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) CreateAssistant(ctx context.Context, cfg CreateAssistantInput) (*Assistant, error) {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not marshal assistant config: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/assistants",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	var assistant Assistant
	if err := json.NewDecoder(resp.Body).Decode(&assistant); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &assistant, nil
}

func (c *Client) AddFilesToAssistant(ctx context.Context, assistantID string, fileIDs []string) error {
	for _, fileID := range fileIDs {
		if err := c.AttachFileToAssistant(ctx, assistantID, fileID); err != nil {
			return fmt.Errorf("failed to attach file %s: %w", fileID, err)
		}
	}
	return nil
}

func (c *Client) AttachFileToAssistant(ctx context.Context, assistantID, fileID string) error {
	input := struct {
		FileID string `json:"file_id"`
	}{FileID: fileID}

	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("could not marshal file attachment: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/assistants/%s/files", c.baseURL, assistantID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to attach file (status %d): %s", resp.StatusCode, body)
	}
	return nil
}

func (c *Client) GetAssistantFiles(ctx context.Context, assistantID string) (*AssistantFiles, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/assistants/%s/files", c.baseURL, assistantID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var files AssistantFiles
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &files, nil
}
