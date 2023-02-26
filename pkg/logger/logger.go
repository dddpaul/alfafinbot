package logger

import (
	"context"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const MESSAGE_ID = "message_id"

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
