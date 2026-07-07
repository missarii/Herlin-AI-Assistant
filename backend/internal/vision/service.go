package vision

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/ai"
)

type Service struct {
	cfg *config.Config
	ai  *ai.Service
}

type ImageAnalysis struct {
	Description string   `json:"description"`
	Objects     []string `json:"objects"`
	Text        string   `json:"text,omitempty"`
	Confidence  float64  `json:"confidence"`
}

type ScreenshotAnalysis struct {
	ErrorExplanation string `json:"error_explanation"`
	SuggestedFix     string `json:"suggested_fix"`
	CodeSnippet      string `json:"code_snippet,omitempty"`
}

type DiagramAnalysis struct {
	Type        string   `json:"type"`
	Components []string `json:"components"`
	Flow       string   `json:"flow"`
	Summary    string   `json:"summary"`
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg, ai: ai.NewService(cfg)}
}

func (s *Service) AnalyzeImage(imageData []byte, imageFormat string, query string) (*ImageAnalysis, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		return s.fallbackImageAnalysis()
	}

	// Convert image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:image/%s;base64,%s", imageFormat, base64Image)

	// Use GPT-4 Vision for image analysis
	type VisionRequest struct {
		Model string `json:"model"`
		Messages []struct {
			Role string `json:"role"`
			Content []interface{} `json:"content"`
		} `json:"messages"`
		MaxTokens int `json:"max_tokens"`
	}

	content := []interface{}{
		map[string]string{"type": "text", "text": "Analyze this image. Describe what you see, identify objects, and extract any text if present."},
		map[string]string{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
	}

	if query != "" {
		content = []interface{}{
			map[string]string{"type": "text", "text": query},
			map[string]string{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
		}
	}

	req := VisionRequest{
		Model: "gpt-4-vision-preview",
		Messages: []struct {
			Role string `json:"role"`
			Content []interface{} `json:"content"`
		}{
			{
				Role:    "user",
				Content: content,
			},
		},
		MaxTokens: 1000,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.OpenAI.APIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// Parse the response to extract structured information
	return s.parseImageAnalysis(result.Choices[0].Message.Content)
}

func (s *Service) AnalyzeScreenshot(imageData []byte, imageFormat string) (*ScreenshotAnalysis, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		return s.fallbackScreenshotAnalysis()
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:image/%s;base64,%s", imageFormat, base64Image)

	type VisionRequest struct {
		Model string `json:"model"`
		Messages []struct {
			Role string `json:"role"`
			Content []interface{} `json:"content"`
		} `json:"messages"`
		MaxTokens int `json:"max_tokens"`
	}

	req := VisionRequest{
		Model: "gpt-4-vision-preview",
		Messages: []struct {
			Role string `json:"role"`
			Content []interface{} `json:"content"`
		}{
			{
				Role: "user",
				Content: []interface{}{
					map[string]string{"type": "text", "text": "This is a screenshot, possibly containing code or an error message. Analyze it and provide: 1) The error explanation, 2) Suggested fix, 3) Any relevant code snippet. Return as JSON with fields: error_explanation, suggested_fix, code_snippet"},
					map[string]string{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
				},
			},
		},
		MaxTokens: 1500,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.OpenAI.APIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// Parse JSON response
	var analysis ScreenshotAnalysis
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &analysis); err != nil {
		// Fallback if not valid JSON
		analysis = ScreenshotAnalysis{
			ErrorExplanation: result.Choices[0].Message.Content,
		}
	}

	return &analysis, nil
}

func (s *Service) AnalyzeDiagram(imageData []byte, imageFormat string) (*DiagramAnalysis, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		return s.fallbackDiagramAnalysis()
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:image/%s;base64,%s", imageFormat, base64Image)

	type VisionRequest struct {
		Model string `json:"model"`
		Messages []struct {
			Role string `json:"role"`
			Content []interface{} `json:"content"`
		} `json:"messages"`
		MaxTokens int `json:"max_tokens"`
	}

	req := VisionRequest{
		Model: "gpt-4-vision-preview",
		Messages: []struct {
			Role string `json:"role"`
			Content []interface{} `json:"content"`
		}{
			{
				Role: "user",
				Content: []interface{}{
					map[string]string{"type": "text", "text": "Analyze this diagram/chart. Identify the type, components, flow, and provide a summary. Return as JSON with fields: type, components (array), flow, summary"},
					map[string]string{"type": "image_url", "image_url": map[string]string{"url": dataURL}},
				},
			},
		},
		MaxTokens: 1000,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.cfg.AI.OpenAI.APIKey)

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	// Parse JSON response
	var analysis DiagramAnalysis
	if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &analysis); err != nil {
		// Fallback if not valid JSON
		analysis = DiagramAnalysis{
			Summary: result.Choices[0].Message.Content,
		}
	}

	return &analysis, nil
}

func (s *Service) parseImageAnalysis(content string) (*ImageAnalysis, error) {
	// Try to parse as JSON first
	var analysis ImageAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err == nil {
		return &analysis, nil
	}

	// Fallback: use AI to extract structured information
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "Extract structured information from the image analysis. Return JSON with fields: description, objects (array), text (optional), confidence (0-1)",
		},
		{
			Role:    "user",
			Content: content,
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		// Ultimate fallback
		return &ImageAnalysis{
			Description: content,
			Objects:     []string{},
			Confidence:  0.5,
		}, nil
	}

	if err := json.Unmarshal([]byte(response.Content), &analysis); err != nil {
		return &ImageAnalysis{
			Description: response.Content,
			Objects:     []string{},
			Confidence:  0.5,
		}, nil
	}

	return &analysis, nil
}

func (s *Service) fallbackImageAnalysis() (*ImageAnalysis, error) {
	return &ImageAnalysis{
		Description: "Image analysis not available. Please configure OpenAI API key with GPT-4 Vision access.",
		Objects:     []string{},
		Confidence:  0.0,
	}, nil
}

func (s *Service) fallbackScreenshotAnalysis() (*ScreenshotAnalysis, error) {
	return &ScreenshotAnalysis{
		ErrorExplanation: "Screenshot analysis not available. Please configure OpenAI API key with GPT-4 Vision access.",
		SuggestedFix:     "Add your OpenAI API key to enable this feature.",
	}, nil
}

func (s *Service) fallbackDiagramAnalysis() (*DiagramAnalysis, error) {
	return &DiagramAnalysis{
		Type:  "unknown",
		Summary: "Diagram analysis not available. Please configure OpenAI API key with GPT-4 Vision access.",
	}, nil
}
