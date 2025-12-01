// Package stock provides functionality to fetch stock price data.
package stock

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// Client is a client for fetching stock data.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// Quote represents a stock quote with price information.
type Quote struct {
	Symbol    string
	Price     float64
	Open      float64
	High      float64
	Low       float64
	Volume    int64
	Timestamp time.Time
}

// HistoricalData represents historical price data for a stock.
type HistoricalData struct {
	Symbol string
	Data   []DailyPrice
}

// DailyPrice represents a single day's price data.
type DailyPrice struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// NewClient creates a new stock data client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://www.alphavantage.co/query",
	}
}

// alphaVantageGlobalQuote represents the API response for global quote.
type alphaVantageGlobalQuote struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
	} `json:"Global Quote"`
}

// alphaVantageTimeSeriesDaily represents the API response for daily time series.
type alphaVantageTimeSeriesDaily struct {
	MetaData struct {
		Symbol string `json:"2. Symbol"`
	} `json:"Meta Data"`
	TimeSeries map[string]struct {
		Open   string `json:"1. open"`
		High   string `json:"2. high"`
		Low    string `json:"3. low"`
		Close  string `json:"4. close"`
		Volume string `json:"5. volume"`
	} `json:"Time Series (Daily)"`
}

// GetQuote fetches the current quote for a symbol.
func (c *Client) GetQuote(symbol string) (*Quote, error) {
	url := fmt.Sprintf("%s?function=GLOBAL_QUOTE&symbol=%s&apikey=%s", c.baseURL, symbol, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var avQuote alphaVantageGlobalQuote
	if err := json.Unmarshal(body, &avQuote); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if avQuote.GlobalQuote.Symbol == "" {
		return nil, fmt.Errorf("no quote data returned for symbol %s (API limit may be reached)", symbol)
	}

	price, _ := strconv.ParseFloat(avQuote.GlobalQuote.Price, 64)
	open, _ := strconv.ParseFloat(avQuote.GlobalQuote.Open, 64)
	high, _ := strconv.ParseFloat(avQuote.GlobalQuote.High, 64)
	low, _ := strconv.ParseFloat(avQuote.GlobalQuote.Low, 64)
	volume, _ := strconv.ParseInt(avQuote.GlobalQuote.Volume, 10, 64)

	tradingDay, _ := time.Parse("2006-01-02", avQuote.GlobalQuote.LatestTradingDay)

	return &Quote{
		Symbol:    symbol,
		Price:     price,
		Open:      open,
		High:      high,
		Low:       low,
		Volume:    volume,
		Timestamp: tradingDay,
	}, nil
}

// GetHistoricalData fetches historical daily price data for a symbol.
func (c *Client) GetHistoricalData(symbol string, days int) (*HistoricalData, error) {
	outputSize := "compact"
	if days > 100 {
		outputSize = "full"
	}

	url := fmt.Sprintf("%s?function=TIME_SERIES_DAILY&symbol=%s&outputsize=%s&apikey=%s",
		c.baseURL, symbol, outputSize, c.apiKey)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var avData alphaVantageTimeSeriesDaily
	if err := json.Unmarshal(body, &avData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(avData.TimeSeries) == 0 {
		return nil, fmt.Errorf("no historical data returned for symbol %s (API limit may be reached)", symbol)
	}

	var prices []DailyPrice
	for dateStr, data := range avData.TimeSeries {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		open, _ := strconv.ParseFloat(data.Open, 64)
		high, _ := strconv.ParseFloat(data.High, 64)
		low, _ := strconv.ParseFloat(data.Low, 64)
		closePrice, _ := strconv.ParseFloat(data.Close, 64)
		volume, _ := strconv.ParseInt(data.Volume, 10, 64)

		prices = append(prices, DailyPrice{
			Date:   date,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		})
	}

	// Sort by date descending (most recent first)
	sort.Slice(prices, func(i, j int) bool {
		return prices[i].Date.After(prices[j].Date)
	})

	// Limit to requested days
	if len(prices) > days {
		prices = prices[:days]
	}

	return &HistoricalData{
		Symbol: symbol,
		Data:   prices,
	}, nil
}
