package code

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/ai"
)

type Service struct {
	cfg *config.Config
	ai  *ai.Service
}

type CodeRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
	Query    string `json:"query"`
}

type CodeResponse struct {
	GeneratedCode string   `json:"generated_code,omitempty"`
	Explanation   string   `json:"explanation,omitempty"`
	Optimizations []string `json:"optimizations,omitempty"`
	Errors        []string `json:"errors,omitempty"`
	Suggestions   []string `json:"suggestions,omitempty"`
	Documentation string   `json:"documentation,omitempty"`
}

type CodeReview struct {
	Score       int      `json:"score"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	BestPractices []string `json:"best_practices"`
	Security    []string `json:"security"`
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg, ai: ai.NewService(cfg)}
}

func (s *Service) GenerateCode(req CodeRequest) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("You are an expert %s developer. Generate clean, efficient, and well-documented code. Provide code only, no explanations unless asked.", req.Language),
		},
		{
			Role:    "user",
			Content: req.Query,
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return &CodeResponse{
		GeneratedCode: response.Content,
	}, nil
}

func (s *Service) DebugCode(req CodeRequest) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are an expert code debugger. Analyze the code, identify errors, and provide fixes. Return your response as JSON with fields: errors (array), suggestions (array), and fixed_code (string).",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Language: %s\n\nCode:\n%s\n\nProblem: %s", req.Language, req.Code, req.Query),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var result struct {
		Errors      []string `json:"errors"`
		Suggestions []string `json:"suggestions"`
		FixedCode   string   `json:"fixed_code"`
	}

	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// Fallback if not valid JSON
		return &CodeResponse{
			Explanation: response.Content,
		}, nil
	}

	return &CodeResponse{
		GeneratedCode: result.FixedCode,
		Errors:        result.Errors,
		Suggestions:   result.Suggestions,
	}, nil
}

func (s *Service) ReviewCode(req CodeRequest) (*CodeReview, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are an expert code reviewer. Analyze the code for quality, best practices, security, and potential issues. Return your response as JSON with fields: score (1-10), issues (array), suggestions (array), best_practices (array), and security (array).",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Language: %s\n\nCode:\n%s", req.Language, req.Code),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var review CodeReview
	if err := json.Unmarshal([]byte(response.Content), &review); err != nil {
		// Fallback if not valid JSON
		return &CodeReview{
			Score:       5,
			Suggestions: []string{response.Content},
		}, nil
	}

	return &review, nil
}

func (s *Service) OptimizeCode(req CodeRequest) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are an expert code optimizer. Analyze the code and provide optimized version with explanations. Return your response as JSON with fields: optimized_code (string), optimizations (array), and explanation (string).",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Language: %s\n\nCode:\n%s", req.Language, req.Code),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	// Try to parse as JSON
	var result struct {
		OptimizedCode string   `json:"optimized_code"`
		Optimizations []string `json:"optimizations"`
		Explanation   string   `json:"explanation"`
	}

	if err := json.Unmarshal([]byte(response.Content), &result); err != nil {
		// Fallback if not valid JSON
		return &CodeResponse{
			Explanation: response.Content,
		}, nil
	}

	return &CodeResponse{
		GeneratedCode: result.OptimizedCode,
		Optimizations: result.Optimizations,
		Explanation:   result.Explanation,
	}, nil
}

func (s *Service) GenerateDocumentation(req CodeRequest) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are an expert technical writer. Generate comprehensive documentation for the provided code. Include function descriptions, parameter explanations, return values, and usage examples.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Language: %s\n\nCode:\n%s", req.Language, req.Code),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return &CodeResponse{
		Documentation: response.Content,
	}, nil
}

func (s *Service) ExplainCode(req CodeRequest) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: "You are an expert code educator. Explain the code in simple terms, covering what it does, how it works, and key concepts.",
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Language: %s\n\nCode:\n%s", req.Language, req.Code),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return &CodeResponse{
		Explanation: response.Content,
	}, nil
}

func (s *Service) ConvertCode(req CodeRequest, targetLanguage string) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("You are an expert in multiple programming languages. Convert the provided code from %s to %s while maintaining the same functionality.", req.Language, targetLanguage),
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Source code (%s):\n%s", req.Language, req.Code),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return &CodeResponse{
		GeneratedCode: response.Content,
	}, nil
}

func (s *Service) GenerateTests(req CodeRequest) (*CodeResponse, error) {
	messages := []ai.Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("You are an expert in test-driven development. Generate comprehensive unit tests for the provided %s code using best practices and popular testing frameworks.", req.Language),
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("Code:\n%s", req.Code),
		},
	}

	response, err := s.ai.GenerateResponse(messages, "")
	if err != nil {
		return nil, err
	}

	return &CodeResponse{
		GeneratedCode: response.Content,
	}, nil
}

func (s *Service) DetectLanguage(code string) string {
	// Simple language detection based on syntax
	if strings.Contains(code, "func ") && strings.Contains(code, "package ") {
		return "go"
	}
	if strings.Contains(code, "def ") && strings.Contains(code, "import ") {
		return "python"
	}
	if strings.Contains(code, "function ") && strings.Contains(code, "const ") {
		return "javascript"
	}
	if strings.Contains(code, "public class ") {
		return "java"
	}
	if strings.Contains(code, "#include") {
		return "cpp"
	}
	if strings.Contains(code, "fn ") && strings.Contains(code, "let ") {
		return "rust"
	}
	return "unknown"
}

func (s *Service) GetSupportedLanguages() []string {
	return []string{
		"go",
		"python",
		"javascript",
		"typescript",
		"java",
		"cpp",
		"c",
		"rust",
		"ruby",
		"php",
		"swift",
		"kotlin",
	}
}
