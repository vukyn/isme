package models

// ActivityItem is a single recent-activity entry returned to the client. The
// server emits a STRUCTURED record (type + meta) and the frontend composes the
// display copy/icon/tone from the type — no server-composed body.
type ActivityItem struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Meta      map[string]any `json:"meta"`
	CreatedAt string         `json:"created_at"`
}
