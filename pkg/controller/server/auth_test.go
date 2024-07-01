package server_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/nounify/pkg/controller/server"
	"github.com/m-mizutani/nounify/pkg/domain/mock"
	"github.com/m-mizutani/nounify/pkg/domain/model"
	"github.com/m-mizutani/nounify/pkg/domain/types"
	"github.com/m-mizutani/nounify/pkg/utils/testutil"
	"github.com/m-mizutani/opac"
)

//go:embed testdata/github_webhook_example.json
var githubWebhookExample []byte

//go:embed testdata/policy_github_auth.rego
var policyGitHubAuth string

//go:embed testdata/policy_google_auth.rego
var policyGoogleAuth string

func TestGitHubAuth(t *testing.T) {
	const testSecret = "test-test-test"
	ucMock := &mock.UseCasesMock{
		HandleMessageFunc: func(ctx context.Context, schema types.Schema, input *model.MessageQueryInput) error {
			return nil
		},
	}
	w := httptest.NewRecorder()
	policy, err := opac.New(opac.Data(map[string]string{"auth": policyGitHubAuth}))
	gt.NoError(t, err)

	h := hmac.New(sha256.New, []byte(testSecret))
	h.Write(githubWebhookExample)
	signature := hex.EncodeToString(h.Sum(nil))

	req := httptest.NewRequest("POST", "/msg/github", bytes.NewReader(githubWebhookExample))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", "sha256="+signature)

	mux := server.New(ucMock,
		server.WithGitHubSecret(testSecret),
		server.WithPolicy(policy),
	)
	mux.ServeHTTP(w, req)

	gt.Equal(t, w.Code, 200)
	gt.A(t, ucMock.HandleMessageCalls()).Length(1)
}

func TestGoogleIDTokenAuth(t *testing.T) {
	token := testutil.LoadEnv(t, "TEST_GOOGLE_ID_TOKEN")

	type testCase struct {
		newReq     func() *http.Request
		expectCode int
		expectCall int
	}

	runTest := func(tc testCase) func(t *testing.T) {
		return func(t *testing.T) {
			ucMock := &mock.UseCasesMock{
				HandleMessageFunc: func(ctx context.Context, schema types.Schema, input *model.MessageQueryInput) error {
					return nil
				},
			}

			w := httptest.NewRecorder()
			policy, err := opac.New(opac.Data(map[string]string{"auth": policyGoogleAuth}))
			gt.NoError(t, err)

			mux := server.New(ucMock,
				server.WithGoogleIDTokenValidation(),
				server.WithPolicy(policy),
			)
			mux.ServeHTTP(w, tc.newReq())

			gt.Equal(t, w.Code, tc.expectCode)
			gt.A(t, ucMock.HandleMessageCalls()).Length(tc.expectCall)
		}
	}

	t.Run("With valid token", runTest(testCase{
		newReq: func() *http.Request {
			req := httptest.NewRequest("POST", "/msg/google", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			return req
		},
		expectCode: http.StatusOK,
		expectCall: 1,
	}))

	t.Run("Without token", runTest(testCase{
		newReq: func() *http.Request {
			return httptest.NewRequest("POST", "/msg/google", nil)
		},
		expectCode: http.StatusForbidden,
		expectCall: 0,
	}))
}
