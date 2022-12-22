package gas

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

type Client struct {
	url *url.URL
}

func NewClient(rawURL string) *Client {
	url, err := url.Parse(rawURL)
	if err != nil {
		log.Panic(err)
	}
	return &Client{url}
}

func (c *Client) Add(p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Local().String())
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))
	c.url.RawQuery = params.Encode()

	res, err := http.Get(c.url.String())
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	res.Body.Close()
	return string(data), nil
}
