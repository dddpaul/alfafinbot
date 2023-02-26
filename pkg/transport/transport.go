package transport

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/dddpaul/alfafin-bot/pkg/logger"
	"github.com/dddpaul/alfafin-bot/pkg/proxy"
)

func New(socks string) http.RoundTripper {
	return &http.Transport{
		Dial: proxy.NewDialer(socks).Dial,
	}
}

func NewTrace(ctx context.Context) *httptrace.ClientTrace {
	var start time.Time
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) { start = time.Now() },
		GotFirstResponseByte: func() {
			logger.Log(ctx, nil).WithField("time_to_first_byte_received", time.Since(start)).Tracef("trace")
		},
	}
}
