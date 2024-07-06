package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/domain/interfaces"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
)

type config struct {
	policy                    interfaces.Policy
	githubSecrets             []string
	validateGitHubActionToken bool
	validateGoogleIDToken     bool
}

type Option func(*config)

func WithPolicy(policy interfaces.Policy) Option {
	return func(cfg *config) {
		cfg.policy = policy
	}
}

func WithGitHubSecret(secret string) Option {
	return func(cfg *config) {
		cfg.githubSecrets = append(cfg.githubSecrets, secret)
	}
}

func WithGitHubActionTokenValidation() Option {
	return func(cfg *config) {
		cfg.validateGitHubActionToken = true
	}
}

func WithGoogleIDTokenValidation() Option {
	return func(cfg *config) {
		cfg.validateGoogleIDToken = true
	}
}

func New(uc interfaces.UseCases, options ...Option) http.Handler {
	var cfg config
	for _, opt := range options {
		opt(&cfg)
	}

	route := chi.NewRouter()
	route.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK:" + types.AppVersion))
	})
	route.Route("/msg", func(r chi.Router) {
		r.Use(logger)
		for _, secret := range cfg.githubSecrets {
			r.Use(authGitHubWebhook(secret))
		}
		if cfg.validateGitHubActionToken {
			r.Use(authGitHubActionToken())
		}
		if cfg.validateGoogleIDToken {
			r.Use(authGoogleIDToken())
		}

		if cfg.policy != nil {
			r.Use(authWithPolicy(cfg.policy))
		}

		r.Post("/*", handleMessage(uc))
	})

	return route
}

func handleError(ctx context.Context, w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	var xErr types.Error
	if errors.As(err, &xErr) {
		code = xErr.Code()
	}

	ctxutil.Logger(ctx).Error("HTTP error", "err", err, "code", code)

	http.Error(w, err.Error(), code)
}

func newMessageQueryInput(r *http.Request) (*model.MessageQueryInput, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).With("method", r.Method).With("path", r.URL.Path)
	}

	var data any
	switch strings.ToLower(r.Header.Get("Content-Type")) {
	case "application/json":
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("body", string(body))
		}

	default:
		data = string(body)
	}

	headers := map[string]string{}
	for key := range r.Header {
		headers[key] = r.Header.Get(key)
	}

	return &model.MessageQueryInput{
		Method: r.Method,
		Path:   r.URL.Path,
		Header: headers,
		Body:   data,
		Auth:   authFromContext(r.Context()),
	}, nil
}

func handleMessage(uc interfaces.UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		schema := strings.Replace(chi.URLParam(r, "*"), "/", ".", -1)

		input, err := newMessageQueryInput(r)
		if err != nil {
			handleError(ctx, w, err)
			return
		}

		if err := uc.HandleMessage(ctx, types.Schema(schema), input); err != nil {
			handleError(ctx, w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
