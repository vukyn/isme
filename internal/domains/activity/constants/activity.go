package constants

// Activity event types — the v1 taxonomy. The frontend maps each type to an
// icon/tone/copy, so these strings are part of the API contract.
const (
	ActivityTypeSignIn          = "sign_in"
	ActivityTypeSignOut         = "sign_out"
	ActivityTypePasswordChanged = "password_changed"
	ActivityTypeInvitationSent  = "invitation_sent"
)

// Limits for the "Recent activity" feed.
const (
	DefaultActivityLimit = 8
	MaxActivityLimit     = 50
)
