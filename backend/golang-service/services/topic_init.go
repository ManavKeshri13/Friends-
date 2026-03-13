package services

import (
	"encoding/json"
	"golang-service/config"
)

func InitTopic(chatID, message, hfKey, groqKey string) error {

	emb, _ := HFEmbedding(hfKey, message)
	info, _ := ExtractStructure(groqKey, message)

	embJSON, _ := json.Marshal(emb)
	entitiesJSON, _ := json.Marshal(info.Entities)

	_, err := config.DB.Exec(`
	INSERT INTO chat_topic_state
	(chat_id, topic_text, topic_embedding, subject, scope, entities)
	VALUES ($1,$2,$3,$4,$5,$6)
	`, chatID, message, string(embJSON), info.Subject, info.Scope, string(entitiesJSON))

	return err
}
