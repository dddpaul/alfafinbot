package stats

import (
	"encoding/json"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"sync"
	"time"
)

var (
	mu sync.Mutex
	df = "2006-01-02"
)

type JsonTime time.Time

func (t JsonTime) MarshalJSON() ([]byte, error) {
	stamp := time.Time(t).Format(df)
	return []byte(stamp), nil
}

type Stats map[JsonTime]float64

type JsonStats struct {
	Stats Stats   `json:"stats"`
	Sum   float64 `json:"sum"`
}

func (s Stats) Add(p *purchases.Purchase) {
	dt := JsonTime(truncateDay(p.Time))
	mu.Lock()
	if _, ok := s[dt]; ok {
		s[dt] = s[dt] + p.PriceRUB
	} else {
		s[dt] = p.PriceRUB
	}
	mu.Unlock()
}

func (s Stats) Get(dt time.Time) float64 {
	return s[JsonTime(truncateDay(dt))]
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
