package voice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/herlin-ai/herlin-assistant/config"
)

type Service struct {
	cfg *config.Config
}

type TranscriptionResult struct {
	Text     string  `json:"text"`
	Language string  `json:"language"`
	Duration float64 `json:"duration"`
}

type SynthesisRequest struct {
	Text     string  `json:"text"`
	Voice    string  `json:"voice"`
	Language string  `json:"language"`
}

type SynthesisResult struct {
	AudioData []byte `json:"audio_data"`
	Format    string `json:"format"`
}

func NewService(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}

// Speech-to-Text using OpenAI Whisper
func (s *Service) SpeechToText(audioData []byte, audioFormat string) (*TranscriptionResult, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		return s.fallbackTranscription()
	}

	// Prepare request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio file
	part, err := writer.CreateFormFile("file", "audio."+audioFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return nil, fmt.Errorf("failed to write audio data: %w", err)
	}

	// Add model parameter
	writer.WriteField("model", "whisper-1")
	writer.WriteField("language", "en")

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req.Body = body
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+s.cfg.AI.OpenAI.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(respBody))
	}

	var result struct {
		Text     string  `json:"text"`
		Language string  `json:"language"`
		Duration float64 `json:"duration"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &TranscriptionResult{
		Text:     result.Text,
		Language: result.Language,
		Duration: result.Duration,
	}, nil
}

// Text-to-Speech using OpenAI TTS
func (s *Service) TextToSpeech(text string, voice string, language string) (*SynthesisResult, error) {
	if s.cfg.AI.OpenAI.APIKey == "" {
		return s.fallbackSynthesis(text)
	}

	type TTSRequest struct {
		Model  string `json:"model"`
		Input  string `json:"input"`
		Voice  string `json:"voice"`
	}

	if voice == "" {
		voice = "alloy"
	}

	req := TTSRequest{
		Model: "tts-1",
		Input: text,
		Voice: voice,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	return &SynthesisResult{
		AudioData: audioData,
		Format:    "mp3",
	}, nil
}

// Alternative TTS using Google Speech API
func (s *Service) TextToSpeechGoogle(text string, voice string, language string) (*SynthesisResult, error) {
	if s.cfg.AI.Gemini.APIKey == "" {
		return s.fallbackSynthesis(text)
	}

	type GoogleTTSRequest struct {
		Input struct {
			Text string `json:"text"`
		} `json:"input"`
		Voice struct {
			LanguageCode string `json:"languageCode"`
			Name         string `json:"name"`
		} `json:"voice"`
		AudioConfig struct {
			AudioEncoding string `json:"audioEncoding"`
		} `json:"audioConfig"`
	}

	if language == "" {
		language = "en-US"
	}
	if voice == "" {
		voice = "en-US-Standard-A"
	}

	req := GoogleTTSRequest{}
	req.Input.Text = text
	req.Voice.LanguageCode = language
	req.Voice.Name = voice
	req.AudioConfig.AudioEncoding = "MP3"

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://texttospeech.googleapis.com/v1/text:synthesize?key=%s", s.cfg.AI.Gemini.APIKey)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google TTS API error: %s", string(respBody))
	}

	var result struct {
		AudioContent string `json:"audioContent"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &SynthesisResult{
		AudioData: []byte(result.AudioContent),
		Format:    "mp3",
	}, nil
}

func (s *Service) fallbackTranscription() (*TranscriptionResult, error) {
	return &TranscriptionResult{
		Text:     "Speech recognition not available. Please configure OpenAI API key.",
		Language: "en",
		Duration: 0,
	}, nil
}

func (s *Service) fallbackSynthesis(text string) (*SynthesisResult, error) {
	return &SynthesisResult{
		AudioData: []byte{},
		Format:    "mp3",
	}, nil
}

// Voice command processing
func (s *Service) ProcessVoiceCommand(transcription string) (string, string, error) {
	// Simple command detection
	commands := map[string]string{
		"hello": "greeting",
		"hi": "greeting",
		"what can you do": "capabilities",
		"help": "help",
		"stop": "stop",
		"cancel": "cancel",
	}

	for cmd, action := range commands {
		if contains(transcription, cmd) {
			return action, transcription, nil
		}
	}

	return "query", transcription, nil
}

func contains(text, substring string) bool {
	return len(text) >= len(substring) && (text == substring || 
		len(text) > len(substring) && (text[:len(substring)] == substring || 
		text[len(text)-len(substring):] == substring))
}
