package usecase

import (
	"context"

	activityModels "github.com/vukyn/isme/internal/domains/activity/models"
)

// fakeActivityUsecase is a test double for the activity recorder. It records each
// Record* call so tests can assert the right event was emitted, and can be made
// to "fail" (recordErr) to prove the audited action still succeeds when the
// recorder errors (the recorder is best-effort and swallows its own errors, so
// recordErr only exercises that the methods return nothing regardless).
type fakeActivityUsecase struct {
	recordErr bool

	signInCalls          []fakeSignInCall
	signOutCalls         []string
	passwordChangedCalls []string
	profileUpdatedCalls  []string
	invitationSentCalls  []fakeInvitationCall

	listItems []activityModels.ActivityItem
	listErr   error
}

type fakeSignInCall struct {
	userID   string
	device   string
	clientIP string
}

type fakeInvitationCall struct {
	inviterID string
	email     string
	roleNames []string
}

func (f *fakeActivityUsecase) RecordSignIn(ctx context.Context, userID, device, clientIP string) {
	f.signInCalls = append(f.signInCalls, fakeSignInCall{userID: userID, device: device, clientIP: clientIP})
}

func (f *fakeActivityUsecase) RecordSignOut(ctx context.Context, userID string) {
	f.signOutCalls = append(f.signOutCalls, userID)
}

func (f *fakeActivityUsecase) RecordPasswordChanged(ctx context.Context, userID string) {
	f.passwordChangedCalls = append(f.passwordChangedCalls, userID)
}

func (f *fakeActivityUsecase) RecordProfileUpdated(ctx context.Context, userID string) {
	f.profileUpdatedCalls = append(f.profileUpdatedCalls, userID)
}

func (f *fakeActivityUsecase) RecordInvitationSent(ctx context.Context, inviterID, email string, roleNames []string) {
	f.invitationSentCalls = append(f.invitationSentCalls, fakeInvitationCall{inviterID: inviterID, email: email, roleNames: roleNames})
}

func (f *fakeActivityUsecase) List(ctx context.Context, userID string, limit int) ([]activityModels.ActivityItem, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listItems, nil
}
