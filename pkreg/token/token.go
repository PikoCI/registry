package token

import "time"

type Token struct {
	ID          string
	NamespaceID string
	Name        string
	TokenHash   string
	Prefix      string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}
