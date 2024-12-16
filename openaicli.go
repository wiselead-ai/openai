package openaicli

import (
	"net/http"
)

// Client represents an OpenAI API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// ClientOption allows configuring the client.
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// New creates a new OpenAI client
func New(apiKey string, httpClient *http.Client, opts ...ClientOption) *Client {
	c := Client{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://api.openai.com/v1",
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}
