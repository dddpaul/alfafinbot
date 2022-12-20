package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/dddpaul/alfafin-bot/pkg/gas"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

const DF = "02.01.2006"
const GAS_ID = "AKfycbyLE6GsIpZL9NvYkgT5odQhS5XTEreq4kZgiaPlRMb3ocZbSUvVlegu88354rNhJwP-zg"

func TestNewPurchase(t *testing.T) {
	p, err := newPurchase("Покупка 527,11 ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.Nil(t, err)
	assert.Equal(t, 527.11, p.Price)
	assert.Equal(t, "Озон", p.Merchant)
	assert.Equal(t, "**1111", p.Card)
	assert.Equal(t, 4506.85, p.Balance)

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

func TestGASClient(t *testing.T) {
	p, err := newPurchase("Покупка 527,11 ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.Nil(t, err)
	client := gas.NewClient(fmt.Sprintf("https://script.google.com/macros/s/%s/exec", GAS_ID))
	client.Request(p)
}

func newPurchase(s string) (*purchases.Purchase, error) {
	return purchases.New(time.Now(), s)
}
