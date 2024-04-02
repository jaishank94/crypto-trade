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
	symbol          = "BTC/USDT"
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
		currentPrice, err := getCurrentPrice(binanceExchange)
		if err != nil {
			log.Printf("Failed to fetch current price: %v", err)
			continue
		}

		// Make trading decision
		if shouldBuy(currentPrice) {
			err := executeTrade(binanceExchange, currentPrice)
			if err != nil {
				log.Printf("Failed to execute trade: %v", err)
			}
		}

		// Sleep for some time before checking again
		time.Sleep(5 * time.Minute)
	}
}

// Fetch current price
func getCurrentPrice(exchange *binance.Binance) (float64, error) {
	ticker, err := exchange.GetTickerPrice(symbol)
	if err != nil {
		return 0, err
	}
	return ticker, nil
}

//simple trend-following strategy
// Determine if we should buy based on the trend-following strategy
func shouldBuy(exchange *binance.Binance) (bool, error) {
	// Fetch historical data for calculating moving averages
	historicalData, err := fetchHistoricalData(exchange, symbol, timeframe, window)
	if err != nil {
		return false, fmt.Errorf("failed to fetch historical data: %w", err)
	}

	// Calculate moving average of closing prices
	var sum float64
	for _, candle := range historicalData {
		sum += candle.Close
	}
	average := sum / float64(len(historicalData))

	// Determine if the current price is above the moving average
	currentPrice, err := getCurrentPrice(exchange)
	if err != nil {
		return false, fmt.Errorf("failed to fetch current price: %w", err)
	}

	return currentPrice > average, nil
}


//simple moving average crossover strategy
// Determine if we should buy based on the trading strategy (simple moving average crossover)
// func shouldBuy(exchange *binance.Binance) (bool, error) {
// 	// Fetch historical data for calculating moving averages
// 	historicalData, err := fetchHistoricalData(exchange, symbol, timeframe, window)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to fetch historical data: %w", err)
// 	}

// 	// Calculate short-term (fast) and long-term (slow) moving averages
// 	var (
// 		shortTermMA float64
// 		longTermMA  float64
// 	)
// 	for _, candle := range historicalData {
// 		shortTermMA += candle.Close
// 		longTermMA += candle.Close
// 	}
// 	shortTermMA /= float64(len(historicalData))
// 	longTermMA /= float64(len(historicalData))

// 	// Determine if there's a crossover
// 	return shortTermMA > longTermMA, nil
// }


// Execute trade
func executeTrade(exchange *binance.Binance, currentPrice float64) error {
	// Calculate position size based on risk percentage
	accountInfo, err := exchange.GetAccountInfo()
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}
	availableBalance := accountInfo.GetBalance("USDT")
	positionSize := availableBalance * riskPercentage

	// Calculate stop loss
	stopLossPrice := currentPrice * (1 - stopLossPercent)

	// Place buy order
	_, err = exchange.CreateOrder(symbol, exchange.Buy, exchange.Market, exchange.IOC, positionSize, currentPrice)
	if err != nil {
		return fmt.Errorf("failed to place buy order: %w", err)
	}

	// Place stop loss order
	_, err = exchange.CreateOrder(symbol, exchange.Sell, exchange.Market, exchange.IOC, positionSize, stopLossPrice)
	if err != nil {
		return fmt.Errorf("failed to place stop loss order: %w", err)
	}

	fmt.Printf("Buy order placed at %.2f\n", currentPrice)
	fmt.Printf("Stop loss order placed at %.2f\n", stopLossPrice)

	return nil
}
