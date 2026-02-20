package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
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
		logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger.Error().Err(err).Msg("discord client failed")
	}
}

func (dc *DiscordClient) SendError(message string) {
	err := dc.sendMessage(message, 16711680)
	if err != nil {
		logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
		logger.Error().Err(err).Msg("discord client failed")
	}
}

type logEvent struct {
	Level string `json:"level"`
}

func (dc *DiscordClient) Write(p []byte) (n int, err error) {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()
	var logEvent logEvent
	if err := json.Unmarshal(p, &logEvent); err != nil {
		logger.Error().Err(err).Msg("cannot unmarshal log event")
		return len(p), err
	}
	level, err := zerolog.ParseLevel(logEvent.Level)
	if err != nil {
		logger.Error().Err(err).Msg("cannot parse log level")
		return len(p), err
	}
	if level == zerolog.ErrorLevel || level == zerolog.FatalLevel || level == zerolog.PanicLevel {
		dc.SendError(string(p))
	}
	return len(p), nil
}

func (dc *DiscordClient) Close() error {
	return nil
}
