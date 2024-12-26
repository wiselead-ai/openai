package openaicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/wiselead-ai/httpclient"
)

func (c *Client) CreateAssistant(ctx context.Context, in *CreateAssistantInput) (*Assistant, error) {
	jsonData, err := json.Marshal(in)
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

	resp, err := httpclient.DoWithRetry(c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code '%d', response: '%s'", resp.StatusCode, string(b))
	}

	var assistant Assistant
	if err := json.NewDecoder(resp.Body).Decode(&assistant); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &assistant, nil
}

func (c *Client) GetAssistant(ctx context.Context, assistantID string) (*Assistant, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.baseURL+"/assistants/"+assistantID,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := httpclient.DoWithRetry(c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code '%d', response: '%s'", resp.StatusCode, string(b))
	}

	var assistant Assistant
	if err := json.NewDecoder(resp.Body).Decode(&assistant); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &assistant, nil
}
