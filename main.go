package main

import (
	"flag"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/dddpaul/alfafin-bot/pkg/telegram"
)

var (
	verbose          bool
	trace            bool
	telegramToken    string
	telegramProxyURL string
	telegramAdmin    string
	gasURL           string
	gasProxyURL      string
	gasClientID      string
	gasClientSecret  string
)

func main() {
	flag.BoolVar(&verbose, "verbose", false, "Enable bot debug")
	flag.BoolVar(&verbose, "trace", false, "Enable network tracing")
	flag.StringVar(&telegramToken, "telegram-token", LookupEnvOrString("TELEGRAM_TOKEN", ""), "Telegram API token")
	flag.StringVar(&telegramProxyURL, "telegram-proxy-url", LookupEnvOrString("TELEGRAM_PROXY_URL", ""), "Telegram SOCKS5 proxy url")
	flag.StringVar(&telegramAdmin, "telegram-admin", LookupEnvOrString("TELEGRAM_ADMIN", ""), "Telegram admin user")
	flag.StringVar(&gasURL, "gas-url", LookupEnvOrString("GAS_URL", ""), "Google App Script URL")
	flag.StringVar(&gasProxyURL, "gas-proxy-url", LookupEnvOrString("GAS_PROXY_URL", ""), "SOCKS5 proxy url for GAS web app")
	flag.StringVar(&gasClientID, "gas-client-id", LookupEnvOrString("GAS_CLIENT_ID", ""), "This app client id for GAS web application")
	flag.StringVar(&gasClientSecret, "gas-client-secret", LookupEnvOrString("GAS_CLIENT_SECRET", ""), "This app client secret for GAS web application")

	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	flag.Parse()
	log.Printf("Configuration %v", getConfig(flag.CommandLine))

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	if trace {
		log.SetLevel(log.TraceLevel)
	}

	if len(telegramToken) == 0 {
		log.Panic("Telegram API token has to be specified")
	}

	bot, err := telegram.NewBot(telegramToken,
		telegram.WithVerbose(verbose),
		telegram.WithAdmin(telegramAdmin),
		telegram.WithSocks(telegramProxyURL),
		telegram.WithGAS(gasURL, gasProxyURL, gasClientID, gasClientSecret))
	if err != nil {
		panic(err)
	}

	bot.Start()
}

func LookupEnvOrString(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getConfig(fs *flag.FlagSet) []string {
	cfg := make([]string, 0, 10)
	fs.VisitAll(func(f *flag.Flag) {
		cfg = append(cfg, fmt.Sprintf("%s:%q", f.Name, f.Value.String()))
	})

	return cfg
}
