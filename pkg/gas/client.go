package gas

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

const DF = "2006-01-02 15:04:05 -0700"

type GASConfig struct {
	Url          string
	ClientID     string
	ClientSecret string
	Verbose      bool
}

type Client struct {
	url     *url.URL
	trace   *httptrace.ClientTrace
	verbose bool
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

	var dns, connect, tlsHandshake time.Time
	trace := &httptrace.ClientTrace{
		DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
		DNSDone: func(ddi httptrace.DNSDoneInfo) {
			log.Printf("DNS Done: %v\n", time.Since(dns))
		},

		TLSHandshakeStart: func() { tlsHandshake = time.Now() },
		TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
			log.Printf("TLS Handshake: %v\n", time.Since(tlsHandshake))
		},

		ConnectStart: func(network, addr string) { connect = time.Now() },
		ConnectDone: func(network, addr string, err error) {
			log.Printf("Connect time: %v\n", time.Since(connect))
		},

		GotFirstResponseByte: func() {
			log.Printf("Time from start to first byte: %v\n", time.Since(dns))
		},
	}

	return &Client{
		url:     u,
		trace:   trace,
		verbose: gas.Verbose,
	}
}

func (c *Client) Add(p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Format(DF))
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))

	if c.verbose {
		log.Printf("REQUEST: %v, BODY: &v", c.url.String(), params)
	}

	resp, err := http.PostForm(c.url.String(), params)
	if err != nil {
		return "", err
	}

	r, err := parse(resp)
	if err != nil {
		return "", err
	}

	if c.verbose {
		log.Printf("RESPONSE: %+v", r)
	}

	return r.Message, nil
}

func (c *Client) Get() (string, error) {
	if c.verbose {
		log.Printf("REQUEST: %v", c.url.String())
	}

	ctx := httptrace.WithClientTrace(context.Background(), c.trace)
	req, _ := http.NewRequestWithContext(ctx, "GET", c.url.String(), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	r, err := parse(resp)
	if err != nil {
		return "", err
	}

	if c.verbose {
		log.Printf("RESPONSE: %+v", r)
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
