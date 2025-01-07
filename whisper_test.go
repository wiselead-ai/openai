package openai

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClient_TranscribeAudio(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        TranscribeAudioInput
		serverResp   []byte
		serverStatus int
		expectError  bool
	}{
		{
			name: "successful transcription",
			input: TranscribeAudioInput{
				Name: "test.mp3",
				Data: bytes.NewReader([]byte("fake audio data")),
			},
			serverResp:   []byte(`{"text": "transcribed text"}`),
			serverStatus: http.StatusOK,
		},
		{
			name: "server error",
			input: TranscribeAudioInput{
				Name: "test.mp3",
				Data: bytes.NewReader([]byte("fake audio data")),
			},
			serverStatus: http.StatusInternalServerError,
			expectError:  true,
		},
		{
			name: "empty audio data",
			input: TranscribeAudioInput{
				Name: "empty.mp3",
				Data: bytes.NewReader([]byte{}),
			},
			serverStatus: http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, "/audio/transcriptions", r.URL.Path)
				require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
				require.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

				w.WriteHeader(tt.serverStatus)
				if tt.serverResp != nil {
					w.Write(tt.serverResp)
				}
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", server.Client(), WithBaseURL(server.URL))

			result, err := client.TranscribeAudio(tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
		})
	}
}
