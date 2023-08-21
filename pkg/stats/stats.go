package stats

import (
	"encoding/json"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"sync"
	"time"
)

var mu sync.Mutex

type Stats map[time.Time]float64

type JsonStats struct {
	Stats Stats   `json:"stats"`
	Sum   float64 `json:"sum"`
}

func (s Stats) Add(p *purchases.Purchase) {
	dt := truncateDay(p.Time)
	mu.Lock()
	if _, ok := s[dt]; ok {
		s[dt] = s[dt] + p.PriceRUB
	} else {
		s[dt] = p.PriceRUB
	}
	mu.Unlock()
}

func (s Stats) Get(dt time.Time) float64 {
	return s[truncateDay(dt)]
}

func (s Stats) Sum() float64 {
	var sum float64 = 0
	for _, v := range s {
		sum = sum + v
	}
	return sum
}

func (s Stats) Stats() (string, error) {
	j := JsonStats{Stats: s, Sum: s.Sum()}
	b, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func New() Stats {
	return make(Stats)
}

func truncateDay(dt time.Time) time.Time {
	return dt.Truncate(time.Hour * 24)
}
