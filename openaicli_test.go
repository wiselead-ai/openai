package openai

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		apiKey    string
		opts      []ClientOption
		wantURL   string
		wantError bool
	}{
		{
			name:    "default client creation",
			apiKey:  "test-key",
			wantURL: "https://api.openai.com/v1",
		},
		{
			name:    "client with custom base URL",
			apiKey:  "test-key",
			opts:    []ClientOption{WithBaseURL("https://custom.api.com")},
			wantURL: "https://custom.api.com",
		},
		{
			name:    "multiple options",
			apiKey:  "test-key",
			opts:    []ClientOption{WithBaseURL("https://custom.api.com")},
			wantURL: "https://custom.api.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, tt.apiKey, http.DefaultClient, tt.opts...)

			require.NotNil(t, client)
			require.Equal(t, tt.apiKey, client.apiKey)
			require.Equal(t, tt.wantURL, client.baseURL)
			require.NotNil(t, client.logger)
			require.NotNil(t, client.httpClient)
		})
	}
}

func TestWithBaseURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "valid URL",
			url:  "https://api.custom.com",
			want: "https://api.custom.com",
		},
		{
			name: "URL with trailing slash",
			url:  "https://api.custom.com/",
			want: "https://api.custom.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			client := New(logger, "test-key", http.DefaultClient, WithBaseURL(tt.url))

			require.Equal(t, tt.want, client.baseURL)
		})
	}
}
