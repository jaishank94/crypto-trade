import time
import logging
from binance.client import Client
import numpy as np
import math

# Constants
exchangeName = "binance"
symbol = "BTCUSDT"
timeframe = "1h"
window = 20
initialBalance = 0.001
riskPercentage = 0.02  # 2% of capital at risk per trade
stopLossPercent = 0.05  # 5% stop loss
short_window = 5  # Short-term EMA window
long_window = 10  # Long-term EMA window

# Logging configuration
logging.basicConfig(level=logging.INFO, format="%(asctime)s - %(levelname)s - %(message)s")

# Binance API test credentials
API_KEY = 'Gp3Y26znlkJNpPoeeHCTrILVILXOU5rDUorPbgiQ4FE916Owm67qQYyM45zWc358'
API_SECRET = 'noLPUROxbsxOwaXjqGNPI6Hss23E76fI8T79YgB8NAbpl9b2RXLXwmDXCqrecBFH'

def main():
    # Initialize Binance client
    binanceClient = Client(api_key=API_KEY, api_secret=API_SECRET, testnet=True)

    # Fetch symbol information
    info = binanceClient.get_symbol_info(symbol)

    while True:
        try:
            # Print current balance
            print_current_balance(binanceClient)

            # Fetch current market data
            currentPrice = get_current_price(binanceClient)

            # Make trading decision
            if should_buy(binanceClient):
                execute_trade(binanceClient, currentPrice)

            # Sleep for some time before checking again
            time.sleep(20)  # 5 minutes
        except Exception as e:
            logging.error(f"An error occurred: {e}")


# Fetch current price
def get_current_price(client):
    ticker = client.get_symbol_ticker(symbol=symbol)
    return float(ticker['price'])

def exponential_moving_average(values, window):
    if len(values) < window:
        return None
    weights = np.exp(np.linspace(-1., 0., window))
    weights /= weights.sum()
    ema = np.convolve(values, weights, mode='full')[:len(values)]
    ema[:window] = ema[window-1]  # Corrected assignment
    return ema[-1]




# Determine if we should buy based on the trend-following strategy using Exponential Moving Averages (EMAs)
def should_buy(client):
    try:
        # Fetch historical data for calculating exponential moving averages
        historicalData = fetch_historical_data(client, symbol, timeframe, long_window)

        # Log the length of historical data and the expected window size
        logging.info(f"Length of historicalData: {len(historicalData)}, Expected window size: {long_window}")

        # Ensure we have enough historical data for the specified window
        if len(historicalData) < long_window:
            logging.warning("Insufficient historical data for calculating moving averages")
            return False

        # Calculate exponential moving average of closing prices
        closes = [candle['close'] for candle in historicalData]
        ema_short = exponential_moving_average(closes, short_window)
        ema_long = exponential_moving_average(closes, long_window)

        # Determine if the short-term EMA is above the long-term EMA
        currentPrice = get_current_price(client)
        print(ema_short, ema_long, currentPrice)
        return ema_short > ema_long and currentPrice > ema_short
    except Exception as e:
        logging.error(f"Failed to determine buy signal: {e}")
        return False




# Fetch historical data
def fetch_historical_data(client, symbol, timeframe, window):
    historicalData = client.get_historical_klines(symbol=symbol, interval=timeframe, limit=window)
    return [{'timestamp': candle[0], 'open': float(candle[1]), 'high': float(candle[2]), 'low': float(candle[3]),
             'close': float(candle[4]), 'volume': float(candle[5])} for candle in historicalData]


# Function to print current balance
def print_current_balance(client):
    try:
        accountInfo = client.get_account()
        balance = float([asset['free'] for asset in accountInfo['balances'] if asset['asset'] == 'USDT'][0])
        logging.info(f"Current balance: {balance:.2f} USDT")
    except Exception as e:
        logging.error(f"Failed to fetch current balance: {e}")

def execute_trade(client, currentPrice):
    try:
        # Get account balance
        accountInfo = client.get_account()
        availableBalance = float([asset['free'] for asset in accountInfo['balances'] if asset['asset'] == 'USDT'][0])
        
        # Calculate position size based on risk percentage
        positionSize = availableBalance * riskPercentage

        # Check if available balance is sufficient
        if positionSize < 10:  # Minimum trade size on Binance is typically 10 USDT
            logging.warning("Available balance is not sufficient for trade.")
            return

        # Fetch symbol info to get lot size constraints
        symbol_info = client.get_symbol_info(symbol)
        lot_size_filter = next((f for f in symbol_info['filters'] if f['filterType'] == 'LOT_SIZE'), None)
        if lot_size_filter is None:
            logging.error("Failed to retrieve lot size information.")
            return

        # Determine the step size for the lot size
        step_size = float(lot_size_filter['stepSize'])

        # Calculate order quantity and round to nearest valid lot size
        order_quantity = round((positionSize / currentPrice) / step_size) * step_size

        # Convert order quantity to string with appropriate precision
        order_quantity_str = '{:.{prec}f}'.format(order_quantity, prec=int(-math.log10(step_size)))

        # Place buy order
        order = client.order_market_buy(symbol=symbol, quantity=order_quantity_str)

        # Calculate stop loss
        stopLossPrice = currentPrice * (1 - stopLossPercent)

        # Place stop loss order
        client.create_order(symbol=symbol, side='SELL', type='STOP_LOSS', quantity=order['executedQty'],
                            stopPrice=str(stopLossPrice))

        logging.info(f"Buy order placed at {currentPrice:.2f}")
        logging.info(f"Stop loss order placed at {stopLossPrice:.2f}")
    except Exception as e:
        logging.error(f"Failed to execute trade: {e}")


if __name__ == "__main__":
    main()
