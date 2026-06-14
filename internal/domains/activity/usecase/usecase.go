package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vukyn/isme/internal/domains/activity/constants"
	"github.com/vukyn/isme/internal/domains/activity/entity"
	"github.com/vukyn/isme/internal/domains/activity/models"
	activityRepo "github.com/vukyn/isme/internal/domains/activity/repository"

	"github.com/vukyn/kuery/log"
)

type usecase struct {
	activityRepo activityRepo.IRepository
}

func NewUsecase(
	activityRepo activityRepo.IRepository,
) IUseCase {
	return &usecase{
		activityRepo: activityRepo,
	}
}

// record is the shared best-effort emit path: it marshals the meta map, writes
// the event, and on any failure logs and returns — it NEVER propagates an error,
// so a recorder failure can never fail the action being audited.
func (u *usecase) record(ctx context.Context, userID, eventType string, meta map[string]any) {
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		log.New().Errorf("activity: failed to marshal meta for %s: %v", eventType, err)
		return
	}
	if err := u.activityRepo.Create(ctx, entity.ActivityEvent{
		UserID: userID,
		Type:   eventType,
		Meta:   string(metaJSON),
	}); err != nil {
		log.New().Errorf("activity: failed to record %s for user %s: %v", eventType, userID, err)
		return
	}
}

func (u *usecase) RecordSignIn(ctx context.Context, userID, device, clientIP string) {
	u.record(ctx, userID, constants.ActivityTypeSignIn, map[string]any{
		"device":    device,
		"client_ip": clientIP,
	})
}

func (u *usecase) RecordSignOut(ctx context.Context, userID string) {
	u.record(ctx, userID, constants.ActivityTypeSignOut, map[string]any{})
}

func (u *usecase) RecordPasswordChanged(ctx context.Context, userID string) {
	u.record(ctx, userID, constants.ActivityTypePasswordChanged, map[string]any{})
}

func (u *usecase) RecordProfileUpdated(ctx context.Context, userID string) {
	u.record(ctx, userID, constants.ActivityTypeProfileUpdated, map[string]any{})
}

func (u *usecase) RecordInvitationSent(ctx context.Context, inviterID, email string, roleNames []string) {
	if roleNames == nil {
		roleNames = []string{}
	}
	u.record(ctx, inviterID, constants.ActivityTypeInvitationSent, map[string]any{
		"email": email,
		"roles": roleNames,
	})
}

func (u *usecase) List(ctx context.Context, userID string, limit int) ([]models.ActivityItem, error) {
	events, err := u.activityRepo.ListByUserID(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	items := make([]models.ActivityItem, 0, len(events))
	for _, event := range events {
		meta := map[string]any{}
		if event.Meta != "" {
			// best-effort: a malformed meta blob yields an empty map rather than
			// failing the whole feed.
			_ = json.Unmarshal([]byte(event.Meta), &meta)
		}
		createdAt := ""
		if !event.CreatedAt.IsZero() {
			createdAt = event.CreatedAt.Format(time.RFC3339)
		}
		items = append(items, models.ActivityItem{
			ID:        event.ID,
			Type:      event.Type,
			Meta:      meta,
			CreatedAt: createdAt,
		})
	}
	return items, nil
}
