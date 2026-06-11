package pkreg_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pikoci/registry/pkreg/namespace"
	"github.com/pikoci/registry/pkreg/orgmember"
	"github.com/pikoci/registry/pkreg/regtype"
	"github.com/pikoci/registry/pkreg/tag"
	"github.com/pikoci/registry/pkreg/token"
	"github.com/pikoci/registry/pkreg/user"
	"github.com/pikoci/registry/pkreg/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetCurrentUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	u := &user.User{ID: "user-1", Username: "testuser"}
	ts.users.EXPECT().FindByID(gomock.Any(), "user-1").Return(u, nil)

	result, err := ts.svc.GetCurrentUser(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, "testuser", result.Username)
}

func TestGetNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", Type: "community"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)

	result, err := ts.svc.GetNamespace(context.Background(), "testns")
	require.NoError(t, err)
	assert.Equal(t, "testns", result.Name)
}

func TestGetNamespace_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "missing").Return(nil, fmt.Errorf("not found"))

	_, err := ts.svc.GetNamespace(context.Background(), "missing")
	assert.Error(t, err)
}

func TestListNamespaceTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FilterByNamespace(gomock.Any(), "ns-1").Return([]*regtype.RegType{
		{ID: "rt-1", Name: "postgres", Kind: "resource_type"},
	}, nil)

	result, err := ts.svc.ListNamespaceTypes(context.Background(), "testns")
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestGetType(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns"}
	rt := &regtype.RegType{ID: "rt-1", Name: "postgres", Kind: "resource_type"}
	versions := []*version.Version{{ID: "v-1", Version: "v1.0.0"}}
	tags := []string{"database", "postgresql"}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "").Return(rt, nil)
	ts.versions.EXPECT().FilterByRegType(gomock.Any(), "rt-1").Return(versions, nil)
	ts.tags.EXPECT().FilterByRegType(gomock.Any(), "rt-1").Return(tags, nil)

	gotRT, gotVersions, gotTags, err := ts.svc.GetType(context.Background(), "testns", "postgres")
	require.NoError(t, err)
	assert.Equal(t, "postgres", gotRT.Name)
	assert.Len(t, gotVersions, 1)
	assert.Equal(t, []string{"database", "postgresql"}, gotTags)
}

func TestSearchTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ts.regTypes.EXPECT().Search(gomock.Any(), regtype.SearchParams{
		Query:   "postgres",
		Kind:    "resource_type",
		Page:    1,
		PerPage: 20,
	}).Return(&regtype.SearchResult{
		Types:      []*regtype.RegType{{ID: "rt-1", Name: "postgres"}},
		TotalCount: 1,
	}, nil)
	ts.tags.EXPECT().ListWithCountsFiltered(gomock.Any(), "resource_type", "postgres", 10).Return(nil, nil)

	result, err := ts.svc.SearchTypes(context.Background(), "postgres", "resource_type", "", "", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalCount)
}

func TestSearchTypes_WithFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ts.regTypes.EXPECT().Search(gomock.Any(), regtype.SearchParams{
		Query:     "test",
		Kind:      "runner_type",
		Tag:       "docker",
		Namespace: "myns",
		Page:      2,
		PerPage:   10,
	}).Return(&regtype.SearchResult{Types: nil, TotalCount: 0}, nil)
	ts.tags.EXPECT().ListWithCountsFiltered(gomock.Any(), "runner_type", "test", 10).Return(nil, nil)

	result, err := ts.svc.SearchTypes(context.Background(), "test", "runner_type", "docker", "myns", 2, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, result.TotalCount)
}

func TestListMyTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1"}
	ts.namespaces.EXPECT().FindByOwnerID(gomock.Any(), "user-1").Return([]*namespace.Namespace{ns}, nil)
	ts.regTypes.EXPECT().FilterByNamespace(gomock.Any(), "ns-1").Return([]*regtype.RegType{
		{ID: "rt-1", Name: "mytype"},
	}, nil)

	result, err := ts.svc.ListMyTypes(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestUpdateTypeTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	rt := &regtype.RegType{ID: "rt-1", Name: "postgres"}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "").Return(rt, nil)
	ts.tags.EXPECT().SetForRegType(gomock.Any(), "rt-1", []string{"db", "sql"}).Return(nil)

	err := ts.svc.UpdateTypeTags(context.Background(), "user-1", "testns", "postgres", []string{"db", "sql"})
	require.NoError(t, err)
}

func TestUpdateTypeTags_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-2"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.orgMembers.EXPECT().FindByOrgAndUser(gomock.Any(), "ns-1", "user-1").Return(nil, fmt.Errorf("not found"))

	err := ts.svc.UpdateTypeTags(context.Background(), "user-1", "testns", "postgres", []string{"db"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestPublishVersion_NewType(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	manifest := []byte(`name = "postgres"
version = "1.0.0"
description = "PostgreSQL resource type"

resource_type "postgres" {
  check {}
}
`)
	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "resource_type").Return(nil, fmt.Errorf("not found"))
	ts.regTypes.EXPECT().Create(gomock.Any(), gomock.Any()).Return("rt-1", nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), gomock.Any(), "1.0.0").Return(nil, fmt.Errorf("not found"))
	ts.versions.EXPECT().Create(gomock.Any(), gomock.Any()).Return("v-1", nil)
	ts.tags.EXPECT().SetForRegType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	v, err := ts.svc.PublishVersion(context.Background(), "user-1", "testns", "", "", manifest, []byte("# README"))
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", v.Version)
}

func TestPublishVersion_ExistingType(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	manifest := []byte(`name = "postgres"
version = "2.0.0"
description = "PostgreSQL resource type v2"

resource_type "postgres" {
  check {}
}
`)
	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	rt := &regtype.RegType{ID: "rt-1", NamespaceID: "ns-1", Name: "postgres", Kind: "resource_type"}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "resource_type").Return(rt, nil)
	ts.regTypes.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), "rt-1", "2.0.0").Return(nil, fmt.Errorf("not found"))
	ts.versions.EXPECT().Create(gomock.Any(), gomock.Any()).Return("v-2", nil)
	ts.tags.EXPECT().SetForRegType(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	v, err := ts.svc.PublishVersion(context.Background(), "user-1", "testns", "", "", manifest, nil)
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", v.Version)
}

func TestPublishVersion_InvalidSemver(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	manifest := []byte(`name = "postgres"
version = "not-a-version"
description = "bad"

resource_type "postgres" {
  check {}
}
`)
	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)

	_, err := ts.svc.PublishVersion(context.Background(), "user-1", "testns", "", "", manifest, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid semver")
}

func TestPublishVersion_DuplicateVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	manifest := []byte(`name = "postgres"
version = "1.0.0"
description = "dup"

resource_type "postgres" {
  check {}
}
`)
	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	rt := &regtype.RegType{ID: "rt-1", NamespaceID: "ns-1", Name: "postgres", Kind: "resource_type"}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "resource_type").Return(rt, nil)
	ts.regTypes.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), "rt-1", "1.0.0").Return(&version.Version{ID: "v-1", Version: "1.0.0"}, nil)

	_, err := ts.svc.PublishVersion(context.Background(), "user-1", "testns", "", "", manifest, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestPublishVersion_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	manifest := []byte(`name = "postgres"
version = "1.0.0"

resource_type "postgres" {
  check {}
}
`)
	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-2"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.orgMembers.EXPECT().FindByOrgAndUser(gomock.Any(), "ns-1", "user-1").Return(nil, fmt.Errorf("not found"))

	_, err := ts.svc.PublishVersion(context.Background(), "user-1", "testns", "", "", manifest, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestGetVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns"}
	rt := &regtype.RegType{ID: "rt-1", Name: "postgres"}
	v := &version.Version{ID: "v-1", Version: "v1.0.0", Content: "resource_type {}"}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "").Return(rt, nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), "rt-1", "v1.0.0").Return(v, nil)

	result, err := ts.svc.GetVersion(context.Background(), "testns", "postgres", "v1.0.0")
	require.NoError(t, err)
	assert.Equal(t, "v1.0.0", result.Version)
}

func TestFetchVersionContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns"}
	rt := &regtype.RegType{ID: "rt-1", Name: "postgres"}
	v := &version.Version{ID: "v-1", Version: "v1.0.0", Downloads: 5}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "").Return(rt, nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), "rt-1", "v1.0.0").Return(v, nil)
	ts.downloads.EXPECT().RecordDownload(gomock.Any(), "v-1", "tok-hash", gomock.Any(), gomock.Any()).Return(true, nil)
	ts.versions.EXPECT().IncrementDownloads(gomock.Any(), "v-1").Return(nil)

	result, err := ts.svc.FetchVersionContent(context.Background(), "testns", "postgres", "v1.0.0", "tok-hash", "1.2.3.4")
	require.NoError(t, err)
	assert.Equal(t, 6, result.Downloads)
}

func TestFetchVersionContent_DeduplicatedDownload(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns"}
	rt := &regtype.RegType{ID: "rt-1", Name: "postgres"}
	v := &version.Version{ID: "v-1", Version: "v1.0.0", Downloads: 5}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "").Return(rt, nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), "rt-1", "v1.0.0").Return(v, nil)
	ts.downloads.EXPECT().RecordDownload(gomock.Any(), "v-1", "tok-hash", gomock.Any(), gomock.Any()).Return(false, nil)

	result, err := ts.svc.FetchVersionContent(context.Background(), "testns", "postgres", "v1.0.0", "tok-hash", "1.2.3.4")
	require.NoError(t, err)
	assert.Equal(t, 5, result.Downloads) // Not incremented
}

func TestYankVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	rt := &regtype.RegType{ID: "rt-1", Name: "postgres"}
	v := &version.Version{ID: "v-1", Version: "v1.0.0"}

	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.regTypes.EXPECT().FindByNamespaceAndName(gomock.Any(), "ns-1", "postgres", "").Return(rt, nil)
	ts.versions.EXPECT().FindByRegTypeAndNumber(gomock.Any(), "rt-1", "v1.0.0").Return(v, nil)
	ts.versions.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, v version.Version) error {
		assert.True(t, v.Yanked)
		return nil
	})

	err := ts.svc.YankVersion(context.Background(), "user-1", "testns", "postgres", "v1.0.0")
	require.NoError(t, err)
}

func TestYankVersion_NotOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-2"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.orgMembers.EXPECT().FindByOrgAndUser(gomock.Any(), "ns-1", "user-1").Return(nil, fmt.Errorf("not found"))

	err := ts.svc.YankVersion(context.Background(), "user-1", "testns", "postgres", "v1.0.0")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied")
}

func TestListTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ts.tags.EXPECT().ListWithCounts(gomock.Any()).Return([]*tag.TagCount{
		{Tag: "database", Count: 5},
		{Tag: "docker", Count: 3},
	}, nil)

	result, err := ts.svc.ListTags(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListTypesByTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ts.tags.EXPECT().FilterRegTypesByTag(gomock.Any(), "database").Return([]string{"rt-1", "rt-2"}, nil)
	ts.regTypes.EXPECT().FindByID(gomock.Any(), "rt-1").Return(&regtype.RegType{ID: "rt-1", NamespaceID: "ns-1", Name: "postgres"}, nil)
	ts.namespaces.EXPECT().FindByID(gomock.Any(), "ns-1").Return(&namespace.Namespace{ID: "ns-1", Name: "testns"}, nil)
	ts.regTypes.EXPECT().FindByID(gomock.Any(), "rt-2").Return(&regtype.RegType{ID: "rt-2", NamespaceID: "ns-1", Name: "mysql"}, nil)
	ts.namespaces.EXPECT().FindByID(gomock.Any(), "ns-1").Return(&namespace.Namespace{ID: "ns-1", Name: "testns"}, nil)

	result, err := ts.svc.ListTypesByTag(context.Background(), "database")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestCreateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns", OwnerID: "user-1"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.tokens.EXPECT().Create(gomock.Any(), gomock.Any()).Return("tok-1", nil)

	tok, rawValue, err := ts.svc.CreateToken(context.Background(), "user-1", "testns", "my-token")
	require.NoError(t, err)
	assert.NotEmpty(t, rawValue)
	assert.Equal(t, "my-token", tok.Name)
	assert.NotEmpty(t, tok.TokenHash)
}

func TestListTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "testns"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "testns").Return(ns, nil)
	ts.tokens.EXPECT().FilterByNamespace(gomock.Any(), "ns-1").Return([]*token.Token{
		{ID: "tok-1", Name: "my-token"},
	}, nil)

	result, err := ts.svc.ListTokens(context.Background(), "testns")
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestRevokeToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ts.tokens.EXPECT().Delete(gomock.Any(), "tok-1").Return(nil)

	err := ts.svc.RevokeToken(context.Background(), "user-1", "tok-1")
	require.NoError(t, err)
}

func TestClaimOrgNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	u := &user.User{ID: "user-1", Username: "testuser"}
	ts.users.EXPECT().FindByID(gomock.Any(), "user-1").Return(u, nil)
	ts.namespaces.EXPECT().Create(gomock.Any(), gomock.Any()).Return("ns-1", nil)
	ts.orgMembers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	ns, err := ts.svc.ClaimOrgNamespace(context.Background(), "user-1", "myorg")
	require.NoError(t, err)
	assert.Equal(t, "myorg", ns.Name)
	assert.Equal(t, "community", ns.Type)
}

func TestInviteOrgMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	ts := newTestService(ctrl)

	ns := &namespace.Namespace{ID: "ns-1", Name: "myorg"}
	ts.namespaces.EXPECT().FindByName(gomock.Any(), "myorg").Return(ns, nil)
	ts.orgMembers.EXPECT().FindByOrgAndUser(gomock.Any(), "ns-1", "user-1").Return(&orgmember.OrgMember{Role: "admin"}, nil)
	ts.users.EXPECT().FindByID(gomock.Any(), "invitee-1").Return(&user.User{ID: "invitee-1"}, nil)
	ts.orgMembers.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

	err := ts.svc.InviteOrgMember(context.Background(), "user-1", "myorg", "invitee-1", "member")
	require.NoError(t, err)
}

// Suppress unused import warnings
var _ = time.Now
