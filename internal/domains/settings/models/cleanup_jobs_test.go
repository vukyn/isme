package models

import "testing"

func TestInvitationCleanupUpdateRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     InvitationCleanupUpdateRequest
		wantErr bool
	}{
		{"disabled ignores cron", InvitationCleanupUpdateRequest{Enabled: false, Cron: ""}, false},
		{"enabled with valid cron", InvitationCleanupUpdateRequest{Enabled: true, Cron: "0 6 * * *"}, false},
		{"enabled with empty cron", InvitationCleanupUpdateRequest{Enabled: true, Cron: ""}, true},
		{"enabled with bad cron", InvitationCleanupUpdateRequest{Enabled: true, Cron: "not a cron"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestUnverifiedUserCleanupUpdateRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     UnverifiedUserCleanupUpdateRequest
		wantErr bool
	}{
		{"floor accepted", UnverifiedUserCleanupUpdateRequest{MinAgeDays: 1}, false},
		{"below floor rejected", UnverifiedUserCleanupUpdateRequest{MinAgeDays: 0}, true},
		{"negative rejected", UnverifiedUserCleanupUpdateRequest{MinAgeDays: -5}, true},
		{"enabled with valid cron", UnverifiedUserCleanupUpdateRequest{Enabled: true, Cron: "0 7 * * *", MinAgeDays: 30}, false},
		{"enabled with empty cron", UnverifiedUserCleanupUpdateRequest{Enabled: true, Cron: "", MinAgeDays: 30}, true},
		{"enabled with bad cron", UnverifiedUserCleanupUpdateRequest{Enabled: true, Cron: "0 7 * *", MinAgeDays: 30}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
