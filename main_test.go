package main

import (
	"encoding/json"
	"github.com/dddpaul/alfafin-bot/pkg/stats"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

var df = "02.01.2006 15:04"

func TestNewPurchaseWithTemplate1(t *testing.T) {
	p, err := newPurchase("Покупка 527,11 ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.Nil(t, err)
	assert.Equal(t, 527.11, p.Price)
	assert.Equal(t, "Озон", p.Merchant)
	assert.Equal(t, "**1111", p.Card)
	assert.Equal(t, 4506.85, p.Balance)
	assert.Equal(t, "₽", p.Currency)

	p, err = newPurchase("Покупка 527.11 ₽, Озон.\nКарта **1111. Баланс: 4506.85 ₽")
	assert.Nil(t, err)
	assert.Equal(t, 527.11, p.Price, "Dot should be accepted as decimal separator")
	assert.Equal(t, 4506.85, p.Balance, "Dot should be accepted as decimal separator")

	p, err = newPurchase("Покупка 527 ₽, Озон.\nКарта **1111. Баланс: 4506 ₽")
	assert.Nil(t, err)
	assert.Equal(t, float64(527), p.Price, "Integer should be parsed to float correctly")
	assert.Equal(t, float64(4506), p.Balance, "Integer should be parsed to float correctly")

	p, err = newPurchase("Покупка 5 271.17 ₽, Озон.\nКарта **1111. Баланс: 4 506.22 ₽")
	assert.Nil(t, err)
	assert.Equal(t, 5271.17, p.Price, "Space should be accepted as thousand separator")
	assert.Equal(t, 4506.22, p.Balance, "Non-breaking space (U+00A0) should be accepted as thousand separator")

	p, err = newPurchase("Отмена операции 527,11 ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.Nil(t, err)
	assert.Equal(t, -527.11, p.Price)

	p, err = newPurchase("")
	assert.NotNil(t, err)

	p, err = newPurchase("ABC")
	assert.NotNil(t, err)

	p, err = newPurchase("Покупка ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.NotNil(t, err)

	p, err = newPurchase("Покупка ABC ₽,nОзон. Карта **1111. Баланс: 4506,85 ₽")
	assert.NotNil(t, err)

	p, err = newPurchase("Деньги пришли! 20 000 ₽ на карту\n**1111. Баланс: 21 945,39 ₽")
	assert.NotNil(t, err)
}

func TestNewPurchaseWithTemplate2(t *testing.T) {
	p, err := newPurchase("**1111 Pokupka 1 234 567 AMD Balans 10 000,12 RUR YANDEX GO 16.08.2023 07:36")
	assert.Nil(t, err)
	dt, _ := time.Parse(df, "16.08.2023 07:36")
	assert.Equal(t, dt, p.Time)
	assert.Equal(t, 1234567.0, p.Price)
	assert.Equal(t, "YANDEX GO", p.Merchant)
	assert.Equal(t, "**1111", p.Card)
	assert.Equal(t, 10000.12, p.Balance)
	assert.Equal(t, "֏", p.Currency)
	assert.Equal(t, roundFloat(p.Price*0.251957, 2), p.PriceRUB)

	p, err = newPurchase("**1111 Pokupka 1 234 567 AMD Balans 10 000,12 RUR YANDEX GO ABC123 16.08.2023 07:36")
	assert.Nil(t, err)
	assert.Equal(t, "YANDEX GO ABC123", p.Merchant)

	p, err = newPurchase("**1111 Pokupka 1 234 567 AMD Balans 10 000,12 RUR YANDEX GO 16.08.2023 25:36")
	assert.NotNil(t, err, "Invalid datetime value should not be parsed")
}

func TestStatsForSeveralPurchases(t *testing.T) {
	p1, _ := newPurchase("Покупка 527,11 ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	p2, _ := newPurchase("**1111 Pokupka 1 234 567 AMD Balans 10 000,12 RUR YANDEX GO 16.08.2023 07:36")
	p3, _ := newPurchase("**1111 Pokupka 1 AMD Balans 10 000,12 RUR YANDEX GO 16.08.2023 14:36")
	e := stats.NewExpenses()
	e.Add(p1)
	e.Add(p2)
	e.Add(p3)
	assert.Equal(t, p1.PriceRUB+p2.PriceRUB+p3.PriceRUB, e.Sum())

	s := stats.Stats{
		Expenses: stats.Expenses{
			truncateDay(time.Now()): stats.Expense{Count: 1, Sum: p1.PriceRUB},
			truncateDay(p2.Time):    stats.Expense{Count: 2, Sum: p2.PriceRUB + p3.PriceRUB},
		},
		Count: 3,
		Sum:   p1.PriceRUB + p2.PriceRUB + p3.PriceRUB,
	}
	assert.Equal(t, s.Count, e.Count())
	j, err := json.Marshal(s)
	assert.Nil(t, err)
	j1, err := e.Stats()
	assert.Nil(t, err)
	assert.Equal(t, string(j), j1)
}

func newPurchase(s string) (*purchases.Purchase, error) {
	return purchases.New(time.Now(), s)
}

func roundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func truncateDay(dt time.Time) time.Time {
	return dt.Truncate(time.Hour * 24)
}
