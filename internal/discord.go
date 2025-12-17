package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/ansi"
)

type DiscordClient struct {
	webhookUrl *string
}

func NewDiscordClient(config AppConfig) *DiscordClient {
	if config.DiscordWebhookUrl == "" {
		return &DiscordClient{webhookUrl: nil}
	}
	return &DiscordClient{webhookUrl: &config.DiscordWebhookUrl}
}

func (dc *DiscordClient) sendMessage(message string, color uint) error {
	if dc.webhookUrl == nil {
		return nil
	}
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       "Puff",
				"description": message,
				"color":       color,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord err: %s", err)
	}

	resp, err := http.Post(*dc.webhookUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("discord err: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("discord err: response code %d", resp.StatusCode)
	}
	return nil
}

func (dc *DiscordClient) SendInfo(message string) {
	err := dc.sendMessage(message, 3447003)
	if err != nil {
		log.New(os.Stdout).Error("discord client failed", "err", err)
	}
}

func (dc *DiscordClient) SendError(message string) {
	err := dc.sendMessage(message, 16711680)
	if err != nil {
		log.New(os.Stdout).Error("discord client failed", "err", err)
	}
}

func (dc *DiscordClient) Write(p []byte) (n int, err error) {
	if regexp.MustCompile(`(?i)(ERRO|FATA)`).Match(p) {
		cleaned := ansi.Strip(string(p))
		dc.SendError(cleaned)
	}
	return len(p), nil
}

func (dc *DiscordClient) Close() error {
	return nil
}
