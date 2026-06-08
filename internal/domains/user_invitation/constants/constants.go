package constants

import "time"

// Invitation status — expired is derived from expires_at, never stored
const (
	InvitationStatusPending  = 1
	InvitationStatusAccepted = 2
	InvitationStatusRevoked  = 3
)

// InvitationTTL is how long an invite link stays valid after creation
const InvitationTTL = 7 * 24 * time.Hour
