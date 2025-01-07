package openai

import (
	"log/slog"
	"net/http"
	"strings"
)

// Client represents an OpenAI API client
type Client struct {
	logger     *slog.Logger
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// ClientOption allows configuring the client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimSuffix(url, "/")
	}
}

// New creates a new OpenAI client
func New(logger *slog.Logger, apiKey string, httpClient *http.Client, opts ...ClientOption) *Client {
	c := Client{
		logger:     logger.WithGroup("openai"),
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://api.openai.com/v1",
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}
