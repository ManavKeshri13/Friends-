package services

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type TopicInfo struct {
	Subject  string   `json:"subject"`
	Entities []string `json:"entities"`
	Scope    string   `json:"scope"`
}

func ExtractStructure(groqKey, text string) (*TopicInfo, error) {

	prompt := `
Return ONLY JSON.

{
 "subject": short main concept,
 "entities": [specific named things],
 "scope": "general | specific | comparison"
}

Text: ` + text

	reqBody := map[string]any{
		"model": "llama3-70b-8192",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	b, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+groqKey)
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var raw map[string]any
	json.NewDecoder(resp.Body).Decode(&raw)

	txt := raw["choices"].([]any)[0].(map[string]any)["message"].(map[string]any)["content"].(string)

	var info TopicInfo
	json.Unmarshal([]byte(txt), &info)

	return &info, nil
}
