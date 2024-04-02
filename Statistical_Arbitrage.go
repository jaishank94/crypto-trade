package main

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/thrasher-corp/gocryptotrader/exchanges/binance"
	"github.com/thrasher-corp/gocryptotrader/exchanges/exchange"
)

const (
	exchangeName    = "binance"
	symbol1         = "BTC/USDT"
	symbol2         = "ETH/USDT"
	timeframe       = "1h"
	window          = 20
	initialBalance  = 0.001
	riskPercentage  = 0.02 // 2% of capital at risk per trade
	stopLossPercent = 0.05 // 5% stop loss
)

func main() {
	// Initialize exchange
	binanceExchange := binance.New(exchange.APIKey, exchange.APISecret)

	// Connect to exchange
	err := binanceExchange.SetupExchangeCredentials(&exchange.APIKeyAndSecret{
		APIKey:    "your_api_key",
		APISecret: "your_api_secret",
	})
	if err != nil {
		log.Fatalf("Failed to connect to %s exchange: %v", exchangeName, err)
	}

	// Run continuously
	for {
		// Fetch current market data
		price1, err := getCurrentPrice(binanceExchange, symbol1)
		if err != nil {
			log.Printf("Failed to fetch current price for %s: %v", symbol1, err)
			continue
		}

		price2, err := getCurrentPrice(binanceExchange, symbol2)
		if err != nil {
			log.Printf("Failed to fetch current price for %s: %v", symbol2, err)
			continue
		}

		// Make trading decision
		historicalSpreadMean, historicalSpreadStdDev := calculateHistoricalSpreadStats(price1, price2)
		if shouldBuy(price1, price2, historicalSpreadMean, historicalSpreadStdDev) {
			err := executeTrade(binanceExchange, price1, price2)
			if err != nil {
				log.Printf("Failed to execute trade: %v", err)
			}
		}

		// Sleep for some time before checking again
		time.Sleep(5 * time.Minute)
	}
}

// Fetch current price
func getCurrentPrice(exchange *binance.Binance, symbol string) (float64, error) {
	ticker, err := exchange.GetTickerPrice(symbol)
	if err != nil {
		return 0, err
	}
	return ticker, nil
}

// Calculate historicalSpreadMean and historicalSpreadStdDev
func calculateHistoricalSpreadStats(price1, price2 float64) (float64, float64) {
	// Fetch historical price data for both assets (example)
	asset1Prices := []float64{100, 105, 110, 115, 120} // Replace with actual historical price data
	asset2Prices := []float64{95, 100, 105, 110, 115}  // Replace with actual historical price data

	// Check if the length of both price slices match
	if len(asset1Prices) != len(asset2Prices) {
		log.Println("Error: Length of price slices does not match")
		return 0, 0
	}

	// Calculate spread for each historical data point
	var spreads []float64
	for i := 0; i < len(asset1Prices); i++ {
		spread := asset1Prices[i] - asset2Prices[i]
		spreads = append(spreads, spread)
	}

	// Calculate mean and standard deviation of the spread
	var sum float64
	for _, spread := range spreads {
		sum += spread
	}
	mean := sum / float64(len(spreads))

	var sumSquares float64
	for _, spread := range spreads {
		sumSquares += math.Pow(spread-mean, 2)
	}
	variance := sumSquares / float64(len(spreads)-1)
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}

// Determine if we should buy based on the Statistical Arbitrage strategy
func shouldBuy(price1, price2, historicalSpreadMean, historicalSpreadStdDev float64) bool {
	// Calculate the current spread between the two prices
	spread := price1 - price2

	// Calculate Z-score to measure the deviation from historical spread
	zScore := (spread - historicalSpreadMean) / historicalSpreadStdDev

	// Determine the buy signal condition based on Z-score thresholds
	return zScore > 2.0 // Example threshold for buy signal (can be adjusted)
}

// Execute trade
func executeTrade(exchange *binance.Binance, price1, price2 float64) error {
	// Calculate position size based on risk percentage
	accountInfo, err := exchange.GetAccountInfo()
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}
	availableBalance := accountInfo.GetBalance("USDT")
	positionSize := availableBalance * riskPercentage

	// Calculate stop loss
	stopLossPrice := price1 * (1 - stopLossPercent)

	// Place buy order
	_, err = exchange.CreateOrder(symbol1, exchange.Buy, exchange.Market, exchange.IOC, positionSize, price1)
	if err != nil {
		return fmt.Errorf("failed to place buy order for %s: %w", symbol1, err)
	}

	// Place sell order for symbol2
	_, err = exchange.CreateOrder(symbol2, exchange.Sell, exchange.Market, exchange.IOC, positionSize, price2)
	if err != nil {
		return fmt.Errorf("failed to place sell order for %s: %w", symbol2, err)
	}

	// Place stop loss order
	_, err = exchange.CreateOrder(symbol1, exchange.Sell, exchange.Market, exchange.IOC, positionSize, stopLossPrice)
	if err != nil {
		return fmt.Errorf("failed to place stop loss order for %s: %w", symbol1, err)
	}

	fmt.Printf("Buy order placed for %s at %.2f\n", symbol1, price1)
	fmt.Printf("Sell order placed for %s at %.2f\n", symbol2, price2)
	fmt.Printf("Stop loss order placed for %s at %.2f\n", symbol1, stopLossPrice)

	return nil
}
