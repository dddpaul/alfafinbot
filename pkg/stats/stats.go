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

type Expense struct {
	Count int64   `json:"count"`
	Sum   float64 `json:"sum"`
}

type Expenses map[time.Time]Expense

type Stats struct {
	Expenses Expenses `json:"expenses"`
	Count    int64    `json:"count"`
	Sum      float64  `json:"sum"`
}

func (e Expenses) MarshalJSON() ([]byte, error) {
	m := make(map[string]Expense)
	for k, v := range e {
		m[k.Format(time.DateOnly)] = v
	}
	return json.Marshal(m)
}

func (e Expenses) Add(p *purchases.Purchase) {
	dt := truncateDay(p.Time)
	mu.Lock()
	if _, ok := e[dt]; ok {
		e[dt] = Expense{Count: e[dt].Count + 1, Sum: e[dt].Sum + p.PriceRUB}
	} else {
		e[dt] = Expense{Count: 1, Sum: p.PriceRUB}
	}
	mu.Unlock()
}

func (e Expenses) Get(dt time.Time) Expense {
	return e[truncateDay(dt)]
}

func (e Expenses) Count() int64 {
	var count int64 = 0
	for _, v := range e {
		count = count + v.Count
	}
	return count
}

func (e Expenses) Sum() float64 {
	var sum float64 = 0
	for _, v := range e {
		sum = sum + v.Sum
	}
	return sum
}

func (e Expenses) Stats() (string, error) {
	s := Stats{Expenses: e, Count: e.Count(), Sum: e.Sum()}
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func NewExpenses() Expenses {
	return make(Expenses)
}

func truncateDay(dt time.Time) time.Time {
	return dt.Truncate(time.Hour * 24)
}
