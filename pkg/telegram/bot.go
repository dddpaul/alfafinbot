package telegram

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dddpaul/alfafin-bot/pkg/gas"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"golang.org/x/net/proxy"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	bot       *tb.Bot
	verbose   bool
	admin     string
	gasClient *gas.Client
	client    *http.Client
}

type BotOption func(b *Bot)

func WithVerbose(v bool) BotOption {
	return func(b *Bot) {
		b.verbose = v
	}
}

func WithSocks(s string) BotOption {
	return func(b *Bot) {
		if len(s) == 0 {
			return
		}

		u, err := url.Parse(s)
		if err != nil {
			log.Panic(err)
		}

		var auth *proxy.Auth
		if u.User != nil {
			auth = &proxy.Auth{
				User: u.User.Username(),
			}
			if p, ok := u.User.Password(); ok {
				auth.Password = p
			}
		}

		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
		if err != nil {
			log.Panic(err)
		}
		httpTransport := &http.Transport{
			Dial: dialer.Dial,
		}
		client := &http.Client{Transport: httpTransport}
		b.client = client
	}
}

func WithAdmin(a string) BotOption {
	return func(b *Bot) {
		b.admin = a
	}
}

func WithGASUrl(u string) BotOption {
	return func(b *Bot) {
		b.gasClient = gas.NewClient(u)
	}
}

func NewBot(telegramToken string, opts ...BotOption) (*Bot, error) {
	b := &Bot{}

	for _, opt := range opts {
		opt(b)
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:  telegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Client: b.client,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Authorized on account %s\n", bot.Me.Username)

	b.bot = bot
	return b, nil
}

func (b *Bot) Start() {
	check := func(cmd string, m *tb.Message) bool {
		log.Printf("Received '%s' command from '%s'", cmd, m.Sender.Username)
		if b.admin != "" && b.admin != m.Sender.Username {
			b.bot.Send(m.Sender, "Access restricted")
			return false
		}
		return true
	}

	b.bot.Handle("/status", func(m *tb.Message) {
		if !check("/status", m) {
			return
		}
		b.bot.Send(m.Sender, fmt.Sprintf("I'm fine"))
	})

	b.bot.Handle(tb.OnText, func(m *tb.Message) {
		if b.verbose {
			log.Printf("Text: %s, forwarded: %t", m.Text, m.IsForwarded())
		}
		p, err := purchases.New(getTime(m), m.Text)
		if err != nil {
			log.Printf("ERROR: %v", err)
		}
		b.gasClient.Add(p)
		log.Printf("Purchase %v have been added to sheet", p)
	})

	b.bot.Handle(tb.OnPhoto, func(m *tb.Message) {
		if b.verbose {
			log.Printf("Photo with caption: %s, forwarded: %t", m.Caption, m.IsForwarded())
		}
		p, err := purchases.New(getTime(m), m.Caption)
		if err != nil {
			log.Printf("ERROR: %v", err)
		}
		b.gasClient.Add(p)
		log.Printf("Purchase %v have been added to sheet", p)
	})

	b.bot.Start()
}

func getTime(m *tb.Message) time.Time {
	if m.IsForwarded() {
		return time.Unix(int64(m.OriginalUnixtime), 0)
	}
	return m.Time()
}
