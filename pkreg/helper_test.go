package pkreg_test

import (
	"log/slog"
	"os"

	"github.com/pikoci/registry/pkreg"
	"github.com/pikoci/registry/pkreg/mock"
	"go.uber.org/mock/gomock"
)

type testService struct {
	svc        pkreg.Service
	users      *mock.UserRepository
	namespaces *mock.NamespaceRepository
	regTypes   *mock.RegTypeRepository
	versions   *mock.VersionRepository
	tags       *mock.TagRepository
	tokens     *mock.TokenRepository
	orgMembers *mock.OrgMemberRepository
	downloads  *mock.DownloadRepository
}

func newTestService(ctrl *gomock.Controller) *testService {
	ts := &testService{
		users:      mock.NewUserRepository(ctrl),
		namespaces: mock.NewNamespaceRepository(ctrl),
		regTypes:   mock.NewRegTypeRepository(ctrl),
		versions:   mock.NewVersionRepository(ctrl),
		tags:       mock.NewTagRepository(ctrl),
		tokens:     mock.NewTokenRepository(ctrl),
		orgMembers: mock.NewOrgMemberRepository(ctrl),
		downloads:  mock.NewDownloadRepository(ctrl),
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	ts.svc = pkreg.New(
		ts.users,
		ts.namespaces,
		ts.regTypes,
		ts.versions,
		ts.tags,
		ts.tokens,
		ts.orgMembers,
		ts.downloads,
		[]byte("test-secret"),
		"test-client-id",
		"test-client-secret",
		logger,
	)

	return ts
}
