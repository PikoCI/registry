package orgmember

import "context"

//go:generate go tool mockgen -destination=../mock/orgmember_repository.go -mock_names=Repository=OrgMemberRepository -package mock github.com/pikoci/registry/pkreg/orgmember Repository

type Repository interface {
	Create(ctx context.Context, om OrgMember) error
	FindByOrgAndUser(ctx context.Context, orgID, userID string) (*OrgMember, error)
	FilterByOrg(ctx context.Context, orgID string) ([]*OrgMember, error)
	Delete(ctx context.Context, orgID, userID string) error
}
