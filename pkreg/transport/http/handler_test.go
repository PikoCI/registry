package http_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/pikoci/registry/pkreg/mock"
	tshttp "github.com/pikoci/registry/pkreg/transport/http"
	"go.uber.org/mock/gomock"
)

type testHandler struct {
	handler http.Handler
	svc     *mock.Service
}

func newTestHandler(ctrl *gomock.Controller) *testHandler {
	svc := mock.NewService(ctrl)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	handler := tshttp.Handler(svc, []byte("test-secret"), logger)
	return &testHandler{handler: handler, svc: svc}
}

func doRequest(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}
