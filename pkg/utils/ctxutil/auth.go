package ctxutil

import (
	"context"

	"github.com/m-mizutani/nounify/pkg/domain/model"
)

type ctxAuthKey string

const (
	ctxAuthGitHubWebhook = ctxAuthKey("github_webhook_auth")
	ctxAuthGoogleIDToken = ctxAuthKey("google_id_token")
)

func WithGitHubWebhookAuth(ctx context.Context, auth *model.GitHubWebhookAuth) context.Context {
	return context.WithValue(ctx, ctxAuthGitHubWebhook, auth)
}

func GitHubWebhookAuth(ctx context.Context) *model.GitHubWebhookAuth {
	auth, ok := ctx.Value(ctxAuthGitHubWebhook).(*model.GitHubWebhookAuth)
	if !ok {
		return nil
	}
	return auth
}

func WithGoogleIDToken(ctx context.Context, token map[string]any) context.Context {
	return context.WithValue(ctx, ctxAuthGoogleIDToken, token)
}

func GoogleIDToken(ctx context.Context) map[string]any {
	token, ok := ctx.Value(ctxAuthGoogleIDToken).(map[string]any)
	if !ok {
		return nil
	}
	return token
}
