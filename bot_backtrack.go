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

	// Fetch historical data for backtesting
	historicalData := fetchHistoricalData(binanceExchange, symbol, timeframe, 200)

	// Backtest the strategy
	backtestResults := backtest(historicalData)

	// Print backtest results
	fmt.Printf("Backtest results:\nTotal Trades: %d\nWinning Trades: %d\nLosing Trades: %d\nProfit Factor: %.2f\n",
		backtestResults.TotalTrades, backtestResults.WinningTrades, backtestResults.LosingTrades, backtestResults.ProfitFactor)

	// Execute live trading
	// liveTrade(binanceExchange)
}

// BacktestResult stores the results of the backtesting
type BacktestResult struct {
	TotalTrades   int
	WinningTrades int
	LosingTrades  int
	ProfitFactor  float64
}

// Fetch historical data for backtesting
func fetchHistoricalData(exchange *binance.Binance, symbol, timeframe string, limit int) []exchange.Candle {
	ohlcv, err := exchange.GetHistoricCandles(symbol, timeframe, time.Now().Add(-time.Duration(limit)*time.Hour), time.Now())
	if err != nil {
		log.Fatalf("Failed to fetch historical data: %v", err)
	}
	return ohlcv
}

// Backtest the strategy
func backtest(historicalData []exchange.Candle) BacktestResult {
	var (
		totalTrades   int
		winningTrades int
		losingTrades  int
		profitFactor  float64
	)

	for i := window; i < len(historicalData); i++ {
		var returnsSum float64
		for j := i - window; j < i; j++ {
			returnsSum += (historicalData[j].Close - historicalData[j-1].Close) / historicalData[j-1].Close
		}
		meanReturns := returnsSum / float64(window)

		if math.IsNaN(meanReturns) {
			continue
		}

		if meanReturns > 0 {
			winningTrades++
		} else if meanReturns < 0 {
			losingTrades++
		}
		totalTrades++
	}

	if losingTrades > 0 {
		profitFactor = float64(winningTrades) / float64(losingTrades)
	} else {
		profitFactor = float64(winningTrades)
	}

	return BacktestResult{
		TotalTrades:   totalTrades,
		WinningTrades: winningTrades,
		LosingTrades:  losingTrades,
		ProfitFactor:  profitFactor,
	}
}

// Execute live trading
func liveTrade(exchange *binance.Binance) {
	// Calculate position size based on risk percentage
	accountInfo, err := exchange.GetAccountInfo()
	if err != nil {
		log.Fatalf("Failed to get account info: %v", err)
	}
	availableBalance := accountInfo.GetBalance("USDT")
	positionSize := availableBalance * riskPercentage

	// Fetch current price
	currentPrice, err := exchange.GetTickerPrice(symbol)
	if err != nil {
		log.Fatalf("Failed to fetch current price: %v", err)
	}

	// Calculate stop loss
	stopLossPrice := currentPrice * (1 - stopLossPercent)

	// Place buy order
	order, err := exchange.CreateOrder(symbol, exchange.Buy, exchange.Market, exchange.IOC, positionSize, currentPrice)
	if err != nil {
		log.Fatalf("Failed to place buy order: %v", err)
	}

	// Place stop loss order
	_, err = exchange.CreateOrder(symbol, exchange.Sell, exchange.Market, exchange.IOC, positionSize, stopLossPrice)
	if err != nil {
		log.Fatalf("Failed to place stop loss order: %v", err)
	}

	fmt.Printf("Buy order placed: %v\n", order)
	fmt.Printf("Stop loss order placed at %.2f\n", stopLossPrice)
}
