package regtype

import "time"

type RegType struct {
	ID            string
	NamespaceID   string
	NamespaceName string
	Name          string
	Kind        string // resource_type, runner_type, service_type, secret_type, notification_type
	Description string
	Repository  string
	Homepage    string
	License     string
	Public      bool
	CreatedAt   time.Time
}
