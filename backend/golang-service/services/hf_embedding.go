package services

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func HFEmbedding(apikey, text string) ([]float32, error) {
	body := map[string]string{"inputs": text}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest(
		"POST",
		"https://api-inference.huggingface.co/pipeline/feature-extraction/sentence-transformers/all-MiniLM-L6-v2",
		bytes.NewBuffer(b),
	)

	req.Header.Set("Authorization", "Bearer "+apikey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var vec [][]float32
	json.NewDecoder(resp.Body).Decode(&vec)

	return vec[0], nil
}
