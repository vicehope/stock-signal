// Package signal provides buy/sell signal indicators for stocks.
package signal

import (
	"fmt"
	"math"
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

// =============================================================================
// Moving Average Indicator (50/200-day with Golden/Death Cross)
// =============================================================================

// MovingAverageIndicator implements Moving Average analysis with Golden/Death Cross.
type MovingAverageIndicator struct {
	shortPeriod int // typically 50
	longPeriod  int // typically 200
}

// NewMovingAverageIndicator creates a new MA indicator with 50/200-day periods.
func NewMovingAverageIndicator() *MovingAverageIndicator {
	return &MovingAverageIndicator{
		shortPeriod: 50,
		longPeriod:  200,
	}
}

// NewCustomMovingAverageIndicator creates a MA indicator with custom periods.
func NewCustomMovingAverageIndicator(shortPeriod, longPeriod int) *MovingAverageIndicator {
	return &MovingAverageIndicator{
		shortPeriod: shortPeriod,
		longPeriod:  longPeriod,
	}
}

// Name returns the name of the indicator.
func (m *MovingAverageIndicator) Name() string {
	return fmt.Sprintf("Moving Average (%d/%d)", m.shortPeriod, m.longPeriod)
}

// Analyze implements the Indicator interface.
func (m *MovingAverageIndicator) Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error) {
	if len(data.Data) < m.longPeriod+1 {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", m.longPeriod+1, len(data.Data))
	}

	// Calculate current SMAs
	shortSMA := calculateSMA(data.Data, m.shortPeriod)
	longSMA := calculateSMA(data.Data, m.longPeriod)

	// Calculate previous day's SMAs for crossover detection
	prevShortSMA := calculateSMAOffset(data.Data, m.shortPeriod, 1)
	prevLongSMA := calculateSMAOffset(data.Data, m.longPeriod, 1)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: m.Name(),
		Timestamp: time.Now(),
	}

	// Check for Golden Cross (bullish) or Death Cross (bearish)
	goldenCross := shortSMA > longSMA && prevShortSMA <= prevLongSMA
	deathCross := shortSMA < longSMA && prevShortSMA >= prevLongSMA
	priceAboveBothMAs := currentPrice > shortSMA && currentPrice > longSMA
	priceBelowBothMAs := currentPrice < shortSMA || currentPrice < longSMA

	if goldenCross {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Golden Cross: %d-day SMA (%.2f) crossed above %d-day SMA (%.2f)", 
			m.shortPeriod, shortSMA, m.longPeriod, longSMA)
		signal.Confidence = 0.8
	} else if deathCross {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Death Cross: %d-day SMA (%.2f) crossed below %d-day SMA (%.2f)", 
			m.shortPeriod, shortSMA, m.longPeriod, longSMA)
		signal.Confidence = 0.8
	} else if priceAboveBothMAs && shortSMA > longSMA {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Price (%.2f) above both MAs. %d-day: %.2f, %d-day: %.2f", 
			currentPrice, m.shortPeriod, shortSMA, m.longPeriod, longSMA)
		signal.Confidence = 0.6
	} else if priceBelowBothMAs && shortSMA < longSMA {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Price (%.2f) below MAs. %d-day: %.2f, %d-day: %.2f", 
			currentPrice, m.shortPeriod, shortSMA, m.longPeriod, longSMA)
		signal.Confidence = 0.6
	} else {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("Price near MAs - indecisive. Price: %.2f, %d-day: %.2f, %d-day: %.2f", 
			currentPrice, m.shortPeriod, shortSMA, m.longPeriod, longSMA)
		signal.Confidence = 0.0
	}

	return signal, nil
}

// =============================================================================
// RSI Indicator (Relative Strength Index)
// =============================================================================

// RSIIndicator implements a Relative Strength Index indicator.
type RSIIndicator struct {
	period     int
	overbought float64
	oversold   float64
}

// NewRSIIndicator creates a new RSI indicator with standard settings (14, 70, 30).
func NewRSIIndicator(period int, overbought, oversold float64) *RSIIndicator {
	return &RSIIndicator{
		period:     period,
		overbought: overbought,
		oversold:   oversold,
	}
}

// NewStandardRSIIndicator creates a new RSI indicator with standard 14-period.
func NewStandardRSIIndicator() *RSIIndicator {
	return &RSIIndicator{
		period:     14,
		overbought: 70,
		oversold:   30,
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
	prevRSI := calculateRSIOffset(data.Data, r.period, 1)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: r.Name(),
		Timestamp: time.Now(),
	}

	// Check momentum direction
	momentumUp := rsi > prevRSI
	momentumDown := rsi < prevRSI

	if rsi <= r.oversold {
		signal.Type = Buy
		momentum := ""
		if momentumUp {
			momentum = " (momentum turning up)"
			signal.Confidence = 0.75
		} else {
			signal.Confidence = 0.6
		}
		signal.Reason = fmt.Sprintf("RSI (%.2f) oversold below %.0f%s", rsi, r.oversold, momentum)
	} else if rsi >= r.overbought {
		signal.Type = Sell
		momentum := ""
		if momentumDown {
			momentum = " (momentum weakening)"
			signal.Confidence = 0.75
		} else {
			signal.Confidence = 0.6
		}
		signal.Reason = fmt.Sprintf("RSI (%.2f) overbought above %.0f%s", rsi, r.overbought, momentum)
	} else {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("RSI (%.2f) neutral between %.0f-%.0f", rsi, r.oversold, r.overbought)
		signal.Confidence = 0.0
	}

	return signal, nil
}

// =============================================================================
// MACD Indicator (Moving Average Convergence Divergence)
// =============================================================================

// MACDIndicator implements MACD with signal line crossover detection.
type MACDIndicator struct {
	fastPeriod   int // typically 12
	slowPeriod   int // typically 26
	signalPeriod int // typically 9
}

// NewMACDIndicator creates a new MACD indicator with standard settings.
func NewMACDIndicator() *MACDIndicator {
	return &MACDIndicator{
		fastPeriod:   12,
		slowPeriod:   26,
		signalPeriod: 9,
	}
}

// NewCustomMACDIndicator creates a MACD indicator with custom periods.
func NewCustomMACDIndicator(fast, slow, signal int) *MACDIndicator {
	return &MACDIndicator{
		fastPeriod:   fast,
		slowPeriod:   slow,
		signalPeriod: signal,
	}
}

// Name returns the name of the indicator.
func (m *MACDIndicator) Name() string {
	return fmt.Sprintf("MACD (%d/%d/%d)", m.fastPeriod, m.slowPeriod, m.signalPeriod)
}

// Analyze implements the Indicator interface.
func (m *MACDIndicator) Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error) {
	minRequired := m.slowPeriod + m.signalPeriod + 1
	if len(data.Data) < minRequired {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", minRequired, len(data.Data))
	}

	// Calculate current MACD values
	macdLine, signalLine, histogram := calculateMACD(data.Data, m.fastPeriod, m.slowPeriod, m.signalPeriod)
	
	// Calculate previous day's values for crossover detection
	prevMACDLine, prevSignalLine, _ := calculateMACDOffset(data.Data, m.fastPeriod, m.slowPeriod, m.signalPeriod, 1)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: m.Name(),
		Timestamp: time.Now(),
	}

	// Detect crossovers
	bullishCrossover := macdLine > signalLine && prevMACDLine <= prevSignalLine
	bearishCrossover := macdLine < signalLine && prevMACDLine >= prevSignalLine
	nearZero := math.Abs(macdLine) < 0.5 && math.Abs(histogram) < 0.3

	if bullishCrossover {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("MACD bullish crossover: MACD (%.2f) crossed above Signal (%.2f)", macdLine, signalLine)
		signal.Confidence = 0.75
	} else if bearishCrossover {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("MACD bearish crossover: MACD (%.2f) crossed below Signal (%.2f)", macdLine, signalLine)
		signal.Confidence = 0.75
	} else if nearZero {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("MACD near zero (%.2f) - no clear momentum", macdLine)
		signal.Confidence = 0.0
	} else if macdLine > signalLine {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("MACD (%.2f) above Signal (%.2f) - bullish momentum", macdLine, signalLine)
		signal.Confidence = 0.5
	} else {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("MACD (%.2f) below Signal (%.2f) - bearish momentum", macdLine, signalLine)
		signal.Confidence = 0.5
	}

	return signal, nil
}

// =============================================================================
// Stochastic Oscillator Indicator
// =============================================================================

// StochasticIndicator implements Stochastic Oscillator with %K and %D.
type StochasticIndicator struct {
	kPeriod    int     // typically 14
	dPeriod    int     // typically 3
	overbought float64 // typically 80
	oversold   float64 // typically 20
}

// NewStochasticIndicator creates a new Stochastic indicator with standard settings.
func NewStochasticIndicator() *StochasticIndicator {
	return &StochasticIndicator{
		kPeriod:    14,
		dPeriod:    3,
		overbought: 80,
		oversold:   20,
	}
}

// NewCustomStochasticIndicator creates a Stochastic indicator with custom settings.
func NewCustomStochasticIndicator(kPeriod, dPeriod int, overbought, oversold float64) *StochasticIndicator {
	return &StochasticIndicator{
		kPeriod:    kPeriod,
		dPeriod:    dPeriod,
		overbought: overbought,
		oversold:   oversold,
	}
}

// Name returns the name of the indicator.
func (s *StochasticIndicator) Name() string {
	return fmt.Sprintf("Stochastic (%d,%d)", s.kPeriod, s.dPeriod)
}

// Analyze implements the Indicator interface.
func (s *StochasticIndicator) Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error) {
	minRequired := s.kPeriod + s.dPeriod + 1
	if len(data.Data) < minRequired {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", minRequired, len(data.Data))
	}

	// Calculate current %K and %D
	k, d := calculateStochastic(data.Data, s.kPeriod, s.dPeriod)
	
	// Calculate previous values for crossover detection
	prevK, prevD := calculateStochasticOffset(data.Data, s.kPeriod, s.dPeriod, 1)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: s.Name(),
		Timestamp: time.Now(),
	}

	// Detect crossovers in oversold/overbought regions
	bullishCrossover := k > d && prevK <= prevD
	bearishCrossover := k < d && prevK >= prevD

	if k <= s.oversold && bullishCrossover {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Stochastic oversold (%%K: %.2f, %%D: %.2f) with bullish crossover", k, d)
		signal.Confidence = 0.75
	} else if k <= s.oversold {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Stochastic oversold (%%K: %.2f below %.0f)", k, s.oversold)
		signal.Confidence = 0.6
	} else if k >= s.overbought && bearishCrossover {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Stochastic overbought (%%K: %.2f, %%D: %.2f) with bearish crossover", k, d)
		signal.Confidence = 0.75
	} else if k >= s.overbought {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Stochastic overbought (%%K: %.2f above %.0f)", k, s.overbought)
		signal.Confidence = 0.6
	} else {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("Stochastic neutral (%%K: %.2f, %%D: %.2f)", k, d)
		signal.Confidence = 0.0
	}

	return signal, nil
}

// =============================================================================
// Price Action Indicator (Support/Resistance, Trend Analysis)
// =============================================================================

// PriceActionIndicator analyzes price action for breakouts and trend direction.
type PriceActionIndicator struct {
	lookbackPeriod int // period to analyze for support/resistance
}

// NewPriceActionIndicator creates a new Price Action indicator.
func NewPriceActionIndicator() *PriceActionIndicator {
	return &PriceActionIndicator{
		lookbackPeriod: 20,
	}
}

// NewCustomPriceActionIndicator creates a Price Action indicator with custom lookback.
func NewCustomPriceActionIndicator(lookbackPeriod int) *PriceActionIndicator {
	return &PriceActionIndicator{
		lookbackPeriod: lookbackPeriod,
	}
}

// Name returns the name of the indicator.
func (p *PriceActionIndicator) Name() string {
	return fmt.Sprintf("Price Action (%d-day)", p.lookbackPeriod)
}

// Analyze implements the Indicator interface.
func (p *PriceActionIndicator) Analyze(data *stock.HistoricalData, currentPrice float64) (*Signal, error) {
	if len(data.Data) < p.lookbackPeriod+5 {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", p.lookbackPeriod+5, len(data.Data))
	}

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: p.Name(),
		Timestamp: time.Now(),
	}

	// Find support and resistance levels
	resistance, support := findSupportResistance(data.Data, p.lookbackPeriod)
	
	// Analyze trend using higher highs/lows or lower highs/lows
	trend, trendStrength := analyzeTrend(data.Data, 10) // Use last 10 days for trend

	// Check for breakout/breakdown
	breakoutThreshold := 0.02 // 2% above resistance
	breakdownThreshold := 0.02 // 2% below support

	if currentPrice > resistance*(1+breakoutThreshold) {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Breakout above resistance (%.2f). Current: %.2f", resistance, currentPrice)
		signal.Confidence = 0.7
	} else if currentPrice < support*(1-breakdownThreshold) {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Breakdown below support (%.2f). Current: %.2f", support, currentPrice)
		signal.Confidence = 0.7
	} else if trend == "uptrend" && trendStrength > 0.6 {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Uptrend confirmed (higher highs/lows). Strength: %.0f%%", trendStrength*100)
		signal.Confidence = 0.6
	} else if trend == "downtrend" && trendStrength > 0.6 {
		signal.Type = Sell
		signal.Reason = fmt.Sprintf("Downtrend confirmed (lower highs/lows). Strength: %.0f%%", trendStrength*100)
		signal.Confidence = 0.6
	} else {
		signal.Type = Hold
		signal.Reason = fmt.Sprintf("Consolidating. Support: %.2f, Resistance: %.2f", support, resistance)
		signal.Confidence = 0.0
	}

	return signal, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

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

// calculateEMA calculates the Exponential Moving Average.
func calculateEMA(data []stock.DailyPrice, period int) float64 {
	return calculateEMAOffset(data, period, 0)
}

// calculateEMAOffset calculates the EMA with an offset.
func calculateEMAOffset(data []stock.DailyPrice, period, offset int) float64 {
	if len(data) < period+offset {
		return 0
	}

	multiplier := 2.0 / float64(period+1)
	
	// Start with SMA as initial EMA
	ema := 0.0
	for i := period + offset - 1; i >= offset; i-- {
		if ema == 0 {
			// Initialize with first value
			sum := 0.0
			for j := i; j < i+period && j < len(data); j++ {
				sum += data[j].Close
			}
			ema = sum / float64(period)
		} else {
			ema = (data[i].Close-ema)*multiplier + ema
		}
	}

	return ema
}

// calculateRSI calculates the Relative Strength Index.
func calculateRSI(data []stock.DailyPrice, period int) float64 {
	return calculateRSIOffset(data, period, 0)
}

// calculateRSIOffset calculates RSI with an offset.
func calculateRSIOffset(data []stock.DailyPrice, period, offset int) float64 {
	if len(data) < period+1+offset {
		return 50 // neutral
	}

	var gains, losses float64

	for i := offset; i < period+offset; i++ {
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

// calculateMACD calculates MACD line, signal line, and histogram.
func calculateMACD(data []stock.DailyPrice, fastPeriod, slowPeriod, signalPeriod int) (macdLine, signalLine, histogram float64) {
	return calculateMACDOffset(data, fastPeriod, slowPeriod, signalPeriod, 0)
}

// calculateMACDOffset calculates MACD with an offset.
func calculateMACDOffset(data []stock.DailyPrice, fastPeriod, slowPeriod, signalPeriod, offset int) (macdLine, signalLine, histogram float64) {
	fastEMA := calculateEMAOffset(data, fastPeriod, offset)
	slowEMA := calculateEMAOffset(data, slowPeriod, offset)
	macdLine = fastEMA - slowEMA

	// Calculate signal line (EMA of MACD line)
	// We need to calculate multiple MACD values for the signal line
	macdValues := make([]float64, signalPeriod)
	for i := 0; i < signalPeriod; i++ {
		fast := calculateEMAOffset(data, fastPeriod, offset+i)
		slow := calculateEMAOffset(data, slowPeriod, offset+i)
		macdValues[i] = fast - slow
	}

	// Calculate EMA of MACD values for signal line
	multiplier := 2.0 / float64(signalPeriod+1)
	signalLine = macdValues[signalPeriod-1] // Start with oldest value
	for i := signalPeriod - 2; i >= 0; i-- {
		signalLine = (macdValues[i]-signalLine)*multiplier + signalLine
	}

	histogram = macdLine - signalLine
	return
}

// calculateStochastic calculates %K and %D values.
func calculateStochastic(data []stock.DailyPrice, kPeriod, dPeriod int) (k, d float64) {
	return calculateStochasticOffset(data, kPeriod, dPeriod, 0)
}

// calculateStochasticOffset calculates Stochastic with an offset.
func calculateStochasticOffset(data []stock.DailyPrice, kPeriod, dPeriod, offset int) (k, d float64) {
	if len(data) < kPeriod+dPeriod+offset {
		return 50, 50
	}

	// Calculate %K values for the %D period
	kValues := make([]float64, dPeriod)
	for i := 0; i < dPeriod; i++ {
		kValues[i] = calculateRawK(data, kPeriod, offset+i)
	}

	k = kValues[0] // Most recent %K
	
	// %D is SMA of %K values
	sum := 0.0
	for _, kv := range kValues {
		sum += kv
	}
	d = sum / float64(dPeriod)

	return
}

// calculateRawK calculates raw %K value.
func calculateRawK(data []stock.DailyPrice, period, offset int) float64 {
	if len(data) < period+offset {
		return 50
	}

	currentClose := data[offset].Close
	
	// Find highest high and lowest low in the period
	highestHigh := data[offset].High
	lowestLow := data[offset].Low
	
	for i := offset; i < period+offset; i++ {
		if data[i].High > highestHigh {
			highestHigh = data[i].High
		}
		if data[i].Low < lowestLow {
			lowestLow = data[i].Low
		}
	}

	if highestHigh == lowestLow {
		return 50
	}

	return ((currentClose - lowestLow) / (highestHigh - lowestLow)) * 100
}

// findSupportResistance finds key support and resistance levels.
func findSupportResistance(data []stock.DailyPrice, period int) (resistance, support float64) {
	if len(data) < period {
		return data[0].High, data[0].Low
	}

	// Find highest high (resistance) and lowest low (support) in period
	resistance = data[0].High
	support = data[0].Low

	for i := 0; i < period; i++ {
		if data[i].High > resistance {
			resistance = data[i].High
		}
		if data[i].Low < support {
			support = data[i].Low
		}
	}

	return
}

// analyzeTrend analyzes the trend direction and strength.
func analyzeTrend(data []stock.DailyPrice, period int) (trend string, strength float64) {
	if len(data) < period {
		return "neutral", 0
	}

	// Count higher highs/lows vs lower highs/lows
	higherHighs := 0
	higherLows := 0
	lowerHighs := 0
	lowerLows := 0

	for i := 0; i < period-1; i++ {
		if data[i].High > data[i+1].High {
			higherHighs++
		} else {
			lowerHighs++
		}
		if data[i].Low > data[i+1].Low {
			higherLows++
		} else {
			lowerLows++
		}
	}

	total := float64(period - 1)
	uptrendScore := (float64(higherHighs) + float64(higherLows)) / (2 * total)
	downtrendScore := (float64(lowerHighs) + float64(lowerLows)) / (2 * total)

	if uptrendScore > 0.6 {
		return "uptrend", uptrendScore
	} else if downtrendScore > 0.6 {
		return "downtrend", downtrendScore
	}
	return "neutral", 0.5
}

// =============================================================================
// Legacy SMA Indicator (for backward compatibility)
// =============================================================================

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
	if len(data.Data) < s.longPeriod+1 {
		return nil, fmt.Errorf("insufficient data: need %d days, got %d", s.longPeriod+1, len(data.Data))
	}

	shortSMA := calculateSMA(data.Data, s.shortPeriod)
	longSMA := calculateSMA(data.Data, s.longPeriod)
	prevShortSMA := calculateSMAOffset(data.Data, s.shortPeriod, 1)
	prevLongSMA := calculateSMAOffset(data.Data, s.longPeriod, 1)

	signal := &Signal{
		Symbol:    data.Symbol,
		Price:     currentPrice,
		Indicator: s.Name(),
		Timestamp: time.Now(),
	}

	if shortSMA > longSMA && prevShortSMA <= prevLongSMA {
		signal.Type = Buy
		signal.Reason = fmt.Sprintf("Golden Cross: Short SMA (%.2f) crossed above Long SMA (%.2f)", shortSMA, longSMA)
		signal.Confidence = 0.7
	} else if shortSMA < longSMA && prevShortSMA >= prevLongSMA {
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

// =============================================================================
// Signal Analyzer
// =============================================================================

// SignalAnalyzer combines multiple indicators to generate signals.
type SignalAnalyzer struct {
	indicators []Indicator
}

// NewSignalAnalyzer creates a new signal analyzer with comprehensive indicators.
func NewSignalAnalyzer() *SignalAnalyzer {
	return &SignalAnalyzer{
		indicators: []Indicator{
			NewMovingAverageIndicator(),   // 50/200-day MA with Golden/Death Cross
			NewStandardRSIIndicator(),     // RSI (14) with 70/30 thresholds
			NewMACDIndicator(),            // MACD (12/26/9)
			NewStochasticIndicator(),      // Stochastic (14,3) with 80/20 thresholds
			NewPriceActionIndicator(),     // Price Action (20-day lookback)
		},
	}
}

// NewBasicSignalAnalyzer creates a signal analyzer with basic indicators only.
func NewBasicSignalAnalyzer() *SignalAnalyzer {
	return &SignalAnalyzer{
		indicators: []Indicator{
			NewSMAIndicator(10, 50),
			NewRSIIndicator(14, 70, 30),
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

// GetSummary returns a summary of all signals.
func (a *SignalAnalyzer) GetSummary(signals []*Signal) (buyCount, sellCount, holdCount int, overallSignal SignalType) {
	for _, sig := range signals {
		switch sig.Type {
		case Buy:
			buyCount++
		case Sell:
			sellCount++
		case Hold:
			holdCount++
		}
	}

	if buyCount > sellCount && buyCount > holdCount {
		overallSignal = Buy
	} else if sellCount > buyCount && sellCount > holdCount {
		overallSignal = Sell
	} else {
		overallSignal = Hold
	}

	return
}
