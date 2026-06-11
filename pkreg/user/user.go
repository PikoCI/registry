package user

import "time"

type User struct {
	ID        string
	GitHubID  string
	Username  string
	AvatarURL string
	CreatedAt time.Time
}
