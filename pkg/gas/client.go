package gas

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dddpaul/alfafin-bot/pkg/logger"
	"github.com/dddpaul/alfafin-bot/pkg/proxy"
	"github.com/dddpaul/alfafin-bot/pkg/purchases"
)

const MAX_RETRIES = 5

type Client struct {
	url    *url.URL
	trace  *httptrace.ClientTrace
	client *http.Client
}

type Status int64

const (
	OK Status = iota
	ERROR
	TEMPORAL_ERROR
)

type Response struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

func (r *Response) isError() bool {
	return r.Status != OK
}

func (r *Response) isTemporalError() bool {
	return r.Status == TEMPORAL_ERROR
}

func NewClient(u string, socks string, id string, secret string) *Client {
	u1, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	params := url.Values{}
	params.Add("client_id", id)
	params.Add("client_secret", secret)
	u1.RawQuery = params.Encode()

	return &Client{
		url:   u1,
		trace: nil,
		client: &http.Client{
			Transport:     proxy.NewTransport(socks),
			CheckRedirect: logger.LogRedirect,
		},
	}
}

func (c *Client) Add(ctx context.Context, p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Format(time.RFC3339))
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))
	params.Add("currency", p.Currency)
	params.Add("priceRUB", strconv.FormatFloat(p.PriceRUB, 'f', 2, 64))
	logger.Log(ctx, nil).WithField("url", c.url.String()).WithField("body", fmt.Sprintf("%+v", params)).Debugf("request")

	retry := 1
	for retry <= MAX_RETRIES {
		ctx = logger.WithRetryAttempt(ctx, retry)
		req, err := http.NewRequestWithContext(
			httptrace.WithClientTrace(ctx, logger.NewTrace(ctx)),
			"POST",
			c.url.String(),
			strings.NewReader(params.Encode()))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.client.Do(req)
		if err != nil {
			logger.Log(ctx, err).Errorf("error")
			retry++
			time.Sleep(5 * time.Second)
			continue
		}

		r := parse(ctx, resp)
		logger.Log(ctx, nil).WithField("body", fmt.Sprintf("%+v", r)).Debugf("response")

		if r.isError() {
			err = fmt.Errorf("code %d", r.Status)
			if r.isTemporalError() {
				retry++
				logger.Log(ctx, err).Errorf("Waiting for %d seconds till next retry attempt %d", 5, retry)
				time.Sleep(5 * time.Second)
				continue
			}
			return "", err
		}

		return r.Message, nil
	}

	return "", fmt.Errorf("all %d retries to add purchase was failed", MAX_RETRIES)
}

func (c *Client) Get(ctx context.Context, command string) (string, error) {
	params := url.Values{}
	params.Add("command", command)
	u := c.url.String() + "&" + params.Encode()
	logger.Log(ctx, nil).WithField("url", u).Debugf("request")

	req, err := http.NewRequestWithContext(
		httptrace.WithClientTrace(ctx, logger.NewTrace(ctx)),
		"GET",
		u,
		nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	r := parse(ctx, resp)
	return r.Message, nil
}

// Parse HTTP response from Google App Script
func parse(ctx context.Context, resp *http.Response) *Response {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{Status: ERROR, Message: err.Error()}
	}
	err = resp.Body.Close()
	if err != nil {
		return &Response{Status: ERROR, Message: err.Error()}
	}

	s := string(data)
	if strings.Contains(s, ".errorMessage") {
		logger.Log(ctx, err).WithField("body", s).Errorf("Google apps HTML error")
		return &Response{Status: ERROR, Message: after(s, "Error: ")}
	}

	r := &Response{}
	if err = json.Unmarshal(data, r); err != nil {
		logger.Log(ctx, err).WithField("body", s).Errorf("error")
		return &Response{Status: ERROR, Message: err.Error()}
	}

	return r
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
