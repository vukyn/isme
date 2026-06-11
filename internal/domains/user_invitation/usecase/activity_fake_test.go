package usecase

import (
	"context"

	activityModels "github.com/vukyn/isme/internal/domains/activity/models"
)

// fakeActivityUsecase is a test double for the activity recorder. It records each
// RecordInvitationSent call so tests can assert the right event was emitted, and
// can be made to "fail" (recordErr) to prove the invite still succeeds when the
// recorder errors (the recorder is best-effort and swallows its own errors).
type fakeActivityUsecase struct {
	recordErr bool

	invitationSentCalls []fakeInvitationCall
}

type fakeInvitationCall struct {
	inviterID string
	email     string
	roleNames []string
}

func (f *fakeActivityUsecase) RecordSignIn(ctx context.Context, userID, device, clientIP string) {
}

func (f *fakeActivityUsecase) RecordSignOut(ctx context.Context, userID string) {}

func (f *fakeActivityUsecase) RecordPasswordChanged(ctx context.Context, userID string) {}

func (f *fakeActivityUsecase) RecordInvitationSent(ctx context.Context, inviterID, email string, roleNames []string) {
	f.invitationSentCalls = append(f.invitationSentCalls, fakeInvitationCall{inviterID: inviterID, email: email, roleNames: roleNames})
}

func (f *fakeActivityUsecase) List(ctx context.Context, userID string, limit int) ([]activityModels.ActivityItem, error) {
	return nil, nil
}
