package gas

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

const DF = "2006-01-02 15:04:05 -0700"

type Client struct {
	u *url.URL
}

func NewClient(rawURL string) *Client {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Panic(err)
	}
	return &Client{u}
}

func (c *Client) Add(p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Format(DF))
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))

	res, err := http.PostForm(c.u.String(), params)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	res.Body.Close()

	s := string(data)
	if strings.Contains(s, ".errorMessage") {
		return "", errors.New("GAS: " + after(s, "Error: "))
	}

	return s, nil
}

func (c *Client) CurrentMonthSum() (string, error) {
	res, err := http.Get(c.u.String())
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	res.Body.Close()

	s := string(data)
	if strings.Contains(s, ".errorMessage") {
		return "", errors.New("GAS: " + after(s, "Error: "))
	}

	return s, nil
}

// Get substring after a string
func after(s string, a string) string {
	pos := strings.LastIndex(s, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(s) {
		return ""
	}
	return s[adjustedPos : len(s)-1]
}
