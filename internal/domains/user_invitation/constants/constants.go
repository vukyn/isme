package constants

import "time"

// Invitation status — expired is derived from expires_at, never stored
const (
	InvitationStatusPending  = 1
	InvitationStatusAccepted = 2
	InvitationStatusRevoked  = 3
)

// Display statuses exposed by the public invite-detail endpoint so the
// AcceptInvite page can distinguish the renderable states.
const (
	DisplayStatusValid    = "valid"
	DisplayStatusExpired  = "expired"
	DisplayStatusAccepted = "accepted"
	DisplayStatusRevoked  = "revoked"
)

// InvitationTTL is how long an invite link stays valid after creation
const InvitationTTL = 7 * 24 * time.Hour
