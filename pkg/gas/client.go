package gas

import (
	"context"
	"encoding/json"
	"errors"
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

const DF = "2006-01-02 15:04:05 -0700"
const MAX_RETRIES = 5

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
	OK Status = iota
	ERROR
)

type Response struct {
	Status  Status `json:"status"`
	Message string `json:"message"`
}

func (r *Response) isError() bool {
	return r.Status != OK
}

type ErrorCode int64

const (
	OTHER ErrorCode = iota
	TIMEOUT
)

type GASError struct {
	message string
	code    ErrorCode
}

func (e *GASError) Error() string {
	return fmt.Sprintf("error message: %s, code: %d", e.message, e.code)
}

func NewClient(ctx context.Context, gas *GASConfig, command string) *Client {
	u, err := url.Parse(gas.Url)
	if err != nil {
		panic(err)
	}
	params := url.Values{}
	params.Add("client_id", gas.ClientID)
	params.Add("client_secret", gas.ClientSecret)
	params.Add("command", command)
	u.RawQuery = params.Encode()

	return &Client{
		url:   u,
		trace: logger.NewTrace(ctx),
		client: &http.Client{
			Transport:     proxy.NewTransport(gas.Socks),
			CheckRedirect: logger.LogRedirect,
		},
	}
}

func (c *Client) Add(ctx context.Context, p *purchases.Purchase) (string, error) {
	params := url.Values{}
	params.Add("time", p.Time.Format(DF))
	params.Add("merchant", p.Merchant)
	params.Add("price", strconv.FormatFloat(p.Price, 'f', 2, 64))
	params.Add("currency", p.Currency)
	params.Add("priceRUB", strconv.FormatFloat(p.PriceRUB, 'f', 2, 64))
	logger.Log(ctx, nil).WithField("url", c.url.String()).WithField("body", fmt.Sprintf("%+v", params)).Debugf("request")

	retry := 1
	for retry <= MAX_RETRIES {
		ctx = logger.WithRetryAttempt(ctx, retry)
		req, err := http.NewRequestWithContext(
			httptrace.WithClientTrace(ctx, c.trace),
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

		r, err := parse(resp)
		if err != nil {
			var gasError *GASError
			if errors.As(err, &gasError) {
				if gasError.code == TIMEOUT {
					logger.Log(ctx, err).Errorf("error")
					retry++
					time.Sleep(5 * time.Second)
					continue
				}
			}
			return "", err
		}
		logger.Log(ctx, nil).WithField("body", fmt.Sprintf("%+v", r)).Debugf("response")

		return r.Message, nil
	}

	return "", fmt.Errorf("all %d retries to add purchase was failed", MAX_RETRIES)
}

func (c *Client) Get(ctx context.Context) (string, error) {
	logger.Log(ctx, nil).WithField("url", c.url.String()).Debugf("request")

	req, err := http.NewRequestWithContext(
		httptrace.WithClientTrace(ctx, c.trace),
		"GET",
		c.url.String(),
		nil)
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
	logger.Log(ctx, nil).WithField("body", fmt.Sprintf("%+v", r)).Debugf("response")

	return r.Message, nil
}

// Parse HTTP response from Google App Script
func parse(resp *http.Response) (*Response, error) {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	s := string(data)
	if strings.Contains(s, ".errorMessage") {
		message := after(s, "Error: ")
		if strings.HasPrefix(message, "Timeout") {
			return nil, &GASError{message: message, code: TIMEOUT}
		}
		return nil, &GASError{message: message, code: OTHER}
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
