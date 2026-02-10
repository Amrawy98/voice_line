package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const systemPrompt = `You are a sales call debrief analyst. You receive transcripts of salespeople describing recent customer interactions.

Your task is to extract structured information from the transcript and return it as JSON.

You MUST return a JSON object with exactly these fields:

{
  "deal_outlook": one of "moving_forward", "at_risk", "stalled",
  "customer_sentiment": one of "positive", "neutral", "negative", "mixed",
  "summary": a 1-3 sentence summary of the meeting,
  "positive_signals": an array of specific positive indicators mentioned (empty array if none),
  "negative_signals": an array of specific concerns or risks mentioned (empty array if none),
  "next_steps": an array of concrete follow-up actions mentioned (empty array if none),
  "deal_details": {
    "company": the customer company name if mentioned (empty string if not),
    "contact": the customer contact name if mentioned (empty string if not),
    "product": the product or service discussed if mentioned (empty string if not)
  }
}

Rules:
- Extract only what is explicitly stated or clearly implied. Do not fabricate.
- deal_outlook must be exactly one of: "moving_forward", "at_risk", "stalled"
- customer_sentiment must be exactly one of: "positive", "neutral", "negative", "mixed"
- Arrays must always be arrays, never null. Use empty array [] if nothing to report.
- deal_details fields must always be strings, never null. Use empty string "" if not mentioned.
- Return ONLY the JSON object, no other text.`

type AnalysisResult struct {
	DealOutlook       string      `json:"deal_outlook"`
	CustomerSentiment string      `json:"customer_sentiment"`
	Summary           string      `json:"summary"`
	PositiveSignals   []string    `json:"positive_signals"`
	NegativeSignals   []string    `json:"negative_signals"`
	NextSteps         []string    `json:"next_steps"`
	DealDetails       DealDetails `json:"deal_details"`
}

type DealDetails struct {
	Company string `json:"company"`
	Contact string `json:"contact"`
	Product string `json:"product"`
}

type AnalysisService struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewAnalysisService(apiKey string, model string) *AnalysisService {
	return &AnalysisService{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	MaxTokens      int             `json:"max_tokens"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
	Provider       *providerPrefs  `json:"provider,omitempty"`
}

type providerPrefs struct {
	RequireParameters bool `json:"require_parameters"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

func (s *AnalysisService) Analyze(ctx context.Context, transcript string) (*AnalysisResult, error) {
	reqBody := chatRequest{
		Model: s.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: transcript},
		},
		ResponseFormat: &responseFormat{
			Type: "json_object",
		},
		Provider: &providerPrefs{
			RequireParameters: true,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://openrouter.ai/api/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openrouter request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return nil, fmt.Errorf("openrouter error (%d): %s", chatResp.Error.Code, chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openrouter returned no choices")
	}

	choice := chatResp.Choices[0]
	log.Printf("LLM finish_reason=%s, model=%s, content_length=%d", choice.FinishReason, chatResp.Model, len(choice.Message.Content))

	if choice.FinishReason != "stop" {
		log.Printf("raw LLM response: %s", choice.Message.Content)
		return nil, fmt.Errorf("LLM response incomplete (finish_reason=%s)", choice.FinishReason)
	}

	raw := choice.Message.Content

	var result AnalysisResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		log.Printf("raw LLM response: %s", raw)
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w", err)
	}

	return &result, nil
}