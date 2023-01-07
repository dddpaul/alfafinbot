package telegram

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dddpaul/alfafin-bot/pkg/gas"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"golang.org/x/net/proxy"
	tb "gopkg.in/tucnak/telebot.v2"
)

type Bot struct {
	bot       *tb.Bot
	verbose   bool
	admin     string
	gasConfig *gas.GASConfig
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

func WithGAS(url string, id string, secret string) BotOption {
	return func(b *Bot) {
		b.gasConfig = &gas.GASConfig{
			Url:          url,
			ClientID:     id,
			ClientSecret: secret,
		}
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

	add := func(p *purchases.Purchase) {
		resp, err := gas.NewClient(b.gasConfig, "").Add(p)
		if err != nil {
			log.Printf("ERROR: %v", err)
			return
		}
		log.Printf("Purchase %v have been added to sheet", resp)
	}

	b.bot.Handle("/status", func(m *tb.Message) {
		if !check("/status", m) {
			return
		}
		b.bot.Send(m.Sender, fmt.Sprintf("I'm fine"))
	})

	b.bot.Handle("/today", func(m *tb.Message) {
		if !check("/today", m) {
			return
		}
		resp, err := gas.NewClient(b.gasConfig, "today").Get()
		if err != nil {
			log.Printf("ERROR: %v", err)
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle("/week", func(m *tb.Message) {
		if !check("/week", m) {
			return
		}
		resp, err := gas.NewClient(b.gasConfig, "week").Get()
		if err != nil {
			log.Printf("ERROR: %v", err)
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle("/month", func(m *tb.Message) {
		if !check("/month", m) {
			return
		}
		resp, err := gas.NewClient(b.gasConfig, "month").Get()
		if err != nil {
			log.Printf("ERROR: %v", err)
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle("/year", func(m *tb.Message) {
		if !check("/year", m) {
			return
		}
		resp, err := gas.NewClient(b.gasConfig, "year").Get()
		if err != nil {
			log.Printf("ERROR: %v", err)
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle(tb.OnText, func(m *tb.Message) {
		if b.verbose {
			log.Printf("Text: \"%s\", forwarded: %t", m.Text, m.IsForwarded())
		}
		p, err := purchases.New(getTime(m), m.Text)
		if err != nil {
			log.Printf("ERROR: %v", err)
			return
		}
		add(p)
	})

	b.bot.Handle(tb.OnPhoto, func(m *tb.Message) {
		if b.verbose {
			log.Printf("Photo with caption: \"%s\", forwarded: %t", m.Caption, m.IsForwarded())
		}
		p, err := purchases.New(getTime(m), m.Caption)
		if err != nil {
			log.Printf("ERROR: %v", err)
			return
		}
		add(p)
	})

	b.bot.Start()
}

func getTime(m *tb.Message) time.Time {
	if m.IsForwarded() {
		return time.Unix(int64(m.OriginalUnixtime), 0)
	}
	return m.Time()
}
