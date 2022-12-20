package main

import (
	"flag"
	"log"

	"github.com/dddpaul/alfafin-bot/pkg/telegram"
)

var (
	verbose          bool
	telegramToken    string
	telegramProxyURL string
	telegramAdmin    string
	gasURL           string
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable bot debug")
	flag.StringVar(&telegramToken, "telegram-token", "", "Telegram API token")
	flag.StringVar(&telegramProxyURL, "telegram-proxy-url", "", "Telegram SOCKS5 proxy url")
	flag.StringVar(&telegramAdmin, "telegram-admin", "", "Telegram admin user")
	flag.StringVar(&gasURL, "gas-url", "", "Google App Script URL")
	flag.Parse()

	if len(telegramToken) == 0 {
		log.Panic("Telegram API token has to be specified")
	}

	bot, err := telegram.NewBot(telegramToken,
		telegram.WithVerbose(verbose),
		telegram.WithAdmin(telegramAdmin),
		telegram.WithSocks(telegramProxyURL),
		telegram.WithGASUrl(gasURL))
	if err != nil {
		log.Panic(err)
	}

	bot.Start()
}
