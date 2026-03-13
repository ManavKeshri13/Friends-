package services

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// SendWhatsAppMessage sends a message via Twilio WhatsApp API
// Phone format: "whatsapp:+919876543210"
func SendWhatsAppMessage(to, message string) error {
	accountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	authToken := os.Getenv("TWILIO_AUTH_TOKEN")
	fromNumber := os.Getenv("TWILIO_WHATSAPP_FROM") // e.g. "whatsapp:+14155238886"

	if accountSID == "" || authToken == "" {
		return fmt.Errorf("Twilio credentials not set")
	}

	// Ensure "whatsapp:" prefix
	if !strings.HasPrefix(to, "whatsapp:") {
		to = "whatsapp:" + to
	}

	apiURL := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSID)

	msgData := url.Values{}
	msgData.Set("To", to)
	msgData.Set("From", fromNumber)
	msgData.Set("Body", message)

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(msgData.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth(accountSID, authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Twilio error: HTTP %d", resp.StatusCode)
	}
	return nil
}
