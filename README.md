Alfabank Finance Bot
=========

Simple Telegram bot for parsing and uploading to Google sheet messages from [t.me/s/AlfaBank](https://t.me/s/AlfaBank).

Install:

```bash
go get -u github.com/dddpaul/alfafin-bot
```

Or grab Docker image:

```bash
docker pull dddpaul/alfafinbot
```

Usage:

```
  -gas-client-id string
    	This app client id for GAS web application
  -gas-client-secret string
    	This app client secret for GAS web application
  -gas-proxy-url string
    	SOCKS5 proxy url for GAS web app
  -gas-url string
    	Google App Script URL
  -telegram-admin string
    	Telegram admin user
  -telegram-proxy-url string
    	Telegram SOCKS5 proxy url
  -telegram-token string
    	Telegram API token
  -trace
    	Enable network tracing
  -verbose
    	Enable bot debug
```
