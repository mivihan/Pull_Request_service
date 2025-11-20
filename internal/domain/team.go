package domain

import (
	"fmt"
	"strings"
	"time"
)

type Team struct {
	TeamName  string
	Members   []*User
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Team) Validate() error {
	if strings.TrimSpace(t.TeamName) == "" {
		return fmt.Errorf("team_name cannot be empty")
	}
	if t.Members == nil {
		return fmt.Errorf("members cannot be nil")
	}
	return nil
}

func (t *Team) GetActiveMembersExcluding(excludeUserIDs ...string) []*User {
	excludeMap := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeMap[id] = true
	}

	activeMembers := make([]*User, 0)
	for _, member := range t.Members {
		if member.IsActive && !excludeMap[member.UserID] {
			activeMembers = append(activeMembers, member)
		}
	}
	return activeMembers
}
