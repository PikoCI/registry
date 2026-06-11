package version

import (
	"encoding/json"
	"time"
)

type Version struct {
	ID                string
	TypeID            string
	Version           string
	Content           string // raw HCL
	Readme            string
	Params            json.RawMessage
	Examples          json.RawMessage
	Yanked            bool
	Deprecated        bool
	DeprecatedMessage string
	SuccessorVersion  string
	Downloads         int
	CreatedAt         time.Time
	CreatedBy         string // user ID
}
