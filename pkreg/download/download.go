package download

import "time"

type Download struct {
	VersionID string
	TokenHash string
	IPHash    string
	Date      time.Time
}
