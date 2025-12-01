// Package signal provides buy/sell signal indicators for stocks.
package signal

import (
	"fmt"
	"time"

	"github.com/vicehope/stock-signal/pkg/stock"
)

// SignalType represents the type of trading signal.
type SignalType string

const (
	// Buy indicates a buy signal.
	Buy SignalType = "BUY"
	// Sell indicates a sell signal.
	Sell SignalType = "SELL"
	// Hold indicates no action should be taken.
	Hold SignalType = "HOLD"
)

// Signal represents a trading signal with metadata.
type Signal struct {
	Type       SignalType
	Symbol     string
	Price      float64
	Reason     string
	Indicator  string
	Timestamp  time.Time
	Confidence float64 // 0.0 to 1.0
}

// Indicator is an interface for signal indicators.
type Indicator interface {
	// Name returns the name of the indicator.
	Name() string
	// Analyze analyzes historical data and returns a signal.
	Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error)
}

// SMAIndicator implements a Simple Moving Average crossover indicator.
type SMAIndicator struct {
	shortPeriod int
	longPeriod  int
}

// NewSMAIndicator creates a new SMA crossover indicator.
func NewSMAIndicator(shortPeriod, longPeriod int) *SMAIndicator {
	return &SMAIndicator{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
	}
}

// Name returns the name of the indicator.
func (s *SMAIndicator) Name() string {
	return fmt.Sprintf("SMA Crossover (%d/%d)", s.shortPeriod, s.longPeriod)
}

// Analyze implements the Indicator interface.
func (s *SMAIndicator) Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error) {
	if len(data.Data) < s.longPeriod {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", s.longPeriod, len(data.Data))
	}

	// Calculate SMAs
	shortSMA := calculateSMA(data.Data, s.shortPeriod)
	longSMA := calculateSMA(data.Data, s.longPeriod)

	// Calculate previous day's SMAs for crossover detection
	prevShortSMA := calculateSMAOffset(data.Data, s.shortPeriod, 1)
	prevLongSMA := calculateSMAOffset(data.Data, s.longPeriod, 1)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: s.Name(),
		Timestamp: time.Now(),
	}

	// Detect crossover
	if shortSMA > longSMA && prevShortSMA <= prevLongSMA {
		// Golden cross - bullish signal
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Golden Cross: Short SMA (%.2f) crossed above Long SMA (%.2f)", shortSMA, longSMA)
		signal.Confidence = 0.7
	} else if shortSMA < longSMA && prevShortSMA >= prevLongSMA {
		// Death cross - bearish signal
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Death Cross: Short SMA (%.2f) crossed below Long SMA (%.2f)", shortSMA, longSMA)
		signal.Confidence = 0.7
	} else {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("No crossover. Short SMA: %.2f, Long SMA: %.2f", shortSMA, longSMA)
		signal.Confidence = 0.0
	}

	return signal, nil
}

// calculateSMA calculates the Simple Moving Average for the most recent period.
func calculateSMA(data []stock.DailyPrice, period int) float64 {
	return calculateSMAOffset(data, period, 0)
}

// calculateSMAOffset calculates the SMA with an offset from the most recent data.
func calculateSMAOffset(data []stock.DailyPrice, period, offset int) float64 {
	if len(data) < period+offset {
		return 0
	}

	sum := 0.0
	for i := offset; i < period+offset; i++ {
		sum += data[i].Close
	}
	return sum / float64(period)
}

// RSIIndicator implements a Relative Strength Index indicator.
type RSIIndicator struct {
	period         int
	overbought     float64
	oversold       float64
}

// NewRSIIndicator creates a new RSI indicator.
func NewRSIIndicator(period int, overbought, oversold float64) *RSIIndicator {
	return &RSIIndicator{
		period:     period,
		overbought: overbought,
		oversold:   oversold,
	}
}

// Name returns the name of the indicator.
func (r *RSIIndicator) Name() string {
	return fmt.Sprintf("RSI (%d)", r.period)
}

// Analyze implements the Indicator interface.
func (r *RSIIndicator) Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error) {
	if len(data.Data) < r.period+1 {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", r.period+1, len(data.Data))
	}

	rsi := calculateRSI(data.Data, r.period)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: r.Name(),
		Timestamp: time.Now(),
	}

	if rsi <= r.oversold {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("RSI (%.2f) is in oversold territory (below %.2f)", rsi, r.oversold)
		signal.Confidence = 0.6 + (r.oversold-rsi)/100*0.4
	} else if rsi >= r.overbought {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("RSI (%.2f) is in overbought territory (above %.2f)", rsi, r.overbought)
		signal.Confidence = 0.6 + (rsi-r.overbought)/100*0.4
	} else {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("RSI (%.2f) is in neutral territory", rsi)
		signal.Confidence = 0.0
	}

	return signal, nil
}

// calculateRSI calculates the Relative Strength Index.
func calculateRSI(data []stock.DailyPrice, period int) float64 {
	if len(data) < period+1 {
		return 50 // neutral
	}

	var gains, losses float64

	// Data is sorted descending (most recent first), so we need to reverse for calculation
	for i := 0; i < period; i++ {
		change := data[i].Close - data[i+1].Close
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		return 100
	}

	rs := avgGain / avgLoss
	rsi := 100 - (100 / (1 + rs))

	return rsi
}

// SignalAnalyzer combines multiple indicators to generate signals.
type SignalAnalyzer struct {
	indicators []Indicator
}

// NewSignalAnalyzer creates a new signal analyzer with default indicators.
func NewSignalAnalyzer() *SignalAnalyzer {
	return &SignalAnalyzer{
		indicators: []Indicator{
			NewSMAIndicator(10, 50),  // Short-term crossover
			NewRSIIndicator(14, 70, 30), // Standard RSI
		},
	}
}

// AddIndicator adds a custom indicator to the analyzer.
func (a *SignalAnalyzer) AddIndicator(indicator Indicator) {
	a.indicators = append(a.indicators, indicator)
}

// AnalyzeAll runs all indicators and returns their signals.
func (a *SignalAnalyzer) AnalyzeAll(data *stock.HistoricalData, currentPrice float64) ([]*Signal, error) {
	var signals []*Signal

	for _, indicator := range a.indicators {
		signal, err := indicator.Analyze(data, currentPrice)
		if err != nil {
			// Log error but continue with other indicators
			continue
		}
		signals = append(signals, signal)
	}

	return signals, nil
}
