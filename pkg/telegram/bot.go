package telegram

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dddpaul/alfafin-bot/pkg/gas"
	"github.com/dddpaul/alfafin-bot/pkg/logger"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"github.com/dddpaul/alfafin-bot/pkg/transport"
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

func WithSocks(socks string) BotOption {
	return func(b *Bot) {
		b.client = &http.Client{
			Transport: transport.NewSocksTransport(socks),
		}
	}
}

func WithAdmin(a string) BotOption {
	return func(b *Bot) {
		b.admin = a
	}
}

func WithGAS(url string, socks string, id string, secret string) BotOption {
	return func(b *Bot) {
		b.gasConfig = &gas.GASConfig{
			Url:          url,
			Socks:        socks,
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
			b.bot.Send(m.Sender, "ERROR: Access restricted")
			return false
		}
		return true
	}

	add := func(ctx context.Context, p *purchases.Purchase) {
		resp, err := gas.NewClient(b.gasConfig, "").Add(ctx, p)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
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
		ctx := newContext(m)
		resp, err := gas.NewClient(b.gasConfig, "today").Get(ctx)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle("/week", func(m *tb.Message) {
		if !check("/week", m) {
			return
		}
		ctx := newContext(m)
		resp, err := gas.NewClient(b.gasConfig, "week").Get(ctx)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle("/month", func(m *tb.Message) {
		if !check("/month", m) {
			return
		}
		ctx := newContext(m)
		resp, err := gas.NewClient(b.gasConfig, "month").Get(ctx)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle("/year", func(m *tb.Message) {
		if !check("/year", m) {
			return
		}
		ctx := newContext(m)
		resp, err := gas.NewClient(b.gasConfig, "year").Get(ctx)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			b.bot.Send(m.Sender, fmt.Sprintf("ERROR: %v", err))
			return
		}
		b.bot.Send(m.Sender, resp)
	})

	b.bot.Handle(tb.OnText, func(m *tb.Message) {
		if b.verbose {
			log.Printf("Text: \"%s\", forwarded: %t", m.Text, m.IsForwarded())
		}
		ctx := newContext(m)
		p, err := purchases.New(getTime(m), m.Text)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			return
		}
		add(newContext(m), p)
	})

	b.bot.Handle(tb.OnPhoto, func(m *tb.Message) {
		if b.verbose {
			log.Printf("Photo with caption: \"%s\", forwarded: %t", m.Caption, m.IsForwarded())
		}
		ctx := newContext(m)
		p, err := purchases.New(getTime(m), m.Caption)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			return
		}
		add(newContext(m), p)
	})

	b.bot.Start()
}

func getTime(m *tb.Message) time.Time {
	if m.IsForwarded() {
		return time.Unix(int64(m.OriginalUnixtime), 0)
	}
	return m.Time()
}

func newContext(m *tb.Message) context.Context {
	return context.WithValue(context.Background(), "message_id", m.ID)
}
