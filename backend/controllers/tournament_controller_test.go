package controllers

import (
	"encoding/json"
	"testing"
	"time"

	"arguehub/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestToTournamentResponse_HostCanSeeManagementFields(t *testing.T) {
	hostID := primitive.NewObjectID()

	tournament := models.Tournament{
		ID:         primitive.NewObjectID(),
		HostID:     hostID,
		Topic:      "AI in Education",
		Visibility: models.VisibilityPrivate,
		InviteCode: "SECRET123",
		JoinRequests: []models.JoinRequest{
			{
				UserID: primitive.NewObjectID(),
				Name:   "Test User",
				Status: models.JoinRequestPending,
			},
		},
		CreatedAt: time.Now(),
	}

	response := toTournamentResponse(tournament, true)

	if response.InviteCode != tournament.InviteCode {
		t.Errorf("host should receive invite code")
	}

	if len(response.JoinRequests) != 1 {
		t.Errorf("host should receive join requests; got %d", len(response.JoinRequests))
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if _, exists := payload["inviteCode"]; !exists {
		t.Error("host response should contain inviteCode")
	}

	if _, exists := payload["joinRequests"]; !exists {
		t.Error("host response should contain joinRequests")
	}
}

func TestToTournamentResponse_NonHostCannotSeeManagementFields(t *testing.T) {
	tournament := models.Tournament{
		ID:         primitive.NewObjectID(),
		HostID:     primitive.NewObjectID(),
		Topic:      "AI in Education",
		Visibility: models.VisibilityPublic,
		InviteCode: "SECRET123",
		JoinRequests: []models.JoinRequest{
			{
				UserID: primitive.NewObjectID(),
				Name:   "Test User",
				Status: models.JoinRequestPending,
			},
		},
		CreatedAt: time.Now(),
	}

	response := toTournamentResponse(tournament, false)

	if response.InviteCode != "" {
		t.Error("non-host response must not contain an invite code")
	}

	if len(response.JoinRequests) != 0 {
		t.Error("non-host response must not contain join requests")
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if _, exists := payload["inviteCode"]; exists {
		t.Error("non-host JSON must omit inviteCode")
	}

	if _, exists := payload["joinRequests"]; exists {
		t.Error("non-host JSON must omit joinRequests")
	}

	if payload["topic"] != tournament.Topic {
		t.Error("non-host response should preserve public tournament fields")
	}
}

func TestCanJoinTournament(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "upcoming tournament allows joining",
			status:   models.StatusUpcoming,
			expected: true,
		},
		{
			name:     "live tournament rejects joining",
			status:   models.StatusLive,
			expected: false,
		},
		{
			name:     "completed tournament rejects joining",
			status:   models.StatusCompleted,
			expected: false,
		},
		{
			name:     "unknown or empty status rejects joining",
			status:   "",
			expected: false,
		},
		{
			name:     "arbitrary invalid status rejects joining",
			status:   "cancelled",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canJoinTournament(tt.status); got != tt.expected {
				t.Errorf("canJoinTournament(%q) = %v, want %v", tt.status, got, tt.expected)
			}
		})
	}
}
