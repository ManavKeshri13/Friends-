package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func ResolveGroqAPIKeyFromRequest(c *gin.Context) string {
	headers := []string{"X-Groq-Api-Key", "X-Api-Key"}
	for _, h := range headers {
		if v := strings.TrimSpace(c.GetHeader(h)); v != "" {
			return v
		}
	}
	return ResolveGroqAPIKeyFromEnv()
}

func ResolveGroqAPIKeyFromEnv() string {
	candidates := []string{"GROQ_API_KEY", "GROQ_API_TOKEN", "API_KEY"}
	for _, name := range candidates {
		if v := strings.TrimSpace(os.Getenv(name)); v != "" {
			return v
		}
	}
	return ""
}

func GenerateGroqReply(ctx context.Context, apiKey string, preferredModel string, topic string, input string) (string, error) {
	if strings.TrimSpace(apiKey) == "" {
		return "", fmt.Errorf("GROQ_API_KEY not set")
	}

	model := "llama-3.1-70b-versatile"
	if m := strings.TrimSpace(preferredModel); m != "" {
		model = m
	} else {
		fallbackModels := []string{
			"llama-3.3-70b-versatile",
			"llama-3.1-70b-versatile",
			"llama-3.1-8b-instant",
			"mixtral-8x7b-32768",
			"gemma2-9b-it",
		}
		for _, m := range fallbackModels {
			if testReply, err := httpGroqGenerate(apiKey, m, topic, input); err == nil && strings.TrimSpace(testReply) != "" {
				model = m
				break
			}
		}
	}

	return httpGroqGenerate(apiKey, model, topic, input)
}

func httpGroqGenerate(apiKey string, model string, topic string, input string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	prompt := fmt.Sprintf("You are a helpful tutor. The chat topic is '%s'. Answer ONLY within this topic. User: %s", topic, input)

	bodyObj := map[string]interface{}{
		"model": model,
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  1024,
	}

	b, err := json.Marshal(bodyObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * 1e9}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if parsed.Error != nil {
		return "", fmt.Errorf("API error: %s", parsed.Error.Message)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

func ExtractTopicFromMessage(message string) string {
	apiKey := ResolveGroqAPIKeyFromEnv()
	if apiKey == "" {
		apiKey = ResolveGroqAPIKeyFromEnv()
	}

	prompt := fmt.Sprintf(`Extract the main topic from this message. 
Return ONLY a short topic name (1-3 words).
Examples: "Python Programming", "JavaScript", "Machine Learning", "History"

Message: %s

Topic:`, message)

	bodyObj := map[string]interface{}{
		"model": "llama-3.1-8b-instant",
		"messages": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.3,
		"max_tokens":  50,
	}

	b, err := json.Marshal(bodyObj)
	if err != nil {
		return ""
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return ""
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 15 * 1e9}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return ""
	}

	if len(parsed.Choices) == 0 {
		return ""
	}

	topic := strings.TrimSpace(parsed.Choices[0].Message.Content)
	topic = strings.Trim(topic, "\"'")
	return topic
}
