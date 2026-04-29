package domain

// Project is a tenant-like principal that owns events and tokens.
type Project struct {
	ID    string
	Token ProjectToken
}
