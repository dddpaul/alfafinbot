package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

const DF = "02.01.2006"

func TestNewPurchase(t *testing.T) {
	p, err := purchases.New("Покупка 527,11 ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.Nil(t, err)
	assert.Equal(t, 527.11, p.Price)
	assert.Equal(t, "Озон", p.Merchant)
	assert.Equal(t, "**1111", p.Card)
	assert.Equal(t, 4506.85, p.Balance)

	p, err = purchases.New("Покупка 527.11 ₽, Озон.\nКарта **1111. Баланс: 4506.85 ₽")
	assert.Nil(t, err)
	assert.Equal(t, 527.11, p.Price, "Dot should be accepted as decimal separator")
	assert.Equal(t, 4506.85, p.Balance, "Dot should be accepted as decimal separator")

	p, err = purchases.New("Покупка 527 ₽, Озон.\nКарта **1111. Баланс: 4506 ₽")
	assert.Nil(t, err)
	assert.Equal(t, float64(527), p.Price, "Integer should be parsed to float correctly")
	assert.Equal(t, float64(4506), p.Balance, "Integer should be parsed to float correctly")

	p, err = purchases.New("")
	assert.NotNil(t, err)

	p, err = purchases.New("ABC")
	assert.NotNil(t, err)

	p, err = purchases.New("Покупка ₽, Озон.\nКарта **1111. Баланс: 4506,85 ₽")
	assert.NotNil(t, err)

	p, err = purchases.New("Покупка ABC ₽,nОзон. Карта **1111. Баланс: 4506,85 ₽")
	assert.NotNil(t, err)

	p, err = purchases.New("Деньги пришли! 20 000 ₽ на карту\n**1111. Баланс: 21 945,39 ₽")
	assert.NotNil(t, err)
}
