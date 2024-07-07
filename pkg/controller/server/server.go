package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/nounify/pkg/domain/interfaces"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/errutil"
)

type config struct {
	policy                    interfaces.Policy
	githubSecrets             []string
	validateGitHubActionToken bool
	validateGoogleIDToken     bool
	validateAmazonSNS         bool
	authErrStatusCode         int
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

func WithAmazonSNSValidation() Option {
	return func(cfg *config) {
		cfg.validateAmazonSNS = true
	}
}

func WithAuthErrStatusCode(code int) Option {
	return func(cfg *config) {
		cfg.authErrStatusCode = code
	}
}

func New(uc interfaces.UseCases, options ...Option) http.Handler {
	cfg := &config{
		authErrStatusCode: http.StatusForbidden,
	}
	for _, opt := range options {
		opt(cfg)
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
		if cfg.validateAmazonSNS {
			r.Use(authAmazonSNS())
		}

		if cfg.policy != nil {
			r.Use(authWithPolicy(cfg.policy, cfg.authErrStatusCode))
		}

		r.Post("/*", handleMessage(uc))
	})

	return route
}

type handleErrorOpt struct {
	forceCode int
}

type handleErrorOption func(o *handleErrorOpt)

func handleErrorWithForceCode(code int) handleErrorOption {
	return func(o *handleErrorOpt) {
		o.forceCode = code
	}
}

func handleError(ctx context.Context, w http.ResponseWriter, err error, options ...handleErrorOption) {
	opt := &handleErrorOpt{}
	for _, o := range options {
		o(opt)
	}

	code := http.StatusInternalServerError
	var xErr types.Error
	if errors.As(err, &xErr) {
		code = xErr.Code()
	}

	errutil.Handle(ctx, "HTTP error", err)

	if opt.forceCode > 0 {
		code = opt.forceCode
	}
	http.Error(w, err.Error(), code)
}

func newMessageQueryInput(r *http.Request) (*model.MessageQueryInput, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).With("method", r.Method).With("path", r.URL.Path)
	}

	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).With("method", r.Method).With("path", r.URL.Path)
	}

	var data any
	switch mediaType {
	case "application/json":
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("body", string(body))
		}

	case "text/plain":
		if r.Header.Get("X-Amz-Sns-Message-Id") != "" {
			if err := json.Unmarshal(body, &data); err != nil {
				return nil, goerr.Wrap(types.ErrInvalidInput.Wrap(err)).
					With("method", r.Method).
					With("path", r.URL.Path).
					With("body", string(body))
			}
		} else {
			data = string(body)
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
