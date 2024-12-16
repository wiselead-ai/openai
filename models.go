package openaicli

import (
	"encoding/json"
)

type Model string

const (
	AssistantModel Model = "gpt-4o-mini"
	EmbeddingModel Model = "text-embedding-3-small"

	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"

	ToolTypeCodeInterpreter = "code_interpreter"
	ToolTypeFileSearch      = "file_search"

	RunStatusQueued         = "queued"
	RunStatusInProgress     = "in_progress"
	RunStatusCompleted      = "completed"
	RunStatusFailed         = "failed"
	RunStatusCancelling     = "cancelling"
	RunStatusCancelled      = "cancelled"
	RunStatusExpired        = "expired"
	RunStatusRequiresAction = "requires_action"
)

type (
	// Meta allows for arbitrary metadata to be attached to objects.
	Meta map[string]any

	// Assistant

	CreateAssistantInput struct {
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Model        Model    `json:"model"`
		Instructions string   `json:"instructions"`
		Tools        []Tool   `json:"tools"`
		FileIDs      []string `json:"file_ids,omitempty"`
		Metadata     Meta     `json:"metadata,omitempty"`
	}

	Assistant struct {
		ID           string   `json:"id"`
		Object       string   `json:"object"`
		CreatedAt    int64    `json:"created_at"`
		Name         string   `json:"name"`
		Description  string   `json:"description"`
		Model        Model    `json:"model"`
		Instructions string   `json:"instructions"`
		Tools        []Tool   `json:"tools"`
		FileIDs      []string `json:"file_ids"`
		Metadata     Meta     `json:"metadata,omitempty"`
	}

	AssistantFiles struct {
		Object string          `json:"object"`
		Data   []AssistantFile `json:"data"`
	}

	AssistantFile struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		CreatedAt int64  `json:"created_at"`
		FileID    string `json:"file_id"`
	}

	// Thread

	Thread struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		CreatedAt int    `json:"created_at"`
		Metadata  Meta   `json:"metadata"`
	}

	ThreadMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	CreateThreadMessageInput struct {
		ThreadID string
		Message  ThreadMessage
	}

	// Run

	Run struct {
		ID             string          `json:"id"`
		Object         string          `json:"object"`
		CreatedAt      int64           `json:"created_at"`
		ThreadID       string          `json:"thread_id"`
		AssistantID    string          `json:"assistant_id"`
		Status         string          `json:"status"`
		StartedAt      int64           `json:"started_at,omitempty"`
		ExpiresAt      int64           `json:"expires_at,omitempty"`
		CancelledAt    int64           `json:"cancelled_at,omitempty"`
		FailedAt       int64           `json:"failed_at,omitempty"`
		CompletedAt    int64           `json:"completed_at,omitempty"`
		LastError      *Error          `json:"last_error,omitempty"`
		Model          string          `json:"model"`
		Instructions   string          `json:"instructions,omitempty"`
		Tools          []Tool          `json:"tools"`
		FileIDs        []string        `json:"file_ids"`
		RequiredAction *RequiredAction `json:"required_action,omitempty"`
	}

	Content struct {
		Type string    `json:"type"`
		Text TextValue `json:"text"`
	}

	TextValue struct {
		Value       string       `json:"value"`
		Annotations []Annotation `json:"annotations,omitempty"`
		Citations   []Citation   `json:"citations,omitempty"`
	}

	Annotation struct {
		Type         string `json:"type"`
		Text         string `json:"text"`
		FileCitation *struct {
			FileID string `json:"file_id"`
			Quote  string `json:"quote"`
		} `json:"file_citation,omitempty"`
	}

	Citation struct {
		FileID string `json:"file_id"`
		Quote  string `json:"quote"`
	}

	RequiredAction struct {
		Type      string     `json:"type"`
		ToolCalls []ToolCall `json:"tool_calls"`
	}

	ToolCall struct {
		ID        string       `json:"id"`
		Type      string       `json:"type"`
		Function  FunctionCall `json:"function"`
		Arguments string       `json:"arguments"`
	}

	FunctionCall struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}

	ToolOutput struct {
		ToolCallID string `json:"tool_call_id"`
		Output     string `json:"output"`
	}

	ThreadMessageList struct {
		Object  string           `json:"object"`
		Data    []MessageContent `json:"data"`
		FirstID string           `json:"first_id"`
		LastID  string           `json:"last_id"`
	}

	MessageContent struct {
		ID        string    `json:"id"`
		Object    string    `json:"object"`
		CreatedAt int64     `json:"created_at"`
		ThreadID  string    `json:"thread_id"`
		Role      string    `json:"role"`
		Content   []Content `json:"content"`
	}

	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	}

	Tool struct {
		Type     string    `json:"type"`
		Function *Function `json:"function,omitempty"`
	}

	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	ListResponse struct {
		Object  string `json:"object"`
		Data    []any  `json:"data"`
		FirstID string `json:"first_id"`
		LastID  string `json:"last_id"`
		HasMore bool   `json:"has_more"`
	}

	FileUploadResponse struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		Purpose   string `json:"purpose"`
		CreatedAt int64  `json:"created_at"`
	}

	FileDetails struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		Purpose   string `json:"purpose"`
		CreatedAt int64  `json:"created_at"`
	}

	RunSteps struct {
		Object string    `json:"object"`
		Data   []RunStep `json:"data"`
	}

	RunStep struct {
		ID          string      `json:"id"`
		Object      string      `json:"object"`
		CreatedAt   int64       `json:"created_at"`
		RunID       string      `json:"run_id"`
		Status      string      `json:"status"`
		StepDetails *StepDetail `json:"step_details"`
	}

	StepDetail struct {
		Type      string     `json:"type"`
		ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	}
)
