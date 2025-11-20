package domain

import (
	"fmt"
	"strings"
	"time"
)

type User struct {
	UserID    string
	Username  string
	TeamName  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.UserID) == "" {
		return fmt.Errorf("user_id cannot be empty")
	}
	if strings.TrimSpace(u.Username) == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if strings.TrimSpace(u.TeamName) == "" {
		return fmt.Errorf("team_name cannot be empty")
	}
	return nil
}

func (u *User) CanBeReviewer() bool {
	return u.IsActive
}
