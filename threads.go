package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/wiselead-ai/httpclient"
)

func (c *Client) CreateThread(ctx context.Context) (*Thread, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads", c.baseURL),
		bytes.NewBuffer([]byte("{}")),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := httpclient.DoWithRetry(c.httpClient, req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: '%d', response: '%s'", resp.StatusCode, string(b))
	}

	var thread Thread
	if err := json.NewDecoder(resp.Body).Decode(&thread); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &thread, nil
}

func (c *Client) AddMessage(ctx context.Context, in CreateMessageInput) error {
	jsonData, err := json.Marshal(in.Message)
	if err != nil {
		return fmt.Errorf("could not marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads/%s/messages", c.baseURL, in.ThreadID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(b), "Can't add messages to thread") {
			time.Sleep(5 * time.Second)
			resp, err = c.httpClient.Do(req)
			if err != nil {
				return fmt.Errorf("could not send request: %w", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("unexpected status code: '%d', response: '%s'", resp.StatusCode, string(b))
			}
		} else {
			return fmt.Errorf("unexpected status code: '%d', response: '%s'", resp.StatusCode, string(b))
		}
	}
	return nil
}

func (c *Client) GetMessages(ctx context.Context, threadID string) (*ThreadMessageList, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/threads/%s/messages", c.baseURL, threadID),
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

	var messages ThreadMessageList
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &messages, nil
}

func (c *Client) RunThread(ctx context.Context, threadID, assistantID string) (*Run, error) {
	jsonData, err := json.Marshal(struct {
		AssistantID string `json:"assistant_id"`
	}{
		AssistantID: assistantID,
	})
	if err != nil {
		return nil, fmt.Errorf("could not marshal run input: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads/%s/runs", c.baseURL, threadID),
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
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(responseBody))
	}

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &run, nil
}

// Add this new method to handle tool outputs
func (c *Client) SubmitToolOutputs(ctx context.Context, threadID string, runID string, outputs []ToolOutput) error {
	input := struct {
		ToolOutputs []ToolOutput `json:"tool_outputs"`
	}{
		ToolOutputs: outputs,
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("could not marshal tool outputs: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads/%s/runs/%s/submit_tool_outputs", c.baseURL, threadID, runID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := httpclient.DoWithRetry(c.httpClient, req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetRun(ctx context.Context, threadID, runID string) (*Run, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/threads/%s/runs/%s", c.baseURL, threadID, runID),
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

	var run Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &run, nil
}

func (c *Client) WaitForRun(ctx context.Context, threadID, runID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			run, err := c.GetRun(ctx, threadID, runID)
			if err != nil {
				return fmt.Errorf("failed to get run: %w", err)
			}

			switch run.Status {
			case RunStatusCompleted:
				return nil
			case RunStatusFailed:
				if run.LastError != nil {
					return fmt.Errorf("run failed: %s - %s", run.LastError.Code, run.LastError.Message)
				}
				return fmt.Errorf("run failed without error details")
			case RunStatusCancelled, RunStatusExpired:
				return fmt.Errorf("run ended with status: %s", run.Status)
			case RunStatusQueued, RunStatusInProgress:
				time.Sleep(time.Second)
				continue
			default:
				return fmt.Errorf("unknown run status: %s", run.Status)
			}
		}
	}
}
