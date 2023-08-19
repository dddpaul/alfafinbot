package purchases

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"slices"

	"github.com/dddpaul/cbr-currency-go"
	"github.com/natekfl/untemplate"
)

type Operation int64

const (
	Buy Operation = iota
	Cancel
)

var (
	ut1           *untemplate.Untemplater
	ut2           *untemplate.Untemplater
	ut3           *untemplate.Untemplater
	mdRegexp      = regexp.MustCompile(`^(.+) (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2})$`)
	df            = "02.01.2006 15:04"
	roubleSymbols = []string{"RUB", "RUR", "₽"}
)

func init() {
	var err error
	ut1, err = untemplate.Create("Покупка {price} {currency}, {merchant}. Карта {card}. Баланс: {balance} ₽")
	if err != nil {
		panic(err)
	}
	ut2, err = untemplate.Create("{card} Pokupka {price} {currency} Balans {balance} RUR {merchant_datetime}")
	if err != nil {
		panic(err)
	}
	ut3, err = untemplate.Create("Отмена операции {price} {currency}, {merchant}. Карта {card}. Баланс: {balance} ₽")
	if err != nil {
		panic(err)
	}
	cbr.UpdateCurrencyRates()
}

type Purchase struct {
	Time     time.Time
	Price    float64
	Merchant string
	Card     string
	Balance  float64
	Currency string
	PriceRUB float64
}

func New(t time.Time, s string) (*Purchase, error) {
	s1 := strings.ReplaceAll(s, "\n", " ")
	op := Buy

	m, err := ut1.Extract(s1)
	if err != nil {
		m, err = ut2.Extract(s1)
		if err != nil {
			op = Cancel
			m, err = ut3.Extract(s1)
			if err != nil {
				return nil, err
			}
		}
	}

	price, err := parseFloat(m["price"])
	if err != nil {
		return nil, err
	}
	if op == Cancel {
		price = -price
	}

	balance, err := parseFloat(m["balance"])
	if err != nil {
		return nil, err
	}

	merchant := m["merchant"]
	if md, ok := m["merchant_datetime"]; ok {
		merchant, t, err = parseMerchantAndDatetime(md)
		if err != nil {
			return nil, err
		}
	}

	priceRUB, err := calcRoublePrice(price, m["currency"])
	if err != nil {
		return nil, err
	}

	return &Purchase{
		Time:     t,
		Price:    price,
		Merchant: merchant,
		Card:     m["card"],
		Balance:  balance,
		Currency: m["currency"],
		PriceRUB: priceRUB,
	}, nil
}

func parseFloat(s string) (float64, error) {
	s1 := strings.Replace(s, ",", ".", 1)
	s1 = strings.ReplaceAll(s1, " ", "")
	s1 = strings.ReplaceAll(s1, "\u00A0", "")
	return strconv.ParseFloat(s1, 64)
}

func parseMerchantAndDatetime(md string) (string, time.Time, error) {
	tokens := mdRegexp.FindStringSubmatch(md)
	if len(tokens) != 3 {
		return "", time.Time{}, fmt.Errorf("incorrent merchant and datetime format: %s", md)
	}
	t, err := time.Parse(df, tokens[2])
	if err != nil {
		return "", time.Time{}, err
	}
	return tokens[1], t, nil
}

func calcRoublePrice(price float64, currency string) (float64, error) {
	if slices.Contains(roubleSymbols, currency) {
		return price, nil
	}
	if rate, ok := cbr.GetCurrencyRates()[currency]; ok {
		return price * rate.Value, nil
	}
	return 0, fmt.Errorf("unknown currency %s", currency)
}
