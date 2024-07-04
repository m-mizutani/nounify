package server_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/gt"

	"github.com/m-mizutani/nounify/pkg/controller/server"
	"github.com/m-mizutani/nounify/pkg/domain/mock"
)

func TestHealthCheck(t *testing.T) {
	ucMock := mock.UseCasesMock{}
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	mux := server.New(&ucMock)
	mux.ServeHTTP(w, r)
	gt.Equal(t, w.Code, http.StatusOK)
}

func TestServerError(t *testing.T) {
	type testCase struct {
		req      func() *http.Request
		expCode  int
		expBody  string
		testMock func(t *testing.T, ucMock *mock.UseCasesMock)
	}

	test := func(tc testCase) func(*testing.T) {
		return func(t *testing.T) {
			ucMock := mock.UseCasesMock{}
			w := httptest.NewRecorder()

			mux := server.New(&ucMock)
			mux.ServeHTTP(w, tc.req())

			gt.Equal(t, w.Code, tc.expCode)
			if tc.expBody != "" {
				gt.Equal(t, w.Body.String(), tc.expBody)
			}
			if tc.testMock != nil {
				tc.testMock(t, &ucMock)
			}
		}
	}

	t.Run("invalid path", test(testCase{
		req: func() *http.Request {
			return httptest.NewRequest(http.MethodGet, "/invalid", nil)
		},
		expCode: http.StatusNotFound,
		testMock: func(t *testing.T, ucMock *mock.UseCasesMock) {
			gt.A(t, ucMock.HandleMessageCalls()).Length(0)
		},
	}))

	t.Run("invalid data", test(testCase{
		req: func() *http.Request {
			r := httptest.NewRequest(http.MethodPost, "/msg/schema", bytes.NewReader([]byte("invalid")))
			r.Header.Set("Content-Type", "application/json")
			return r
		},
		expCode: http.StatusBadRequest,
		expBody: ": invalid input: invalid character 'i' looking for beginning of value\n",
		testMock: func(t *testing.T, ucMock *mock.UseCasesMock) {
			gt.A(t, ucMock.HandleMessageCalls()).Length(0)
		},
	}))
}
