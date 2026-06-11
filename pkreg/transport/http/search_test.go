package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pikoci/registry/pkreg/regtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSearchTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	th := newTestHandler(ctrl)

	th.svc.EXPECT().SearchTypes(gomock.Any(), "postgres", "", "", "", 1, 20).Return(&regtype.SearchResult{
		Types: []*regtype.RegType{
			{ID: "rt-1", Name: "postgres", Kind: "resource_type"},
		},
		TotalCount: 1,
	}, nil)

	rr := doRequest(th.handler, http.MethodGet, "/api/plugins?q=postgres")
	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
}

func TestSearchTypes_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	th := newTestHandler(ctrl)

	th.svc.EXPECT().SearchTypes(gomock.Any(), "", "", "", "", 1, 20).Return(&regtype.SearchResult{
		Types:      nil,
		TotalCount: 0,
	}, nil)

	rr := doRequest(th.handler, http.MethodGet, "/api/plugins")
	assert.Equal(t, http.StatusOK, rr.Code)
}
