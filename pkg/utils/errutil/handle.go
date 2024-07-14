package errutil

import (
	"context"

	"github.com/getsentry/sentry-go"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
)

func Handle(ctx context.Context, msg string, err error) {
	var goErr *goerr.Error
	if err != nil {
		goErr = goerr.Unwrap(err)
	}

	// Sending error to Sentry
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		if goErr != nil {
			for k, v := range goErr.Values() {
				scope.SetExtra(k, v)
			}
		}
	})
	evID := hub.CaptureException(err)

	ctxutil.Logger(ctx).Error(msg,
		"error", err,
		"sentry.EventID", evID,
	)
}
