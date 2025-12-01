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

	// The signal type depends on the trend and crossover
	if signal.Type != Buy && signal.Type != Sell && signal.Type != Hold {
		t.Errorf("Invalid signal type: %s", signal.Type)
	}
}

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

func TestSignalAnalyzer(t *testing.T) {
	analyzer := NewSignalAnalyzer()
	data := generateTestData(60, 100, "up")

	signals, err := analyzer.AnalyzeAll(data, 150)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(signals) != 2 { // Should have SMA and RSI signals
		t.Errorf("Expected 2 signals, got %d", len(signals))
	}
}

func TestSignalAnalyzer_AddIndicator(t *testing.T) {
	analyzer := NewSignalAnalyzer()
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
