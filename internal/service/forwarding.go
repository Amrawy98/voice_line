package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	notionAPIURL  = "https://api.notion.com/v1/pages"
	notionVersion = "2022-06-28"
)

type ForwardingService struct {
	apiKey   string
	parentID string
	httpClient *http.Client
}

func NewForwardingService(apiKey, parentID string) *ForwardingService {
	return &ForwardingService{
		apiKey:     apiKey,
		parentID:   parentID,
		httpClient: &http.Client{},
	}
}

type PageInput struct {
	Summary           string
	DealOutlook       string
	CustomerSentiment string
	Company           string
	Contact           string
	Product           string
	PositiveSignals   []string
	NegativeSignals   []string
	NextSteps         []string
}

func (s *ForwardingService) CreatePage(ctx context.Context, input PageInput) error {
	date := time.Now().Format("Jan 2, 2006")
	title := "Sales Call Summary - " + date
	if input.Company != "" {
		title = "Sales Call - " + input.Company + " - " + date
	}

	children := []map[string]interface{}{
		{
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": input.Summary}},
				},
			},
		},
		{
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": "Deal Outlook"}},
				},
			},
		},
		{
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": input.DealOutlook}},
				},
			},
		},
		{
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": "Customer Sentiment"}},
				},
			},
		},
		{
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": input.CustomerSentiment}},
				},
			},
		},
	}

	if len(input.PositiveSignals) > 0 {
		children = append(children, map[string]interface{}{
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": "Positive Signals"}},
				},
			},
		})
		for _, signal := range input.PositiveSignals {
			children = append(children, map[string]interface{}{
				"bulleted_list_item": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{"text": map[string]string{"content": signal}},
					},
				},
			})
		}
	}

	if len(input.NegativeSignals) > 0 {
		children = append(children, map[string]interface{}{
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": "Negative Signals"}},
				},
			},
		})
		for _, signal := range input.NegativeSignals {
			children = append(children, map[string]interface{}{
				"bulleted_list_item": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{"text": map[string]string{"content": signal}},
					},
				},
			})
		}
	}

	if len(input.NextSteps) > 0 {
		children = append(children, map[string]interface{}{
			"heading_2": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{"text": map[string]string{"content": "Next Steps"}},
				},
			},
		})
		for _, step := range input.NextSteps {
			children = append(children, map[string]interface{}{
				"bulleted_list_item": map[string]interface{}{
					"rich_text": []map[string]interface{}{
						{"text": map[string]string{"content": step}},
					},
				},
			})
		}
	}

	body := map[string]interface{}{
		"parent": map[string]string{
			"page_id": s.parentID,
		},
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"title": []map[string]interface{}{
					{"text": map[string]string{"content": title}},
				},
			},
		},
		"children": children,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal Notion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, notionAPIURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create Notion request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", notionVersion)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Notion API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Notion API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}