// Package main is the entry point for the stock signal bot.
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vicehope/stock-signal/pkg/config"
	sigindicator "github.com/vicehope/stock-signal/pkg/signal"
	"github.com/vicehope/stock-signal/pkg/stock"
	"github.com/vicehope/stock-signal/pkg/telegram"
)

// RequiredHistoricalDays is the number of days of historical data needed.
// This must be at least 201 days to support the 200-day moving average indicator,
// with extra buffer for crossover detection and other calculations.
const RequiredHistoricalDays = 250

func main() {
	log.Println("Starting Stock Signal Bot...")

	// Load configuration
	cfg := config.LoadFromEnv()

	// Validate configuration
	if missing := cfg.Validate(); len(missing) > 0 {
		log.Fatalf("Missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// Initialize clients
	stockClient := stock.NewClient(cfg.AlphaVantageAPIKey)
	telegramBot := telegram.NewBot(cfg.TelegramBotToken, cfg.TelegramChatID)
	analyzer := sigindicator.NewSignalAnalyzer()

	// Send startup message
	startupMsg := fmt.Sprintf("🚀 Stock Signal Bot started!\n\n"+
		"📊 Tracking: %s\n"+
		"⏰ Check interval: %d minutes\n"+
		"🔧 Indicators: MA(50/200), RSI(14), MACD, Stochastic, Price Action\n"+
		"📅 Started at: %s",
		cfg.Symbol,
		cfg.CheckIntervalMinutes,
		time.Now().Format("2006-01-02 15:04:05 MST"))

	if err := telegramBot.SendMessage(startupMsg); err != nil {
		log.Printf("Warning: Failed to send startup message: %v", err)
	}

	// Run initial check
	checkSignals(cfg.Symbol, stockClient, telegramBot, analyzer)

	// Set up ticker for regular checks
	ticker := time.NewTicker(time.Duration(cfg.CheckIntervalMinutes) * time.Minute)
	defer ticker.Stop()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Running signal checks every %d minutes for %s", cfg.CheckIntervalMinutes, cfg.Symbol)

	for {
		select {
		case <-ticker.C:
			checkSignals(cfg.Symbol, stockClient, telegramBot, analyzer)
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			if err := telegramBot.SendMessage("🛑 Stock Signal Bot is shutting down"); err != nil {
				log.Printf("Warning: Failed to send shutdown message: %v", err)
			}
			return
		}
	}
}

func checkSignals(symbol string, stockClient *stock.Client, telegramBot *telegram.Bot, analyzer *sigindicator.SignalAnalyzer) {
	log.Printf("Checking signals for %s...", symbol)

	// Get current quote
	quote, err := stockClient.GetQuote(symbol)
	if err != nil {
		log.Printf("Error fetching quote: %v", err)
		return
	}

	log.Printf("Current price for %s: $%.2f", symbol, quote.Price)

	// Get historical data (need enough for longest indicator - 200+ days for MA)
	historicalData, err := stockClient.GetHistoricalData(symbol, RequiredHistoricalDays)
	if err != nil {
		log.Printf("Error fetching historical data: %v", err)
		return
	}

	// Run all indicators
	signals, err := analyzer.AnalyzeAll(historicalData, quote.Price)
	if err != nil {
		log.Printf("Error analyzing signals: %v", err)
		return
	}

	// Get summary
	buyCount, sellCount, holdCount, overallSignal := analyzer.GetSummary(signals)

	// Format and send message
	message := formatSignalMessage(symbol, quote, signals, buyCount, sellCount, holdCount, overallSignal)
	if err := telegramBot.SendMessage(message); err != nil {
		log.Printf("Error sending message: %v", err)
	} else {
		log.Println("Signal message sent successfully")
	}
}

func formatSignalMessage(symbol string, quote *stock.Quote, signals []*sigindicator.Signal, buyCount, sellCount, holdCount int, overallSignal sigindicator.SignalType) string {
	var sb strings.Builder

	// Header with overall recommendation
	overallEmoji := getSignalEmoji(overallSignal)
	sb.WriteString(fmt.Sprintf("📊 %s Signal Report\n", symbol))
	sb.WriteString("═══════════════════════════\n\n")

	// Overall Signal Summary
	sb.WriteString(fmt.Sprintf("%s OVERALL: %s\n", overallEmoji, overallSignal))
	sb.WriteString(fmt.Sprintf("   🟢 Buy: %d | 🔴 Sell: %d | 🟡 Hold: %d\n\n", buyCount, sellCount, holdCount))

	// Price Information
	sb.WriteString("💰 Price Data:\n")
	sb.WriteString(fmt.Sprintf("   Current: $%.2f\n", quote.Price))
	sb.WriteString(fmt.Sprintf("   Open: $%.2f | High: $%.2f | Low: $%.2f\n", quote.Open, quote.High, quote.Low))
	sb.WriteString(fmt.Sprintf("   Volume: %d\n", quote.Volume))
	sb.WriteString(fmt.Sprintf("   Date: %s\n\n", quote.Timestamp.Format("2006-01-02")))

	// Detailed Signal Analysis
	sb.WriteString("🔍 Indicator Signals:\n")
	sb.WriteString("───────────────────────────\n")

	for _, sig := range signals {
		emoji := getSignalEmoji(sig.Type)
		sb.WriteString(fmt.Sprintf("%s %s: %s\n", emoji, sig.Indicator, sig.Type))
		sb.WriteString(fmt.Sprintf("   └ %s\n", sig.Reason))
		if sig.Confidence > 0 {
			sb.WriteString(fmt.Sprintf("   └ Confidence: %.0f%%\n", sig.Confidence*100))
		}
	}

	sb.WriteString(fmt.Sprintf("\n⏰ Updated: %s", time.Now().Format("2006-01-02 15:04:05 MST")))

	return sb.String()
}

func getSignalEmoji(signalType sigindicator.SignalType) string {
	switch signalType {
	case sigindicator.Buy:
		return "🟢"
	case sigindicator.Sell:
		return "🔴"
	default:
		return "🟡"
	}
}
