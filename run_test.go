package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_GetRunSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		threadID       string
		runID          string
		serverResponse *RunSteps
		serverStatus   int
		expectedError  bool
	}{
		{
			name:     "successful retrieval",
			threadID: "thread_123",
			runID:    "run_456",
			serverResponse: &RunSteps{
				Object: "list",
				Data: []RunStep{
					{
						ID:        "step_789",
						Object:    "thread.run.step",
						CreatedAt: 1699009709,
						RunID:     "run_456",
						Status:    RunStatusCompleted,
						StepDetails: &StepDetail{
							Type: "message_creation",
							ToolCalls: []ToolCall{
								{
									ID:   "call_abc",
									Type: "function",
									Function: FunctionCall{
										Name:      "test_function",
										Arguments: `{"key": "value"}`,
									},
								},
							},
						},
					},
				},
			},
			serverStatus: http.StatusOK,
		},
		{
			name:          "not found",
			threadID:      "thread_invalid",
			runID:         "run_invalid",
			serverStatus:  http.StatusNotFound,
			expectedError: true,
		},
		{
			name:          "server error",
			threadID:      "thread_123",
			runID:         "run_456",
			serverStatus:  http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/threads/"+tt.threadID+"/runs/"+tt.runID+"/steps", r.URL.Path)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Equal(t, "assistants=v2", r.Header.Get("OpenAI-Beta"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			client := &Client{
				httpClient: server.Client(),
				baseURL:    server.URL,
				apiKey:     "test-key",
			}

			result, err := client.GetRunSteps(context.Background(), tt.threadID, tt.runID)
			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}
