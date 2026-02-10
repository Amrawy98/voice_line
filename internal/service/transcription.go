package service

import (
	"bytes"
	"context"

	openai "github.com/sashabaranov/go-openai"
)

const groqBaseURL = "https://api.groq.com/openai/v1"

type TranscriptionService struct {
	openaiClient *openai.Client
}

func NewTranscriptionService(apiKey string) *TranscriptionService {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = groqBaseURL

	return &TranscriptionService{
		openaiClient: openai.NewClientWithConfig(config),
	}
}

func (s *TranscriptionService) Transcribe(ctx context.Context, fileName string, data []byte) (string, error) {
	req := openai.AudioRequest{
		Model:    "whisper-large-v3-turbo",
		FilePath: fileName,
		Reader:   bytes.NewReader(data),
	}

	resp, err := s.openaiClient.CreateTranscription(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}
