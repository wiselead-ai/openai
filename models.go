package openaicli

import "io"

const (
	AssistantModel Model = "gpt-4o-mini"

	RoleUser = "user"

	RunStatusQueued         = "queued"
	RunStatusInProgress     = "in_progress"
	RunStatusCompleted      = "completed"
	RunStatusFailed         = "failed"
	RunStatusCancelling     = "cancelling"
	RunStatusCancelled      = "cancelled"
	RunStatusExpired        = "expired"
	RunStatusRequiresAction = "requires_action"
	RunStatusPending        = "pending"

	// Tool types
	ToolTypeFunction        = "function"
	ToolTypeCodeInterpreter = "code_interpreter"
	ToolTypeFileSearch      = "file_search"
)

type (
	Model string

	// Common types
	Meta map[string]any

	// Assistant

	CreateAssistantInput struct {
		Metadata      Meta          `json:"metadata,omitempty"`
		Name          string        `json:"name"`
		Description   string        `json:"description"`
		Model         Model         `json:"model"`
		Instructions  string        `json:"instructions"`
		Tools         []Tool        `json:"tools"`
		ToolResources ToolResources `json:"tool_resources,omitempty"`
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

	Tool struct {
		Type     string              `json:"type"`
		Function *FunctionDefinition `json:"function,omitempty"`
	}

	ToolResources struct {
		CodeInterpreter *CodeInterpreter `json:"code_interpreter,omitempty"`
		FileSearch      *FileSearch      `json:"file_search,omitempty"`
	}

	CodeInterpreter struct {
		FileIDs []string `json:"file_ids"`
	}

	FileSearch struct {
		VectorStoreIDs []string `json:"vector_store_ids"`
	}

	FunctionDefinition struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}

	// Vector Store

	CreateVectorStoreInput struct {
		Name        string   `json:"name"`
		Description string   `json:"description,omitempty"`
		FileIDs     []string `json:"file_ids"`
	}

	VectorStore struct {
		ID         string `json:"id"`
		Object     string `json:"object"`
		Name       string `json:"name"`
		Status     string `json:"status"`
		UsageBytes int    `json:"usage_bytes"`
		CreatedAt  int64  `json:"created_at"`
		FileCounts struct {
			InProgress int `json:"in_progress"`
			Completed  int `json:"completed"`
			Failed     int `json:"failed"`
			Cancelled  int `json:"cancelled"`
			Total      int `json:"total"`
		} `json:"file_counts"`
		Metadata     map[string]interface{} `json:"metadata"`
		ExpiresAfter interface{}            `json:"expires_after"`
		ExpiresAt    interface{}            `json:"expires_at"`
		LastActiveAt int64                  `json:"last_active_at"`
	}

	// WhisperAI

	TranscribeAudioInput struct {
		Name string
		Data io.Reader
	}

	// Yet to organize the below types

	CreateMessageInput struct {
		ThreadID string
		Message  ThreadMessage
	}

	ThreadMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
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

	Thread struct {
		ID        string `json:"id"`
		Object    string `json:"object"`
		CreatedAt int    `json:"created_at"`
		Metadata  Meta   `json:"metadata"`
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

	ToolOutput struct {
		ToolCallID string `json:"tool_call_id"`
		Output     string `json:"output"`
	}

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
		LastError      *RunError       `json:"last_error,omitempty"`
		Model          string          `json:"model"`
		Instructions   string          `json:"instructions,omitempty"`
		Tools          []Tool          `json:"tools"`
		FileIDs        []string        `json:"file_ids"`
		RequiredAction *RequiredAction `json:"required_action,omitempty"`
	}

	RunError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}

	RequiredAction struct {
		Type      string     `json:"type"`
		ToolCalls []ToolCall `json:"tool_calls"`
	}
)
