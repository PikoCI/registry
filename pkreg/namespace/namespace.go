package namespace

import "time"

type Namespace struct {
	ID        string
	Name      string
	Type      string // official, verified, community
	GitHubID  string
	OwnerID   string
	Public    bool
	CreatedAt time.Time
}
