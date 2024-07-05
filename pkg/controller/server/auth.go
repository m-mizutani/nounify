package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v62/github"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/domain/interfaces"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
)

type middlewareFunc func(next http.Handler) http.Handler

func trimToken(token string) string {
	e := min(len(token), 8)
	return token[:e] + "..."
}

func authGitHubWebhook(secret string) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if not GitHub webhook
			if r.Header.Get("X-GitHub-Event") == "" {
				next.ServeHTTP(w, r)
				return
			}

			payload, err := github.ValidatePayload(r, []byte(secret))
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			auth := model.NewGitHubAppAuth(r)
			r.Body = io.NopCloser(bytes.NewReader(payload))
			r = r.WithContext(ctxutil.WithGitHubAppAuth(r.Context(), auth))

			next.ServeHTTP(w, r)
		})
	}
}

func validateGitHubActionToken(authHdr string) (model.GitHubActionToken, error) {
	hdr := strings.SplitN(authHdr, " ", 2)

	// Skip if not Bearer token
	if len(hdr) != 2 || hdr[0] != "Bearer" {
		return nil, nil
	}

	jwksURL := "https://token.actions.githubusercontent.com/.well-known/jwks"

	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	token, err := jwt.ParseString(hdr[1], jwt.WithKeySet(set))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to parse JWT token as GitHub Action token").With("token", trimToken(hdr[1]))
	}

	claims, err := token.AsMap(context.Background())
	if err != nil {
		return nil, goerr.Wrap(err, "failed to convert JWT token to map").With("token", trimToken(hdr[1]))
	}

	return claims, nil
}

func authGitHubActionToken() middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := validateGitHubActionToken(r.Header.Get("Authorization"))
			if claims == nil {
				if err != nil {
					ctxutil.Logger(r.Context()).Warn("failed to parse JWT token", "err", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(ctxutil.WithGitHubActionToken(r.Context(), claims))
			next.ServeHTTP(w, r)
		})
	}
}

func validateGoogleIDToken(authHdr string) (map[string]any, error) {
	hdr := strings.SplitN(authHdr, " ", 2)

	// Skip if not Bearer token
	if len(hdr) != 2 || hdr[0] != "Bearer" {
		return nil, nil
	}

	jwksURL := "https://www.googleapis.com/oauth2/v3/certs"

	set, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		return nil, goerr.Wrap(err)
	}

	token, err := jwt.ParseString(hdr[1], jwt.WithKeySet(set))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to parse JWT token as Google ID Token").With("token", trimToken(hdr[1]))
	}

	claims, err := token.AsMap(context.Background())
	if err != nil {
		return nil, goerr.Wrap(err, "failed to convert JWT token to map").With("token", trimToken(hdr[1]))
	}

	return claims, nil
}

func authGoogleIDToken() middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := validateGoogleIDToken(r.Header.Get("Authorization"))
			if claims == nil {
				if err != nil {
					ctxutil.Logger(r.Context()).Warn("failed to fetch JWK set", "err", err)
				}
				next.ServeHTTP(w, r)
				return
			}

			r = r.WithContext(ctxutil.WithGoogleIDToken(r.Context(), claims))
			next.ServeHTTP(w, r)
		})
	}
}

func authWithPolicy(policy interfaces.Policy) middlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			input := model.AuthQueryInput{
				Method: r.Method,
				Path:   r.URL.Path,
				Header: map[string]string{},
			}

			for key := range r.Header {
				input.Header[key] = r.Header.Get(key)
			}

			if claims := ctxutil.GoogleIDToken(r.Context()); claims != nil {
				input.Auth.Google = claims
			}
			if auth := ctxutil.GitHubAppAuth(r.Context()); auth != nil {
				input.Auth.GitHub.App = auth
			}
			if auth := ctxutil.GitHubActionToken(r.Context()); auth != nil {
				input.Auth.GitHub.Action = auth
			}

			var output model.AuthQueryOutput
			if err := policy.Query(r.Context(), "data.auth", input, &output); err != nil {
				handleError(w, err)
				return
			}
			ctxutil.Logger(r.Context()).Debug("auth query result", "input", input, "output", output)

			if !output.Allow {
				handleError(w, types.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
