package constants

import "time"

// Invitation status. Expired is stored once an invitation is swept by the
// cleanup job or superseded by a new invite; a pending row whose expires_at
// has passed is still displayed as expired before that happens.
const (
	InvitationStatusPending  = 1
	InvitationStatusAccepted = 2
	InvitationStatusRevoked  = 3
	InvitationStatusExpired  = 4
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
