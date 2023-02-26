package transport

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/dddpaul/alfafin-bot/pkg/logger"
	"github.com/dddpaul/alfafin-bot/pkg/proxy"
)

func NewSocksTransport(socks string) http.RoundTripper {
	return &http.Transport{
		Dial: proxy.NewDialer(socks).Dial,
	}
}

func LogRedirect(req *http.Request, via []*http.Request) error {
	logger.Log(req.Context(), nil).WithField("url", req.URL.String()).Debugf("redirect")
	return nil
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
