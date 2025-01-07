package openai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_CreateVectorStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          *CreateVectorStoreInput
		serverResponse *VectorStore
		serverStatus   int
		expectedError  bool
	}{
		{
			name: "successful creation",
			input: &CreateVectorStoreInput{
				Name:        "Test Store",
				Description: "Test description",
				FileIDs:     []string{"file-123"},
				Metadata:    map[string]any{"key": "value"},
			},
			serverResponse: &VectorStore{
				ID:           "vec_123",
				Object:       "vector_store",
				Name:         "Test Store",
				Description:  "Test description",
				Status:       "active",
				CreatedAt:    1699009709,
				LastActiveAt: 1699009709,
				Metadata:     map[string]any{"key": "value"},
			},
			serverStatus: http.StatusOK,
		},
		{
			name: "invalid request",
			input: &CreateVectorStoreInput{
				Name: "Invalid Store",
			},
			serverStatus:  http.StatusBadRequest,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/vector_stores", r.URL.Path) // Fixed: changed from vector-stores to vector_stores
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

				w.WriteHeader(tt.serverStatus)
				if tt.serverResponse != nil {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			result, err := client.CreateVectorStore(context.Background(), tt.input)
			if tt.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, result)
		})
	}
}
