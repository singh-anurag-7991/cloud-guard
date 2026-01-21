package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type SlackWrapper struct {
	WebhookURL string
}

func NewSlackWrapper(webhookURL string) *SlackWrapper {
	return &SlackWrapper{WebhookURL: webhookURL}
}

func (s *SlackWrapper) SendAlert(finding models.Finding) error {
	if s.WebhookURL == "" {
		return nil
	}

	color := "#36a64f" // Green (Info)
	if finding.RiskLevel == "HIGH" {
		color = "#FF0000" // Red
	} else if finding.RiskLevel == "MEDIUM" {
		color = "#FFA500" // Orange
	} else if finding.RiskLevel == "LOW" {
		color = "#FFD700" // Gold
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":   color,
				"pretext": fmt.Sprintf("🚨 Cloud Guard Alert: %s Risk", finding.RiskLevel),
				"title":   fmt.Sprintf("%s (%s)", finding.Description, finding.ResourceID),
				"text":    fmt.Sprintf("%s\n\n*Recommendation:* %s", finding.Description, finding.Recommendation),
				"footer":  "Cloud Guard Scanner",
				"ts":      finding.GeneratedAt.Unix(),
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(s.WebhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send slack alert: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack api returned status: %d", resp.StatusCode)
	}

	return nil
}
