package orgmember

import "time"

type OrgMember struct {
	OrgID     string
	UserID    string
	Role      string // admin, member
	CreatedAt time.Time
}
