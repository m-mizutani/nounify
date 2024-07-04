package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/nounify/pkg/domain/interfaces"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
)

type config struct {
	policy                interfaces.Policy
	githubSecrets         []string
	validateGoogleIDToken bool
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
		if cfg.validateGoogleIDToken {
			r.Use(authGoogleIDToken())
		}

		if cfg.policy != nil {
			r.Use(authWithPolicy(cfg.policy))
		}

		r.Post("/{schema}", handleMessage(uc))
	})

	return route
}

func handleError(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	var xErr types.Error
	if errors.As(err, &xErr) {
		code = xErr.Code()
	}

	http.Error(w, err.Error(), code)
}

func handleMessage(uc interfaces.UseCases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schema := chi.URLParam(r, "schema")

		input, err := model.NewMessageQueryInput(r)
		if err != nil {
			handleError(w, err)
			return
		}

		if err := uc.HandleMessage(r.Context(), types.Schema(schema), input); err != nil {
			handleError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
