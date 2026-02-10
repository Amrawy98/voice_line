package transcription

import (
	"bytes"
	"context"

	openai "github.com/sashabaranov/go-openai"
)

const groqBaseURL = "https://api.groq.com/openai/v1"

type Client struct {
	openaiClient *openai.Client
}

func NewClient(apiKey string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = groqBaseURL

	return &Client{
		openaiClient: openai.NewClientWithConfig(config),
	}
}

func (c *Client) Transcribe(ctx context.Context, fileName string, data []byte) (string, error) {
	req := openai.AudioRequest{
		Model:    "whisper-large-v3-turbo",
		FilePath: fileName,
		Reader:   bytes.NewReader(data),
	}

	resp, err := c.openaiClient.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}