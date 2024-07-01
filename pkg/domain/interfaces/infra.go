package interfaces

import (
	"context"

	"github.com/m-mizutani/opac"
	"github.com/slack-go/slack"
)

type Slack interface {
	PostMessageContext(ctx context.Context, channelID string, options ...slack.MsgOption) (string, string, error)
}

type Policy interface {
	Query(ctx context.Context, query string, input, output any, options ...opac.QueryOption) error
}
