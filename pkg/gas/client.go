package gas

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/dddpaul/alfafin-bot/pkg/purchases"
	"github.com/dddpaul/alfafin-bot/pkg/transport"
)

const DF = "2006-01-02 15:04:05 -0700"

type GASConfig struct {
	Url          string
	Socks        string
	ClientID     string
	ClientSecret string
}

type Client struct {
	url    *url.URL
	trace  *httptrace.ClientTrace
	client *http.Client
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

func NewClient(gas *GASConfig, command string) *Client {
	u, err := url.Parse(gas.Url)
	if err != nil {
		log.Panic(err)
	}
	params := url.Values{}
	params.Add("client_id", gas.ClientID)
	params.Add("client_secret", gas.ClientSecret)
	params.Add("command", command)
	u.RawQuery = params.Encode()

	return &Client{
		url:   u,
		trace: transport.NewTrace(),
		client: &http.Client{
			Transport:     transport.NewSocksTransport(gas.Socks),
			CheckRedirect: transport.LogRedirect,
		},
	}
}

func (c *Client) Add(p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Format(DF))
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))
	log.Debugf("REQUEST: %v, BODY: &v", c.url.String(), params)

	resp, err := http.PostForm(c.url.String(), params)
	if err != nil {
		return "", err
	}

	r, err := parse(resp)
	if err != nil {
		return "", err
	}
	log.Debugf("RESPONSE: %+v", r)

	return r.Message, nil
}

func (c *Client) Get() (string, error) {
	log.Debugf("REQUEST: %v", c.url.String())

	ctx := httptrace.WithClientTrace(context.Background(), c.trace)
	req, err := http.NewRequestWithContext(ctx, "GET", c.url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	r, err := parse(resp)
	if err != nil {
		return "", err
	}
	log.Debugf("RESPONSE: %+v", r)

	return r.Message, nil
}

func logRedirect(req *http.Request, via []*http.Request) error {
	log.Debugf("REDIRECT: %v", req.URL.String())
	return nil
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
