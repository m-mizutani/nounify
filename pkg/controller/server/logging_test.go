package server_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/nounify/pkg/controller/server"
	"github.com/m-mizutani/nounify/pkg/utils/ctxutil"
	"github.com/m-mizutani/nounify/pkg/utils/logging"
)

func TestRedactAuthorization(t *testing.T) {
	route := chi.NewRouter()
	route.Route("/msg", func(r chi.Router) {
		r.Use(server.Logger)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
	})

	var buf bytes.Buffer
	logger := gt.R1(logging.New(&buf, "debug", "json")).NoError(t)
	ctx := ctxutil.WithLogger(context.Background(), logger)
	r := httptest.NewRequest(http.MethodGet, "/msg", nil).WithContext(ctx)
	r.Header.Set("Authorization", "Bearer secret-token")
	r.Header.Set("X-Api-Key", "api-key")

	w := httptest.NewRecorder()
	route.ServeHTTP(w, r)
	gt.Equal(t, w.Code, http.StatusOK)
	gt.S(t, buf.String()).Contains("api-key").NotContains("secret-token")
}
