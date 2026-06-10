package models

import "testing"

func TestUpdateRequestValidate(t *testing.T) {
	cases := []struct {
		name    string
		req     UpdateRequest
		wantErr bool
	}{
		{"disabled ignores cron", UpdateRequest{Enabled: false, Cron: ""}, false},
		{"disabled ignores bad cron", UpdateRequest{Enabled: false, Cron: "nonsense"}, false},
		{"enabled with valid cron", UpdateRequest{Enabled: true, Cron: "0 3 * * *"}, false},
		{"enabled with empty cron", UpdateRequest{Enabled: true, Cron: ""}, true},
		{"enabled with bad cron", UpdateRequest{Enabled: true, Cron: "not a cron"}, true},
		{"enabled with too-many fields", UpdateRequest{Enabled: true, Cron: "0 3 * * * *"}, true},
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
