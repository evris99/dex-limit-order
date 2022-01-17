# Dex Limit Order

A bot that provides limit and stop orders on PancakeSwap. It is controlled using a Telegram bot.

[![API Reference](https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667)](https://pkg.go.dev/github.com/evris99/dex-limit-order)

## Installation

### Dependencies

You need to have go and gcc installed.

### Building

To build clone this repository and run:
```
go build -o dex-limit-order .
```

## Configuration

Rename the configuration.example.toml file to configuration.toml.
Then open it with your text editor and add your telegram API key and your websocket BSC endpoint URL.

## Usage

You can add a user by running
```
dex-limit-order -new_user_id <userID>
```
To find your Telegram ID, you can follow this [guide](https://www.alphr.com/telegram-find-user-id/).

You also need to generate a wallet by running
```
dex-limit-order -wallet wallet.json
```
You can import this wallet to Metamask and add funds to make trades and pay for gas fees.