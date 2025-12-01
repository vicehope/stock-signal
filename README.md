# Stock Signal Bot

A Go application that monitors VOO (Vanguard S&P 500 ETF) stock prices and sends buy/sell signal indicators to Telegram.

## Features

- **Real-time Stock Data**: Fetches current and historical stock prices using Alpha Vantage API
- **Multiple Signal Indicators**:
  - Simple Moving Average (SMA) Crossover (10/50 day)
  - Relative Strength Index (RSI) with overbought/oversold detection
- **Telegram Notifications**: Sends formatted signal updates to your Telegram channel
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

### SMA Crossover (10/50)
- **Buy Signal**: When short-term SMA crosses above long-term SMA (Golden Cross)
- **Sell Signal**: When short-term SMA crosses below long-term SMA (Death Cross)

### RSI (14-day)
- **Buy Signal**: RSI drops below 30 (oversold)
- **Sell Signal**: RSI rises above 70 (overbought)

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
analyzer := signal.NewSignalAnalyzer()
analyzer.AddIndicator(signal.NewSMAIndicator(5, 20))
```

## API Rate Limits

Alpha Vantage free tier allows 25 API calls per day. The bot checks signals hourly by default, which uses 2 calls per check (quote + historical data).

## License

MIT