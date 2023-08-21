package stats

import (
	"encoding/json"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	log "github.com/sirupsen/logrus"
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

// Expenses has Unix timestamp key instead of time.Time because of https://pkg.go.dev/time#Time:
// Note that the Go == operator compares not just the time instant but also the Location and the
// monotonic clock reading. Therefore, Time values should not be used as map or database keys
// without first guaranteeing that the identical Location has been set for all values, which can
// be achieved through use of the UTC or Local method, and that the monotonic clock reading has
// been stripped by setting t = t.Round(0)
type Expenses map[int64]Expense

func (e Expenses) MarshalJSON() ([]byte, error) {
	m := make(map[string]Expense)
	for k, v := range e {
		m[time.Unix(k, 0).Format(time.DateOnly)] = v
	}
	return json.Marshal(m)
}

func (e Expenses) Add(p *purchases.Purchase) {
	dt := truncateDay(p.Time).Unix()
	mu.Lock()
	if _, ok := e[dt]; ok {
		e[dt] = Expense{Count: e[dt].Count + 1, Sum: e[dt].Sum + p.PriceRUB}
	} else {
		e[dt] = Expense{Count: 1, Sum: p.PriceRUB}
	}
	mu.Unlock()
}

func (e Expenses) Get(dt time.Time) Expense {
	return e[truncateDay(dt).Unix()]
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

type Stats struct {
	Expenses Expenses `json:"expenses"`
	Count    int64    `json:"count"`
	Sum      float64  `json:"sum"`
}

func (e Expenses) Stats() (string, error) {
	s := Stats{Expenses: e, Count: e.Count(), Sum: e.Sum()}
	log.Debugf("expenses = %+v, or %v", e, e)
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
