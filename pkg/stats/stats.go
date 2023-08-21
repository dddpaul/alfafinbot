package stats

import (
	"encoding/json"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"sync"
	"time"
)

var (
	mu sync.Mutex
)

type Expenses map[time.Time]float64

type Stats struct {
	Expenses Expenses `json:"expenses"`
	Sum      float64  `json:"sum"`
}

func (e Expenses) MarshalJSON() ([]byte, error) {
	m := make(map[string]float64)
	for k, v := range e {
		m[k.Format(time.DateOnly)] = v
	}
	return json.Marshal(m)
}

func (e Expenses) Add(p *purchases.Purchase) {
	dt := truncateDay(p.Time)
	mu.Lock()
	if _, ok := e[dt]; ok {
		e[dt] = e[dt] + p.PriceRUB
	} else {
		e[dt] = p.PriceRUB
	}
	mu.Unlock()
}

func (e Expenses) Get(dt time.Time) float64 {
	return e[truncateDay(dt)]
}

func (e Expenses) Sum() float64 {
	var sum float64 = 0
	for _, v := range e {
		sum = sum + v
	}
	return sum
}

func (e Expenses) Stats() (string, error) {
	s := Stats{Expenses: e, Sum: e.Sum()}
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func New() Expenses {
	return make(Expenses)
}

func truncateDay(dt time.Time) time.Time {
	return dt.Truncate(time.Hour * 24)
}
