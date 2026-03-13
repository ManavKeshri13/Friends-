// services/scheduler.go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang-service/config"
	"golang-service/models"
)

func StartScheduler() {
	fmt.Println("✅ Scheduler started")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		processDueSchedules()
	}
}

func processDueSchedules() {
	var schedules []models.Schedule
	err := config.DB.Select(&schedules, `
		SELECT * FROM schedules
		WHERE active = true
		AND scheduled_time <= NOW()
		AND scheduled_time >= NOW() - INTERVAL '1 hour'
	`)
	if err != nil || len(schedules) == 0 {
		return
	}

	fmt.Printf("📅 %d due schedule(s) found\n", len(schedules))
	for _, s := range schedules {
		go triggerQuiz(s)
	}
}

func triggerQuiz(schedule models.Schedule) {
	var user models.User
	if err := config.DB.Get(&user, "SELECT * FROM users WHERE id=$1", schedule.UserID); err != nil {
		fmt.Printf("User not found for schedule %d\n", schedule.ID)
		return
	}
	if user.Phone == "" {
		fmt.Printf("No phone for user %d\n", schedule.UserID)
		return
	}

	quizID, totalQuestions, err := generateAndStoreQuiz(schedule)
	if err != nil {
		fmt.Printf("Failed to generate quiz: %v\n", err)
		return
	}

	if err := advanceSchedule(schedule); err != nil {
		fmt.Printf("Failed to advance schedule: %v\n", err)
	}

	n8nWebhookURL := os.Getenv("N8N_WEBHOOK_URL")
	if n8nWebhookURL == "" {
		fmt.Println("⚠️ N8N_WEBHOOK_URL not set, skipping n8n trigger")
		return
	}

	payload := map[string]interface{}{
		"phone":           user.Phone,
		"quiz_id":         quizID,
		"topic":           schedule.Topic,
		"total_questions": totalQuestions,
		"schedule_id":     schedule.ID,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(n8nWebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil || resp.StatusCode >= 400 {
		fmt.Printf("Failed to trigger n8n for quiz %d: %v\n", quizID, err)
		return
	}

	fmt.Printf("✅ Quiz %d handed to n8n for phone %s\n", quizID, user.Phone)
}

func generateAndStoreQuiz(schedule models.Schedule) (int, int, error) {
	duration := 10
	numQuestions := duration / 3
	if numQuestions < 3 {
		numQuestions = 3
	}
	if numQuestions > 20 {
		numQuestions = 20
	}

	apiKey := ResolveGroqAPIKeyFromEnv()
	ctx := context.Background()

	questions, err := generateMCQQuestions(ctx, apiKey, schedule.Topic, numQuestions)
	if err != nil || len(questions) == 0 {
		questions = generateSimpleMCQQuestions(schedule.Topic, numQuestions)
	}

	if len(questions) < numQuestions {
		fallback := generateSimpleMCQQuestions(schedule.Topic, numQuestions-len(questions))
		questions = append(questions, fallback...)
	}
	if len(questions) > numQuestions {
		questions = questions[:numQuestions]
	}

	config.DB.Exec(`DELETE FROM quizzes WHERE chat_id=$1 AND status != 'completed'`, schedule.ChatID)

	var quizID int
	err = config.DB.QueryRow(`
		INSERT INTO quizzes (user_id, chat_id, topic, status, total_questions, created_at)
		VALUES ($1, $2, $3, 'pending', $4, $5)
		RETURNING id
	`, schedule.UserID, schedule.ChatID, schedule.Topic, len(questions), time.Now()).Scan(&quizID)
	if err != nil {
		return 0, 0, err
	}

	for i, q := range questions {
		optionsJSON, _ := json.Marshal(q.Options)
		_, err = config.DB.Exec(`
			INSERT INTO quiz_questions (quiz_id, question, answer, options, order_num)
			VALUES ($1, $2, $3, $4, $5)
		`, quizID, q.Question, q.Answer, string(optionsJSON), i+1)
		if err != nil {
			return 0, 0, err
		}
	}

	return quizID, len(questions), nil
}

func advanceSchedule(schedule models.Schedule) error {
	if schedule.RecurrenceType == "once" {
		_, err := config.DB.Exec("UPDATE schedules SET active=false WHERE id=$1", schedule.ID)
		return err
	}

	var nextTime time.Time
	now := time.Now()

	t, err := time.Parse("15:04", schedule.ReminderTime)
	if err != nil {
		return fmt.Errorf("invalid reminder_time in schedule: %w", err)
	}
	hour := t.Hour()
	minute := t.Minute()

	switch schedule.RecurrenceType {
	case "daily":
		tomorrow := now.AddDate(0, 0, 1)
		nextTime = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), hour, minute, 0, 0, now.Location())

	case "weekly":
		nextDay := now.AddDate(0, 0, 1)
		nextTime = calculateNextWeeklyReminder(nextDay, schedule.DaysOfWeek, hour, minute)

	default:
		return fmt.Errorf("unknown recurrence_type: %s", schedule.RecurrenceType)
	}

	_, err = config.DB.Exec(
		"UPDATE schedules SET scheduled_time=$1 WHERE id=$2",
		nextTime, schedule.ID,
	)
	return err
}

func calculateNextWeeklyReminder(now time.Time, daysOfWeek string, hour, minute int) time.Time {
	var days []int
	for _, dayStr := range splitAndTrim(daysOfWeek, ",") {
		var day int
		fmt.Sscanf(dayStr, "%d", &day)
		if day >= 0 && day <= 6 {
			days = append(days, day)
		}
	}

	if len(days) == 0 {
		return now.AddDate(0, 0, 1)
	}

	currentWeekday := int(now.Weekday())

	for i := 0; i < len(days)-1; i++ {
		for j := i + 1; j < len(days); j++ {
			if days[i] > days[j] {
				days[i], days[j] = days[j], days[i]
			}
		}
	}

	for _, day := range days {
		daysUntil := (day - currentWeekday + 7) % 7

		if daysUntil > 0 {
			nextDate := now.AddDate(0, 0, daysUntil)
			return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), hour, minute, 0, 0, now.Location())
		}

		if daysUntil == 0 {
			today := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
			if today.After(now) {
				return today
			}
		}
	}

	firstDay := days[0]
	daysUntil := (firstDay - currentWeekday + 7) % 7
	if daysUntil == 0 {
		daysUntil = 7
	}
	nextDate := now.AddDate(0, 0, daysUntil)
	return time.Date(nextDate.Year(), nextDate.Month(), nextDate.Day(), hour, minute, 0, 0, now.Location())
}

func generateMCQQuestions(ctx context.Context, apiKey string, topic string, numQuestions int) ([]struct {
	Question string
	Answer   string
	Options  []string
}, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY not set")
	}

	prompt := fmt.Sprintf(`Generate EXACTLY %d multiple choice questions (MCQ) about "%s". 
Return JSON array with: [{"question": "...", "options": ["A", "B", "C", "D"], "answer": "A"}, ...]`, numQuestions, topic)

	response, err := GenerateGroqReply(ctx, apiKey, "", topic, prompt)
	if err != nil {
		return nil, err
	}

	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(strings.TrimSuffix(response, "```"), "```json")
		response = strings.TrimSpace(response)
	}

	var questionsJSON []struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
		Answer   string   `json:"answer"`
	}

	if err := json.Unmarshal([]byte(response), &questionsJSON); err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}

	questions := make([]struct {
		Question string
		Answer   string
		Options  []string
	}, len(questionsJSON))

	for i, q := range questionsJSON {
		questions[i] = struct {
			Question string
			Answer   string
			Options  []string
		}{
			Question: q.Question,
			Answer:   q.Answer,
			Options:  q.Options,
		}
	}

	return questions, nil
}

func generateSimpleMCQQuestions(topic string, numQuestions int) []struct {
	Question string
	Answer   string
	Options  []string
} {
	topic = strings.ToLower(topic)
	baseQuestions := []struct {
		Question string
		Answer   string
		Options  []string
	}{
		{Question: fmt.Sprintf("What is the main topic about %s?", topic), Options: []string{topic, "Something else", "Random", "Other"}, Answer: "A"},
		{Question: fmt.Sprintf("Which is relevant to %s?", topic), Options: []string{topic + " concepts", "Cooking", "Sports", "Weather"}, Answer: "A"},
		{Question: fmt.Sprintf("What did you learn about %s?", topic), Options: []string{"Key concepts", "Nothing", "Random", "Other"}, Answer: "A"},
	}

	questions := make([]struct {
		Question string
		Answer   string
		Options  []string
	}, numQuestions)
	for i := 0; i < numQuestions; i++ {
		questions[i] = baseQuestions[i%len(baseQuestions)]
	}
	return questions
}

func splitAndTrim(s, sep string) []string {
	parts := make([]string, 0)
	for _, p := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}
