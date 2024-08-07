package purchases

import (
	"fmt"
	"math"
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
	templates       []*untemplate.Untemplater
	mdRegexp        = regexp.MustCompile(`^(.+) (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2})`)
	df              = "02.01.2006 15:04"
	ddmmyyyy        = "02.01.2006"
	digitsRegexp    = regexp.MustCompile(`\d+`)
	currencySymbols = map[string]string{"RUB": "₽", "RUR": "₽", "₽": "₽", "USD": "$", "EUR": "€", "AMD": "֏", "BYN": "Br"}
	roubleSymbols   = []string{"RUB", "RUR", "₽"}
)

func init() {
	templateStrings := []string{
		"Покупка {price} {currency}, {merchant}. Карта {card}. Баланс: {balance} ₽",         // Alfabank template before 2023-08
		"{card} Pokupka {price} {currency} Balans {balance} RUR {merchant_datetime}",        // Alfabank template after 2023-08
		"Покупка {card}: {price} {currency} в {merchant} Баланс: {balance}",                 // Alfabank template after 2024-07
		"{date} {price} {currency} - {merchant}",                                            // Custom template for adding purchases manually
		"Отмена операции {price} {currency}, {merchant}. Карта {card}. Баланс: {balance} ₽", // Alfabank cancel template
	}

	for _, tmplStr := range templateStrings {
		tmpl, err := untemplate.Create(tmplStr)
		if err != nil {
			panic(err)
		}
		templates = append(templates, tmpl)
	}

	cbr.UpdateCurrencyRates()
}

type Purchase struct {
	Time     time.Time
	Price    float64
	Merchant string
	Card     string
	Currency string
	PriceRUB float64
}

func New(dt time.Time, s string) (*Purchase, error) {
	s1 := strings.ReplaceAll(s, "\n", " ")
	op := Buy

	var m map[string]string
	var err error

	for i, tmpl := range templates {
		m, err = tmpl.Extract(s1)
		if err == nil {
			if i == len(templates)-1 { // Last template is cancel
				op = Cancel
			}
			break
		}
	}

	if err != nil {
		return nil, err
	}

	price, err := parseFloat(m["price"])
	if err != nil {
		return nil, err
	}
	price = roundFloat(price, 2)
	if op == Cancel {
		price = -price
	}

	if date, ok := m["date"]; ok {
		dt, err = time.ParseInLocation(ddmmyyyy, date, time.Local)
		if err != nil {
			return nil, err
		}
	}

	merchant := m["merchant"]
	if md, ok := m["merchant_datetime"]; ok {
		merchant, dt, err = parseMerchantAndDatetime(md)
		if err != nil {
			return nil, err
		}
	}
	merchant = digitsRegexp.ReplaceAllString(merchant, "")
	merchant = strings.Trim(merchant, " ")

	currencySymbol, ok := currencySymbols[m["currency"]]
	if !ok {
		return nil, fmt.Errorf("unknown currency %s", m["currency"])
	}

	priceRUB, err := calcRoublePrice(price, m["currency"], dt)
	if err != nil {
		return nil, err
	}

	return &Purchase{
		Time:     dt,
		Price:    price,
		Merchant: merchant,
		Card:     m["card"],
		Currency: currencySymbol,
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
		return "", time.Time{}, fmt.Errorf("incorrect merchant and datetime format: %s", md)
	}
	dt, err := time.ParseInLocation(df, tokens[2], time.Local)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokens[1], dt, nil
}

func calcRoublePrice(price float64, currency string, dt time.Time) (float64, error) {
	if slices.Contains(roubleSymbols, currency) {
		return roundFloat(price, 2), nil
	}
	rates := cbr.GetCurrencyRates()
	if truncateDay(time.Now()) != truncateDay(dt) {
		var err error
		rates, err = cbr.FetchCurrencyRates(dt)
		if err != nil {
			return 0, err
		}
	}
	if rate, ok := rates[currency]; ok {
		return roundFloat(price*rate.Value, 2), nil
	}
	return 0, fmt.Errorf("unknown currency: %s", currency)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func truncateDay(dt time.Time) time.Time {
	return dt.Truncate(time.Hour * 24)
}
