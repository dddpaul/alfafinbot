package logger

import (
	"context"
	"net/http"
	"net/http/httptrace"
	"time"

	log "github.com/sirupsen/logrus"
)

const MESSAGE_ID = "message_id"

func WithMessageID(id int) context.Context {
	return context.WithValue(context.Background(), MESSAGE_ID, id)
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
	return entry
}

func LogRedirect(req *http.Request, via []*http.Request) error {
	Log(req.Context(), nil).WithField("url", req.URL.String()).Debugf("redirect")
	return nil
}
