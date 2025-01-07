package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestClient_CreateThread(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse *Thread
		serverStatus   int
		expectError    bool
	}{
		{
			name: "successful creation",
			serverResponse: &Thread{
				ID:        "thread_123",
				Object:    "thread",
				CreatedAt: 1699009709,
				Metadata:  map[string]any{"key": "value"},
			},
			serverStatus: http.StatusOK,
		},
		{
			name:         "server error",
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads", r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			result, err := client.CreateThread(context.Background())
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}

func TestClient_AddMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        CreateMessageInput
		serverStatus int
		retryCase    bool
		expectError  bool
		responses    []int // Multiple response codes for retry scenarios
	}{
		{
			name: "successful message addition",
			input: CreateMessageInput{
				ThreadID: "thread_123",
				Message: ThreadMessage{
					Role:    RoleUser,
					Content: "Hello",
				},
			},
			serverStatus: http.StatusOK,
		},
		{
			name: "retry successful",
			input: CreateMessageInput{
				ThreadID: "thread_123",
				Message: ThreadMessage{
					Role:    RoleUser,
					Content: "Retry me",
				},
			},
			retryCase:    true,
			serverStatus: http.StatusOK,
		},
		{
			name: "retry exhausted",
			input: CreateMessageInput{
				ThreadID: "thread_123",
				Message:  ThreadMessage{Role: RoleUser, Content: "test"},
			},
			responses:   []int{http.StatusBadRequest, http.StatusBadRequest, http.StatusBadRequest},
			expectError: true,
		},
		{
			name: "invalid json",
			input: CreateMessageInput{
				ThreadID: "thread_123",
				Message:  ThreadMessage{Role: "invalid", Content: string([]byte{0x7f})},
			},
			serverStatus: http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads/"+tt.input.ThreadID+"/messages", r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))
				require.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var message ThreadMessage
				err := json.NewDecoder(r.Body).Decode(&message)
				require.NoError(t, err)
				require.Equal(t, tt.input.Message, message)

				if tt.retryCase && callCount == 0 {
					callCount++
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]interface{}{
						"error": "Can't add messages to thread",
					})
					return
				}

				if len(tt.responses) > 0 {
					w.WriteHeader(tt.responses[callCount])
					callCount++
					return
				}

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			err := client.AddMessage(context.Background(), tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.retryCase {
				require.Greater(t, callCount, 0, "retry should have occurred")
			}
		})
	}
}

func TestClient_WaitForRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		threadID    string
		runID       string
		responses   []Run
		expectError bool
	}{
		{
			name:     "successful completion",
			threadID: "thread_123",
			runID:    "run_456",
			responses: []Run{
				{Status: RunStatusQueued},
				{Status: RunStatusInProgress},
				{Status: RunStatusCompleted},
			},
		},
		{
			name:     "run fails",
			threadID: "thread_123",
			runID:    "run_456",
			responses: []Run{
				{Status: RunStatusQueued},
				{Status: RunStatusInProgress},
				{Status: RunStatusFailed, LastError: &RunError{Code: "error", Message: "Failed"}},
			},
			expectError: true,
		},
		{
			name:     "run cancelled",
			threadID: "thread_123",
			runID:    "run_456",
			responses: []Run{
				{Status: RunStatusQueued},
				{Status: RunStatusCancelled},
			},
			expectError: true,
		},
		{
			name:     "requires action then completes",
			threadID: "thread_123",
			runID:    "run_456",
			responses: []Run{
				{Status: RunStatusRequiresAction},
				{Status: RunStatusInProgress},
				{Status: RunStatusCompleted},
			},
		},
		{
			name:     "expires after queued",
			threadID: "thread_123",
			runID:    "run_456",
			responses: []Run{
				{Status: RunStatusQueued},
				{Status: RunStatusExpired},
			},
			expectError: true,
		},
		{
			name:     "unknown status",
			threadID: "thread_123",
			runID:    "run_456",
			responses: []Run{
				{Status: "unknown_status"},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads/"+tt.threadID+"/runs/"+tt.runID, r.URL.Path)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				w.WriteHeader(http.StatusOK)
				if callCount < len(tt.responses) {
					json.NewEncoder(w).Encode(tt.responses[callCount])
					callCount++
				} else {
					// Return last response for any additional calls
					json.NewEncoder(w).Encode(tt.responses[len(tt.responses)-1])
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			// Use a longer timeout for tests
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := client.WaitForRun(ctx, tt.threadID, tt.runID)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestClient_GetMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		threadID       string
		serverResponse *ThreadMessageList
		serverStatus   int
		expectError    bool
	}{
		{
			name:     "successful retrieval",
			threadID: "thread_123",
			serverResponse: &ThreadMessageList{
				Object: "list",
				Data: []MessageContent{
					{
						ID:        "msg_123",
						Object:    "message",
						CreatedAt: 1699009709,
						ThreadID:  "thread_123",
						Role:      RoleUser,
						Content: []Content{
							{
								Type: "text",
								Text: TextValue{Value: "test message"},
							},
						},
					},
				},
			},
			serverStatus: http.StatusOK,
		},
		{
			name:         "not found",
			threadID:     "thread_nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads/"+tt.threadID+"/messages", r.URL.Path)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			result, err := client.GetMessages(context.Background(), tt.threadID)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}

func TestClient_RunThread(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		threadID       string
		assistantID    string
		serverResponse *Run
		serverStatus   int
		expectError    bool
	}{
		{
			name:        "successful run",
			threadID:    "thread_123",
			assistantID: "asst_123",
			serverResponse: &Run{
				ID:          "run_123",
				Object:      "run",
				ThreadID:    "thread_123",
				AssistantID: "asst_123",
				Status:      RunStatusQueued,
			},
			serverStatus: http.StatusOK,
		},
		{
			name:         "invalid thread",
			threadID:     "thread_invalid",
			assistantID:  "asst_123",
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads/"+tt.threadID+"/runs", r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				var input struct {
					AssistantID string `json:"assistant_id"`
				}
				err := json.NewDecoder(r.Body).Decode(&input)
				require.NoError(t, err)
				require.Equal(t, tt.assistantID, input.AssistantID)

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			result, err := client.RunThread(context.Background(), tt.threadID, tt.assistantID)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}

func TestClient_SubmitToolOutputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		threadID     string
		runID        string
		outputs      []ToolOutput
		serverStatus int
		expectError  bool
	}{
		{
			name:     "successful submission",
			threadID: "thread_123",
			runID:    "run_123",
			outputs: []ToolOutput{
				{
					ToolCallID: "call_123",
					Output:     "function result",
				},
			},
			serverStatus: http.StatusOK,
		},
		{
			name:         "invalid run",
			threadID:     "thread_123",
			runID:        "run_invalid",
			outputs:      []ToolOutput{},
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, fmt.Sprintf("/threads/%s/runs/%s/submit_tool_outputs", tt.threadID, tt.runID), r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				var input struct {
					ToolOutputs []ToolOutput `json:"tool_outputs"`
				}
				err := json.NewDecoder(r.Body).Decode(&input)
				require.NoError(t, err)
				require.Equal(t, tt.outputs, input.ToolOutputs)

				w.WriteHeader(tt.serverStatus)
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			err := client.SubmitToolOutputs(context.Background(), tt.threadID, tt.runID, tt.outputs)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestClient_GetRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		threadID       string
		runID          string
		serverResponse *Run
		serverStatus   int
		expectError    bool
	}{
		{
			name:     "successful retrieval",
			threadID: "thread_123",
			runID:    "run_123",
			serverResponse: &Run{
				ID:          "run_123",
				Object:      "run",
				ThreadID:    "thread_123",
				AssistantID: "asst_123",
				Status:      RunStatusCompleted,
			},
			serverStatus: http.StatusOK,
		},
		{
			name:         "not found",
			threadID:     "thread_123",
			runID:        "run_nonexistent",
			serverStatus: http.StatusNotFound,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads/"+tt.threadID+"/runs/"+tt.runID, r.URL.Path)
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			result, err := client.GetRun(context.Background(), tt.threadID, tt.runID)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}
