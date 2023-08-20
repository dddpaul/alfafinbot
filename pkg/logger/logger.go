package logger

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const MESSAGE_ID = "message_id"
const RETRY_ATTEMPT = "retry"

func WithMessageID(id int) context.Context {
	return context.WithValue(context.Background(), MESSAGE_ID, id)
}

func WithRetryAttempt(ctx context.Context, r int) context.Context {
	return context.WithValue(ctx, RETRY_ATTEMPT, r)
}

func NewTrace(ctx context.Context) *httptrace.ClientTrace {
	var start time.Time
	return &httptrace.ClientTrace{
		GetConn: func(hostPort string) { start = time.Now() },
		GotFirstResponseByte: func() {
			Log(ctx, nil).WithField("time_to_first_byte_received", time.Since(start)).Tracef("trace")
		},
	}
}

func Log(ctx context.Context, err error) *log.Entry {
	entry := log.WithContext(ctx)
	if err != nil {
		entry = entry.WithField("error", err)
	}
	if messageID := ctx.Value(MESSAGE_ID); messageID != nil {
		entry = entry.WithField(MESSAGE_ID, messageID)
	}
	if retry := ctx.Value(RETRY_ATTEMPT); retry != nil {
		entry = entry.WithField(RETRY_ATTEMPT, retry)
	}
	return entry
}

func LogRedirect(req *http.Request, via []*http.Request) error {
	Log(req.Context(), nil).WithField("url", escape(req.URL.String())).Debugf("redirect")
	return nil
}

func escape(s string) string {
	return strings.Replace(strings.Replace(s, "\n", "", -1), "\r", "", -1)
}
