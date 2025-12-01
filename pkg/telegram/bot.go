// Package telegram provides functionality to send messages via Telegram bot.
package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Bot represents a Telegram bot client.
type Bot struct {
	token      string
	chatID     string
	httpClient *http.Client
	baseURL    string
}

// NewBot creates a new Telegram bot client.
func NewBot(token, chatID string) *Bot {
	return &Bot{
		token:  token,
		chatID: chatID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.telegram.org",
	}
}

// sendMessageRequest represents a request to send a message.
type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

// telegramResponse represents the API response from Telegram.
type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description,omitempty"`
}

// SendMessage sends a text message to the configured chat.
func (b *Bot) SendMessage(text string) error {
	return b.sendMessageWithParseMode(text, "")
}

// SendMarkdownMessage sends a message with Markdown formatting.
func (b *Bot) SendMarkdownMessage(text string) error {
	return b.sendMessageWithParseMode(text, "Markdown")
}

// SendHTMLMessage sends a message with HTML formatting.
func (b *Bot) SendHTMLMessage(text string) error {
	return b.sendMessageWithParseMode(text, "HTML")
}

func (b *Bot) sendMessageWithParseMode(text, parseMode string) error {
	url := fmt.Sprintf("%s/bot%s/sendMessage", b.baseURL, b.token)

	reqBody := sendMessageRequest{
		ChatID:    b.chatID,
		Text:      text,
		ParseMode: parseMode,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := b.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var telegramResp telegramResponse
	if err := json.Unmarshal(body, &telegramResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if !telegramResp.OK {
		return fmt.Errorf("telegram API error: %s", telegramResp.Description)
	}

	return nil
}
