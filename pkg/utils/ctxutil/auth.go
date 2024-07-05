package ctxutil

import (
	"context"

	"github.com/m-mizutani/nounify/pkg/domain/model"
)

type ctxAuthKey string

const (
	ctxAuthGitHubApp     = ctxAuthKey("github_app_auth")
	ctxAuthGitHubAction  = ctxAuthKey("github_action_auth")
	ctxAuthGoogleIDToken = ctxAuthKey("google_id_token")
)

func WithGitHubAppAuth(ctx context.Context, auth *model.GitHubAppAuth) context.Context {
	return context.WithValue(ctx, ctxAuthGitHubApp, auth)
}

func GitHubAppAuth(ctx context.Context) *model.GitHubAppAuth {
	auth, ok := ctx.Value(ctxAuthGitHubApp).(*model.GitHubAppAuth)
	if !ok {
		return nil
	}
	return auth
}

func WithGitHubActionToken(ctx context.Context, auth model.GitHubActionToken) context.Context {
	return context.WithValue(ctx, ctxAuthGitHubAction, auth)
}

func GitHubActionToken(ctx context.Context) model.GitHubActionToken {
	auth, ok := ctx.Value(ctxAuthGitHubAction).(model.GitHubActionToken)
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
