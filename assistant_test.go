package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_CreateAssistant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          *CreateAssistantInput
		serverResponse *Assistant
		serverStatus   int
		expectedError  bool
	}{
		{
			name: "successful creation",
			input: &CreateAssistantInput{
				Model:        "gpt-4",
				Name:         "Test Assistant",
				Description:  "Test description",
				Instructions: "Test instructions",
				Tools:        []Tool{{Type: "code_interpreter"}},
				Metadata:     map[string]any{"key": "value"},
			},
			serverResponse: &Assistant{
				ID:           "asst_123",
				Object:       "assistant",
				CreatedAt:    1699009709,
				Name:         "Test Assistant",
				Description:  "Test description",
				Model:        "gpt-4",
				Instructions: "Test instructions",
				Tools:        []Tool{{Type: "code_interpreter"}},
				FileIDs:      []string{"file-123"},
				Metadata:     map[string]any{"key": "value"},
			},
			serverStatus: http.StatusOK,
		},
		{
			name: "server error",
			input: &CreateAssistantInput{
				Model: "gpt-4",
				Name:  "Test Assistant",
			},
			serverStatus:  http.StatusInternalServerError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/assistants", r.URL.Path)
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

			result, err := client.CreateAssistant(context.Background(), tt.input)
			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}

func TestClient_GetAssistant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		assistantID    string
		serverResponse *Assistant
		serverStatus   int
		expectedError  bool
	}{
		{
			name:        "successful retrieval",
			assistantID: "asst_123",
			serverResponse: &Assistant{
				ID:           "asst_123",
				Object:       "assistant",
				CreatedAt:    1699009709,
				Name:         "Test Assistant",
				Description:  "Test description",
				Model:        "gpt-4",
				Instructions: "Test instructions",
				Tools:        []Tool{{Type: "code_interpreter"}},
				FileIDs:      []string{"file-123"},
				Metadata:     map[string]any{"key": "value"},
			},
			serverStatus: http.StatusOK,
		},
		{
			name:          "not found",
			assistantID:   "asst_nonexistent",
			serverStatus:  http.StatusNotFound,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/assistants/"+tt.assistantID, r.URL.Path)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

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

			result, err := client.GetAssistant(context.Background(), tt.assistantID)
			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}

func TestClient_ModifyAssistant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		assistantID    string
		input          *ModifyAssistantInput
		serverResponse *Assistant
		serverStatus   int
		expectedError  bool
	}{
		{
			name:        "successful modification",
			assistantID: "asst_123",
			input: &ModifyAssistantInput{
				Description:  "Updated description",
				Instructions: "Updated instructions",
				Tools:        []Tool{{Type: "code_interpreter"}},
				Metadata:     map[string]any{"updated": "value"},
			},
			serverResponse: &Assistant{
				ID:           "asst_123",
				Object:       "assistant",
				CreatedAt:    1699009709,
				Name:         "Updated Assistant",
				Description:  "Updated description",
				Model:        "gpt-4",
				Instructions: "Updated instructions",
				Tools:        []Tool{{Type: "code_interpreter"}},
				FileIDs:      []string{"file-456"},
				Metadata:     map[string]any{"updated": "value"},
			},
			serverStatus: http.StatusOK,
		},
		{
			name:        "invalid modification",
			assistantID: "asst_123",
			input: &ModifyAssistantInput{
				Description: "Invalid",
			},
			serverStatus:  http.StatusBadRequest,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/assistants/"+tt.assistantID, r.URL.Path)
				require.Equal(t, http.MethodPost, r.Method)
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

			result, err := client.ModifyAssistant(context.Background(), tt.assistantID, tt.input)
			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}
