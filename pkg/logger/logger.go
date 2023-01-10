package logger

import (
	"context"

	log "github.com/sirupsen/logrus"
)

func ServerLog(ctx context.Context, err error) *log.Entry {
	entry := log.WithField("package", "server")
	if err != nil {
		entry = entry.WithField("error", err)
	}
	if messageID := ctx.Value("message_id"); messageID != nil {
		entry = entry.WithField("message_id", messageID)
	}
	return entry
}
