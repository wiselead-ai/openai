# OpenAI Client

A comprehensive Go client for OpenAI's API services, with focus on the Assistants API (Beta v2).

## Features

### Assistants API (Beta v2)

- Create and manage assistants
- Thread management and messaging
- Run execution and monitoring
- Tool outputs submission
- Run steps tracking

### File Management

- Upload files
- List available files
- Retrieve file content

### Audio Services

- Audio transcription (Whisper AI)

### Vector Store Operations

- Create vector stores
- Monitor store creation progress

## Usage Examples

```go
// Initialize client
client := openai.New(logger, "your-api-key", httpClient)

// Create an assistant
assistant, err := client.CreateAssistant(ctx, &openai.CreateAssistantInput{
    Name: "Research Assistant",
    Model: "gpt-4-1106-preview",
    Tools: []Tool{
        {Type: "code_interpreter"},
    },
})

// Create a thread
thread, err := client.CreateThread(ctx)

// Add a message
err = client.AddMessage(ctx, CreateMessageInput{
    ThreadID: thread.ID,
    Message: Message{
        Role: "user",
        Content: "Hello!",
    },
})

// Run the thread
run, err := client.RunThread(ctx, thread.ID, assistant.ID)

// Wait for completion
err = client.WaitForRun(ctx, thread.ID, run.ID)
```

## API Reference

This implementation follows the OpenAI API specifications:

- Assistants API: Beta v2 (`OpenAI-Beta: assistants=v2`)
- Files API: v1
- Audio API: v1

For detailed API documentation, visit:
[OpenAI API Reference for Assistants](https://platform.openai.com/docs/api-reference/assistants)

## Configuration

The client can be configured with options:

```go
client := openai.New(
    logger,
    apiKey,
    httpClient,
    openai.WithBaseURL("https://custom-url.com"),
)
```
