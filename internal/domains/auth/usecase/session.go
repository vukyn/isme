package usecase

import (
	"context"
	"sort"
	"time"

	appServiceEntity "github.com/vukyn/isme/internal/domains/app_service/entity"
	"github.com/vukyn/isme/internal/domains/auth/models"
	userSessionEntity "github.com/vukyn/isme/internal/domains/user_session/entity"
	pkgCtx "github.com/vukyn/kuery/ctx"
	pkgErr "github.com/vukyn/kuery/http/errors"
)

// ListMySessions returns the active sessions for the caller. The session whose
// token ID matches the request's token ID is flagged as Current. RefreshToken
// and TokenID are never exposed.
func (u *usecase) ListMySessions(ctx context.Context) ([]models.MySessionItem, error) {
	userID := pkgCtx.GetUserID(ctx)
	if userID == "" {
		return nil, pkgErr.InvalidRequest("user not found")
	}

	sessions, err := u.userSessionRepo.GetListActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Enrich SSO sessions with their owning app's name/icon/color via a single
	// batched lookup. First-party isme sessions (empty AppServiceID) are skipped
	// entirely — the frontend renders the isme chip for them — and if every
	// session is first-party the lookup is never issued (no wasted query).
	appServices, err := u.appServicesForSessions(ctx, sessions)
	if err != nil {
		return nil, err
	}

	currentTokenID := pkgCtx.GetTokenID(ctx)
	items := make([]models.MySessionItem, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, mapMySessionItem(session, currentTokenID, appServices))
	}
	// Surface the caller's current session at the top; preserve the existing
	// order for the rest.
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].Current && !items[j].Current
	})
	return items, nil
}

// appServicesForSessions collects the DISTINCT non-empty AppServiceIDs across the
// sessions and resolves them in ONE batched lookup, keyed by id. It returns a nil
// map (and no error) when no session carries an app, so callers never issue a
// query for an all-first-party session set.
func (u *usecase) appServicesForSessions(ctx context.Context, sessions []userSessionEntity.UserSession) (map[string]appServiceEntity.AppService, error) {
	idSet := make(map[string]struct{}, len(sessions))
	for _, session := range sessions {
		if session.AppServiceID != "" {
			idSet[session.AppServiceID] = struct{}{}
		}
	}
	if len(idSet) == 0 {
		return nil, nil
	}

	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	return u.appServiceRepo.GetByIDs(ctx, ids)
}

// CountMySessions returns the total active session count for the caller plus
// how many of those were created in the last 24 hours.
func (u *usecase) CountMySessions(ctx context.Context) (models.MySessionCount, error) {
	userID := pkgCtx.GetUserID(ctx)
	if userID == "" {
		return models.MySessionCount{}, pkgErr.InvalidRequest("user not found")
	}

	counts, err := u.userSessionRepo.CountActiveByUserIDs(ctx, []string{userID})
	if err != nil {
		return models.MySessionCount{}, err
	}

	since := time.Now().UTC().Add(-24 * time.Hour)
	newIn24h, err := u.userSessionRepo.CountActiveByUserIDCreatedAfter(ctx, userID, since)
	if err != nil {
		return models.MySessionCount{}, err
	}

	// Accurate sliding-24h rotation count from the event log (never a stored counter).
	rotations24h, err := u.userSessionRepo.CountRotationsByUserIDSince(ctx, userID, since)
	if err != nil {
		return models.MySessionCount{}, err
	}

	// Most-recent rotation across the user's active sessions drives the card's
	// "last refreshed {when}" delta. Derived from the session list so no extra query.
	sessions, err := u.userSessionRepo.GetListActiveByUserID(ctx, userID)
	if err != nil {
		return models.MySessionCount{}, err
	}
	lastRefreshedAt := ""
	var latest time.Time
	for _, session := range sessions {
		if session.LastRefreshedAt != nil && session.LastRefreshedAt.After(latest) {
			latest = *session.LastRefreshedAt
		}
	}
	if !latest.IsZero() {
		lastRefreshedAt = latest.Format(time.RFC3339)
	}

	return models.MySessionCount{
		Count:           counts[userID],
		NewIn24h:        newIn24h,
		Rotations24h:    rotations24h,
		LastRefreshedAt: lastRefreshedAt,
	}, nil
}

// RevokeMySession revokes a single session owned by the caller. It refuses to
// reveal or revoke a session that does not belong to the caller (returns
// not-found), and refuses to revoke the current session (use logout instead).
func (u *usecase) RevokeMySession(ctx context.Context, sessionID string) error {
	userID := pkgCtx.GetUserID(ctx)
	if userID == "" {
		return pkgErr.InvalidRequest("user not found")
	}
	if sessionID == "" {
		return pkgErr.InvalidRequest("session_id is required")
	}

	session, err := u.userSessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}
	// GetByID returns a zero-value entity on not-found; treat a missing row or
	// a row owned by a different user identically so we never leak another
	// user's session existence.
	if session.ID == "" || session.UserID != userID {
		return pkgErr.NotFound("session not found")
	}

	// The current session is ended via logout, not the session manager.
	if session.TokenID == pkgCtx.GetTokenID(ctx) {
		return pkgErr.InvalidRequest("use logout to end the current session")
	}

	return u.userSessionRepo.InactiveSessionByID(ctx, sessionID)
}

// RevokeMyOtherSessions revokes every active session of the caller except the
// current one.
func (u *usecase) RevokeMyOtherSessions(ctx context.Context) error {
	userID := pkgCtx.GetUserID(ctx)
	if userID == "" {
		return pkgErr.InvalidRequest("user not found")
	}

	currentTokenID := pkgCtx.GetTokenID(ctx)
	if currentTokenID == "" {
		return pkgErr.InvalidRequest("token not found")
	}

	return u.userSessionRepo.InactiveAllUserSessionExcept(ctx, userID, currentTokenID)
}

// mapMySessionItem maps a session entity to the API DTO, guarding zero times to
// empty strings so the JSON never carries a Go zero-time literal. The app_* fields
// are filled from appServices when the session carries a known AppServiceID; for
// first-party isme sessions (empty id) or a deleted app_service (id absent from
// the map) they are left empty and the frontend decides what to render.
func mapMySessionItem(session userSessionEntity.UserSession, currentTokenID string, appServices map[string]appServiceEntity.AppService) models.MySessionItem {
	lastLoginAt := ""
	if !session.LastLoginAt.IsZero() {
		lastLoginAt = session.LastLoginAt.Format(time.RFC3339)
	}
	expiresAt := ""
	if !session.ExpiresAt.IsZero() {
		expiresAt = session.ExpiresAt.Format(time.RFC3339)
	}
	lastRefreshedAt := ""
	if session.LastRefreshedAt != nil && !session.LastRefreshedAt.IsZero() {
		lastRefreshedAt = session.LastRefreshedAt.Format(time.RFC3339)
	}

	var appName, appIcon, appColor string
	if session.AppServiceID != "" {
		if appService, ok := appServices[session.AppServiceID]; ok {
			appName = appService.AppName
			appIcon = appService.Icon
			appColor = appService.Color
		}
	}

	return models.MySessionItem{
		ID:           session.ID,
		ClientIP:     session.ClientIP,
		UserAgent:    session.UserAgent,
		LastLoginAt:  lastLoginAt,
		ExpiresAt:    expiresAt,
		AppServiceID: session.AppServiceID,
		AppName:         appName,
		AppIcon:         appIcon,
		AppColor:        appColor,
		Current:         session.TokenID == currentTokenID,
		RefreshCount:    session.RefreshCount,
		LastRefreshedAt: lastRefreshedAt,
	}
}
