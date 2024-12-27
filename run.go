package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wiselead-ai/httpclient"
)

func (c *Client) GetRunSteps(ctx context.Context, threadID, runID string) (*RunSteps, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/threads/%s/runs/%s/steps", c.baseURL, threadID, runID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := httpclient.DoWithRetry(c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	var steps RunSteps
	if err := json.NewDecoder(resp.Body).Decode(&steps); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &steps, nil
}
