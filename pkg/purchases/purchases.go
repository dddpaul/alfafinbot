package purchases

import (
	"strconv"
	"strings"
	"time"

	"github.com/natekfl/untemplate"
)

var ut *untemplate.Untemplater
var df string

func init() {
	var err error
	ut, err = untemplate.Create("Покупка {price} ₽, {merchant}. Карта {card}. Баланс: {balance} ₽")
	if err != nil {
		panic(err)
	}
}

type Purchase struct {
	Time     time.Time
	Price    float64
	Merchant string
	Card     string
	Balance  float64
}

func New(t time.Time, s string) (*Purchase, error) {
	s1 := strings.ReplaceAll(s, "\n", " ")

	m, err := ut.Extract(s1)
	if err != nil {
		return nil, err
	}
	price, err := parseFloat(m["price"])
	if err != nil {
		return nil, err
	}
	balance, err := parseFloat(m["balance"])
	if err != nil {
		return nil, err
	}

	return &Purchase{
		Time:     t,
		Price:    price,
		Merchant: m["merchant"],
		Card:     m["card"],
		Balance:  balance,
	}, nil
}

func parseFloat(s string) (float64, error) {
	s1 := strings.Replace(s, ",", ".", 1)
	s1 = strings.ReplaceAll(s1, " ", "")
	s1 = strings.ReplaceAll(s1, "\u00A0", "")
	return strconv.ParseFloat(s1, 64)
}
