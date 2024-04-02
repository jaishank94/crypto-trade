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

		// Calculate moving averages for price1 and price2
		ma1, ma2 := calculateMovingAverages(binanceExchange, symbol1, symbol2, timeframe, window)

		// Make trading decision
		if shouldBuy(price1, price2, ma1, ma2) {
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

// Calculate moving averages
func calculateMovingAverages(exchange *binance.Binance, symbol1, symbol2, timeframe string, window int) (float64, float64) {
	// Fetch historical data for symbol1
	historicalData1, err := fetchHistoricalData(exchange, symbol1, timeframe, window)
	if err != nil {
		log.Printf("Failed to fetch historical data for %s: %v", symbol1, err)
		return 0, 0
	}

	// Fetch historical data for symbol2
	historicalData2, err := fetchHistoricalData(exchange, symbol2, timeframe, window)
	if err != nil {
		log.Printf("Failed to fetch historical data for %s: %v", symbol2, err)
		return 0, 0
	}

	// Calculate moving averages for symbol1 and symbol2
	ma1 := calculateMovingAverage(historicalData1)
	ma2 := calculateMovingAverage(historicalData2)

	return ma1, ma2
}

// Calculate moving average
func calculateMovingAverage(data []float64) float64 {
	var sum float64
	for _, price := range data {
		sum += price
	}
	return sum / float64(len(data))
}

// Determine if we should buy based on Momentum Trading strategy
func shouldBuy(price1, price2, ma1, ma2 float64) bool {
	// Check if the price of symbol1 is higher than symbol2
	// and if the short-term moving average (ma1) is above the long-term moving average (ma2)
	if price1 > price2 && ma1 > ma2 {
		// Check for recent positive price momentum
		recentPriceChange := price1 - ma1 // Calculate recent price change
		if recentPriceChange > 0 {
			// Consider buying if recent price change indicates positive momentum
			return true
		}
	}

	// If conditions are not met, do not buy
	return false
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
