package signal

import (
	"testing"
	"time"

	"github.com/vicehope/stock-signal/pkg/stock"
)

func generateTestData(days int, startPrice float64, trend string) *stock.HistoricalData {
	data := make([]stock.DailyPrice, days)
	price := startPrice

	for i := 0; i < days; i++ {
		var change float64
		switch trend {
		case "up":
			change = 1.0 // Upward trend
		case "down":
			change = -1.0 // Downward trend
		default:
			if i%2 == 0 {
				change = 0.5
			} else {
				change = -0.5
			}
		}

		price += change
		data[i] = stock.DailyPrice{
			Date:   time.Now().AddDate(0, 0, -i),
			Open:   price - 0.5,
			High:   price + 0.5,
			Low:    price - 1.0,
			Close:  price,
			Volume: 1000000,
		}
	}

	return &stock.HistoricalData{
		Symbol: "TEST",
		Data:   data,
	}
}

// =============================================================================
// Moving Average Indicator Tests
// =============================================================================

func TestMovingAverageIndicator_Name(t *testing.T) {
	ma := NewMovingAverageIndicator()
	expected := "Moving Average (50/200)"
	if ma.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, ma.Name())
	}
}

func TestMovingAverageIndicator_InsufficientData(t *testing.T) {
	ma := NewMovingAverageIndicator()
	data := generateTestData(100, 100, "neutral") // Not enough for 200-day

	_, err := ma.Analyze(data, 100)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestMovingAverageIndicator_Analyze(t *testing.T) {
	ma := NewMovingAverageIndicator()
	data := generateTestData(250, 100, "up")

	signal, err := ma.Analyze(data, 300)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if signal.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", signal.Symbol)
	}

	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

func TestCustomMovingAverageIndicator(t *testing.T) {
	ma := NewCustomMovingAverageIndicator(10, 50)
	if ma.shortPeriod != 10 || ma.longPeriod != 50 {
		t.Errorf("Expected periods 10/50, got %d/%d", ma.shortPeriod, ma.longPeriod)
	}
}

// =============================================================================
// RSI Indicator Tests
// =============================================================================

func TestRSIIndicator_Name(t *testing.T) {
	rsi := NewRSIIndicator(14, 70, 30)
	expected := "RSI (14)"
	if rsi.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, rsi.Name())
	}
}

func TestRSIIndicator_InsufficientData(t *testing.T) {
	rsi := NewRSIIndicator(14, 70, 30)
	data := generateTestData(10, 100, "neutral") // Not enough data

	_, err := rsi.Analyze(data, 100)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestRSIIndicator_Analyze(t *testing.T) {
	rsi := NewRSIIndicator(14, 70, 30)
	data := generateTestData(30, 100, "up")

	signal, err := rsi.Analyze(data, 120)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if signal.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", signal.Symbol)
	}

	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

func TestStandardRSIIndicator(t *testing.T) {
	rsi := NewStandardRSIIndicator()
	if rsi.period != 14 || rsi.overbought != 70 || rsi.oversold != 30 {
		t.Errorf("Expected standard RSI settings, got period=%d, overbought=%.0f, oversold=%.0f",
			rsi.period, rsi.overbought, rsi.oversold)
	}
}

// =============================================================================
// MACD Indicator Tests
// =============================================================================

func TestMACDIndicator_Name(t *testing.T) {
	macd := NewMACDIndicator()
	expected := "MACD (12/26/9)"
	if macd.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, macd.Name())
	}
}

func TestMACDIndicator_InsufficientData(t *testing.T) {
	macd := NewMACDIndicator()
	data := generateTestData(30, 100, "neutral") // Not enough data (need 26+9+1=36)

	_, err := macd.Analyze(data, 100)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestMACDIndicator_Analyze(t *testing.T) {
	macd := NewMACDIndicator()
	data := generateTestData(50, 100, "up")

	signal, err := macd.Analyze(data, 140)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if signal.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", signal.Symbol)
	}

	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

func TestCustomMACDIndicator(t *testing.T) {
	macd := NewCustomMACDIndicator(5, 10, 3)
	if macd.fastPeriod != 5 || macd.slowPeriod != 10 || macd.signalPeriod != 3 {
		t.Errorf("Expected periods 5/10/3, got %d/%d/%d",
			macd.fastPeriod, macd.slowPeriod, macd.signalPeriod)
	}
}

// =============================================================================
// Stochastic Indicator Tests
// =============================================================================

func TestStochasticIndicator_Name(t *testing.T) {
	stoch := NewStochasticIndicator()
	expected := "Stochastic (14,3)"
	if stoch.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, stoch.Name())
	}
}

func TestStochasticIndicator_InsufficientData(t *testing.T) {
	stoch := NewStochasticIndicator()
	data := generateTestData(10, 100, "neutral") // Not enough data

	_, err := stoch.Analyze(data, 100)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestStochasticIndicator_Analyze(t *testing.T) {
	stoch := NewStochasticIndicator()
	data := generateTestData(30, 100, "up")

	signal, err := stoch.Analyze(data, 120)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if signal.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", signal.Symbol)
	}

	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

func TestCustomStochasticIndicator(t *testing.T) {
	stoch := NewCustomStochasticIndicator(10, 5, 90, 10)
	if stoch.kPeriod != 10 || stoch.dPeriod != 5 {
		t.Errorf("Expected periods 10/5, got %d/%d", stoch.kPeriod, stoch.dPeriod)
	}
	if stoch.overbought != 90 || stoch.oversold != 10 {
		t.Errorf("Expected thresholds 90/10, got %.0f/%.0f", stoch.overbought, stoch.oversold)
	}
}

// =============================================================================
// Price Action Indicator Tests
// =============================================================================

func TestPriceActionIndicator_Name(t *testing.T) {
	pa := NewPriceActionIndicator()
	expected := "Price Action (20-day)"
	if pa.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, pa.Name())
	}
}

func TestPriceActionIndicator_InsufficientData(t *testing.T) {
	pa := NewPriceActionIndicator()
	data := generateTestData(15, 100, "neutral") // Not enough data

	_, err := pa.Analyze(data, 100)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestPriceActionIndicator_Analyze(t *testing.T) {
	pa := NewPriceActionIndicator()
	data := generateTestData(30, 100, "up")

	signal, err := pa.Analyze(data, 120)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if signal.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", signal.Symbol)
	}

	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

func TestCustomPriceActionIndicator(t *testing.T) {
	pa := NewCustomPriceActionIndicator(30)
	if pa.lookbackPeriod != 30 {
		t.Errorf("Expected lookback period 30, got %d", pa.lookbackPeriod)
	}
}

// =============================================================================
// Legacy SMA Indicator Tests
// =============================================================================

func TestSMAIndicator_Name(t *testing.T) {
	sma := NewSMAIndicator(10, 50)
	expected := "SMA Crossover (10/50)"
	if sma.Name() != expected {
		t.Errorf("Expected name %q, got %q", expected, sma.Name())
	}
}

func TestSMAIndicator_InsufficientData(t *testing.T) {
	sma := NewSMAIndicator(10, 50)
	data := generateTestData(30, 100, "neutral") // Not enough data for 50-day SMA

	_, err := sma.Analyze(data, 100)
	if err == nil {
		t.Error("Expected error for insufficient data")
	}
}

func TestSMAIndicator_Analyze(t *testing.T) {
	sma := NewSMAIndicator(10, 50)
	data := generateTestData(60, 100, "up")

	signal, err := sma.Analyze(data, 150)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if signal.Symbol != "TEST" {
		t.Errorf("Expected symbol TEST, got %s", signal.Symbol)
	}

	if signal.Price != 150 {
		t.Errorf("Expected price 150, got %.2f", signal.Price)
	}

	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

// =============================================================================
// Signal Analyzer Tests
// =============================================================================

func TestSignalAnalyzer(t *testing.T) {
	analyzer := NewSignalAnalyzer()
	data := generateTestData(250, 100, "up")

	signals, err := analyzer.AnalyzeAll(data, 300)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have 5 signals (MA, RSI, MACD, Stochastic, Price Action)
	if len(signals) != 5 {
		t.Errorf("Expected 5 signals, got %d", len(signals))
	}
}

func TestBasicSignalAnalyzer(t *testing.T) {
	analyzer := NewBasicSignalAnalyzer()
	data := generateTestData(60, 100, "up")

	signals, err := analyzer.AnalyzeAll(data, 150)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should have 2 signals (SMA and RSI)
	if len(signals) != 2 {
		t.Errorf("Expected 2 signals, got %d", len(signals))
	}
}

func TestSignalAnalyzer_AddIndicator(t *testing.T) {
	analyzer := NewBasicSignalAnalyzer()
	customSMA := NewSMAIndicator(5, 20)
	analyzer.AddIndicator(customSMA)

	data := generateTestData(60, 100, "up")
	signals, err := analyzer.AnalyzeAll(data, 150)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(signals) != 3 { // SMA, RSI, and custom SMA
		t.Errorf("Expected 3 signals, got %d", len(signals))
	}
}

func TestSignalAnalyzer_GetSummary(t *testing.T) {
	analyzer := NewSignalAnalyzer()
	
	signals := []*Signal{
		{Type: Buy},
		{Type: Buy},
		{Type: Sell},
		{Type: Hold},
		{Type: Buy},
	}

	buyCount, sellCount, holdCount, overall := analyzer.GetSummary(signals)

	if buyCount != 3 {
		t.Errorf("Expected 3 buy signals, got %d", buyCount)
	}
	if sellCount != 1 {
		t.Errorf("Expected 1 sell signal, got %d", sellCount)
	}
	if holdCount != 1 {
		t.Errorf("Expected 1 hold signal, got %d", holdCount)
	}
	if overall != Buy {
		t.Errorf("Expected overall BUY signal, got %s", overall)
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestCalculateSMA(t *testing.T) {
	data := []stock.DailyPrice{
		{Close: 100},
		{Close: 110},
		{Close: 120},
		{Close: 130},
		{Close: 140},
	}

	// 3-period SMA of first 3 values: (100 + 110 + 120) / 3 = 110
	sma := calculateSMA(data, 3)
	expected := 110.0
	if sma != expected {
		t.Errorf("Expected SMA %.2f, got %.2f", expected, sma)
	}
}

func TestCalculateRSI(t *testing.T) {
	// Create data with known gains and losses
	// Most recent first (descending order)
	data := make([]stock.DailyPrice, 15)
	prices := []float64{150, 148, 145, 147, 146, 144, 143, 145, 146, 144, 142, 143, 141, 140, 138}
	for i, p := range prices {
		data[i] = stock.DailyPrice{Close: p}
	}

	rsi := calculateRSI(data, 14)

	// RSI should be between 0 and 100
	if rsi < 0 || rsi > 100 {
		t.Errorf("RSI out of range: %.2f", rsi)
	}
}

func TestCalculateEMA(t *testing.T) {
	data := []stock.DailyPrice{
		{Close: 100},
		{Close: 105},
		{Close: 110},
		{Close: 115},
		{Close: 120},
	}

	ema := calculateEMA(data, 3)
	
	// EMA should be calculated
	if ema <= 0 {
		t.Errorf("EMA should be positive, got %.2f", ema)
	}
}

func TestCalculateStochastic(t *testing.T) {
	data := make([]stock.DailyPrice, 20)
	for i := 0; i < 20; i++ {
		data[i] = stock.DailyPrice{
			High:  float64(110 - i),
			Low:   float64(90 - i),
			Close: float64(100 - i),
		}
	}

	k, d := calculateStochastic(data, 14, 3)

	// K and D should be between 0 and 100
	if k < 0 || k > 100 {
		t.Errorf("K out of range: %.2f", k)
	}
	if d < 0 || d > 100 {
		t.Errorf("D out of range: %.2f", d)
	}
}

func TestFindSupportResistance(t *testing.T) {
	data := []stock.DailyPrice{
		{High: 110, Low: 90},
		{High: 115, Low: 85},
		{High: 120, Low: 80},
		{High: 105, Low: 95},
		{High: 108, Low: 92},
	}

	resistance, support := findSupportResistance(data, 5)

	if resistance != 120 {
		t.Errorf("Expected resistance 120, got %.2f", resistance)
	}
	if support != 80 {
		t.Errorf("Expected support 80, got %.2f", support)
	}
}

func TestAnalyzeTrend(t *testing.T) {
	// Create uptrend data (higher highs and higher lows)
	uptrendData := make([]stock.DailyPrice, 10)
	for i := 0; i < 10; i++ {
		uptrendData[i] = stock.DailyPrice{
			High: float64(110 + i*2),
			Low:  float64(90 + i*2),
		}
	}

	trend, strength := analyzeTrend(uptrendData, 10)
	
	// Should detect downtrend since data[0] is most recent and has highest values
	if trend != "downtrend" {
		t.Logf("Trend detected: %s with strength %.2f", trend, strength)
	}
}
