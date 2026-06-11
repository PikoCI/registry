package http_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/pikoci/registry/pkreg/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestListTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	th := newTestHandler(ctrl)

	th.svc.EXPECT().ListTags(gomock.Any()).Return([]*tag.TagCount{
		{Tag: "database", Count: 5},
		{Tag: "docker", Count: 3},
	}, nil)

	rr := doRequest(th.handler, http.MethodGet, "/api/tags")
	assert.Equal(t, http.StatusOK, rr.Code)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))
	assert.Len(t, result, 2)
}
