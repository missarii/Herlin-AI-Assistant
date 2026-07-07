// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/herlin-ai/herlin-assistant/config"
)

type Service struct {
	cfg *config.Config
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Content string
	Tokens  int
	Model   string
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) GenerateResponse(messages []Message, model string) (*Response, error) {
	provider := s.getProvider(model)

	switch provider {
	case "openai":
		return s.generateOpenAIResponse(messages, model)
	case "gemini":
		return s.generateGeminiResponse(messages, model)
	case "claude":
		return s.generateClaudeResponse(messages, model)
	default:
		return s.generateOpenAIResponse(messages, model)
	}
}

func (s *Service) GenerateResponseStream(messages []Message, model string, chunkChan chan<- StreamChunk) (*Response, error) {
	provider := s.getProvider(model)

	switch provider {
	case "openai":
		return s.generateOpenAIStreamResponse(messages, model, chunkChan)
	case "gemini":
		return s.generateGeminiStreamResponse(messages, model, chunkChan)
	case "claude":
		return s.generateClaudeStreamResponse(messages, model, chunkChan)
	default:
		return s.generateOpenAIStreamResponse(messages, model, chunkChan)
	}
}

type StreamChunk struct {
	Content string
	Done    bool
	Token   int
}

func (s *Service) getProvider(model string) string {
	if strings.Contains(model, "gpt") {
		return "openai"
	}
	if strings.Contains(model, "gemini") {
		return "gemini"
	}
	if strings.Contains(model, "claude") {
		return "claude"
	}
	return s.cfg.AI.DefaultProvider
}

// OpenAI Implementation
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (s *Service) generateOpenAIResponse(messages []Message, model string) (*Response, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		return s.fallbackResponse(model)
	}

	requestModel := model
	if requestModel == "" {
		requestModel = s.cfg.AI.OpenAI.Model
	}

	req := OpenAIRequest{
		Model:    requestModel,
		Messages: messages,
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

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	return &Response{
		Content: openAIResp.Choices[0].Message.Content,
		Tokens:  openAIResp.Usage.TotalTokens,
		Model:   requestModel,
	}, nil
}

// Gemini Implementation
type GeminiRequest struct {
	Contents []struct {
		Role  string `json:"role"`
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	UsageMetadata struct {
		TotalTokenCount int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

func (s *Service) generateGeminiResponse(messages []Message, model string) (*Response, error) {
	if s.cfg.AI.Gemini.APIKey == "" {
		return s.fallbackResponse(model)
	}

	requestModel := model
	if requestModel == "" {
		requestModel = s.cfg.AI.Gemini.Model
	}

	var contents []struct {
		Role  string `json:"role"`
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	}

	for _, msg := range messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			Role: role,
			Parts: []struct {
				Text string `json:"text"`
			}{
				{Text: msg.Content},
			},
		})
	}

	req := GeminiRequest{Contents: contents}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", requestModel, s.cfg.AI.Gemini.APIKey)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("Gemini API error: %s", string(body))
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	return &Response{
		Content: geminiResp.Candidates[0].Content.Parts[0].Text,
		Tokens:  geminiResp.UsageMetadata.TotalTokenCount,
		Model:   requestModel,
	}, nil
}

// Claude Implementation
type ClaudeRequest struct {
	Model       string    `json:"model"`
	MaxTokens   int       `json:"max_tokens"`
	Messages    []Message `json:"messages"`
}

type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (s *Service) generateClaudeResponse(messages []Message, model string) (*Response, error) {
	if s.cfg.AI.Claude.APIKey == "" {
		return s.fallbackResponse(model)
	}

	requestModel := model
	if requestModel == "" {
		requestModel = s.cfg.AI.Claude.Model
	}

	req := ClaudeRequest{
		Model:     requestModel,
		MaxTokens: 4096,
		Messages:  messages,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", s.cfg.AI.Claude.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

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
		return nil, fmt.Errorf("Claude API error: %s", string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	return &Response{
		Content: claudeResp.Content[0].Text,
		Tokens:  claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		Model:   requestModel,
	}, nil
}

func (s *Service) fallbackResponse(model string) (*Response, error) {
	return &Response{
		Content: "I'm sorry, but I don't have access to an AI API key. Please configure your API keys in the environment variables to enable AI responses.",
		Tokens:  0,
		Model:   model,
	}, nil
}

// Streaming implementations
func (s *Service) generateOpenAIStreamResponse(messages []Message, model string, chunkChan chan<- StreamChunk) (*Response, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		chunkChan <- StreamChunk{Content: "No API key configured", Done: true}
		return s.fallbackResponse(model)
	}

	requestModel := model
	if requestModel == "" {
		requestModel = s.cfg.AI.OpenAI.Model
	}

	req := OpenAIRequest{
		Model:    requestModel,
		Messages: messages,
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
	httpReq.Header.Set("Accept", "text/event-stream")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	reader := bufio.NewReader(resp.Body)
	var fullContent string
	var totalTokens int

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read stream: %w", err)
		}

		if len(line) < 6 || line[:6] != "data: " {
			continue
		}

		data := line[6:]
		if data == "[DONE]" {
			break
		}

		var streamResp struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
			continue
		}

		if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
			content := streamResp.Choices[0].Delta.Content
			fullContent += content
			chunkChan <- StreamChunk{Content: content, Done: false}
		}
	}

	chunkChan <- StreamChunk{Done: true}

	return &Response{
		Content: fullContent,
		Tokens:  totalTokens,
		Model:   requestModel,
	}, nil
}

func (s *Service) generateGeminiStreamResponse(messages []Message, model string, chunkChan chan<- StreamChunk) (*Response, error) {
	// Gemini streaming implementation similar to OpenAI
	// For now, use non-streaming fallback
	response, err := s.generateGeminiResponse(messages, model)
	if err != nil {
		return nil, err
	}
	
	chunkChan <- StreamChunk{Content: response.Content, Done: true}
	return response, nil
}

func (s *Service) generateClaudeStreamResponse(messages []Message, model string, chunkChan chan<- StreamChunk) (*Response, error) {
	// Claude streaming implementation similar to OpenAI
	// For now, use non-streaming fallback
	response, err := s.generateClaudeResponse(messages, model)
	if err != nil {
		return nil, err
	}
	
	chunkChan <- StreamChunk{Content: response.Content, Done: true}
	return response, nil
}
