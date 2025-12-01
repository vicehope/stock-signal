// Package config provides configuration management for the stock signal application.
package config

import (
	"os"
)

// Config holds all configuration for the application.
type Config struct {
	// TelegramBotToken is the Telegram bot API token
	TelegramBotToken string
	// TelegramChatID is the chat ID to send messages to
	TelegramChatID string
	// AlphaVantageAPIKey is the API key for Alpha Vantage stock data
	AlphaVantageAPIKey string
	// Symbol is the stock symbol to track (default: VOO)
	Symbol string
	// CheckIntervalMinutes is how often to check for signals in minutes
	CheckIntervalMinutes int
}

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *Config {
	symbol := os.Getenv("STOCK_SYMBOL")
	if symbol == "" {
		symbol = "VOO"
	}

	return &Config{
		TelegramBotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:       os.Getenv("TELEGRAM_CHAT_ID"),
		AlphaVantageAPIKey:   os.Getenv("ALPHA_VANTAGE_API_KEY"),
		Symbol:               symbol,
		CheckIntervalMinutes: 60, // Default to hourly checks
	}
}

// Validate checks if all required configuration is present.
func (c *Config) Validate() []string {
	var missing []string
	if c.TelegramBotToken == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if c.TelegramChatID == "" {
		missing = append(missing, "TELEGRAM_CHAT_ID")
	}
	if c.AlphaVantageAPIKey == "" {
		missing = append(missing, "ALPHA_VANTAGE_API_KEY")
	}
	return missing
}
