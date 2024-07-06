package usecase

import (
	"context"
	"fmt"

	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
	"github.com/slack-go/slack"
)

func (x *UseCases) HandleMessage(ctx context.Context, schema types.Schema, input *model.MessageQueryInput) error {

	var output model.MessageQueryOutput
	if err := x.policy.Query(ctx, schema.ToQuery(), input, &output); err != nil {
		return goerr.Wrap(err).With("query", schema.ToQuery())
	}
	ctxutil.Logger(ctx).Info("msg query result", "input", input, "output", output)

	for _, msg := range output.Messages {
		attachment := buildSlackMessage(msg)
		options := []slack.MsgOption{
			slack.MsgOptionAttachments(attachment),
		}

		if msg.Emoji != "" { // Emoji has higher priority than Icon
			options = append(options, slack.MsgOptionIconEmoji(msg.Emoji))
		} else if msg.Icon != "" {
			options = append(options, slack.MsgOptionIconURL(msg.Icon))
		}

		if _, _, err := x.slack.PostMessageContext(ctx, msg.Channel, options...); err != nil {
			return goerr.Wrap(err).With("msg", msg)
		}
	}

	return nil
}

var preservedColors = map[string]string{
	"info":    "#2EB67D",
	"warning": "#FFA500",
	"error":   "#FF0000",
}

func buildSlackMessage(msg model.Message) slack.Attachment {
	color := "#2EB67D"
	if msg.Color != "" {
		if preserved, ok := preservedColors[msg.Color]; ok {
			color = preserved
		} else {
			color = msg.Color
		}
	}

	var blockSet []slack.Block

	if msg.Title != "" {
		txt := slack.NewTextBlockObject("plain_text", msg.Title, false, false)
		blockSet = append(blockSet, slack.NewHeaderBlock(txt))
	}

	var body *slack.TextBlockObject
	if msg.Body != "" {
		body = slack.NewTextBlockObject("mrkdwn", msg.Body, false, false)
	}

	fields := make([]*slack.TextBlockObject, len(msg.Fields))
	for i, field := range msg.Fields {
		value := field.Value
		if field.Link != "" {
			value = "<" + field.Link + "|" + field.Value + ">"
		}
		mrkdwn := fmt.Sprintf("*%s*\n%s", field.Name, value)
		fields[i] = slack.NewTextBlockObject("mrkdwn", mrkdwn, false, false)
	}

	blockSet = append(blockSet, slack.NewSectionBlock(body, fields, nil))

	return slack.Attachment{
		Color: color,
		Blocks: slack.Blocks{
			BlockSet: blockSet,
		},
	}
}
