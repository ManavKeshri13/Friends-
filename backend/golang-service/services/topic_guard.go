package services

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"golang-service/config"
)

func CheckTopic(chatID, message, hfKey, groqKey string) (bool, string, error) {
	var chat struct {
		Topic string `db:"topic"`
	}
	err := config.DB.Get(&chat, "SELECT topic FROM chats WHERE id=$1", chatID)
	if err != nil {
		return false, "", err
	}

	if chat.Topic == "" {
		return true, "", nil
	}

	isOnTopic, reason := IsMessageOnTopic(groqKey, chat.Topic, message)

	if !isOnTopic {
		lockChat(chatID, reason)
		return false, reason, nil
	}

	return true, "", nil
}

func IsMessageOnTopic(groqKey, topic, message string) (bool, string) {
	if groqKey == "" {
		groqKey = ResolveGroqAPIKeyFromEnv()
	}

	if groqKey == "" {
		return true, ""
	}

	prompt := fmt.Sprintf(`You are a topic checker.

Current Topic: %s
User Message: %s

Determine if the user message is related to the current topic.

Rules:
- A message is ON-TOPIC if it asks about or discusses anything related to "%s"
- A message is OFF-TOPIC if it completely changes the subject to something unrelated

Return ONLY JSON:
{"on_topic": true/false, "reason": "brief explanation if off-topic"}

Example off-topic: Topic="Python" Message="What's for lunch?" -> {"on_topic": false, "reason": "Asking about food, not Python"}
Example on-topic: Topic="Python" Message="How do I use functions?" -> {"on_topic": true, "reason": ""}`, topic, message, topic)

	bodyObj := map[string]interface{}{
		"model":       "llama-3.1-8b-instant",
		"messages":    []interface{}{map[string]interface{}{"role": "user", "content": prompt}},
		"temperature": 0.3,
		"max_tokens":  100,
	}

	b, _ := json.Marshal(bodyObj)
	req, _ := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+groqKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return true, ""
	}
	defer resp.Body.Close()

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return true, ""
	}

	if len(parsed.Choices) == 0 {
		return true, ""
	}

	var result struct {
		OnTopic bool   `json:"on_topic"`
		Reason  string `json:"reason"`
	}

	content := parsed.Choices[0].Message.Content
	json.Unmarshal([]byte(content), &result)

	return result.OnTopic, result.Reason
}

func lockChat(chatID, reason string) {
	config.DB.Exec("UPDATE chat_topic_state SET locked=$1, locked_reason=$2 WHERE chat_id=$3", true, reason, chatID)
}

func getChatTopic(chatID string) (string, error) {
	var topic string
	err := config.DB.Get(&topic, "SELECT topic FROM chats WHERE id=$1", chatID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return topic, nil
}
