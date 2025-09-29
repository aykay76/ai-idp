package teams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aykay76/ai-idp/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Common errors
var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrInvalidTeamData   = errors.New("invalid team data")
)

// Service provides team management operations
type Service struct {
	db *database.Pool
}

// NewService creates a new team service
func NewService(db *database.Pool) *Service {
	return &Service{db: db}
}

// Team represents a team in the platform
type Team struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	TenantID           uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name               string                 `json:"name" db:"name"`
	DisplayName        string                 `json:"display_name" db:"display_name"`
	Description        *string                `json:"description,omitempty" db:"description"`
	LeadEmail          string                 `json:"lead_email" db:"lead_email"`
	Members            []Member               `json:"members" db:"members"`
	Contacts           map[string]interface{} `json:"contacts" db:"contacts"`
	Department         *string                `json:"department,omitempty" db:"department"`
	Organization       *string                `json:"organization,omitempty" db:"organization"`
	ManagerEmail       *string                `json:"manager_email,omitempty" db:"manager_email"`
	OwnedApplications  []string               `json:"owned_applications" db:"owned_applications"`
	OwnedDomains       []string               `json:"owned_domains" db:"owned_domains"`
	OwnedRepositories  []string               `json:"owned_repositories" db:"owned_repositories"`
	Policies           map[string]interface{} `json:"policies" db:"policies"`
	BudgetConfig       map[string]interface{} `json:"budget_config" db:"budget_config"`
	MemberCount        int                    `json:"member_count" db:"member_count"`
	ActiveApplications int                    `json:"active_applications" db:"active_applications"`
	MonthlySpend       *float64               `json:"monthly_spend,omitempty" db:"monthly_spend"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy          string                 `json:"created_by" db:"created_by"`
	UpdatedBy          *string                `json:"updated_by,omitempty" db:"updated_by"`
}

// Member represents a team member
type Member struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"` // owner, maintainer, developer, viewer
	JoinedAt time.Time `json:"joined_at"`
	Status   string    `json:"status"` // active, inactive, pending
}

// CreateTeam creates a new team
func (s *Service) CreateTeam(ctx context.Context, team Team) (Team, error) {
	// Set default values
	if team.ID == uuid.Nil {
		team.ID = uuid.New()
	}

	// TODO: Extract tenant ID from context
	if team.TenantID == uuid.Nil {
		// For now, use a default tenant ID - this should be extracted from auth context
		team.TenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	}

	team.CreatedAt = time.Now().UTC()
	team.UpdatedAt = team.CreatedAt

	// TODO: Extract user from auth context
	if team.CreatedBy == "" {
		team.CreatedBy = "system"
	}

	// Validate required fields
	if team.Name == "" {
		return Team{}, fmt.Errorf("%w: name is required", ErrInvalidTeamData)
	}
	if team.DisplayName == "" {
		team.DisplayName = team.Name
	}
	if team.LeadEmail == "" {
		return Team{}, fmt.Errorf("%w: lead_email is required", ErrInvalidTeamData)
	}

	// Initialize empty slices and maps if nil
	if team.Members == nil {
		team.Members = []Member{}
	}
	if team.Contacts == nil {
		team.Contacts = make(map[string]interface{})
	}
	if team.OwnedApplications == nil {
		team.OwnedApplications = []string{}
	}
	if team.OwnedDomains == nil {
		team.OwnedDomains = []string{}
	}
	if team.OwnedRepositories == nil {
		team.OwnedRepositories = []string{}
	}
	if team.Policies == nil {
		team.Policies = make(map[string]interface{})
	}
	if team.BudgetConfig == nil {
		team.BudgetConfig = make(map[string]interface{})
	}

	// Convert complex fields to JSON
	membersJSON, err := json.Marshal(team.Members)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal members: %w", err)
	}

	contactsJSON, err := json.Marshal(team.Contacts)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal contacts: %w", err)
	}

	ownedAppsJSON, err := json.Marshal(team.OwnedApplications)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal owned applications: %w", err)
	}

	ownedDomainsJSON, err := json.Marshal(team.OwnedDomains)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal owned domains: %w", err)
	}

	ownedReposJSON, err := json.Marshal(team.OwnedRepositories)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal owned repositories: %w", err)
	}

	policiesJSON, err := json.Marshal(team.Policies)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal policies: %w", err)
	}

	budgetConfigJSON, err := json.Marshal(team.BudgetConfig)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal budget config: %w", err)
	}

	// Insert team into database
	query := `
		INSERT INTO resource_management.teams (
			id, tenant_id, name, display_name, description, lead_email, members,
			contacts, department, organization, manager_email, owned_applications,
			owned_domains, owned_repositories, policies, budget_config,
			member_count, active_applications, monthly_spend, created_at,
			updated_at, created_by, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23
		)
	`

	_, err = s.db.Exec(ctx, query,
		team.ID, team.TenantID, team.Name, team.DisplayName, team.Description,
		team.LeadEmail, string(membersJSON), string(contactsJSON), team.Department,
		team.Organization, team.ManagerEmail, string(ownedAppsJSON),
		string(ownedDomainsJSON), string(ownedReposJSON), string(policiesJSON),
		string(budgetConfigJSON), team.MemberCount, team.ActiveApplications,
		team.MonthlySpend, team.CreatedAt, team.UpdatedAt, team.CreatedBy, team.UpdatedBy,
	)

	if err != nil {
		return Team{}, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

// GetTeam retrieves a team by ID
func (s *Service) GetTeam(ctx context.Context, teamID uuid.UUID) (Team, error) {
	var team Team
	var membersJSON, contactsJSON, ownedAppsJSON, ownedDomainsJSON, ownedReposJSON, policiesJSON, budgetConfigJSON string

	query := `
		SELECT id, tenant_id, name, display_name, description, lead_email, members,
			   contacts, department, organization, manager_email, owned_applications,
			   owned_domains, owned_repositories, policies, budget_config,
			   member_count, active_applications, monthly_spend, created_at,
			   updated_at, created_by, updated_by
		FROM resource_management.teams
		WHERE id = $1
	`

	err := s.db.QueryRow(ctx, query, teamID).Scan(
		&team.ID, &team.TenantID, &team.Name, &team.DisplayName, &team.Description,
		&team.LeadEmail, &membersJSON, &contactsJSON, &team.Department,
		&team.Organization, &team.ManagerEmail, &ownedAppsJSON,
		&ownedDomainsJSON, &ownedReposJSON, &policiesJSON,
		&budgetConfigJSON, &team.MemberCount, &team.ActiveApplications,
		&team.MonthlySpend, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy, &team.UpdatedBy,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return Team{}, ErrTeamNotFound
		}
		return Team{}, fmt.Errorf("failed to get team: %w", err)
	}

	// Parse JSON fields
	if err := json.Unmarshal([]byte(membersJSON), &team.Members); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal members: %w", err)
	}

	if err := json.Unmarshal([]byte(contactsJSON), &team.Contacts); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal contacts: %w", err)
	}

	if err := json.Unmarshal([]byte(ownedAppsJSON), &team.OwnedApplications); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal owned applications: %w", err)
	}

	if err := json.Unmarshal([]byte(ownedDomainsJSON), &team.OwnedDomains); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal owned domains: %w", err)
	}

	if err := json.Unmarshal([]byte(ownedReposJSON), &team.OwnedRepositories); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal owned repositories: %w", err)
	}

	if err := json.Unmarshal([]byte(policiesJSON), &team.Policies); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal policies: %w", err)
	}

	if err := json.Unmarshal([]byte(budgetConfigJSON), &team.BudgetConfig); err != nil {
		return Team{}, fmt.Errorf("failed to unmarshal budget config: %w", err)
	}

	return team, nil
}

// ListTeams retrieves a paginated list of teams
func (s *Service) ListTeams(ctx context.Context, limit, offset int) ([]Team, int, error) {
	var teams []Team
	var totalCount int

	// Get total count
	countQuery := `SELECT COUNT(*) FROM resource_management.teams`
	err := s.db.QueryRow(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get teams count: %w", err)
	}

	// Get teams with pagination
	query := `
		SELECT id, tenant_id, name, display_name, description, lead_email, members,
			   contacts, department, organization, manager_email, owned_applications,
			   owned_domains, owned_repositories, policies, budget_config,
			   member_count, active_applications, monthly_spend, created_at,
			   updated_at, created_by, updated_by
		FROM resource_management.teams
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var team Team
		var membersJSON, contactsJSON, ownedAppsJSON, ownedDomainsJSON, ownedReposJSON, policiesJSON, budgetConfigJSON string

		err := rows.Scan(
			&team.ID, &team.TenantID, &team.Name, &team.DisplayName, &team.Description,
			&team.LeadEmail, &membersJSON, &contactsJSON, &team.Department,
			&team.Organization, &team.ManagerEmail, &ownedAppsJSON,
			&ownedDomainsJSON, &ownedReposJSON, &policiesJSON,
			&budgetConfigJSON, &team.MemberCount, &team.ActiveApplications,
			&team.MonthlySpend, &team.CreatedAt, &team.UpdatedAt, &team.CreatedBy, &team.UpdatedBy,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan team row: %w", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal([]byte(membersJSON), &team.Members); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal members: %w", err)
		}

		if err := json.Unmarshal([]byte(contactsJSON), &team.Contacts); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal contacts: %w", err)
		}

		if err := json.Unmarshal([]byte(ownedAppsJSON), &team.OwnedApplications); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal owned applications: %w", err)
		}

		if err := json.Unmarshal([]byte(ownedDomainsJSON), &team.OwnedDomains); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal owned domains: %w", err)
		}

		if err := json.Unmarshal([]byte(ownedReposJSON), &team.OwnedRepositories); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal owned repositories: %w", err)
		}

		if err := json.Unmarshal([]byte(policiesJSON), &team.Policies); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal policies: %w", err)
		}

		if err := json.Unmarshal([]byte(budgetConfigJSON), &team.BudgetConfig); err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal budget config: %w", err)
		}

		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating team rows: %w", err)
	}

	return teams, totalCount, nil
}

// UpdateTeam updates an existing team
func (s *Service) UpdateTeam(ctx context.Context, team Team) (Team, error) {
	// Validate required fields
	if team.ID == uuid.Nil {
		return Team{}, fmt.Errorf("%w: team ID is required", ErrInvalidTeamData)
	}
	if team.Name == "" {
		return Team{}, fmt.Errorf("%w: name is required", ErrInvalidTeamData)
	}
	if team.LeadEmail == "" {
		return Team{}, fmt.Errorf("%w: lead_email is required", ErrInvalidTeamData)
	}

	// Set update timestamp
	team.UpdatedAt = time.Now().UTC()

	// TODO: Extract user from auth context
	if team.UpdatedBy == nil {
		updatedBy := "system"
		team.UpdatedBy = &updatedBy
	}

	// Convert complex fields to JSON
	membersJSON, err := json.Marshal(team.Members)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal members: %w", err)
	}

	contactsJSON, err := json.Marshal(team.Contacts)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal contacts: %w", err)
	}

	ownedAppsJSON, err := json.Marshal(team.OwnedApplications)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal owned applications: %w", err)
	}

	ownedDomainsJSON, err := json.Marshal(team.OwnedDomains)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal owned domains: %w", err)
	}

	ownedReposJSON, err := json.Marshal(team.OwnedRepositories)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal owned repositories: %w", err)
	}

	policiesJSON, err := json.Marshal(team.Policies)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal policies: %w", err)
	}

	budgetConfigJSON, err := json.Marshal(team.BudgetConfig)
	if err != nil {
		return Team{}, fmt.Errorf("failed to marshal budget config: %w", err)
	}

	// Update team in database
	query := `
		UPDATE resource_management.teams SET
			name = $2, display_name = $3, description = $4, lead_email = $5,
			members = $6, contacts = $7, department = $8, organization = $9,
			manager_email = $10, owned_applications = $11, owned_domains = $12,
			owned_repositories = $13, policies = $14, budget_config = $15,
			member_count = $16, active_applications = $17, monthly_spend = $18,
			updated_at = $19, updated_by = $20
		WHERE id = $1
	`

	result, err := s.db.Exec(ctx, query,
		team.ID, team.Name, team.DisplayName, team.Description, team.LeadEmail,
		string(membersJSON), string(contactsJSON), team.Department, team.Organization,
		team.ManagerEmail, string(ownedAppsJSON), string(ownedDomainsJSON),
		string(ownedReposJSON), string(policiesJSON), string(budgetConfigJSON),
		team.MemberCount, team.ActiveApplications, team.MonthlySpend,
		team.UpdatedAt, team.UpdatedBy,
	)

	if err != nil {
		return Team{}, fmt.Errorf("failed to update team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return Team{}, ErrTeamNotFound
	}

	return team, nil
}

// DeleteTeam deletes a team by ID
func (s *Service) DeleteTeam(ctx context.Context, teamID uuid.UUID) error {
	query := `DELETE FROM resource_management.teams WHERE id = $1`

	result, err := s.db.Exec(ctx, query, teamID)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrTeamNotFound
	}

	return nil
}
