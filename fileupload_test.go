package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_ListFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		serverResponse *ListResponse
		serverStatus   int
		expectError    bool
	}{
		{
			name: "success",
			serverResponse: &ListResponse{
				Object: "list",
				Data:   []any{map[string]any{"id": "file-123"}},
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
				require.Equal(t, "/files", r.URL.Path)
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

			resp, err := client.ListFiles(context.Background())
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, resp)
		})
	}
}

func TestClient_UploadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		purpose        string
		data           []byte
		serverResponse *FileUploadResponse
		serverStatus   int
		expectError    bool
	}{
		{
			name:    "success",
			purpose: "fine-tune",
			data:    []byte(`test file content`),
			serverResponse: &FileUploadResponse{
				ID:     "file-123",
				Object: "file",
			},
			serverStatus: http.StatusOK,
		},
		{
			name:         "bad request",
			purpose:      "fine-tune",
			data:         []byte(`test file content`),
			serverStatus: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:         "empty file",
			purpose:      "fine-tune",
			data:         []byte{},
			serverStatus: http.StatusBadRequest,
			expectError:  true,
		},
		{
			name:    "large file",
			purpose: "fine-tune",
			data:    bytes.Repeat([]byte("x"), 1024*1024), // 1MB file
			serverResponse: &FileUploadResponse{
				ID:     "file-large",
				Object: "file",
			},
			serverStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/files", r.URL.Path)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

				err := r.ParseMultipartForm(32 << 20)
				require.NoError(t, err)

				_, _, err = r.FormFile("file")
				require.NoError(t, err)

				require.Equal(t, tt.purpose, r.FormValue("purpose"))

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

			resp, err := client.UploadFile(context.Background(), bytes.NewReader(tt.data), tt.purpose)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.serverResponse, resp)
		})
	}
}

func TestClient_GetFileContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fileID      string
		fileDetails *FileDetails
		body        []byte
		status1     int
		status2     int
		expectError bool
	}{
		{
			name:        "success",
			fileID:      "file-123",
			fileDetails: &FileDetails{Purpose: "finetuning"},
			body:        []byte("file content"),
			status1:     http.StatusOK,
			status2:     http.StatusOK,
		},
		{
			name:        "invalid file",
			fileID:      "file-999",
			fileDetails: &FileDetails{Purpose: "assistants"},
			status1:     http.StatusOK,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch callCount {
				case 0:
					require.Equal(t, "/files/"+tt.fileID, r.URL.Path)
					w.WriteHeader(tt.status1)
					if tt.fileDetails != nil {
						json.NewEncoder(w).Encode(tt.fileDetails)
					}
				default:
					require.Equal(t, fmt.Sprintf("/files/%s/content", tt.fileID), r.URL.Path)
					w.WriteHeader(tt.status2)
					if len(tt.body) > 0 {
						w.Write(tt.body)
					}
				}
				callCount++
			}))
			defer server.Close()

			client := &Client{
				httpClient: server.Client(),
				baseURL:    server.URL,
				apiKey:     "test-key",
			}

			data, err := client.GetFileContent(context.Background(), tt.fileID)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.body, data)
		})
	}
}
