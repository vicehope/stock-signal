# Stock Signal Bot

A Go application that monitors VOO (Vanguard S&P 500 ETF) stock prices and sends buy/sell signal indicators to Telegram.

## Features

- **Real-time Stock Data**: Fetches current and historical stock prices using Alpha Vantage API
- **Comprehensive Signal Indicators**:
  - Moving Averages (50/200-day) with Golden/Death Cross detection
  - Relative Strength Index (RSI) with overbought/oversold detection
  - MACD (Moving Average Convergence Divergence) with signal line crossover
  - Stochastic Oscillator with %K/%D crossover detection
  - Price Action analysis (breakout/breakdown, trend analysis)
- **Telegram Notifications**: Sends formatted signal updates to your Telegram channel
- **Signal Summary**: Aggregates all indicators to provide overall buy/sell/hold recommendation
- **Extensible Framework**: Easy to add new indicators

## Prerequisites

- Go 1.24 or later
- Alpha Vantage API key (free tier available at https://www.alphavantage.co/support/#api-key)
- Telegram Bot token (create via [@BotFather](https://t.me/botfather))
- Telegram Chat ID where messages will be sent

## Installation

```bash
# Clone the repository
git clone https://github.com/vicehope/stock-signal.git
cd stock-signal

# Build the application
go build -o signal-bot ./cmd/signal-bot
```

## Configuration

Set the following environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `TELEGRAM_BOT_TOKEN` | Yes | Your Telegram bot token |
| `TELEGRAM_CHAT_ID` | Yes | Chat ID to send messages to |
| `ALPHA_VANTAGE_API_KEY` | Yes | Your Alpha Vantage API key |
| `STOCK_SYMBOL` | No | Stock symbol to track (default: VOO) |

## Usage

```bash
# Set environment variables
export TELEGRAM_BOT_TOKEN="your-bot-token"
export TELEGRAM_CHAT_ID="your-chat-id"
export ALPHA_VANTAGE_API_KEY="your-api-key"

# Run the bot
./signal-bot
```

### Docker (Optional)

```bash
docker build -t stock-signal .
docker run -e TELEGRAM_BOT_TOKEN=xxx -e TELEGRAM_CHAT_ID=xxx -e ALPHA_VANTAGE_API_KEY=xxx stock-signal
```

## Signal Indicators

### Moving Averages (50/200-day)

| Signal | Condition |
|--------|-----------|
| **Buy** | Golden Cross: 50-day SMA crosses above 200-day SMA, or price above both MAs |
| **Sell** | Death Cross: 50-day SMA crosses below 200-day SMA, or price below MAs |
| **Hold** | Price near moving averages - indecisive trend |

### RSI (14-day)

| Signal | Condition |
|--------|-----------|
| **Buy** | RSI below 30 (oversold), especially if momentum is turning up |
| **Sell** | RSI above 70 (overbought), especially if momentum is weakening |
| **Hold** | RSI between 30-70 (neutral/consolidating) |

### MACD (12/26/9)

| Signal | Condition |
|--------|-----------|
| **Buy** | MACD line crosses above signal line (bullish crossover) |
| **Sell** | MACD line crosses below signal line (bearish crossover) |
| **Hold** | MACD near zero or flat - no clear momentum |

### Stochastic Oscillator (14,3)

| Signal | Condition |
|--------|-----------|
| **Buy** | %K below 20 (oversold) with bullish crossover from below |
| **Sell** | %K above 80 (overbought) with bearish crossover from above |
| **Hold** | %K between 20-80 (neutral) |

### Price Action (20-day)

| Signal | Condition |
|--------|-----------|
| **Buy** | Breakout above resistance, or uptrend with higher highs/lows |
| **Sell** | Breakdown below support, or downtrend with lower highs/lows |
| **Hold** | Consolidating between support and resistance |

## Project Structure

```
stock-signal/
├── cmd/
│   └── signal-bot/      # Main application entry point
│       └── main.go
├── pkg/
│   ├── config/          # Configuration management
│   │   └── config.go
│   ├── signal/          # Signal indicators
│   │   ├── indicator.go
│   │   └── indicator_test.go
│   ├── stock/           # Stock data client
│   │   └── client.go
│   └── telegram/        # Telegram bot client
│       └── bot.go
├── go.mod
└── README.md
```

## Adding Custom Indicators

Implement the `Indicator` interface to add custom indicators:

```go
type Indicator interface {
    Name() string
    Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error)
}
```

Example:
```go
// Use the comprehensive analyzer (default)
analyzer := signal.NewSignalAnalyzer()

// Or use basic analyzer with just SMA and RSI
analyzer := signal.NewBasicSignalAnalyzer()

// Add custom indicators
analyzer.AddIndicator(signal.NewCustomMovingAverageIndicator(10, 50))
analyzer.AddIndicator(signal.NewCustomMACDIndicator(8, 17, 9))
analyzer.AddIndicator(signal.NewCustomStochasticIndicator(10, 5, 85, 15))
```

## API Rate Limits

Alpha Vantage free tier allows 25 API calls per day. The bot checks signals hourly by default, which uses 2 calls per check (quote + historical data).

## Sample Telegram Message

```
📊 VOO Signal Report
═══════════════════════════

🟢 OVERALL: BUY
   🟢 Buy: 3 | 🔴 Sell: 1 | 🟡 Hold: 1

💰 Price Data:
   Current: $450.25
   Open: $448.50 | High: $451.00 | Low: $447.80
   Volume: 5234567
   Date: 2024-01-15

🔍 Indicator Signals:
───────────────────────────
🟢 Moving Average (50/200): BUY
   └ Price ($450.25) above both MAs. 50-day: $445.30, 200-day: $430.15
   └ Confidence: 60%
🟡 RSI (14): HOLD
   └ RSI (55.32) neutral between 30-70
🟢 MACD (12/26/9): BUY
   └ MACD (2.15) above Signal (1.80) - bullish momentum
   └ Confidence: 50%
🔴 Stochastic (14,3): SELL
   └ Stochastic overbought (%K: 82.50 above 80)
   └ Confidence: 60%
🟢 Price Action (20-day): BUY
   └ Uptrend confirmed (higher highs/lows). Strength: 70%
   └ Confidence: 60%

⏰ Updated: 2024-01-15 14:30:00 EST
```

## License

MIT