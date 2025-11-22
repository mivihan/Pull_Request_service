package domain

import (
    "fmt"
    "strings"
    "time"
)

type Team struct {
    TeamName  string
    CreatedAt time.Time
}

func (t *Team) Validate() error {
    if strings.TrimSpace(t.TeamName) == "" {
        return fmt.Errorf("team_name cannot be empty")
    }
    return nil
}