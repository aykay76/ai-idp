package teams

import (
	"context"

	"github.com/google/uuid"
)

// TeamService defines the interface for team operations
type TeamService interface {
	CreateTeam(ctx context.Context, team Team) (Team, error)
	GetTeam(ctx context.Context, teamID uuid.UUID) (Team, error)
	ListTeams(ctx context.Context, limit, offset int) ([]Team, int, error)
	UpdateTeam(ctx context.Context, team Team) (Team, error)
	DeleteTeam(ctx context.Context, teamID uuid.UUID) error
}