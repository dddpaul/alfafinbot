package gas

import (
	"encoding/json"
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

type Status int64

const (
	OK    Status = 0
	ERROR        = 1
)

type Response struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

func (r *Response) isError() bool {
	return r.Status != OK
}

func NewClient(rawURL string, id string, secret string) *Client {
	u, err := url.Parse(rawURL)
	if err != nil {
		log.Panic(err)
	}
	params := url.Values{}
	params.Add("client_id", id)
	params.Add("client_secret", secret)
	u.RawQuery = params.Encode()
	return &Client{u}
}

func (c *Client) Add(p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Format(DF))
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))

	resp, err := http.PostForm(c.u.String(), params)
	if err != nil {
		return "", err
	}

	r, err := parse(resp)
	if err != nil {
		return "", err
	}

	return r.Message, nil
}

func (c *Client) CurrentMonthSum() (string, error) {
	params := url.Values{}
	params.Add("command", "month")
	c.u.RawQuery += "&" + params.Encode()

	resp, err := http.Get(c.u.String())
	if err != nil {
		return "", err
	}

	r, err := parse(resp)
	if err != nil {
		return "", err
	}

	return r.Message, nil
}

// Parse HTTP response from Google App Script
func parse(resp *http.Response) (*Response, error) {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	s := string(data)
	if strings.Contains(s, ".errorMessage") {
		return nil, errors.New("GAS: " + after(s, "Error: "))
	}

	r := &Response{}
	if err = json.Unmarshal(data, r); err != nil {
		return nil, err
	}

	if r.isError() {
		return nil, errors.New("GAS: " + r.Message)
	}

	return r, nil
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
